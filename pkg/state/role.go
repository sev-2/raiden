package state

import (
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/supabase/objects"
)

func ToRoles(roleStates []RoleState, appRoles []raiden.Role, withNativeRole bool) (roles []objects.Role, err error) {
	mapRoleState := map[string]RoleState{}
	for i := range roleStates {
		r := roleStates[i]
		mapRoleState[r.Role.Name] = r
	}

	for _, role := range appRoles {
		state, isStateExist := mapRoleState[role.Name()]
		if !isStateExist {
			return
		}

		if state.IsNative && withNativeRole {
			roles = append(roles, state.Role)
		}

		if !state.IsNative {
			sr, err := createRoleFromState(state, role)
			if err != nil {
				return roles, err
			}
			roles = append(roles, sr)
		}
	}

	return
}

func createRoleFromState(rs RoleState, role raiden.Role) (r objects.Role, err error) {
	r = rs.Role
	r.ConnectionLimit = role.ConnectionLimit()
	r.CanBypassRLS = role.CanBypassRls()
	r.CanCreateDB = role.CanCreateDB()
	r.CanCreateRole = role.CanCreateRole()
	r.CanLogin = role.CanLogin()
	r.InheritRole = role.InheritRole()
	r.IsReplicationRole = role.IsReplicationRole()
	r.IsSuperuser = role.IsSuperuser()

	return
}
