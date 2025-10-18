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

func TestGenerateStorages(t *testing.T) {
	dir, err := os.MkdirTemp("", "storage")
	assert.NoError(t, err)

	storagePath := filepath.Join(dir, "internal")
	err1 := utils.CreateFolder(storagePath)
	assert.NoError(t, err1)

	var fileSize = new(int)
	*fileSize = 1048576 // 1MB

	storages := []*generator.GenerateStorageInput{
		{
			Bucket: objects.Bucket{
				Name:             "test_bucket",
				Public:           true,
				FileSizeLimit:    fileSize,
				AllowedMimeTypes: []string{"image/jpeg", "image/png"},
			},
			Policies: objects.Policies{
				{
					Table:  "objects",
					Action: "select",
				},
			},
		},
	}

	err2 := generator.GenerateStorages(dir, "test-project", storages, nil, nil, generator.GenerateFn(generator.Generate))
	assert.NoError(t, err2)
	assert.FileExists(t, dir+"/internal/storages/test_bucket.go")
}
