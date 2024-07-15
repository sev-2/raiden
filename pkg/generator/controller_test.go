package generator_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sev-2/raiden/pkg/generator"
	"github.com/sev-2/raiden/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestGenerateController(t *testing.T) {
	dir, err := os.MkdirTemp("", "controller")
	assert.NoError(t, err)

	controllerPath := filepath.Join(dir, "internal")
	err1 := utils.CreateFolder(controllerPath)
	assert.NoError(t, err1)

	err2 := generator.GenerateHelloWorldController(dir, generator.GenerateFn(generator.Generate))
	assert.NoError(t, err2)
	assert.FileExists(t, dir+"/internal/controllers/hello.go")
}
