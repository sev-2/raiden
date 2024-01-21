
package roles

import (
	"github.com/sev-2/raiden/pkg/postgres"
)

var SupabaseFunctionsAdmin = &postgres.Role{
	ActiveConnections : 0,
	CanBypassRLS : false,
	CanCreateDB : false,
	CanCreateRole : true,
	CanLogin : true,
	Config : map[string]any{"search_path": "supabase_functions"},
	ConnectionLimit : 100,
	ID : 17037,
	InheritRole : false,
	IsReplicationRole : false,
	IsSuperuser : false,
	Name : "supabase_functions_admin",
	ValidUntil : nil,
}
