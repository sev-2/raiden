package roles

import (
	"github.com/sev-2/raiden"
)

type Authenticated struct {
	raiden.RoleBase
}

func (r *Authenticated) Name() string {
	return "authenticated"
}
