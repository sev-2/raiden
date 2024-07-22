package generator_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sev-2/raiden/pkg/generator"
	"github.com/sev-2/raiden/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestGenerateImport(t *testing.T) {
	conf := loadConfig()

	dir, err := os.MkdirTemp("", "import")
	assert.NoError(t, err)

	importPath := filepath.Join(dir, "cmd")
	err1 := utils.CreateFolder(importPath)
	assert.NoError(t, err1)

	err2 := generator.GenerateImportMainFunction(dir, conf, generator.GenerateFn(generator.Generate))
	assert.NoError(t, err2)
	assert.FileExists(t, dir+"/cmd/import/main.go")
}
