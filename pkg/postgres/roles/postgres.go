package roles

import (
	"github.com/sev-2/raiden"
)

type Postgres struct {
	raiden.RoleBase
}

func (r *Postgres) Name() string {
	return "postgres"
}

func (r *Postgres) IsReplicationRole() bool {
	return true
}

func (r *Postgres) CanBypassRls() bool {
	return true
}

func (r *Postgres) CanCreateDB() bool {
	return true
}

func (r *Postgres) CanCreateRole() bool {
	return true
}

func (r *Postgres) CanLogin() bool {
	return true
}
