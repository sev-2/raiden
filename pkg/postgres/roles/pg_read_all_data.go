package roles

import (
	"github.com/sev-2/raiden"
)

type PgReadAllData struct {
	raiden.RoleBase
}

func (r *PgReadAllData) Name() string {
	return "pg_read_all_data"
}
