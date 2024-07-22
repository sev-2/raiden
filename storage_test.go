package raiden_test

import (
	"testing"

	"github.com/sev-2/raiden"
	"github.com/stretchr/testify/assert"
)

type MockBucket struct {
	raiden.BucketBase
	name string
}

func (m *MockBucket) Name() string {
	return m.name
}

func TestBucketBase_Public(t *testing.T) {
	bucket := &raiden.BucketBase{}
	assert.False(t, bucket.Public(), "Expected Public() to return false")
}

func TestBucketBase_AllowedMimeTypes(t *testing.T) {
	bucket := &raiden.BucketBase{}
	assert.Nil(t, bucket.AllowedMimeTypes(), "Expected AllowedMimeTypes() to return nil")
}

func TestBucketBase_FileSizeLimit(t *testing.T) {
	bucket := &raiden.BucketBase{}
	assert.Equal(t, 0, bucket.FileSizeLimit(), "Expected FileSizeLimit() to return 0")
}

func TestBucketBase_AvifAutoDetection(t *testing.T) {
	bucket := &raiden.BucketBase{}
	assert.False(t, bucket.AvifAutoDetection(), "Expected AvifAutoDetection() to return false")
}

func TestMockBucket_Name(t *testing.T) {
	mockName := "test-bucket"
	bucket := &MockBucket{name: mockName}
	assert.Equal(t, mockName, bucket.Name(), "Expected Name() to return the correct bucket name")
}
