package roles

import (
	"github.com/sev-2/raiden"
)

type PgWriteAllData struct {
	raiden.RoleBase
}

func (r *PgWriteAllData) Name() string {
	return "pg_write_all_data"
}
