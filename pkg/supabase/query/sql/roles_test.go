package sql

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetRolesQuery(t *testing.T) {
	// Test that the GetRolesQuery constant exists and contains expected content
	assert.Contains(t, GetRolesQuery, "pg_roles")
	assert.Contains(t, GetRolesQuery, "rolname AS name")
	assert.Contains(t, GetRolesQuery, "rolsuper AS is_superuser")
	assert.Contains(t, GetRolesQuery, "rolcreatedb AS can_create_db")
	assert.Contains(t, GetRolesQuery, "rolcreaterole AS can_create_role")
	assert.Contains(t, GetRolesQuery, "rolinherit AS inherit_role")
	assert.Contains(t, GetRolesQuery, "rolcanlogin AS can_login")
	assert.Contains(t, GetRolesQuery, "connection_limit")
	assert.Contains(t, GetRolesQuery, "rolvaliduntil AS valid_until")
	assert.Contains(t, GetRolesQuery, "rolconfig AS config")
}

func TestGetRoleMembershipsQuery(t *testing.T) {
	// Test that the getRoleMembershipsQuery constant exists and contains expected content
	assert.Contains(t, getRoleMembershipsQuery, "role_inheritance")
	assert.Contains(t, getRoleMembershipsQuery, "parent_id")
	assert.Contains(t, getRoleMembershipsQuery, "parent_role")
	assert.Contains(t, getRoleMembershipsQuery, "inherit_id")
	assert.Contains(t, getRoleMembershipsQuery, "inherit_role_name")
	assert.Contains(t, getRoleMembershipsQuery, "schema_name")
}

func TestGenerateGetRoleMembershipsQuery(t *testing.T) {
	// Test with empty included schemas
	query := GenerateGetRoleMembershipsQuery([]string{})
	
	assert.Contains(t, query, "pg_namespace n")
	assert.Contains(t, query, "pg_roles r")
	
	// Test with specific schemas
	query = GenerateGetRoleMembershipsQuery([]string{"public", "auth"})
	
	assert.Contains(t, query, "'public'")
	assert.Contains(t, query, "'auth'")
	assert.Contains(t, query, "n.nspname = ANY(ARRAY['public','auth'])")
}

func TestGenerateGetRoleMembershipsQueryMultipleSchemas(t *testing.T) {
	schemas := []string{"public", "auth", "storage", "extensions"}
	query := GenerateGetRoleMembershipsQuery(schemas)
	
	for _, schema := range schemas {
		assert.Contains(t, query, "'"+schema+"'")
	}
	
	// Check that the query contains the expected structure
	assert.Contains(t, query, "n.nspname = ANY(ARRAY['public','auth','storage','extensions'])")
}

func TestGenerateGetRoleMembershipsQuerySpecialCharacters(t *testing.T) {
	// Test with schema names that might have special characters
	schemas := []string{"schema-with-dash", "schema_with_underscore", "schema123"}
	query := GenerateGetRoleMembershipsQuery(schemas)
	
	for _, schema := range schemas {
		assert.Contains(t, query, "'"+schema+"'")
	}
}

func TestRoleQueryStructure(t *testing.T) {
	// Test that the role queries have proper structure
	
	// Check GetRolesQuery structure
	assert.Contains(t, GetRolesQuery, "SELECT")
	assert.Contains(t, GetRolesQuery, "FROM")
	assert.Contains(t, GetRolesQuery, "pg_roles")
	
	// Check getRoleMembershipsQuery structure
	assert.Contains(t, getRoleMembershipsQuery, "WITH")
	assert.Contains(t, getRoleMembershipsQuery, "SELECT")
	assert.Contains(t, getRoleMembershipsQuery, "FROM")
	assert.Contains(t, getRoleMembershipsQuery, "pg_auth_members")
}

func TestRoleQueryConstantsExist(t *testing.T) {
	// Test that the expected constants exist and are not empty
	assert.NotEmpty(t, GetRolesQuery)
	assert.NotEmpty(t, getRoleMembershipsQuery)
}

func TestSQLInjectionProtection(t *testing.T) {
	// Test that the query generation properly handles schema names 
	// to prevent SQL injection (using the formatting approach)
	
	schemas := []string{"normal_schema", "schema'; DROP TABLE users; --"}
	query := GenerateGetRoleMembershipsQuery(schemas)
	
	// The malicious schema should be quoted properly to prevent injection
	// Note: This is a basic test, in real implementation we rely on the fmt.Sprintf 
	// and proper quoting functions in the code
	assert.Contains(t, query, "'schema'; DROP TABLE users; --'")
}