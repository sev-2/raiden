package generator_test

import (
	"os"
	"testing"

	"github.com/sev-2/raiden/pkg/generator"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/stretchr/testify/assert"
)

func TestGenerateRestControllers(t *testing.T) {
	dir, err := os.MkdirTemp("", "rest_controller")
	assert.NoError(t, err)

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
					{Name: "location", DataType: "point", IsNullable: true},
				},
				RLSEnabled: true,
				RLSForced:  false,
			},
		},
	}

	err2 := generator.GenerateRestControllers(dir, "test-project", tables, generator.GenerateFn(generator.Generate))
	assert.NoError(t, err2)
	assert.FileExists(t, dir+"/internal/controllers/rest/v1/test_table/rest.go")

}
