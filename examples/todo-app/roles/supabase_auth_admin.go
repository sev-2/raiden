
package roles

import (
	"github.com/sev-2/raiden/pkg/postgres"
)

var SupabaseAuthAdmin = &postgres.Role{
	ActiveConnections : 0,
	CanBypassRLS : false,
	CanCreateDB : false,
	CanCreateRole : true,
	CanLogin : true,
	Config : map[string]any{"idle_in_transaction_session_timeout": "60000","search_path": "auth"},
	ConnectionLimit : 100,
	ID : 16504,
	InheritRole : false,
	IsReplicationRole : false,
	IsSuperuser : false,
	Name : "supabase_auth_admin",
	ValidUntil : nil,
}
