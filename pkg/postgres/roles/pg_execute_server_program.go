package roles

import (
	"github.com/sev-2/raiden"
)

type PgExecuteServerProgram struct {
	raiden.RoleBase
}

func (r *PgExecuteServerProgram) Name() string {
	return "pg_execute_server_program"
}
