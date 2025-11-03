package roles_test

import (
	"testing"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/resource/roles"
	"github.com/sev-2/raiden/pkg/state"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/stretchr/testify/assert"
)

type stubRole struct {
	raiden.RoleBase
	name string
}

func (r *stubRole) Name() string {
	return r.name
}

func TestGetNewCountData(t *testing.T) {
	supabaseRoles := []objects.Role{
		{Name: "role1"},
		{Name: "role2"},
		{Name: "role3"},
	}

	extractResult := state.ExtractRoleResult{
		Delete: []objects.Role{
			{Name: "role1"},
			{Name: "role2"},
		},
	}

	count := roles.GetNewCountData(supabaseRoles, extractResult)
	assert.Equal(t, 2, count)
}

func TestGetNewCountDataNoMatch(t *testing.T) {
	supabaseRoles := []objects.Role{
		{Name: "role1"},
		{Name: "role2"},
	}

	extractResult := state.ExtractRoleResult{
		Delete: []objects.Role{
			{Name: "role3"},
			{Name: "role4"},
		},
	}

	count := roles.GetNewCountData(supabaseRoles, extractResult)
	assert.Equal(t, 0, count)
}

func TestGetNewCountDataEmpty(t *testing.T) {
	supabaseRoles := []objects.Role{}

	extractResult := state.ExtractRoleResult{
		Delete: []objects.Role{},
	}

	count := roles.GetNewCountData(supabaseRoles, extractResult)
	assert.Equal(t, 0, count)
}

func TestGetNewCountDataPartialMatch(t *testing.T) {
	supabaseRoles := []objects.Role{
		{Name: "role1"},
		{Name: "role2"},
		{Name: "role3"},
	}

	extractResult := state.ExtractRoleResult{
		Delete: []objects.Role{
			{Name: "role1"},
			{Name: "role4"},
		},
	}

	count := roles.GetNewCountData(supabaseRoles, extractResult)
	assert.Equal(t, 1, count)
}

func TestAttachInherithRole_NoMemberships(t *testing.T) {
	nativeRoles := map[string]raiden.Role{}

	supabaseRoles := objects.Roles{
		{ID: 1, Name: "user_role"},
	}

	roleMemberships := objects.RoleMemberships{} // Empty

	result := roles.AttachInherithRole(nativeRoles, supabaseRoles, roleMemberships)
	assert.Len(t, result, 1)
	assert.Empty(t, result[0].InheritRoles) // No memberships, so no inheritance
}

func TestAttachInherithRole_MixedMemberships(t *testing.T) {
	nativeRoles := map[string]raiden.Role{
		"native_parent": &stubRole{name: "native_parent"},
		"native_target": &stubRole{name: "native_target"},
	}

	supabaseRoles := objects.Roles{
		{ID: 1, Name: "target_role"},
		{ID: 2, Name: "parent_role"},
		{ID: 3, Name: "native_parent"},
		{ID: 4, Name: "native_target"},
	}

	roleMemberships := objects.RoleMemberships{
		{ParentID: 2, ParentRole: "parent_role", InheritID: 1, InheritRole: "target_role"},
		{ParentID: 3, ParentRole: "native_parent", InheritID: 1, InheritRole: "target_role"},
		{ParentID: 999, ParentRole: "ghost_parent", InheritID: 1, InheritRole: "target_role"},
		{ParentID: 2, ParentRole: "parent_role", InheritID: 4, InheritRole: "native_target"},
	}

	result := roles.AttachInherithRole(nativeRoles, supabaseRoles, roleMemberships)

	var target objects.Role
	var found bool
	for _, r := range result {
		if r.Name == "target_role" {
			target = r
			found = true
		}
		if r.Name == "native_target" {
			assert.Empty(t, r.InheritRoles)
		}
	}

	assert.True(t, found)
	if assert.Len(t, target.InheritRoles, 1) {
		assert.Equal(t, "parent_role", target.InheritRoles[0].Name)
	}
}
