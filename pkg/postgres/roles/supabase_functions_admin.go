package roles

import (
	"github.com/sev-2/raiden"
)

type SupabaseFunctionsAdmin struct {
	raiden.RoleBase
}

func (r *SupabaseFunctionsAdmin) Name() string {
	return "supabase_functions_admin"
}

func (r *SupabaseFunctionsAdmin) InheritRole() bool {
	return false
}

func (r *SupabaseFunctionsAdmin) CanCreateRole() bool {
	return true
}

func (r *SupabaseFunctionsAdmin) CanLogin() bool {
	return true
}
