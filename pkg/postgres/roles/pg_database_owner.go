package roles

import (
	"github.com/sev-2/raiden"
)

type PgDatabaseOwner struct {
	raiden.RoleBase
}

func (r *PgDatabaseOwner) Name() string {
	return "pg_database_owner"
}
