package roles

import (
	"github.com/sev-2/raiden"
)

type SupabaseReadOnlyUser struct {
	raiden.RoleBase
}

func (r *SupabaseReadOnlyUser) Name() string {
	return "supabase_read_only_user"
}

func (r *SupabaseReadOnlyUser) CanBypassRls() bool {
	return true
}

func (r *SupabaseReadOnlyUser) CanLogin() bool {
	return true
}
