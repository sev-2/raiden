package acl

import (
	"testing"
	"time"

	"github.com/sev-2/raiden/pkg/state"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/stretchr/testify/require"
)

func TestRolesAccessorsAndQueries(t *testing.T) {
	validUntil := objects.NewSupabaseTime(time.Date(2024, time.January, 2, 15, 4, 5, 0, time.UTC))

	inherit := &objects.Role{Name: "child", ConnectionLimit: 5}
	role := objects.Role{
		Name:            "parent",
		ConnectionLimit: 10,
		InheritRole:     true,
		CanBypassRLS:    true,
		CanCreateDB:     true,
		CanCreateRole:   true,
		CanLogin:        true,
		ValidUntil:      validUntil,
		InheritRoles:    []*objects.Role{inherit},
	}

	native := objects.Role{Name: "postgres", ConnectionLimit: -1}

	st := &state.State{
		Roles: []state.RoleState{
			{Role: role, IsNative: false},
			{Role: native, IsNative: true},
		},
	}

	prepareState(t, st)

	got, err := GetRole("parent")
	require.NoError(t, err)
	require.Equal(t, "parent", got.Name())
	require.Equal(t, 10, got.ConnectionLimit())
	require.True(t, got.IsInheritRole())
	require.True(t, got.CanBypassRls())
	require.True(t, got.CanCreateDB())
	require.True(t, got.CanCreateRole())
	require.True(t, got.CanLogin())
	require.Equal(t, validUntil, got.ValidUntil())

	inheritRoles := got.InheritRoles()
	require.Len(t, inheritRoles, 1)
	require.Equal(t, "child", inheritRoles[0].Name())

	rrWithNil := newRaidenRole(objects.Role{Name: "with-nil", InheritRoles: []*objects.Role{inherit, nil}})
	require.Len(t, rrWithNil.InheritRoles(), 1)

	_, errMissing := GetRole("missing")
	require.EqualError(t, errMissing, "role missing is not available")

	roles, err := GetRoles()
	require.NoError(t, err)
	require.Len(t, roles, 1)
	require.Equal(t, "parent", roles[0].Name())

	available, err := GetAvailableRole()
	require.NoError(t, err)
	require.Len(t, available, 2)

	require.NoError(t, validateRole("parent"))
	require.NoError(t, validateRole("postgres"))
	require.Error(t, validateRole("ghost"))
}
