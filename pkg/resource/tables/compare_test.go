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

	err := tables.Compare(tables.CompareModeImport, source, target)
	assert.NoError(t, err)
}

func TestCompareList(t *testing.T) {
	source := []objects.Table{
		{
			ID:         1,
			Name:       "table1",
			Schema:     "public",
			RLSEnabled: true,
			RLSForced:  true,
		},
		{
			ID:   2,
			Name: "table2",
		},
	}

	target := []objects.Table{
		{
			ID:         1,
			Name:       "table1_updated",
			Schema:     "private",
			RLSEnabled: false,
			RLSForced:  false,
		},
		{
			ID:   2,
			Name: "table2",
		},
	}

	diffResult, err := tables.CompareList(tables.CompareModeImport, source, target)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(diffResult))
	assert.Equal(t, "table1", diffResult[0].SourceResource.Name)
	assert.Equal(t, "table1_updated", diffResult[0].TargetResource.Name)
}

func TestCompareItem(t *testing.T) {

	sourceRelationshipAction := objects.TablesRelationshipAction{
		UpdateAction:   "c",
		DeletionAction: "c",
	}

	targetRelationshipAction := objects.TablesRelationshipAction{
		UpdateAction:   "a",
		DeletionAction: "a",
	}

	source := objects.Table{
		ID:          1,
		Name:        "table1",
		Schema:      "public",
		RLSEnabled:  true,
		RLSForced:   true,
		PrimaryKeys: []objects.PrimaryKey{{Name: "id", Schema: "public", TableName: "table1"}},
		Columns: []objects.Column{
			{Name: "id", DataType: "int", IsNullable: false},
			{Name: "name", DataType: "varchar", IsNullable: true},
			{Name: "nullable", DataType: "varchar", IsNullable: true},
			{Name: "changeable", DataType: "varchar", IsNullable: true},
			{Name: "uniqueness", DataType: "varchar", IsNullable: false, IsUnique: true},
			{Name: "identity", DataType: "varchar", IsNullable: false, IsIdentity: true},
		},
		Relationships: []objects.TablesRelationship{
			{
				ConstraintName:    "constraint1",
				SourceSchema:      "public",
				SourceTableName:   "table1",
				SourceColumnName:  "id",
				TargetTableSchema: "public",
				TargetTableName:   "table2",
				TargetColumnName:  "id",
				Index:             &objects.Index{Schema: "public", Table: "table1", Name: "index1", Definition: "index1"},
				Action:            &sourceRelationshipAction,
			},
		},
	}

	target := objects.Table{
		ID:          1,
		Name:        "table1_updated",
		Schema:      "private",
		RLSEnabled:  false,
		RLSForced:   false,
		PrimaryKeys: []objects.PrimaryKey{{Name: "id", Schema: "public", TableName: "table1"}},
		Columns: []objects.Column{
			{Name: "id", DataType: "int", IsNullable: false},
			{Name: "name", DataType: "varchar", IsNullable: false},
			{Name: "description", DataType: "text", IsNullable: true},
			{Name: "nullable", DataType: "varchar", IsNullable: false},
			{Name: "changeable", DataType: "json", IsNullable: true},
			{Name: "uniqueness", DataType: "varchar", IsNullable: false, IsUnique: false},
			{Name: "identity", DataType: "varchar", IsNullable: false, IsIdentity: false},
		},
		Relationships: []objects.TablesRelationship{
			{
				ConstraintName:    "constraint1",
				SourceSchema:      "public",
				SourceTableName:   "table1",
				SourceColumnName:  "id",
				TargetTableSchema: "public",
				TargetTableName:   "table2",
				TargetColumnName:  "id",
				Index:             &objects.Index{Schema: "public", Table: "table1", Name: "index1", Definition: "index1"},
				Action:            &targetRelationshipAction,
			},
			{
				ConstraintName:    "constraint2",
				SourceSchema:      "public",
				SourceTableName:   "table1",
				SourceColumnName:  "name",
				TargetTableSchema: "public",
				TargetTableName:   "table2",
				TargetColumnName:  "name",
				Action:            &targetRelationshipAction,
			},
		},
	}

	diffResult := tables.CompareItem(tables.CompareModeImport, source, target)
	assert.True(t, diffResult.IsConflict)
	assert.Equal(t, "table1", diffResult.SourceResource.Name)
	assert.Equal(t, "table1_updated", diffResult.TargetResource.Name)
	assert.Equal(t, []objects.UpdateColumnType{objects.UpdateColumnNullable}, diffResult.DiffItems.ChangeColumnItems[0].UpdateItems)
	assert.Equal(t, []objects.UpdateColumnType{objects.UpdateColumnNullable}, diffResult.DiffItems.ChangeColumnItems[1].UpdateItems)
}

