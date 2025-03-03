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

func TestGenerateTypes(t *testing.T) {
	dir, err := os.MkdirTemp("", "role")
	assert.NoError(t, err)

	rolePath := filepath.Join(dir, "internal")
	err1 := utils.CreateFolder(rolePath)
	assert.NoError(t, err1)

	types := []objects.Type{
		{
			Name:       "test_type",
			Schema:     "public",
			Format:     "",
			Enums:      []string{"test_1", "test_2"},
			Attributes: []objects.TypeAttribute{},
			Comment:    nil,
		},
	}

	err2 := generator.GenerateTypes(dir, types, generator.GenerateFn(generator.Generate))
	assert.NoError(t, err2)
	assert.FileExists(t, dir+"/internal/types/test_type.go")
}
