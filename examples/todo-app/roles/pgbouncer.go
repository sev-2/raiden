
package roles

import (
	"github.com/sev-2/raiden/pkg/postgres"
)

var Pgbouncer = &postgres.Role{
	ActiveConnections : 0,
	CanBypassRLS : false,
	CanCreateDB : false,
	CanCreateRole : false,
	CanLogin : true,
	Config : map[string]any{},
	ConnectionLimit : 100,
	ID : 16384,
	InheritRole : true,
	IsReplicationRole : false,
	IsSuperuser : false,
	Name : "pgbouncer",
	ValidUntil : nil,
}
