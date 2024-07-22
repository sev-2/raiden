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
