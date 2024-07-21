package roles_test

import (
	"testing"

	"github.com/sev-2/raiden/pkg/resource/roles"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/stretchr/testify/assert"
)

func TestCompare(t *testing.T) {
	source := []objects.Role{
		{Name: "role4"},
		{Name: "role4"},
	}

	target := []objects.Role{
		{Name: "role3"},
		{Name: "role4"},
	}

	err := roles.Compare(source, target)
	assert.NoError(t, err)
}

func TestCompareList(t *testing.T) {
	source := []objects.Role{
		{
			Name:            "role1",
			CanBypassRLS:    true,
			CanLogin:        true,
			ConnectionLimit: 10,
			CanCreateDB:     true,
			CanCreateRole:   true,
			Config:          map[string]interface{}{"key": "value"},
			InheritRole:     true,
			ValidUntil:      &objects.SupabaseTime{},
		},
		{
			Name:   "role2",
			Config: map[string]interface{}{"not-key": "value"},
		},
	}

	target := []objects.Role{
		{
			Name:            "role1_updated",
			CanBypassRLS:    false,
			CanLogin:        false,
			ConnectionLimit: 20,
			CanCreateDB:     false,
			CanCreateRole:   false,
			Config:          map[string]interface{}{"key": "new-value"},
			InheritRole:     false,
			ValidUntil:      &objects.SupabaseTime{},
		},
		{
			Name:   "role2",
			Config: map[string]interface{}{"key": "value"},
		},
	}

	diffResult, err := roles.CompareList(source, target)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(diffResult))
	assert.Equal(t, "role1", diffResult[0].SourceResource.Name)
	assert.Equal(t, "role2", diffResult[0].TargetResource.Name)
}

func TestCompareItem(t *testing.T) {
	source := objects.Role{
		Name: "role1",
	}

	target := objects.Role{
		Name: "role1_updated",
	}

	diffResult := roles.CompareItem(source, target)
	assert.True(t, diffResult.IsConflict)
	assert.Equal(t, "role1", diffResult.SourceResource.Name)
	assert.Equal(t, "role1_updated", diffResult.TargetResource.Name)
}
