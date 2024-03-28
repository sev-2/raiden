package state

import (
	"reflect"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/utils"
)

type ExtractRoleResult struct {
	Existing []objects.Role
	New      []objects.Role
	Delete   []objects.Role
}

func ExtractRole(roleStates []RoleState, appRoles []raiden.Role, withNativeRole bool) (result ExtractRoleResult, err error) {
	mapRoleState := map[string]RoleState{}
	for i := range roleStates {
		r := roleStates[i]
		mapRoleState[r.Role.Name] = r
	}

	for _, role := range appRoles {
		state, isStateExist := mapRoleState[role.Name()]
		if !isStateExist {
			r := objects.Role{}
			BindToSupabaseRole(&r, role)
			result.New = append(result.New, r)
			continue
		}

		if state.IsNative && withNativeRole {
			result.Existing = append(result.Existing, state.Role)
		}

		if !state.IsNative {
			sr := BuildRoleFromState(state, role)
			result.Existing = append(result.Existing, sr)
		}

		delete(mapRoleState, role.Name())
	}

	for _, state := range mapRoleState {
		result.Delete = append(result.Delete, state.Role)
	}

	return
}

func BindToSupabaseRole(r *objects.Role, role raiden.Role) {
	name := role.Name()
	if name == "" {
		rv := reflect.TypeOf(role)
		name = utils.ToSnakeCase(rv.Name())
	}

	r.Name = name
	r.ConnectionLimit = role.ConnectionLimit()
	r.CanBypassRLS = role.CanBypassRls()
	r.CanCreateDB = role.CanCreateDB()
	r.CanCreateRole = role.CanCreateRole()
	r.CanLogin = role.CanLogin()
	r.InheritRole = role.InheritRole()
	r.ValidUntil = role.ValidUntil()

	// need role with superuser to create new superuser role and set replication
	// r.IsReplicationRole = role.IsReplicationRole()
	// r.IsSuperuser = role.IsSuperuser()
}

func BuildRoleFromState(rs RoleState, role raiden.Role) (r objects.Role) {
	r = rs.Role
	BindToSupabaseRole(&r, role)
	return
}
