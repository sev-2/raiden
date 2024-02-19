package roles

import (
	"github.com/sev-2/raiden"
)

type SupabaseAuthAdmin struct {
	raiden.RoleBase
}

func (r *SupabaseAuthAdmin) Name() string {
	return "supabase_auth_admin"
}

func (r *SupabaseAuthAdmin) InheritRole() bool {
	return false
}

func (r *SupabaseAuthAdmin) CanCreateRole() bool {
	return true
}

func (r *SupabaseAuthAdmin) CanLogin() bool {
	return true
}
