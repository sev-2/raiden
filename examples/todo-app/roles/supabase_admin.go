
package roles

import (
	"github.com/sev-2/raiden/pkg/postgres"
)

var SupabaseAdmin = &postgres.Role{
	ActiveConnections : 5,
	CanBypassRLS : true,
	CanCreateDB : true,
	CanCreateRole : true,
	CanLogin : true,
	Config : map[string]any{"search_path": "\"\\$user\", public, auth, extensions"},
	ConnectionLimit : 100,
	ID : 16388,
	InheritRole : true,
	IsReplicationRole : true,
	IsSuperuser : true,
	Name : "supabase_admin",
	ValidUntil : nil,
}
