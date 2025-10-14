package state_test

import (
	"testing"
	"time"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/state"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/stretchr/testify/assert"
)

type MockRole struct {
	raiden.RoleBase
}

type MockChildRole struct {
	raiden.RoleBase
}

type mockUnnamedRole struct {
	raiden.RoleBase
}

type mockExplicitRole struct {
	raiden.RoleBase
}

type mockEdgeRole struct {
	raiden.RoleBase
}

func (r *MockRole) Name() string {
	return "test_role"
}

func (r *MockRole) ConnectionLimit() int {
	return 10
}

func (r *MockRole) IsInheritRole() bool {
	return true
}

func (r *MockRole) InheritRoles() []raiden.Role {
	return []raiden.Role{&MockChildRole{}}
}

func (r *MockRole) IsReplicationRole() bool {
	return true
}

func (r *MockRole) CanBypassRls() bool {
	return true
}

func (r *MockRole) CanCreateDB() bool {
	return true
}

func (r *MockRole) CanCreateRole() bool {
	return true
}

func (r *MockRole) CanLogin() bool {
	return true
}

func (r *MockRole) ValidUntil() *objects.SupabaseTime {
	return objects.NewSupabaseTime(time.Now())
}

func (r *MockChildRole) Name() string {
	return "child_role"
}

func (r *mockUnnamedRole) Name() string {
	return ""
}

func (r *mockExplicitRole) Name() string {
	return "explicit_role"
}

func (r *mockEdgeRole) Name() string {
	return "edge_role"
}

func (r *mockEdgeRole) IsInheritRole() bool {
	return true
}

func (r *mockEdgeRole) InheritRoles() []raiden.Role {
	return []raiden.Role{
		nil,
		&mockUnnamedRole{},
		&mockUnnamedRole{},
		&mockExplicitRole{},
	}
}

func TestExtractRole(t *testing.T) {
	roleStates := []state.RoleState{
		{Role: objects.Role{Name: "existing_role"}, IsNative: false},
		{Role: objects.Role{Name: "native_role"}, IsNative: true},
	}

	appRoles := []raiden.Role{
		&MockRole{},
	}

	result, err := state.ExtractRole(roleStates, appRoles, true)
	assert.NoError(t, err)
	assert.Len(t, result.Existing, 0)
	assert.Len(t, result.New, 1)
	assert.Len(t, result.Delete, 2)

	result, err = state.ExtractRole(roleStates, appRoles, false)
	assert.NoError(t, err)
	assert.Len(t, result.Existing, 0)
	assert.Len(t, result.New, 1)
	assert.Len(t, result.Delete, 1)
}

func TestBindToSupabaseRole(t *testing.T) {
	role := MockRole{}
	r := objects.Role{}

	state.BindToSupabaseRole(&r, &role)
	assert.Equal(t, "test_role", r.Name)
	assert.Equal(t, 10, r.ConnectionLimit)
	assert.True(t, r.CanBypassRLS)
	assert.True(t, r.CanCreateDB)
	assert.True(t, r.CanCreateRole)
	assert.True(t, r.CanLogin)
	assert.True(t, r.InheritRole)
	assert.NotNil(t, r.ValidUntil)
	if assert.Len(t, r.InheritRoles, 1) {
		assert.Equal(t, "child_role", r.InheritRoles[0].Name)
	}
}

func TestBuildRoleFromState(t *testing.T) {
	timeNow := objects.NewSupabaseTime(time.Now())
	rs := state.RoleState{
		Role: objects.Role{
			Name:            "test_role",
			ConnectionLimit: 10,
			CanBypassRLS:    true,
			CanCreateDB:     true,
			CanCreateRole:   true,
			CanLogin:        true,
			InheritRole:     true,
			ValidUntil:      timeNow,
		},
	}
	role := &MockRole{}

	r := state.BuildRoleFromState(rs, role)
	assert.Equal(t, "test_role", r.Name)
	assert.Equal(t, 10, r.ConnectionLimit)
	assert.True(t, r.CanBypassRLS)
	assert.True(t, r.CanCreateDB)
	assert.True(t, r.CanCreateRole)
	assert.True(t, r.CanLogin)
	assert.True(t, r.InheritRole)
	assert.NotNil(t, r.ValidUntil)
}

func TestBindToSupabaseRole_InheritRoleEdgeCases(t *testing.T) {
	role := mockEdgeRole{}
	var res objects.Role

	state.BindToSupabaseRole(&res, &role)

	assert.Equal(t, "edge_role", res.Name)
	assert.True(t, res.InheritRole)
	if assert.Len(t, res.InheritRoles, 2) {
		names := []string{res.InheritRoles[0].Name, res.InheritRoles[1].Name}
		assert.Equal(t, "mock_unnamed_role", names[0])
		assert.Equal(t, "explicit_role", names[1])
	}
}

func TestExtractRoleResult_ToDeleteFlatMap(t *testing.T) {
	extractRoleResult := state.ExtractRoleResult{
		Delete: []objects.Role{
			{Name: "role1"},
			{Name: "role2"},
		},
	}

	mapData := extractRoleResult.ToDeleteFlatMap()
	assert.Len(t, mapData, 2)
	assert.Contains(t, mapData, "role1")
	assert.Contains(t, mapData, "role2")
	assert.Equal(t, "role1", mapData["role1"].Name)
	assert.Equal(t, "role2", mapData["role2"].Name)
}
