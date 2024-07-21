package tables_test

import (
	"bytes"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/resource/migrator"
	"github.com/sev-2/raiden/pkg/resource/tables"
	"github.com/sev-2/raiden/pkg/state"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/stretchr/testify/assert"
)

var (
	MigratedItems = objects.UpdateTableParam{
		OldData: objects.Table{Name: "old_table"},
		ChangeItems: []objects.UpdateTableType{
			objects.UpdateTableName,
			objects.UpdateTableSchema,
			objects.UpdateTableRlsEnable,
			objects.UpdateTableRlsForced,
			objects.UpdateTablePrimaryKey,
			objects.UpdateTableReplicaIdentity,
		},
		ChangeRelationItems: []objects.UpdateRelationItem{
			{
				Type: objects.UpdateRelationCreate,
				Data: objects.TablesRelationship{
					ConstraintName:    "constraint1",
					SourceSchema:      "public",
					SourceTableName:   "table1",
					SourceColumnName:  "id",
					TargetTableSchema: "public",
					TargetTableName:   "table2",
					TargetColumnName:  "id",
				},
			},
			{
				Type: objects.UpdateRelationUpdate,
				Data: objects.TablesRelationship{
					ConstraintName:    "constraint1",
					SourceSchema:      "public",
					SourceTableName:   "table1",
					SourceColumnName:  "id",
					TargetTableSchema: "public",
					TargetTableName:   "table2",
					TargetColumnName:  "id",
				},
			},
			{
				Type: objects.UpdateRelationDelete,
				Data: objects.TablesRelationship{
					ConstraintName:    "constraint1",
					SourceSchema:      "public",
					SourceTableName:   "table1",
					SourceColumnName:  "id",
					TargetTableSchema: "public",
					TargetTableName:   "table2",
					TargetColumnName:  "id",
				},
			},
		},
		ChangeColumnItems: []objects.UpdateColumnItem{
			{
				Name: "name",
				UpdateItems: []objects.UpdateColumnType{
					objects.UpdateColumnNullable,
				},
			},
			{
				Name: "old",
				UpdateItems: []objects.UpdateColumnType{
					objects.UpdateColumnDelete,
				},
			},
			{
				Name: "description",
				UpdateItems: []objects.UpdateColumnType{
					objects.UpdateColumnNew,
				},
			},
			{
				Name: "nullable",
				UpdateItems: []objects.UpdateColumnType{
					objects.UpdateColumnNullable,
				},
			},
			{
				Name: "uniqueness",
				UpdateItems: []objects.UpdateColumnType{
					objects.UpdateColumnUnique,
				},
			},
			{
				Name: "changeable",
				UpdateItems: []objects.UpdateColumnType{
					objects.UpdateColumnDataType,
					objects.UpdateColumnDefaultValue,
				},
			},
			{
				Name: "identity",
				UpdateItems: []objects.UpdateColumnType{
					objects.UpdateColumnIdentity,
				},
			},
		},
	}

	SourceTable = objects.Table{
		ID:          1,
		Name:        "table1",
		Schema:      "public",
		RLSEnabled:  true,
		RLSForced:   true,
		PrimaryKeys: []objects.PrimaryKey{{Name: "id", Schema: "public", TableName: "table1"}},
		Columns: []objects.Column{
			{Name: "id", DataType: "int", IsNullable: false},
			{Name: "name", DataType: "varchar", IsNullable: true},
			{Name: "old", DataType: "varchar", IsNullable: true},
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
			},
		},
	}

	TargetTable = objects.Table{
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
)

// TestPrintDiffResult tests the PrintDiffResult function
func TestPrintDiffResult(t *testing.T) {
	diffResult := []tables.CompareDiffResult{
		{
			Name:           "test_table",
			IsConflict:     true,
			SourceResource: SourceTable,
			TargetResource: TargetTable,
			DiffItems:      MigratedItems,
		},
	}

	sRelation := tables.MapRelations{
		"public.table1": []*state.Relation{
			{
				Table:        "table1",
				Type:         "some_type",
				RelationType: raiden.RelationTypeHasOne,
			},
		},
	}
	tRelation := tables.MapRelations{
		"public.table1_updated": []*state.Relation{
			{
				Table:        "table1_updated",
				Type:         "some_type",
				RelationType: raiden.RelationTypeHasOne,
			},
		},
	}

	err := tables.PrintDiffResult(diffResult, sRelation, tRelation)
	assert.EqualError(t, err, "canceled import process, you have conflict table. please fix it first")
}

