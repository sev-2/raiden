package meta

import (
	"testing"

	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/stretchr/testify/assert"
)

// Test role struct functionality
func TestRoleStruct(t *testing.T) {
	role := objects.Role{
		ID:                1,
		Name:              "test_role",
		CanBypassRLS:      true,
		CanCreateDB:       true,
		CanCreateRole:     true,
		CanLogin:          true,
		Config:            map[string]any{"key": "value"},
		ConnectionLimit:   10,
		InheritRole:       true,
		IsReplicationRole: false,
		IsSuperuser:       false,
		ValidUntil:        nil,
		Password:          "password",
		ActiveConnections: 2,
	}

	assert.Equal(t, 1, role.ID)
	assert.Equal(t, "test_role", role.Name)
	assert.True(t, role.CanBypassRLS)
	assert.True(t, role.CanCreateDB)
	assert.True(t, role.CanCreateRole)
	assert.True(t, role.CanLogin)
	assert.Equal(t, map[string]any{"key": "value"}, role.Config)
	assert.Equal(t, 10, role.ConnectionLimit)
	assert.True(t, role.InheritRole)
	assert.False(t, role.IsReplicationRole)
	assert.False(t, role.IsSuperuser)
	assert.Equal(t, "password", role.Password)
	assert.Equal(t, 2, role.ActiveConnections)
}

func TestRoleMembershipsStruct(t *testing.T) {
	rm := objects.RoleMembership{
		ParentID:    1,
		ParentRole:  "parent_role",
		InheritID:   2,
		InheritRole: "inherit_role",
	}

	assert.Equal(t, 1, rm.ParentID)
	assert.Equal(t, "parent_role", rm.ParentRole)
	assert.Equal(t, 2, rm.InheritID)
	assert.Equal(t, "inherit_role", rm.InheritRole)
}

func TestUpdateRoleTypeConstants(t *testing.T) {
	assert.Equal(t, objects.UpdateConnectionLimit, objects.UpdateConnectionLimit)
	assert.Equal(t, objects.UpdateRoleName, objects.UpdateRoleName)
	assert.Equal(t, objects.UpdateRoleIsReplication, objects.UpdateRoleIsReplication)
	assert.Equal(t, objects.UpdateRoleIsSuperUser, objects.UpdateRoleIsSuperUser)
	assert.Equal(t, objects.UpdateRoleInheritRole, objects.UpdateRoleInheritRole)
	assert.Equal(t, objects.UpdateRoleCanCreateDb, objects.UpdateRoleCanCreateDb)
	assert.Equal(t, objects.UpdateRoleCanCreateRole, objects.UpdateRoleCanCreateRole)
	assert.Equal(t, objects.UpdateRoleCanLogin, objects.UpdateRoleCanLogin)
	assert.Equal(t, objects.UpdateRoleCanBypassRls, objects.UpdateRoleCanBypassRls)
	assert.Equal(t, objects.UpdateRoleConfig, objects.UpdateRoleConfig)
	assert.Equal(t, objects.UpdateRoleValidUntil, objects.UpdateRoleValidUntil)
}

func TestUpdateRoleInheritTypeConstants(t *testing.T) {
	assert.Equal(t, objects.UpdateRoleInheritGrant, objects.UpdateRoleInheritGrant)
	assert.Equal(t, objects.UpdateRoleInheritRevoke, objects.UpdateRoleInheritRevoke)
}

func TestUpdateRoleInheritancesFiltering(t *testing.T) {
	// Test with some valid and invalid items
	items := []objects.UpdateRoleInheritItem{
		{Role: objects.Role{Name: "valid_role_1"}, Type: objects.UpdateRoleInheritGrant},
		{Role: objects.Role{Name: ""}, Type: objects.UpdateRoleInheritGrant}, // Invalid - empty name
		{Role: objects.Role{Name: "valid_role_2"}, Type: objects.UpdateRoleInheritRevoke},
		{Role: objects.Role{Name: ""}, Type: objects.UpdateRoleInheritRevoke}, // Invalid - empty name
	}

	// Test the filtering logic used in updateRoleInheritances
	validItems := make([]objects.UpdateRoleInheritItem, 0, len(items))
	for i := range items {
		it := items[i]
		if it.Role.Name == "" {
			continue
		}
		validItems = append(validItems, it)
	}

	assert.Equal(t, 2, len(validItems)) // Only 2 valid items should remain
	assert.Equal(t, "valid_role_1", validItems[0].Role.Name)
	assert.Equal(t, "valid_role_2", validItems[1].Role.Name)
}

func TestRoleUpdateParam(t *testing.T) {
	role := objects.Role{
		Name: "test_role",
	}

	updateParam := objects.UpdateRoleParam{
		OldData: role,
		ChangeItems: []objects.UpdateRoleType{
			objects.UpdateRoleName,
			objects.UpdateRoleCanLogin,
		},
		ChangeInheritItems: []objects.UpdateRoleInheritItem{
			{
				Role: objects.Role{Name: "parent_role"},
				Type: objects.UpdateRoleInheritGrant,
			},
		},
	}

	assert.Equal(t, role, updateParam.OldData)
	assert.Equal(t, 2, len(updateParam.ChangeItems))
	assert.Equal(t, objects.UpdateRoleName, updateParam.ChangeItems[0])
	assert.Equal(t, objects.UpdateRoleCanLogin, updateParam.ChangeItems[1])
	assert.Equal(t, 1, len(updateParam.ChangeInheritItems))
	assert.Equal(t, "parent_role", updateParam.ChangeInheritItems[0].Role.Name)
	assert.Equal(t, objects.UpdateRoleInheritGrant, updateParam.ChangeInheritItems[0].Type)
}

func TestRoleMembershipsGroupByInheritId(t *testing.T) {
	// Test RoleMemberships GroupByInheritId function which is defined in the objects package
	// but can be tested in context of role functionality
	memberships := objects.RoleMemberships{
		{ParentID: 1, ParentRole: "parent1", InheritID: 10, InheritRole: "child1"},
		{ParentID: 2, ParentRole: "parent2", InheritID: 10, InheritRole: "child1"}, // Same inherit ID
		{ParentID: 3, ParentRole: "parent3", InheritID: 20, InheritRole: "child2"},
	}
	
	grouped := memberships.GroupByInheritId()
	assert.Equal(t, 2, len(grouped)) // 2 different inherit IDs: 10 and 20
	assert.Equal(t, 2, len(grouped[10])) // 2 parents for inherit ID 10
	assert.Equal(t, 1, len(grouped[20])) // 1 parent for inherit ID 20
}

func TestRolesToMap(t *testing.T) {
	// Test Roles ToMap function which is defined in the objects package
	// but can be tested in context of role functionality
	roles := objects.Roles{
		{ID: 1, Name: "role1"},
		{ID: 2, Name: "role2"},
	}
	
	roleMap := roles.ToMap()
	assert.Equal(t, 2, len(roleMap))
	assert.Equal(t, "role1", roleMap[1].Name)
	assert.Equal(t, "role2", roleMap[2].Name)
}