package raiden_test

import (
	"testing"

	"github.com/sev-2/raiden"
	"github.com/stretchr/testify/assert"
)

func TestRoleBase_ConnectionLimit(t *testing.T) {
	roleBase := raiden.RoleBase{}
	assert.Equal(t, 60, roleBase.ConnectionLimit())
}

func TestRoleBase_InheritRole(t *testing.T) {
	roleBase := raiden.RoleBase{}
	assert.Equal(t, true, roleBase.InheritRole())
}

func TestRoleBase_CanBypassRls(t *testing.T) {
	roleBase := raiden.RoleBase{}
	assert.Equal(t, false, roleBase.CanBypassRls())
}

func TestRoleBase_CanCreateDB(t *testing.T) {
	roleBase := raiden.RoleBase{}
	assert.Equal(t, false, roleBase.CanCreateDB())
}

func TestRoleBase_CanCreateRole(t *testing.T) {
	roleBase := raiden.RoleBase{}
	assert.Equal(t, false, roleBase.CanCreateRole())
}

func TestRoleBase_CanLogin(t *testing.T) {
	roleBase := raiden.RoleBase{}
	assert.Equal(t, false, roleBase.CanLogin())
}

func TestRoleBase_ValidUntil(t *testing.T) {
	roleBase := raiden.RoleBase{}
	assert.Nil(t, roleBase.ValidUntil())
}

func TestDefaultRoleValidUntilLayout(t *testing.T) {
	assert.Equal(t, "2006-01-02", raiden.DefaultRoleValidUntilLayout)
}

func TestDefaultRoleConnectionLimit(t *testing.T) {
	assert.Equal(t, 60, raiden.DefaultRoleConnectionLimit)
}
