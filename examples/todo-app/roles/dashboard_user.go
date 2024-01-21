
package roles

import (
	"github.com/sev-2/raiden/pkg/postgres"
)

var DashboardUser = &postgres.Role{
	ActiveConnections : 0,
	CanBypassRLS : false,
	CanCreateDB : true,
	CanCreateRole : true,
	CanLogin : false,
	Config : map[string]any{},
	ConnectionLimit : 100,
	ID : 16564,
	InheritRole : true,
	IsReplicationRole : true,
	IsSuperuser : false,
	Name : "dashboard_user",
	ValidUntil : nil,
}
