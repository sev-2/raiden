
package roles

import (
	"github.com/sev-2/raiden/pkg/postgres"
)

var SupabaseReadOnlyUser = &postgres.Role{
	ActiveConnections : 0,
	CanBypassRLS : true,
	CanCreateDB : false,
	CanCreateRole : false,
	CanLogin : true,
	Config : map[string]any{},
	ConnectionLimit : 100,
	ID : 16390,
	InheritRole : true,
	IsReplicationRole : false,
	IsSuperuser : false,
	Name : "supabase_read_only_user",
	ValidUntil : nil,
}
