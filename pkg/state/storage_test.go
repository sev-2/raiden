package state_test

import (
	"testing"

	"github.com/sev-2/raiden"
	st "github.com/sev-2/raiden/pkg/builder"
	"github.com/sev-2/raiden/pkg/state"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/stretchr/testify/assert"
)

// Mock types
type MockBucket struct {
	raiden.BucketBase
	name              string
	public            bool
	allowedMimeTypes  []string
	avifAutoDetection bool
	fileSizeLimit     int
	Acl               raiden.Acl
}

func (m *MockBucket) Name() string               { return m.name }
func (m *MockBucket) Public() bool               { return m.public }
func (m *MockBucket) AllowedMimeTypes() []string { return m.allowedMimeTypes }
func (m *MockBucket) AvifAutoDetection() bool    { return m.avifAutoDetection }
func (m *MockBucket) FileSizeLimit() int         { return m.fileSizeLimit }

func (m *MockBucket) ConfigureAcl() {
	m.Acl.Enable()
	m.Acl.Define(
		raiden.Rule("read_bucket").For("authenticated").To(raiden.CommandSelect).
			Using(st.True).
			WithPermissive(),
	)
}

func TestExtractStorage(t *testing.T) {
	storageStates := []state.StorageState{
		{
			Storage: objects.Bucket{Name: "storage1"},
		},
		{
			Storage: objects.Bucket{Name: "storage2"},
		},
	}

	appStorages := []raiden.Bucket{
		&MockBucket{name: "storage1"},
		&MockBucket{name: "storage3"},
	}

	result, err := state.ExtractStorage(storageStates, appStorages)
	assert.NoError(t, err)
	assert.Len(t, result.Existing, 1)
	assert.Len(t, result.New, 1)
	assert.Len(t, result.Delete, 1)
}

func TestBindToSupabaseStorage(t *testing.T) {
	mockStorage := &MockBucket{
		name:              "storage1",
		public:            true,
		allowedMimeTypes:  []string{"image/png", "image/jpeg"},
		avifAutoDetection: true,
		fileSizeLimit:     1024,
	}

	var bucket objects.Bucket
	state.BindToSupabaseStorage(&bucket, mockStorage)

	assert.Equal(t, "storage1", bucket.Name)
	assert.True(t, bucket.Public)
	assert.Equal(t, []string{"image/png", "image/jpeg"}, bucket.AllowedMimeTypes)
	assert.True(t, bucket.AvifAutoDetection)
	assert.NotNil(t, bucket.FileSizeLimit)
	assert.Equal(t, 1024, *bucket.FileSizeLimit)
}

func TestBuildStorageFromApp(t *testing.T) {
	mockStorage := &MockBucket{name: "storage1"}
	result := state.BuildStorageFromApp(mockStorage)
	assert.Equal(t, "storage1", result.Storage.Name)
}

func TestBuildStorageFromState(t *testing.T) {
	mockStorage := &MockBucket{name: "storage1"}
	storageState := state.StorageState{
		Storage: objects.Bucket{
			Name: "storage1",
		},
	}
	result := state.BuildStorageFromState(storageState, mockStorage)
	assert.Equal(t, "storage1", result.Storage.Name)
}

func TestExtractStorageResult_ToDeleteFlatMap(t *testing.T) {
	storageResult := state.ExtractStorageResult{
		Delete: []state.ExtractStorageItem{
			{Storage: objects.Bucket{Name: "storage1"}},
		},
	}

	result := storageResult.ToDeleteFlatMap()
	assert.Len(t, result, 1)
	assert.Contains(t, result, "storage1")
}
