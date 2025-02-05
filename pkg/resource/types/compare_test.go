package types_test

import (
	"testing"

	"github.com/sev-2/raiden/pkg/resource/types"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/stretchr/testify/assert"
)

func TestCompare(t *testing.T) {
	source := []objects.Type{
		{
			Name:       "type1_updated",
			Schema:     "public",
			Format:     "",
			Enums:      []string{"test_1", "test_2"},
			Attributes: []string{"attribute_1"},
			Comment:    nil,
		},
	}

	target := []objects.Type{
		{
			Name:       "type1_updated",
			Schema:     "public",
			Format:     "",
			Enums:      []string{"test_1", "test_2"},
			Attributes: []string{"attribute_1"},
			Comment:    nil,
		},
	}

	err := types.Compare(source, target)
	assert.NoError(t, err)
}

func TestCompareList(t *testing.T) {
	source := []objects.Type{
		{
			Name:       "type1",
			Schema:     "public",
			Format:     "",
			Enums:      []string{"test_1", "test_2"},
			Attributes: []string{},
			Comment:    nil,
		},
		{
			Name:   "type2",
			Schema: "auth",
		},
	}

	target := []objects.Type{
		{
			Name:       "type1_updated",
			Schema:     "public",
			Format:     "",
			Enums:      []string{"test_1", "test_2", "test_3"},
			Attributes: []string{"attribute_1"},
			Comment:    nil,
		},
		{
			Name:   "type2",
			Schema: "public",
		},
	}

	diffResult, err := types.CompareList(source, target)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(diffResult))
	assert.Equal(t, "type1", diffResult[0].SourceResource.Name)
	assert.Equal(t, "type2", diffResult[0].TargetResource.Name)
}

func TestCompareItem(t *testing.T) {
	source := objects.Type{
		Name: "type1",
	}

	target := objects.Type{
		Name: "type1_updated",
	}

	diffResult := types.CompareItem(source, target)
	assert.True(t, diffResult.IsConflict)
	assert.Equal(t, "type1", diffResult.SourceResource.Name)
	assert.Equal(t, "type1_updated", diffResult.TargetResource.Name)
}
