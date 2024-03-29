package roles

import (
	"github.com/sev-2/raiden"
)

type ServiceRole struct {
	raiden.RoleBase
}

func (r *ServiceRole) Name() string {
	return "service_role"
}

func (r *ServiceRole) CanBypassRls() bool {
	return true
}
