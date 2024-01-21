
package roles

import (
	"github.com/sev-2/raiden/pkg/postgres"
)

var Anon = &postgres.Role{
	ActiveConnections : 0,
	CanBypassRLS : false,
	CanCreateDB : false,
	CanCreateRole : false,
	CanLogin : false,
	Config : map[string]any{"statement_timeout": "3s"},
	ConnectionLimit : 100,
	ID : 16447,
	InheritRole : true,
	IsReplicationRole : false,
	IsSuperuser : false,
	Name : "anon",
	ValidUntil : nil,
}
