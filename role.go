package raiden

import (
	"github.com/sev-2/raiden/pkg/supabase/objects"
)

type (
	Role interface {
		// name
		Name() string

		// default 60
		ConnectionLimit() int

		// default true
		InheritRole() bool

		// Disable for now, because need super user role for set it
		// // default false
		// IsReplicationRole() bool
		// // default false
		// IsSuperuser() bool

		// default false
		CanBypassRls() bool

		// default false
		CanCreateDB() bool

		// default false
		CanCreateRole() bool

		// default false
		CanLogin() bool

		// default nil
		ValidUntil() *objects.ValidUntil
	}

	RoleBase struct {
	}
)

const (
	DefaultRoleValidUntilLayout = "2006-01-02"
	DefaultRoleConnectionLimit  = 60
)

// ----- Base Role Default Func -----
func (r *RoleBase) ConnectionLimit() int {
	return DefaultRoleConnectionLimit
}

func (r *RoleBase) InheritRole() bool {
	return true
}

// func (r *RoleBase) IsReplicationRole() bool {
// 	return false
// }

// func (r *RoleBase) IsSuperuser() bool {
// 	return false
// }

func (r *RoleBase) CanBypassRls() bool {
	return false
}

func (r *RoleBase) CanCreateDB() bool {
	return false
}

func (r *RoleBase) CanCreateRole() bool {
	return false
}

func (r *RoleBase) CanLogin() bool {
	return false
}

func (r *RoleBase) ValidUntil() *objects.ValidUntil {
	return nil
}
