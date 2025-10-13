package storages_test

import (
	"testing"

	"github.com/sev-2/raiden/pkg/resource/storages"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/stretchr/testify/assert"
)

func TestCompare(t *testing.T) {
	source := []objects.Bucket{
		{Name: "bucket1"},
		{Name: "bucket2"},
	}

	target := []objects.Bucket{
		{Name: "bucket1"},
		{Name: "bucket2"},
	}

	err := storages.Compare(source, target)
	assert.NoError(t, err)
}

func TestCompareList(t *testing.T) {
	source := []objects.Bucket{
		{Name: "bucket1"},
		{Name: "bucket2"},
	}

	target := []objects.Bucket{
		{Name: "bucket1_updated"},
		{Name: "bucket2"},
	}

	diffResult, err := storages.CompareList(source, target)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(diffResult))
	assert.Equal(t, "bucket1", diffResult[0].SourceResource.Name)
	assert.Equal(t, "bucket2", diffResult[0].TargetResource.Name)
}

func TestCompareItem(t *testing.T) {
	source := objects.Bucket{
		Name:   "bucket1",
		Public: true,
	}

	target := objects.Bucket{
		Name:   "bucket1_updated",
		Public: false,
	}

	diffResult := storages.CompareItem(source, target)
	assert.True(t, diffResult.IsConflict)
	assert.Equal(t, "bucket1", diffResult.SourceResource.Name)
	assert.Equal(t, "bucket1_updated", diffResult.TargetResource.Name)
	assert.Equal(t, objects.UpdateBucketIsPublic, diffResult.DiffItems.ChangeItems[0])
}

func TestCompareItem_WithFileSizeLimit(t *testing.T) {
	// Test nil vs non-nil case
	size100 := 100
	source1 := objects.Bucket{
		Name:   "bucket1",
		FileSizeLimit: &size100,
	}
	
	target1 := objects.Bucket{
		Name:   "bucket1", 
		FileSizeLimit: nil,
	}
	
	result1 := storages.CompareItem(source1, target1)
	assert.True(t, result1.IsConflict)
	
	// Test different values
	size200 := 200
	source2 := objects.Bucket{
		Name:   "bucket2",
		FileSizeLimit: &size100,
	}
	
	target2 := objects.Bucket{
		Name:   "bucket2",
		FileSizeLimit: &size200,
	}
	
	result2 := storages.CompareItem(source2, target2)
	assert.True(t, result2.IsConflict)
	
	// Test both nil (no conflict)
	source3 := objects.Bucket{
		Name:   "bucket3",
		FileSizeLimit: nil,
	}
	
	target3 := objects.Bucket{
		Name:   "bucket3",
		FileSizeLimit: nil,
	}
	
	result3 := storages.CompareItem(source3, target3)
	assert.False(t, result3.IsConflict)
}

func TestCompareItem_WithAllowedMimeTypes(t *testing.T) {
	// Test different lengths
	source1 := objects.Bucket{
		Name:   "bucket1",
		AllowedMimeTypes: []string{"image/png"},
	}
	
	target1 := objects.Bucket{
		Name:   "bucket1",
		AllowedMimeTypes: []string{"image/png", "image/jpg"},
	}
	
	result1 := storages.CompareItem(source1, target1)
	assert.True(t, result1.IsConflict)
	
	// Test same length but different content
	source2 := objects.Bucket{
		Name:   "bucket2",
		AllowedMimeTypes: []string{"image/png", "image/gif"},
	}
	
	target2 := objects.Bucket{
		Name:   "bucket2",
		AllowedMimeTypes: []string{"image/png", "image/jpg"},
	}
	
	result2 := storages.CompareItem(source2, target2)
	assert.True(t, result2.IsConflict)
	
	// Test same content
	source3 := objects.Bucket{
		Name:   "bucket3",
		AllowedMimeTypes: []string{"image/png", "image/jpg"},
	}
	
	target3 := objects.Bucket{
		Name:   "bucket3",
		AllowedMimeTypes: []string{"image/png", "image/jpg"},
	}
	
	result3 := storages.CompareItem(source3, target3)
	assert.False(t, result3.IsConflict)
}


