package roles_test

import (
	"testing"

	"github.com/sev-2/raiden/pkg/resource/roles"
	"github.com/sev-2/raiden/pkg/state"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/stretchr/testify/assert"
)

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
