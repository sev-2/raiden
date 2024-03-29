package roles

import (
	"github.com/sev-2/raiden"
)

type PgMonitor struct {
	raiden.RoleBase
}

func (r *PgMonitor) Name() string {
	return "pg_monitor"
}
