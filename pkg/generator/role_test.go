package generator_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sev-2/raiden/pkg/generator"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestGenerateRoles(t *testing.T) {
	dir, err := os.MkdirTemp("", "role")
	assert.NoError(t, err)

	rolePath := filepath.Join(dir, "internal")
	err1 := utils.CreateFolder(rolePath)
	assert.NoError(t, err1)

	roles := []objects.Role{
		{
			Name:              "test_role",
			ConnectionLimit:   10,
			InheritRole:       true,
			IsReplicationRole: false,
			IsSuperuser:       false,
			CanBypassRLS:      false,
			CanCreateDB:       true,
			CanCreateRole:     false,
			CanLogin:          true,
		},
	}

	err2 := generator.GenerateRoles(dir, roles, generator.GenerateFn(generator.Generate))
	assert.NoError(t, err2)
	assert.FileExists(t, dir+"/internal/roles/test_role.go")
}

func TestGenerateRoles_WithInheritance(t *testing.T) {
	dir, err := os.MkdirTemp("", "role-inherit")
	assert.NoError(t, err)

	rolePath := filepath.Join(dir, "internal")
	assert.NoError(t, utils.CreateFolder(rolePath))

	roles := []objects.Role{
		{
			Name:          "child_role",
			CanCreateRole: true,
		},
		{
			Name:        "parent_role",
			InheritRole: true,
			InheritRoles: []*objects.Role{
				{Name: "child_role"},
			},
			CanLogin: true,
		},
	}

	assert.NoError(t, generator.GenerateRoles(dir, roles, generator.GenerateFn(generator.Generate)))

	generatedParent := filepath.Join(dir, "internal", "roles", "parent_role.go")
	assert.FileExists(t, generatedParent)

	content, readErr := os.ReadFile(generatedParent)
	assert.NoError(t, readErr)
	generated := string(content)

	assert.Contains(t, generated, "func (r *ParentRole) InheritRoles() []raiden.Role")
	assert.Contains(t, generated, "return []raiden.Role{ &ChildRole{} }")
	assert.Contains(t, generated, "github.com/sev-2/raiden")
}
