
package roles

import (
	"github.com/sev-2/raiden/pkg/postgres"
)

var SupabaseStorageAdmin = &postgres.Role{
	ActiveConnections : 0,
	CanBypassRLS : false,
	CanCreateDB : false,
	CanCreateRole : true,
	CanLogin : true,
	Config : map[string]any{"search_path": "storage"},
	ConnectionLimit : 100,
	ID : 16559,
	InheritRole : false,
	IsReplicationRole : false,
	IsSuperuser : false,
	Name : "supabase_storage_admin",
	ValidUntil : nil,
}
