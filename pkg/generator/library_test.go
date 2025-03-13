package generator_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sev-2/raiden/pkg/generator"
	"github.com/sev-2/raiden/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestGenerateLibRegister(t *testing.T) {
	dir, err := os.MkdirTemp("", "lib_register")
	assert.NoError(t, err)

	libPath := filepath.Join(dir, "internal")
	err1 := utils.CreateFolder(libPath)
	assert.NoError(t, err1)

	err2 := generator.GenerateLibRegister(dir, "test", generator.GenerateFn(generator.Generate))
	assert.NoError(t, err2)
	assert.Equal(t, true, utils.IsFolderExists(dir+"/internal/bootstrap"))
	assert.FileExists(t, dir+"/internal/bootstrap/libs.go")

	foundFiles, err3 := generator.WalkScanLib(dir + "/internal/libs")
	assert.NoError(t, err3)
	assert.Empty(t, foundFiles)
}

func TestGenerateLibRegister_empty(t *testing.T) {
	testPath, err := utils.GetAbsolutePath("/testdata")
	assert.NoError(t, err)

	libs, err := generator.WalkScanLib(testPath)
	assert.NoError(t, err)

	assert.Equal(t, 0, len(libs))
}

func TestGenerateLibRegister_nil(t *testing.T) {
	err1 := generator.GenerateLibRegister("", "test", generator.GenerateFn(generator.Generate))
	assert.Error(t, err1)

	libs, err := generator.WalkScanLib("")
	assert.Error(t, err)
	assert.Equal(t, 0, len(libs))
}