func TestCompareItemWithoutIndex(t *testing.T) {
	targetRelationshipAction := objects.TablesRelationshipAction{
		UpdateAction:   "c",
		DeletionAction: "c",
	}

	source := objects.Table{
		ID:          1,
		Name:        "table1",
		Schema:      "public",
		RLSEnabled:  true,
		RLSForced:   true,
		PrimaryKeys: []objects.PrimaryKey{{Name: "id", Schema: "public", TableName: "table1"}},
		Columns: []objects.Column{
			{Name: "id", DataType: "int", IsNullable: false},
			{Name: "name", DataType: "varchar", IsNullable: true},
			{Name: "nullable", DataType: "varchar", IsNullable: true},
			{Name: "changeable", DataType: "varchar", IsNullable: true},
			{Name: "uniqueness", DataType: "varchar", IsNullable: false, IsUnique: true},
			{Name: "identity", DataType: "varchar", IsNullable: false, IsIdentity: true},
		},
		Relationships: []objects.TablesRelationship{
			{
				ConstraintName:    "constraint1",
				SourceSchema:      "public",
				SourceTableName:   "table1",
				SourceColumnName:  "id",
				TargetTableSchema: "public",
				TargetTableName:   "table2",
				TargetColumnName:  "id",
				Action:            nil,
			},
		},
	}

	target := objects.Table{
		ID:          1,
		Name:        "table1_updated",
		Schema:      "private",
		RLSEnabled:  false,
		RLSForced:   false,
		PrimaryKeys: []objects.PrimaryKey{{Name: "id", Schema: "public", TableName: "table1"}},
		Columns: []objects.Column{
			{Name: "id", DataType: "int", IsNullable: false},
			{Name: "name", DataType: "varchar", IsNullable: false},
			{Name: "description", DataType: "text", IsNullable: true},
			{Name: "nullable", DataType: "varchar", IsNullable: false},
			{Name: "changeable", DataType: "json", IsNullable: true},
			{Name: "uniqueness", DataType: "varchar", IsNullable: false, IsUnique: false},
			{Name: "identity", DataType: "varchar", IsNullable: false, IsIdentity: false},
		},
		Relationships: []objects.TablesRelationship{
			{
				ConstraintName:    "constraint1",
				SourceSchema:      "public",
				SourceTableName:   "table1",
				SourceColumnName:  "id",
				TargetTableSchema: "public",
				TargetTableName:   "table2",
				TargetColumnName:  "id",
				Action:            &targetRelationshipAction,
			},
			{
				ConstraintName:    "constraint2",
				SourceSchema:      "public",
				SourceTableName:   "table1",
				SourceColumnName:  "name",
				TargetTableSchema: "public",
				TargetTableName:   "table2",
				TargetColumnName:  "name",
			},
		},
	}

	diffResult := tables.CompareItem(tables.CompareModeImport, source, target)
	assert.True(t, diffResult.IsConflict)
	assert.Equal(t, "table1", diffResult.SourceResource.Name)
	assert.Equal(t, "table1_updated", diffResult.TargetResource.Name)
	assert.Equal(t, []objects.UpdateColumnType{objects.UpdateColumnNullable}, diffResult.DiffItems.ChangeColumnItems[0].UpdateItems)
	assert.Equal(t, []objects.UpdateColumnType{objects.UpdateColumnNullable}, diffResult.DiffItems.ChangeColumnItems[1].UpdateItems)
}
