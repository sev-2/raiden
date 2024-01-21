
package roles

import (
	"github.com/sev-2/raiden/pkg/postgres"
)

var SupabaseReplicationAdmin = &postgres.Role{
	ActiveConnections : 0,
	CanBypassRLS : false,
	CanCreateDB : false,
	CanCreateRole : false,
	CanLogin : true,
	Config : map[string]any{},
	ConnectionLimit : 100,
	ID : 16389,
	InheritRole : true,
	IsReplicationRole : true,
	IsSuperuser : false,
	Name : "supabase_replication_admin",
	ValidUntil : nil,
}
