package roles

import (
	"github.com/sev-2/raiden"
)

type SupabaseReplicationAdmin struct {
	raiden.RoleBase
}

func (r *SupabaseReplicationAdmin) Name() string {
	return "supabase_replication_admin"
}

func (r *SupabaseReplicationAdmin) IsReplicationRole() bool {
	return true
}

func (r *SupabaseReplicationAdmin) CanLogin() bool {
	return true
}
