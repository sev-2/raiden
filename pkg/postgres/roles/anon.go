package roles

import (
	"github.com/sev-2/raiden"
)

type Anon struct {
	raiden.RoleBase
}

func (r *Anon) Name() string {
	return "anon"
}
