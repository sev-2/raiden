package acl

import (
	"fmt"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/state"
	"github.com/sev-2/raiden/pkg/supabase/objects"
)

// custom reiden role

func newRaidenRole(role objects.Role) raidenRole {
	return raidenRole{
		role: role,
	}
}

type raidenRole struct {
	raiden.RoleBase
	role objects.Role
}

func (r *raidenRole) Name() string {
	return r.role.Name
}

func (r *raidenRole) ConnectionLimit() int {
	return r.role.ConnectionLimit
}

func (r *raidenRole) IsInheritRole() bool {
	return r.role.InheritRole
}

func (r *raidenRole) InheritRoles() []raiden.Role {
	rr := make([]raiden.Role, 0)

	if len(r.role.InheritRoles) > 0 {
		for _, ir := range r.role.InheritRoles {
			if ir != nil {
				nr := newRaidenRole(*ir)
				rr = append(rr, &nr)
			}
		}
	}

	return rr
}

func (r *raidenRole) CanBypassRls() bool {
	return r.role.CanBypassRLS
}

func (r *raidenRole) CanCreateDB() bool {
	return r.role.CanCreateDB
}

func (r *raidenRole) CanCreateRole() bool {
	return r.role.CanCreateRole
}

func (r *raidenRole) CanLogin() bool {
	return r.role.CanLogin
}

func (r *raidenRole) ValidUntil() *objects.SupabaseTime {
	return r.role.ValidUntil
}

// GetRole
func GetRole(name string) (raiden.Role, error) {
	// get available role
	avRole, err := GetAvailableRole()
	if err != nil {
		return nil, err
	}

	// validate role
	var foundRole *objects.Role
	for i := range avRole {
		r := avRole[i]
		if r.Name == name {
			foundRole = &r
			break
		}
	}

	if foundRole == nil {
		return nil, fmt.Errorf("role %s is not available", name)
	}

	raidenRole := newRaidenRole(*foundRole)
	return &raidenRole, nil
}

func GetRoles() ([]raiden.Role, error) {
	st, err := state.Load()
	if err != nil {
		return nil, err
	}

	// validate role
	roles := make([]raiden.Role, 0)
	for _, r := range st.Roles {
		if r.IsNative {
			continue
		}
		rr := newRaidenRole(r.Role)
		roles = append(roles, &rr)
	}

	return roles, nil
}

func GetAvailableRole() (roles []objects.Role, err error) {
	st, err := state.Load()
	if err != nil {
		return roles, err
	}

	for i := range st.Roles {
		r := st.Roles[i]
		roles = append(roles, r.Role)
	}
	return
}

func validateRole(roleName string) (err error) {
	avRoles, err := GetAvailableRole()
	if err != nil {
		return err
	}

	found := false
	for i := range avRoles {
		r := avRoles[i]
		if r.Name == roleName {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("role %s is not available in database", roleName)
	}

	return nil
}
