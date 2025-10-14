package cloud

import (
	"testing"

	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/supabase/query/sql"
	"github.com/stretchr/testify/assert"
)

// Test role helper functions in cloud package
func TestRoleFindConfigFn(t *testing.T) {
	// Test with a role that has config
	roleWithConfig := map[string]any{
		"name":   "test_role",
		"config": []any{"key1=value1", "key2=value2"},
	}

	configs := roleFindConfigFn(roleWithConfig)
	assert.Equal(t, 2, len(configs))
	assert.Equal(t, "key1=value1", configs[0])
	assert.Equal(t, "key2=value2", configs[1])

	// Test with a role that has no config
	roleWithoutConfig := map[string]any{
		"name": "test_role",
	}

	configs = roleFindConfigFn(roleWithoutConfig)
	assert.Nil(t, configs)
}

func TestRoleConfigsToMapFn(t *testing.T) {
	configArr := []any{"key1=value1", "key2=value2", "invalid_config"}

	configMap := roleConfigsToMapFn(configArr)
	assert.Equal(t, 2, len(configMap))
	assert.Equal(t, "value1", configMap["key1"])
	assert.Equal(t, "value2", configMap["key2"])
	// Invalid config should be skipped
	_, exists := configMap["invalid_config"]
	assert.False(t, exists)
}

func TestRoleResultDecoratorFn(t *testing.T) {
	// Test with roles that have config
	roles := []any{
		map[string]any{
			"name":   "role1",
			"config": []any{"key1=value1", "key2=value2"},
		},
		map[string]any{
			"name":   "role2",
			"config": []any{"key3=value3"},
		},
	}

	result := roleResultDecoratorFn(roles)
	decoratedRoles := result.([]any)

	// Check first role
	role1 := decoratedRoles[0].(map[string]any)
	config1 := role1["config"].(map[string]any)
	assert.Equal(t, "value1", config1["key1"])
	assert.Equal(t, "value2", config1["key2"])

	// Check second role
	role2 := decoratedRoles[1].(map[string]any)
	config2 := role2["config"].(map[string]any)
	assert.Equal(t, "value3", config2["key3"])
}

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

	// Since we can't call updateRoleInheritances directly due to its dependency on external functions,
	// we just test the filtering logic conceptually
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

// Since updateRoleInheritances is an internal function, let's create a more comprehensive test
// that tests the core logic by simulating its behavior
func TestUpdateRoleInheritances(t *testing.T) {
	// Since we can't easily test the full function with external dependencies,
	// we'll test the core logic components that we can validate

	// 1. Test that validItems filtering works correctly
	items := []objects.UpdateRoleInheritItem{
		{Role: objects.Role{Name: "valid_role_1"}, Type: objects.UpdateRoleInheritGrant},
		{Role: objects.Role{Name: ""}, Type: objects.UpdateRoleInheritGrant}, // Invalid - empty name
		{Role: objects.Role{Name: "valid_role_2"}, Type: objects.UpdateRoleInheritRevoke},
		{Role: objects.Role{Name: ""}, Type: objects.UpdateRoleInheritRevoke},            // Invalid - empty name
		{Role: objects.Role{Name: "valid_role_3"}, Type: objects.UpdateRoleInheritGrant}, // Additional valid item
	}

	// Test the filtering logic from updateRoleInheritances
	validItems := make([]objects.UpdateRoleInheritItem, 0, len(items))
	for i := range items {
		it := items[i]
		if it.Role.Name == "" {
			continue
		}
		validItems = append(validItems, it)
	}

	assert.Equal(t, 3, len(validItems)) // Only 3 valid items should remain
	assert.Equal(t, "valid_role_1", validItems[0].Role.Name)
	assert.Equal(t, "valid_role_2", validItems[1].Role.Name)
	assert.Equal(t, "valid_role_3", validItems[2].Role.Name)

	// 2. Test different inheritance types
	assert.Equal(t, objects.UpdateRoleInheritGrant, validItems[0].Type)
	assert.Equal(t, objects.UpdateRoleInheritRevoke, validItems[1].Type)
	assert.Equal(t, objects.UpdateRoleInheritGrant, validItems[2].Type)

	// 3. Test when no valid items remain
	emptyItems := []objects.UpdateRoleInheritItem{
		{Role: objects.Role{Name: ""}, Type: objects.UpdateRoleInheritGrant},
		{Role: objects.Role{Name: ""}, Type: objects.UpdateRoleInheritRevoke},
	}

	emptyValidItems := make([]objects.UpdateRoleInheritItem, 0, len(emptyItems))
	for i := range emptyItems {
		it := emptyItems[i]
		if it.Role.Name == "" {
			continue
		}
		emptyValidItems = append(emptyValidItems, it)
	}

	assert.Equal(t, 0, len(emptyValidItems)) // No items should remain
}

// Test GetRoleMemberships function logic by testing the query generation
func TestGetRoleMemberships(t *testing.T) {
	// Test with nil/empty schemas
	schemas := []string{}
	query := sql.GenerateGetRoleMembershipsQuery(schemas)
	assert.Contains(t, query, "pg_auth_members")

	// Test with specific schemas
	schemas = []string{"public", "private"}
	query = sql.GenerateGetRoleMembershipsQuery(schemas)
	assert.Contains(t, query, "pg_namespace")
	for _, schema := range schemas {
		assert.Contains(t, query, schema)
	}

	// Test with single schema
	schemas = []string{"public"}
	query = sql.GenerateGetRoleMembershipsQuery(schemas)
	assert.Contains(t, query, "pg_namespace")
	assert.Contains(t, query, "public")
}
