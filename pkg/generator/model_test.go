package generator_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sev-2/raiden/pkg/generator"
	"github.com/sev-2/raiden/pkg/state"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestGenerateModels(t *testing.T) {
	dir, err := os.MkdirTemp("", "model")
	assert.NoError(t, err)

	modelPath := filepath.Join(dir, "internal")
	err1 := utils.CreateFolder(modelPath)
	assert.NoError(t, err1)

	tables := []*generator.GenerateModelInput{
		{
			Table: objects.Table{
				Name:   "test_table",
				Schema: "public",
				PrimaryKeys: []objects.PrimaryKey{
					{Name: "id"},
				},
				Columns: []objects.Column{
					{Name: "id", DataType: "integer", IsNullable: false},
					{Name: "name", DataType: "text", IsNullable: true},
				},
				RLSEnabled: true,
				RLSForced:  false,
			},
			Relations: []state.Relation{
				{
					RelationType: "one_to_many",
					Table:        "related_table",
					ForeignKey:   "test_table_id",
					PrimaryKey:   "id",
				},
			},
			Policies: objects.Policies{},
			ValidationTags: state.ModelValidationTag{
				"id":   "required",
				"name": "required",
			},
		},
	}

	err2 := generator.GenerateModels(dir, tables, generator.GenerateFn(generator.Generate))
	assert.NoError(t, err2)
	assert.FileExists(t, dir+"/internal/models/test_table.go")
}
