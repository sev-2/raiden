package roles

import (
	"github.com/sev-2/raiden"
)

type PgsodiumKeyholder struct {
	raiden.RoleBase
}

func (r *PgsodiumKeyholder) Name() string {
	return "pgsodium_keyholder"
}
