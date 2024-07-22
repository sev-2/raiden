package generator_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sev-2/raiden/pkg/generator"
	"github.com/sev-2/raiden/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestGenerateJob(t *testing.T) {
	dir, err := os.MkdirTemp("", "job")
	assert.NoError(t, err)

	jobPath := filepath.Join(dir, "internal")
	err1 := utils.CreateFolder(jobPath)
	assert.NoError(t, err1)

	err2 := generator.GenerateHelloWorldJob(dir, generator.GenerateFn(generator.Generate))
	assert.NoError(t, err2)
	assert.FileExists(t, dir+"/internal/jobs/hello.go")
}
