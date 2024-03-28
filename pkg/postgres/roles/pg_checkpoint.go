package roles

import (
	"github.com/sev-2/raiden"
)

type PgCheckpoint struct {
	raiden.RoleBase
}

func (r *PgCheckpoint) Name() string {
	return "pg_checkpoint"
}
