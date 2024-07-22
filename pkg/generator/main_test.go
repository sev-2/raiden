package generator_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/sev-2/raiden/pkg/generator"
	"github.com/sev-2/raiden/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestGenerateMainFunction(t *testing.T) {
	dir, err := os.MkdirTemp("", "main")
	assert.NoError(t, err)

	config := loadConfig()

	err1 := generator.GenerateMainFunction(dir, config, generator.GenerateFn(generator.Generate))
	assert.NoError(t, err1)
	assert.FileExists(t, fmt.Sprintf("%s/cmd/%s/%s.go", dir, config.ProjectName, utils.ToSnakeCase(config.ProjectName)))
}
