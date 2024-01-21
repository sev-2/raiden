
package roles

import (
	"github.com/sev-2/raiden/pkg/postgres"
)

var Authenticated = &postgres.Role{
	ActiveConnections : 0,
	CanBypassRLS : false,
	CanCreateDB : false,
	CanCreateRole : false,
	CanLogin : false,
	Config : map[string]any{"statement_timeout": "8s"},
	ConnectionLimit : 100,
	ID : 16448,
	InheritRole : true,
	IsReplicationRole : false,
	IsSuperuser : false,
	Name : "authenticated",
	ValidUntil : nil,
}
