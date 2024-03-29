package roles

import (
	"github.com/sev-2/raiden"
)

type PgReadAllStats struct {
	raiden.RoleBase
}

func (r *PgReadAllStats) Name() string {
	return "pg_read_all_stats"
}
