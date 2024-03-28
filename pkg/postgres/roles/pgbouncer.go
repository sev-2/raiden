package roles

import (
	"github.com/sev-2/raiden"
)

type Pgbouncer struct {
	raiden.RoleBase
}

func (r *Pgbouncer) Name() string {
	return "pgbouncer"
}

func (r *Pgbouncer) CanLogin() bool {
	return true
}
