package roles

import (
	"github.com/sev-2/raiden"
)

type SupabaseStorageAdmin struct {
	raiden.RoleBase
}

func (r *SupabaseStorageAdmin) Name() string {
	return "supabase_storage_admin"
}

func (r *SupabaseStorageAdmin) InheritRole() bool {
	return false
}

func (r *SupabaseStorageAdmin) CanCreateRole() bool {
	return true
}

func (r *SupabaseStorageAdmin) CanLogin() bool {
	return true
}
