package storages_test

import (
	"testing"

	"github.com/sev-2/raiden/pkg/resource/storages"
	"github.com/sev-2/raiden/pkg/state"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/stretchr/testify/assert"
)

func TestGetNewCountData(t *testing.T) {
	supabaseBuckets := []objects.Bucket{
		{Name: "bucket1"},
		{Name: "bucket2"},
		{Name: "bucket3"},
	}

	extractResult := state.ExtractStorageResult{
		Delete: []state.ExtractStorageItem{
			{Storage: objects.Bucket{Name: "bucket1"}},
			{Storage: objects.Bucket{Name: "bucket2"}},
		},
	}

	count := storages.GetNewCountData(supabaseBuckets, extractResult)
	assert.Equal(t, 2, count)
}

func TestGetNewCountDataNoMatch(t *testing.T) {
	supabaseBuckets := []objects.Bucket{
		{Name: "bucket1"},
		{Name: "bucket2"},
	}

	extractResult := state.ExtractStorageResult{
		Delete: []state.ExtractStorageItem{
			{Storage: objects.Bucket{Name: "bucket3"}},
			{Storage: objects.Bucket{Name: "bucket4"}},
		},
	}

	count := storages.GetNewCountData(supabaseBuckets, extractResult)
	assert.Equal(t, 0, count)
}

func TestGetNewCountDataEmpty(t *testing.T) {
	supabaseBuckets := []objects.Bucket{}

	extractResult := state.ExtractStorageResult{
		Delete: []state.ExtractStorageItem{},
	}

	count := storages.GetNewCountData(supabaseBuckets, extractResult)
	assert.Equal(t, 0, count)
}

func TestGetNewCountDataPartialMatch(t *testing.T) {
	supabaseBuckets := []objects.Bucket{
		{Name: "bucket1"},
		{Name: "bucket2"},
		{Name: "bucket3"},
	}

	extractResult := state.ExtractStorageResult{
		Delete: []state.ExtractStorageItem{
			{Storage: objects.Bucket{Name: "bucket1"}},
			{Storage: objects.Bucket{Name: "bucket4"}},
		},
	}

	count := storages.GetNewCountData(supabaseBuckets, extractResult)
	assert.Equal(t, 1, count)
}
