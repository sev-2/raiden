package roles

import (
	"github.com/sev-2/raiden"
)

type Authenticator struct {
	raiden.RoleBase
}

func (r *Authenticator) Name() string {
	return "authenticator"
}

func (r *Authenticator) InheritRole() bool {
	return false
}

func (r *Authenticator) CanLogin() bool {
	return true
}
