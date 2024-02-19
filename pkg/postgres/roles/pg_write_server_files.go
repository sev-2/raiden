package roles

import (
	"github.com/sev-2/raiden"
)

type PgWriteServerFiles struct {
	raiden.RoleBase
}

func (r *PgWriteServerFiles) Name() string {
	return "pg_write_server_files"
}
