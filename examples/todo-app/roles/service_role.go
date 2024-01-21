
package roles

import (
	"github.com/sev-2/raiden/pkg/postgres"
)

var ServiceRole = &postgres.Role{
	ActiveConnections : 0,
	CanBypassRLS : true,
	CanCreateDB : false,
	CanCreateRole : false,
	CanLogin : false,
	Config : map[string]any{},
	ConnectionLimit : 100,
	ID : 16449,
	InheritRole : true,
	IsReplicationRole : false,
	IsSuperuser : false,
	Name : "service_role",
	ValidUntil : nil,
}
