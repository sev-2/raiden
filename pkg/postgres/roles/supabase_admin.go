package roles

import (
	"github.com/sev-2/raiden"
)

type SupabaseAdmin struct {
	raiden.RoleBase
}

func (r *SupabaseAdmin) Name() string {
	return "supabase_admin"
}

func (r *SupabaseAdmin) IsReplicationRole() bool {
	return true
}
func (r *SupabaseAdmin) IsSuperuser() bool {
	return true
}

func (r *SupabaseAdmin) CanBypassRls() bool {
	return true
}

func (r *SupabaseAdmin) CanCreateDB() bool {
	return true
}

func (r *SupabaseAdmin) CanCreateRole() bool {
	return true
}

func (r *SupabaseAdmin) CanLogin() bool {
	return true
}
