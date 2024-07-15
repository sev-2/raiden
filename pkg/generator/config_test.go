package generator_test

import (
	"os"
	"testing"

	"github.com/sev-2/raiden/pkg/generator"
	"github.com/stretchr/testify/assert"
)

func TestGenerateConfig(t *testing.T) {
	dir, err := os.MkdirTemp("", "config")
	assert.NoError(t, err)

	conf := loadConfig()

	err1 := generator.GenerateConfig(dir, conf, generator.GenerateFn(generator.Generate))
	assert.NoError(t, err1)
	assert.FileExists(t, dir+"/configs/app.yaml")
}
