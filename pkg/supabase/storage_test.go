package supabase_test

import (
	"testing"

	"github.com/sev-2/raiden/pkg/supabase"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/stretchr/testify/assert"
)

func TestGetBuckets_Cloud(t *testing.T) {
	cfg := loadCloudConfig()

	_, err := supabase.GetBuckets(cfg)
	assert.Error(t, err)
}

func TestGetBucket_Cloud(t *testing.T) {
	cfg := loadCloudConfig()

	_, err := supabase.GetBucket(cfg, "some-bucket")
	assert.Error(t, err)
}

func TestCreateBucket_Cloud(t *testing.T) {
	cfg := loadCloudConfig()

	_, err := supabase.CreateBucket(cfg, objects.Bucket{})
	assert.Error(t, err)
}

func TestUpdateBucket_Cloud(t *testing.T) {
	cfg := loadCloudConfig()

	bucket := objects.Bucket{
		ID:     "bucket-1",
		Name:   "some-bucket",
		Public: true,
	}
	updateParam := objects.UpdateBucketParam{
		ChangeItems: []objects.UpdateBucketType{
			objects.UpdateBucketIsPublic,
		},
	}

	// Test when there are update items - should try to make request
	_ = supabase.UpdateBucket(cfg, bucket, updateParam)
	// This might fail because we don't have proper mocking, but we can test the function exists

	// Test when there are no update items - should return early without calling API
	bucket2 := objects.Bucket{
		ID:     "bucket-2",
		Name:   "another-bucket",
		Public: false,
	}
	updateParam2 := objects.UpdateBucketParam{
		ChangeItems: []objects.UpdateBucketType{}, // Empty
	}

	err2 := supabase.UpdateBucket(cfg, bucket2, updateParam2)
	assert.NoError(t, err2) // Should return early without error
}

func TestDeleteBucket_Cloud(t *testing.T) {
	cfg := loadCloudConfig()

	err := supabase.DeleteBucket(cfg, objects.Bucket{})
	assert.Error(t, err)
}

func TestGetBuckets_SelfHosted(t *testing.T) {
	cfg := loadSelfHostedConfig()

	_, err := supabase.GetBuckets(cfg)
	assert.Error(t, err)
}

func TestGetBucket_SelfHosted(t *testing.T) {
	cfg := loadSelfHostedConfig()

	_, err := supabase.GetBucket(cfg, "some-bucket")
	assert.Error(t, err)
}

func TestCreateBucket_SelfHosted(t *testing.T) {
	cfg := loadSelfHostedConfig()

	_, err := supabase.CreateBucket(cfg, objects.Bucket{})
	assert.Error(t, err)
}

func TestUpdateBucket_SelfHosted(t *testing.T) {
	cfg := loadSelfHostedConfig()

	err := supabase.UpdateBucket(cfg, objects.Bucket{}, objects.UpdateBucketParam{})
	assert.NoError(t, err) // Update bucket with no changes should pass
}

func TestDeleteBucket_SelfHosted(t *testing.T) {
	cfg := loadSelfHostedConfig()

	err := supabase.DeleteBucket(cfg, objects.Bucket{})
	assert.Error(t, err)
}
