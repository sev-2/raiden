package generator

import (
	"testing"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/stretchr/testify/assert"
)

type mockRole struct {
	raiden.RoleBase
	name string
}

func (m *mockRole) Name() string                      { return m.name }
func (m *mockRole) InheritRoles() []raiden.Role       { return nil }
func (m *mockRole) CanBypassRls() bool                { return false }
func (m *mockRole) CanCreateDB() bool                 { return false }
func (m *mockRole) CanCreateRole() bool               { return false }
func (m *mockRole) CanLogin() bool                    { return false }
func (m *mockRole) ValidUntil() *objects.SupabaseTime { return nil }

func TestResolvePolicyRoles_SkipsPublicPseudoRole(t *testing.T) {
	decls := make(map[string]*modelRoleRef)
	existing := make(map[string]bool)
	roleMap := map[string]string{"admin": "admin"}
	nativeRoleMap := map[string]raiden.Role{}

	args, useRoles, useNative, err := resolvePolicyRoles(
		[]string{"public", "admin"},
		decls, existing, roleMap, nativeRoleMap,
	)

	assert.NoError(t, err)
	assert.Len(t, args, 1)
	assert.True(t, useRoles)
	assert.False(t, useNative)
	// "public" should not appear in decls
	_, hasPublic := decls["public"]
	assert.False(t, hasPublic)
}

func TestResolvePolicyRoles_SkipsEmptyAndDuplicates(t *testing.T) {
	decls := make(map[string]*modelRoleRef)
	existing := make(map[string]bool)
	roleMap := map[string]string{"editor": "editor"}
	nativeRoleMap := map[string]raiden.Role{}

	args, _, _, err := resolvePolicyRoles(
		[]string{"editor", "", "editor", "  "},
		decls, existing, roleMap, nativeRoleMap,
	)

	assert.NoError(t, err)
	assert.Len(t, args, 1)
}

func TestResolvePolicyRoles_NativeRole(t *testing.T) {
	decls := make(map[string]*modelRoleRef)
	existing := make(map[string]bool)
	roleMap := map[string]string{}
	nativeRoleMap := map[string]raiden.Role{
		"anon": &mockRole{name: "anon"},
	}

	args, useRoles, useNative, err := resolvePolicyRoles(
		[]string{"anon"},
		decls, existing, roleMap, nativeRoleMap,
	)

	assert.NoError(t, err)
	assert.Len(t, args, 1)
	assert.False(t, useRoles)
	assert.True(t, useNative)
}

func TestResolvePolicyRoles_OnlyPublic(t *testing.T) {
	decls := make(map[string]*modelRoleRef)
	existing := make(map[string]bool)
	roleMap := map[string]string{}
	nativeRoleMap := map[string]raiden.Role{}

	args, useRoles, useNative, err := resolvePolicyRoles(
		[]string{"public"},
		decls, existing, roleMap, nativeRoleMap,
	)

	assert.NoError(t, err)
	assert.Len(t, args, 0)
	assert.False(t, useRoles)
	assert.False(t, useNative)
}
