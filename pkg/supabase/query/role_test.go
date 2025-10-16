package query

import (
	"testing"
	"time"

	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/stretchr/testify/assert"
)

func TestBuildCreateRoleQuery(t *testing.T) {
	role := objects.Role{
		Name:            "test_role",
		CanCreateDB:     true,
		CanCreateRole:   true,
		CanBypassRLS:    true,
		CanLogin:        true,
		InheritRole:     true,
		ConnectionLimit: 5,
		ValidUntil:      nil, // This would be a pointer to time, so let's keep it nil for now
		Config:          map[string]interface{}{"search_path": "public"},
	}

	query := BuildCreateRoleQuery(role)

	// Check that the query contains expected parts
	assert.Contains(t, query, "CREATE ROLE \"test_role\"")
	assert.Contains(t, query, "CREATEDB")
	assert.Contains(t, query, "CREATEROLE")
	assert.Contains(t, query, "BYPASSRLS")
	assert.Contains(t, query, "LOGIN")
	assert.Contains(t, query, "INHERIT")
	assert.Contains(t, query, "CONNECTION LIMIT 5")
	assert.Contains(t, query, "ALTER ROLE \"test_role\" SET \"search_path\" = 'public'")
}

func TestBuildCreateRoleQueryWithDifferentValues(t *testing.T) {
	role := objects.Role{
		Name:            "limited_role",
		CanCreateDB:     false,
		CanCreateRole:   false,
		CanBypassRLS:    false,
		CanLogin:        false,
		InheritRole:     false,
		ConnectionLimit: 1,
		Config:          nil,
	}

	query := BuildCreateRoleQuery(role)

	// Check that the query contains expected parts for default (false) values
	assert.Contains(t, query, "CREATE ROLE \"limited_role\"")
	assert.Contains(t, query, "NOCREATEDB")
	assert.Contains(t, query, "NOCREATEROLE")
	assert.Contains(t, query, "NOBYPASSRLS")
	assert.Contains(t, query, "NOLOGIN")
	assert.Contains(t, query, "NOINHERIT")
	assert.Contains(t, query, "CONNECTION LIMIT 1")
	// Should not contain config clauses since Config is nil
}

func TestBuildUpdateRoleQuery(t *testing.T) {
	newRole := objects.Role{
		Name:            "updated_role",
		ConnectionLimit: 10,
		CanLogin:        false,
		IsSuperuser:     true,
		InheritRole:     false,
		CanBypassRLS:    true,
		CanCreateDB:     false,
		CanCreateRole:   true,
		ValidUntil:      nil,
		Config:          map[string]interface{}{"log_statement": "all"},
	}

	oldRole := objects.Role{
		Name: "original_role",
	}

	updateParam := objects.UpdateRoleParam{
		OldData: oldRole,
		ChangeItems: []objects.UpdateRoleType{
			objects.UpdateConnectionLimit,
			objects.UpdateRoleCanLogin,
			objects.UpdateRoleIsSuperUser,
			objects.UpdateRoleInheritRole,
			objects.UpdateRoleCanBypassRls,
			objects.UpdateRoleCanCreateDb,
			objects.UpdateRoleCanCreateRole,
			objects.UpdateRoleConfig,
		},
	}

	query := BuildUpdateRoleQuery(newRole, updateParam)

	// Check that the query contains expected update parts
	assert.Contains(t, query, "ALTER ROLE \"original_role\"")
	assert.Contains(t, query, "CONNECTION LIMIT 10")
	assert.Contains(t, query, "NOLOGIN") // since CanLogin is false
	assert.Contains(t, query, "SUPERUSER")
	assert.Contains(t, query, "NOINHERIT") // since InheritRole is false
	assert.Contains(t, query, "BYPASSRLS")
	assert.Contains(t, query, "NOCREATEDB")                // since CanCreateDB is false
	assert.Contains(t, query, "CREATEROLE")                // since CanCreateRole is true
	assert.Contains(t, query, "\"log_statement\" = 'all'") // config part
}

func TestBuildUpdateRoleQueryWithRename(t *testing.T) {
	newRole := objects.Role{
		Name: "new_role_name",
	}

	oldRole := objects.Role{
		Name: "old_role_name",
	}

	updateParam := objects.UpdateRoleParam{
		OldData: oldRole,
		ChangeItems: []objects.UpdateRoleType{
			objects.UpdateRoleName,
		},
	}

	query := BuildUpdateRoleQuery(newRole, updateParam)

	// Check that the query contains rename command
	assert.Contains(t, query, "RENAME TO \"new_role_name\"")
}

func TestBuildUpdateRoleQueryWithValidUntil(t *testing.T) {
	future := time.Now().Add(24 * time.Hour)
	newRole := objects.Role{
		Name:       "original_role",
		ValidUntil: objects.NewSupabaseTime(future),
	}

	updateParam := objects.UpdateRoleParam{
		OldData: objects.Role{Name: "original_role"},
		ChangeItems: []objects.UpdateRoleType{
			objects.UpdateRoleValidUntil,
		},
	}

	query := BuildUpdateRoleQuery(newRole, updateParam)
	assert.Contains(t, query, "VALID UNTIL '")
}

func TestBuildDeleteRoleQuery(t *testing.T) {
	role := objects.Role{
		Name: "role_to_delete",
	}

	query := BuildDeleteRoleQuery(role)

	// Check that the query contains expected parts (checking for parts that are actually in the generated query)
	assert.Contains(t, query, "role_to_delete")
	assert.Contains(t, query, "RAISE EXCEPTION 'Role")
	assert.Contains(t, query, "REVOKE %I FROM authenticator")
	assert.Contains(t, query, "REVOKE anon FROM %I")
	assert.Contains(t, query, "DROP ROLE %I")
}

func TestBuildRoleInheritQuery(t *testing.T) {
	// Test GRANT case
	query, err := BuildRoleInheritQuery("parent_role", "child_role", objects.UpdateRoleInheritGrant)
	assert.NoError(t, err)
	assert.Equal(t, `GRANT "child_role" TO "parent_role";`, query)

	// Test REVOKE case
	query, err = BuildRoleInheritQuery("parent_role", "child_role", objects.UpdateRoleInheritRevoke)
	assert.NoError(t, err)
	assert.Equal(t, `REVOKE "child_role" FROM "parent_role";`, query)
}

func TestBuildRoleInheritQueryErrorCases(t *testing.T) {
	// Test empty role name
	_, err := BuildRoleInheritQuery("", "child_role", objects.UpdateRoleInheritGrant)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "role name is required")

	// Test empty inherit role name
	_, err = BuildRoleInheritQuery("parent_role", "", objects.UpdateRoleInheritGrant)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "inherit role name is required")

	// Test unsupported action
	_, err = BuildRoleInheritQuery("parent_role", "child_role", "unsupported_action")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported role inherit action")
}

func TestRoleStructInQueryContext(t *testing.T) {
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

func TestUpdateRoleParamInQueryContext(t *testing.T) {
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