// TestPrintDiff tests the PrintDiff function
func TestPrintDiff(t *testing.T) {
	if os.Getenv("TEST_RUN") == "1" {
		diffData := tables.CompareDiffResult{
			Name:           "test_table",
			IsConflict:     true,
			SourceResource: objects.Table{Name: "source_table"},
			TargetResource: objects.Table{Name: "target_table"},
		}

		sRelation := tables.MapRelations{}
		tRelation := tables.MapRelations{}

		tables.PrintDiff(diffData, sRelation, tRelation)
		return
	}

	var outb, errb bytes.Buffer
	cmd := exec.Command(os.Args[0], "-test.run=TestPrintDiff")
	cmd.Stdout = &outb
	cmd.Stderr = &errb

	cmd.Env = append(os.Environ(), "TEST_RUN=1")
	err := cmd.Start()
	assert.NoError(t, err)

	time.Sleep(1 * time.Second)
	err1 := cmd.Process.Signal(syscall.SIGTERM)
	assert.NoError(t, err1)

	assert.Contains(t, outb.String(), "Found diff")
	assert.Contains(t, outb.String(), "End found diff")

	successDiffData := tables.CompareDiffResult{
		Name:           "test_table",
		IsConflict:     false,
		SourceResource: SourceTable,
		TargetResource: TargetTable,
		DiffItems:      MigratedItems,
	}

	sRelation := tables.MapRelations{
		"public.table1": []*state.Relation{
			{
				Table:        "table1",
				Type:         "some_type",
				RelationType: raiden.RelationTypeHasOne,
			},
		},
	}
	tRelation := tables.MapRelations{
		"public.table1_updated": []*state.Relation{
			{
				Table:        "table1_updated",
				Type:         "some_type",
				RelationType: raiden.RelationTypeHasOne,
			},
		},
	}

	tables.PrintDiff(successDiffData, sRelation, tRelation)
}

// TestGetDiffChangeMessage tests the GetDiffChangeMessage function
func TestGetDiffChangeMessage(t *testing.T) {
	items := []tables.MigrateItem{
		{
			Type:    migrator.MigrateTypeCreate,
			NewData: objects.Table{Name: "new_table"},
		},
		{
			Type:    migrator.MigrateTypeUpdate,
			NewData: objects.Table{Name: "update_table"},
		},
		{
			Type:    migrator.MigrateTypeDelete,
			OldData: objects.Table{Name: "delete_table"},
		},
	}

	diffMessage := tables.GetDiffChangeMessage(items)
	assert.Contains(t, diffMessage, "New Table")
	assert.Contains(t, diffMessage, "Update Table")
	assert.Contains(t, diffMessage, "Delete Table")
}

// TestGenerateDiffMessage tests the GenerateDiffMessage function
func TestGenerateDiffMessage(t *testing.T) {
	diffData := tables.CompareDiffResult{
		Name:           "test_table",
		IsConflict:     false,
		SourceResource: SourceTable,
		TargetResource: TargetTable,
		DiffItems:      MigratedItems,
	}

	sRelation := tables.MapRelations{}
	tRelation := tables.MapRelations{}

	diffMessage, err := tables.GenerateDiffMessage(diffData, sRelation, tRelation)
	assert.NoError(t, err)
	assert.NotNil(t, diffMessage)
}

// TestGenerateDiffChangeMessage tests the GenerateDiffChangeMessage function
func TestGenerateDiffChangeMessage(t *testing.T) {
	newTable := []string{"new_table1", "new_table2"}
	updateTable := []string{"update_table1", "update_table2"}
	deleteTable := []string{"delete_table1", "delete_table2"}

	diffMessage, err := tables.GenerateDiffChangeMessage(newTable, updateTable, deleteTable)
	assert.NoError(t, err)
	assert.Contains(t, diffMessage, "New Table")
	assert.Contains(t, diffMessage, "Update Table")
	assert.Contains(t, diffMessage, "Delete Table")
}

// TestGenerateDiffChangeUpdateMessage tests the GenerateDiffChangeUpdateMessage function
func TestGenerateDiffChangeUpdateMessage(t *testing.T) {
	item := tables.MigrateItem{
		NewData: objects.Table{Name: "new_table"},
		OldData: objects.Table{Name: "old_table"},
		MigrationItems: objects.UpdateTableParam{
			ChangeItems: []objects.UpdateTableType{objects.UpdateTableName},
		},
	}

	diffMessage, err := tables.GenerateDiffChangeUpdateMessage("test_table", item)
	assert.NoError(t, err)
	assert.Contains(t, diffMessage, "Update Table test_table")
}
