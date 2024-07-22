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
