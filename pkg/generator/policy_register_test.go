package generator_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sev-2/raiden/pkg/generator"
	"github.com/sev-2/raiden/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestGeneratePolicyRegister(t *testing.T) {
	dir, err := os.MkdirTemp("", "policy_register")
	assert.NoError(t, err)

	policyPath := filepath.Join(dir, "internal")
	err1 := utils.CreateFolder(policyPath)
	assert.NoError(t, err1)

	err2 := generator.GeneratePolicyRegister(dir, "test", generator.GenerateFn(generator.Generate))
	assert.NoError(t, err2)
	assert.Equal(t, true, utils.IsFolderExists(dir+"/internal/bootstrap"))
	assert.FileExists(t, dir+"/internal/bootstrap/policies.go")
	assert.Equal(t, true, utils.IsFolderExists(dir+"/internal/policies"))
}
