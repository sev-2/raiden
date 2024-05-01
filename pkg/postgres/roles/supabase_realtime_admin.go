package roles

import (
	"github.com/sev-2/raiden"
)

type SupabaseRealtimeAdmin struct {
	raiden.RoleBase
}

func (r *SupabaseRealtimeAdmin) Name() string {
	return "supabase_realtime_admin"
}

func (r *SupabaseRealtimeAdmin) IsReplicationRole() bool {
	return true
}

func (r *SupabaseRealtimeAdmin) CanLogin() bool {
	return true
}
