package roles

import (
	"github.com/sev-2/raiden"
)

type DashboardUser struct {
	raiden.RoleBase
}

func (r *DashboardUser) Name() string {
	return "dashboard_user"
}

func (r *DashboardUser) IsReplicationRole() bool {
	return true
}

func (r *DashboardUser) CanCreateDB() bool {
	return true
}

func (r *DashboardUser) CanCreateRole() bool {
	return true
}
