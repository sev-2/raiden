package tables_test

import (
	"testing"

	"github.com/sev-2/raiden/pkg/resource/tables"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/stretchr/testify/assert"
)

func TestCompare(t *testing.T) {
	source := []objects.Table{
		{ID: 1, Name: "table1"},
		{ID: 2, Name: "table2"},
	}

	target := []objects.Table{
		{ID: 1, Name: "table1"},
		{ID: 2, Name: "table2"},
	}

	err := tables.Compare(source, target)
	assert.NoError(t, err)
}

func TestCompareList(t *testing.T) {
	source := []objects.Table{
		{ID: 1, Name: "table1"},
		{ID: 2, Name: "table2"},
	}

	target := []objects.Table{
		{ID: 1, Name: "table1_updated"},
		{ID: 2, Name: "table2"},
	}

	diffResult, err := tables.CompareList(source, target)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(diffResult))
	assert.Equal(t, "table1", diffResult[0].SourceResource.Name)
	assert.Equal(t, "table1_updated", diffResult[0].TargetResource.Name)
}

func TestCompareItem(t *testing.T) {
	source := objects.Table{
		ID:          1,
		Name:        "table1",
		Schema:      "public",
		PrimaryKeys: []objects.PrimaryKey{{Name: "id", Schema: "public", TableName: "table1"}},
		Columns: []objects.Column{
			{Name: "id", DataType: "int", IsNullable: false},
			{Name: "name", DataType: "varchar", IsNullable: true},
		},
	}

	target := objects.Table{
		ID:          1,
		Name:        "table1_updated",
		Schema:      "public",
		PrimaryKeys: []objects.PrimaryKey{{Name: "id", Schema: "public", TableName: "table1"}},
		Columns: []objects.Column{
			{Name: "id", DataType: "int", IsNullable: false},
			{Name: "name", DataType: "varchar", IsNullable: false},
			{Name: "description", DataType: "text", IsNullable: true},
		},
	}

	diffResult := tables.CompareItem(source, target)
	assert.True(t, diffResult.IsConflict)
	assert.Equal(t, "table1", diffResult.SourceResource.Name)
	assert.Equal(t, "table1_updated", diffResult.TargetResource.Name)
	assert.Equal(t, []objects.UpdateColumnType{objects.UpdateColumnNullable}, diffResult.DiffItems.ChangeColumnItems[0].UpdateItems)
	assert.Equal(t, []objects.UpdateColumnType{objects.UpdateColumnDelete}, diffResult.DiffItems.ChangeColumnItems[1].UpdateItems)
}
