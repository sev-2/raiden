
package roles

import (
	"github.com/sev-2/raiden/pkg/postgres"
)

var Postgres = &postgres.Role{
	ActiveConnections : 5,
	CanBypassRLS : true,
	CanCreateDB : true,
	CanCreateRole : true,
	CanLogin : true,
	Config : map[string]any{"search_path": "\"\\$user\", public, extensions"},
	ConnectionLimit : 100,
	ID : 10,
	InheritRole : true,
	IsReplicationRole : true,
	IsSuperuser : false,
	Name : "postgres",
	ValidUntil : nil,
}
