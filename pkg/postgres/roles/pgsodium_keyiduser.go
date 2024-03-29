package roles

import (
	"github.com/sev-2/raiden"
)

type PgsodiumKeyiduser struct {
	raiden.RoleBase
}

func (r *PgsodiumKeyiduser) Name() string {
	return "pgsodium_keyiduser"
}
