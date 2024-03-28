package roles

import (
	"github.com/sev-2/raiden"
)

type PgStatScanTables struct {
	raiden.RoleBase
}

func (r *PgStatScanTables) Name() string {
	return "pg_stat_scan_tables"
}
