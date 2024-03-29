package roles

import (
	"github.com/sev-2/raiden"
)

type PgsodiumKeymaker struct {
	raiden.RoleBase
}

func (r *PgsodiumKeymaker) Name() string {
	return "pgsodium_keymaker"
}
