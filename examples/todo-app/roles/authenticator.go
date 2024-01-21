
package roles

import (
	"github.com/sev-2/raiden/pkg/postgres"
)

var Authenticator = &postgres.Role{
	ActiveConnections : 1,
	CanBypassRLS : false,
	CanCreateDB : false,
	CanCreateRole : false,
	CanLogin : true,
	Config : map[string]any{"statement_timeout": "8s","lock_timeout": "8s","session_preload_libraries": "supautils, safeupdate"},
	ConnectionLimit : 100,
	ID : 16450,
	InheritRole : false,
	IsReplicationRole : false,
	IsSuperuser : false,
	Name : "authenticator",
	ValidUntil : nil,
}
