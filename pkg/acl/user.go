package acl

import (
	"fmt"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/state"
	"github.com/sev-2/raiden/pkg/supabase"
	"github.com/sev-2/raiden/pkg/supabase/objects"
)

func SetUserRole(cfg *raiden.Config, userId string, role raiden.Role) error {
	if err := validateRole(role.Name()); err != nil {
		return err
	}
	return doSetRole(cfg, userId, role.Name())
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

func GetUserRole() (roleName string, err error) {
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

func doSetRole(cfg *raiden.Config, userId string, roleName string) error {
	// call supabase with admin privilege for update user, user service account for this case
	data := objects.User{Role: roleName}
	_, err := supabase.AdminUpdateUserData(cfg, userId, data)
	if err != nil {
		return err
	}
	return nil
}
