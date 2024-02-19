package roles

import (
	"github.com/sev-2/raiden"
)

type PgReadServerFiles struct {
	raiden.RoleBase
}

func (r *PgReadServerFiles) Name() string {
	return "pg_read_server_files"
}
