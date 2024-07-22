package generator_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sev-2/raiden/pkg/generator"
	"github.com/sev-2/raiden/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestGenerateModelRegister(t *testing.T) {
	dir, err := os.MkdirTemp("", "model_register")
	assert.NoError(t, err)

	modelPath := filepath.Join(dir, "internal")
	err1 := utils.CreateFolder(modelPath)
	assert.NoError(t, err1)

	err2 := generator.GenerateModelRegister(dir, "test", generator.GenerateFn(generator.Generate))
	assert.NoError(t, err2)
	assert.Equal(t, true, utils.IsFolderExists(dir+"/internal/bootstrap"))
	assert.Equal(t, true, utils.IsFolderExists(dir+"/internal/models"))
}
