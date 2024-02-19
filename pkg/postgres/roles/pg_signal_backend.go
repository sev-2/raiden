package roles

import (
	"github.com/sev-2/raiden"
)

type PgSignalBackend struct {
	raiden.RoleBase
}

func (r *PgSignalBackend) Name() string {
	return "pg_signal_backend"
}
