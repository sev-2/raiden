package roles_test

import (
	"testing"
	"time"

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
			ID:              1, // Same ID to match with target
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
			ID:     2, // Same ID to match with target
			Name:   "role2",
			Config: map[string]interface{}{"not-key": "value"},
		},
	}

	target := []objects.Role{
		{
			ID:              1, // Same ID to match with source (same logical role, different name)
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
			ID:     2, // Same ID to match with source
			Name:   "role2",
			Config: map[string]interface{}{"key": "value"},
		},
	}

	diffResult, err := roles.CompareList(source, target)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(diffResult))
	assert.Equal(t, "role1", diffResult[0].SourceResource.Name)
	assert.Equal(t, "role1_updated", diffResult[0].TargetResource.Name) // Updated assertion to match target name
	assert.Equal(t, "role2", diffResult[1].SourceResource.Name)
	assert.Equal(t, "role2", diffResult[1].TargetResource.Name)
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

func TestCompareItem_InheritRolesCaseInsensitive(t *testing.T) {
	source := objects.Role{
		ID:   1,
		Name: "role1",
		InheritRoles: []*objects.Role{
			{Name: "Parent_Role"},
			{Name: "another_role"},
		},
	}

	target := objects.Role{
		ID:   1,
		Name: "role1",
		InheritRoles: []*objects.Role{
			{Name: "parent_role"},
		},
	}

	diffResult := roles.CompareItem(source, target)
	assert.True(t, diffResult.IsConflict)
	if assert.Len(t, diffResult.DiffItems.ChangeInheritItems, 1) {
		assert.Equal(t, objects.UpdateRoleInheritGrant, diffResult.DiffItems.ChangeInheritItems[0].Type)
		assert.Equal(t, "another_role", diffResult.DiffItems.ChangeInheritItems[0].Role.Name)
	}
}

func TestCompareItem_FieldDifferences(t *testing.T) {
	ptrA := "value-a"
	ptrB := "value-b"
	future := objects.NewSupabaseTime(time.Now().Add(48 * time.Hour))
	past := objects.NewSupabaseTime(time.Now().Add(-48 * time.Hour))

	source := objects.Role{
		Name:            "role",
		ConnectionLimit: 5,
		CanBypassRLS:    true,
		CanCreateDB:     true,
		CanCreateRole:   true,
		CanLogin:        true,
		Config:          map[string]any{"ptr": &ptrA},
		InheritRole:     true,
		InheritRoles:    []*objects.Role{{Name: "child_a"}},
		ValidUntil:      future,
	}

	target := objects.Role{
		Name:            "role",
		ConnectionLimit: 7,
		CanBypassRLS:    false,
		CanCreateDB:     false,
		CanCreateRole:   false,
		CanLogin:        false,
		Config:          map[string]any{"ptr": &ptrB},
		InheritRole:     false,
		InheritRoles:    []*objects.Role{{Name: "child_b"}},
		ValidUntil:      past,
	}

	diffResult := roles.CompareItem(source, target)
	assert.True(t, diffResult.IsConflict)
	assert.Contains(t, diffResult.DiffItems.ChangeItems, objects.UpdateConnectionLimit)
	assert.Contains(t, diffResult.DiffItems.ChangeItems, objects.UpdateRoleCanBypassRls)
	assert.Contains(t, diffResult.DiffItems.ChangeItems, objects.UpdateRoleCanCreateDb)
	assert.Contains(t, diffResult.DiffItems.ChangeItems, objects.UpdateRoleCanCreateRole)
	assert.Contains(t, diffResult.DiffItems.ChangeItems, objects.UpdateRoleCanLogin)
	assert.Contains(t, diffResult.DiffItems.ChangeItems, objects.UpdateRoleConfig)
	assert.Contains(t, diffResult.DiffItems.ChangeItems, objects.UpdateRoleInheritRole)
	assert.Contains(t, diffResult.DiffItems.ChangeItems, objects.UpdateRoleValidUntil)
	if assert.Len(t, diffResult.DiffItems.ChangeInheritItems, 2) {
		typesFound := []objects.UpdateRoleInheritType{
			diffResult.DiffItems.ChangeInheritItems[0].Type,
			diffResult.DiffItems.ChangeInheritItems[1].Type,
		}
		assert.Contains(t, typesFound, objects.UpdateRoleInheritGrant)
		assert.Contains(t, typesFound, objects.UpdateRoleInheritRevoke)
	}
}

func TestCompareNoConflict(t *testing.T) {
	// Test case with no conflicts that should return no error
	source := []objects.Role{
		{
			ID:           1,
			Name:         "role1",
			CanBypassRLS: true,
			CanLogin:     true,
		},
	}

	target := []objects.Role{
		{
			ID:           1,
			Name:         "role1",
			CanBypassRLS: true, // Same as source
			CanLogin:     true, // Same as source
		},
	}

	err := roles.Compare(source, target)
	assert.NoError(t, err) // Should not return error because no conflicts
}

func TestCompareWithError(t *testing.T) {
	// Test case with conflicts that should return an error
	source := []objects.Role{
		{
			ID:              1,
			Name:            "role1",
			CanBypassRLS:    true,
			CanLogin:        true,
			ConnectionLimit: 10,
		},
	}

	target := []objects.Role{
		{
			ID:              1,
			Name:            "role1", // Same name, but other fields different
			CanBypassRLS:    false,
			CanLogin:        false,
			ConnectionLimit: 20, // Different from source
		},
	}

	err := roles.Compare(source, target)
	assert.Error(t, err) // Should return error because of conflicts
	assert.Contains(t, err.Error(), "canceled import process, you have conflict in role")
}
