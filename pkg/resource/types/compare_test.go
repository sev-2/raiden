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
			Name:   "type1_updated",
			Schema: "public",
			Format: "",
			Enums:  []string{"test_1", "test_2"},
			Attributes: []objects.TypeAttribute{{
				Name: "type_attribute_1",
			}},
			Comment: nil,
		},
	}

	target := []objects.Type{
		{
			Name:   "type1_updated",
			Schema: "public",
			Format: "",
			Enums:  []string{"test_1", "test_2"},
			Attributes: []objects.TypeAttribute{{
				Name: "type_attribute_1",
			}},
			Comment: nil,
		},
	}

	err := types.Compare(source, target)
	assert.NoError(t, err)
}

func TestCompareList(t *testing.T) {
	sourceComment := "source comment"
	source := []objects.Type{
		{
			Name:       "type1_updated",
			Schema:     "public",
			Format:     "",
			Enums:      []string{"test_1", "test_2"},
			Attributes: []objects.TypeAttribute{},
			Comment:    nil,
		},
		{
			Name:    "type2",
			Schema:  "auth",
			Comment: nil,
		},
		{
			Name:       "type3",
			Schema:     "auth",
			Comment:    &sourceComment,
			Attributes: []objects.TypeAttribute{{Name: "type_attribute_1"}},
		},
		{
			Name:       "type4",
			Schema:     "auth",
			Comment:    &sourceComment,
			Enums:      []string{"enum_x"},
			Attributes: []objects.TypeAttribute{{Name: "type_attribute_x"}},
		},
	}

	targetComment := "comment test"
	target := []objects.Type{
		{
			Name:       "type1_updated",
			Schema:     "public",
			Format:     "",
			Enums:      []string{"test_1", "test_2", "test_3"},
			Attributes: []objects.TypeAttribute{{Name: "type_attribute_1"}},
			Comment:    nil,
		},
		{
			Name:       "type2",
			Schema:     "public",
			Comment:    &targetComment,
			Attributes: []objects.TypeAttribute{{Name: "type_attribute_1"}},
		},
		{
			Name:       "type3",
			Schema:     "auth",
			Comment:    &targetComment,
			Attributes: []objects.TypeAttribute{{Name: "type_attribute_1"}, {Name: "type_attribute_2"}},
		},
		{
			Name:       "type4",
			Schema:     "auth",
			Comment:    &sourceComment,
			Enums:      []string{"enum_x"},
			Attributes: []objects.TypeAttribute{{Name: "type_attribute_x"}},
		},
	}

	diffResult, err := types.CompareList(source, target)
	assert.NoError(t, err)
	assert.Equal(t, 4, len(diffResult))
	assert.Equal(t, "type1_updated", diffResult[0].SourceResource.Name)
	assert.Equal(t, "type4", diffResult[0].TargetResource.Name)
}

func TestCompareItem(t *testing.T) {
	source := objects.Type{
		Name:   "type1",
		Schema: "public",
	}

	target := objects.Type{
		Name:   "type1_updated",
		Schema: "auth",
	}

	diffResult := types.CompareItem(source, target)
	assert.True(t, diffResult.IsConflict)
	assert.Equal(t, "type1", diffResult.SourceResource.Name)
	assert.Equal(t, "type1_updated", diffResult.TargetResource.Name)
}
