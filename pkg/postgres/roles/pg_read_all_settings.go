package roles

import (
	"github.com/sev-2/raiden"
)

type PgReadAllSettings struct {
	raiden.RoleBase
}

func (r *PgReadAllSettings) Name() string {
	return "pg_read_all_settings"
}
