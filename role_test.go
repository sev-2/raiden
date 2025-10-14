package raiden_test

import (
	"testing"

	"github.com/sev-2/raiden"
	"github.com/stretchr/testify/assert"
)

type sampleRole struct {
	raiden.RoleBase
	name string
}

func (r *sampleRole) Name() string {
	return r.name
}

func TestRoleBase_ConnectionLimit(t *testing.T) {
	roleBase := raiden.RoleBase{}
	assert.Equal(t, 60, roleBase.ConnectionLimit())
}

func TestRoleBase_InheritRole(t *testing.T) {
	roleBase := raiden.RoleBase{}
	assert.Equal(t, true, roleBase.IsInheritRole())
}

func TestRoleBase_InheritRoles(t *testing.T) {
	roleBase := raiden.RoleBase{}
	roles := roleBase.InheritRoles()
	assert.NotNil(t, roles)
	assert.Len(t, roles, 0)
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

func TestRoleInterfaceDefaults(t *testing.T) {
	var r raiden.Role = &sampleRole{name: "reader"}
	assert.Equal(t, "reader", r.Name())
	assert.Equal(t, raiden.DefaultRoleConnectionLimit, r.ConnectionLimit())
	assert.True(t, r.IsInheritRole())
	assert.Len(t, r.InheritRoles(), 0)
	assert.Nil(t, r.ValidUntil())
}
