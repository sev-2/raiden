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

func TestGenerateTypes_WithComment(t *testing.T) {
	dir, err := os.MkdirTemp("", "type_comment")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	rolePath := filepath.Join(dir, "internal")
	err1 := utils.CreateFolder(rolePath)
	assert.NoError(t, err1)

	comment := "Creator Source"
	types := []objects.Type{
		{
			Name:       "creator_source",
			Schema:     "public",
			Format:     "",
			Enums:      []string{"tiktok", "instagram"},
			Attributes: []objects.TypeAttribute{},
			Comment:    &comment,
		},
	}

	err2 := generator.GenerateTypes(dir, types, generator.GenerateFn(generator.Generate))
	assert.NoError(t, err2)

	filePath := filepath.Join(dir, "internal", "types", "creator_source.go")
	assert.FileExists(t, filePath)

	content, err := os.ReadFile(filePath)
	assert.NoError(t, err)
	// Comment should be properly quoted in the generated code
	assert.Contains(t, string(content), `"Creator Source"`)
	assert.Contains(t, string(content), "comment :=")
}

func TestGenerateTypes_WithoutComment(t *testing.T) {
	dir, err := os.MkdirTemp("", "type_no_comment")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	rolePath := filepath.Join(dir, "internal")
	err1 := utils.CreateFolder(rolePath)
	assert.NoError(t, err1)

	types := []objects.Type{
		{
			Name:       "user_role",
			Schema:     "public",
			Format:     "",
			Enums:      []string{"admin", "user"},
			Attributes: []objects.TypeAttribute{},
			Comment:    nil,
		},
	}

	err2 := generator.GenerateTypes(dir, types, generator.GenerateFn(generator.Generate))
	assert.NoError(t, err2)

	filePath := filepath.Join(dir, "internal", "types", "user_role.go")
	assert.FileExists(t, filePath)

	content, err := os.ReadFile(filePath)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "return nil")
}
