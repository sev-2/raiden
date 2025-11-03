package query

import (
	"testing"

	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/stretchr/testify/assert"
)

func TestBuildUpdateTableQueryHandlesForceRLS(t *testing.T) {
	newTable := objects.Table{Schema: "public", Name: "courses", RLSForced: true}
	oldTable := objects.Table{Schema: "public", Name: "courses"}
	stmt := BuildUpdateTableQuery(newTable, objects.UpdateTableParam{
		OldData:     oldTable,
		ChangeItems: []objects.UpdateTableType{objects.UpdateTableRlsForced},
	})

	assert.Contains(t, stmt, "FORCE ROW LEVEL SECURITY")
}

func TestBuildUpdateColumnQueryRename(t *testing.T) {
	oldCol := objects.Column{Schema: "public", Table: "courses", Name: "old_name", DataType: "text"}
	newCol := objects.Column{Schema: "public", Table: "courses", Name: "new_name", DataType: "text"}
	stmt := BuildUpdateColumnQuery(oldCol, newCol, objects.UpdateColumnItem{
		UpdateItems: []objects.UpdateColumnType{objects.UpdateColumnName},
	})

	assert.Contains(t, stmt, "RENAME COLUMN \"old_name\" TO \"new_name\"")
}

func TestBuildFkQuery(t *testing.T) {
	relation := &objects.TablesRelationship{
		SourceSchema:      "public",
		SourceTableName:   "courses",
		SourceColumnName:  "category_id",
		TargetTableSchema: "public",
		TargetTableName:   "categories",
		TargetColumnName:  "id",
		ConstraintName:    "courses_category_id_fkey",
		Action: &objects.TablesRelationshipAction{
			UpdateAction:   "c",
			DeletionAction: "r",
		},
	}

	stmt, err := BuildFkQuery(objects.UpdateRelationCreate, relation)
	assert.NoError(t, err)
	assert.Contains(t, stmt, "ALTER TABLE IF EXISTS \"public\".\"courses\" ADD CONSTRAINT \"courses_category_id_fkey\"")
	assert.Contains(t, stmt, "FOREIGN KEY (\"category_id\") REFERENCES \"public\".\"categories\" (\"id\")")
	assert.Contains(t, stmt, "ON UPDATE CASCADE")
	assert.Contains(t, stmt, "ON DELETE RESTRICT")
}

func TestBuildFkIndexQuery(t *testing.T) {
	relation := &objects.TablesRelationship{
		SourceSchema:     "public",
		SourceTableName:  "courses",
		SourceColumnName: "category_id",
	}

	createStmt, err := BuildFKIndexQuery(objects.UpdateRelationCreate, relation)
	assert.NoError(t, err)
	assert.Equal(t, "CREATE INDEX IF NOT EXISTS \"ix_courses_category_id\" ON \"public\".\"courses\" (\"category_id\");", createStmt)

	dropStmt, err := BuildFKIndexQuery(objects.UpdateRelationDelete, relation)
	assert.NoError(t, err)
	assert.Equal(t, "DROP INDEX IF EXISTS \"public\".\"ix_courses_category_id\";", dropStmt)
}

func TestBuildCreateColumnQueryPrimary(t *testing.T) {
	stmt, err := BuildCreateColumnQuery(objects.Column{Schema: "public", Table: "courses", Name: "id", DataType: "uuid"}, true)
	assert.NoError(t, err)
	assert.Contains(t, stmt, "ADD COLUMN \"id\" uuid")
	assert.Contains(t, stmt, "PRIMARY KEY")
}

func TestBuildCreateTableQuery(t *testing.T) {
	table := objects.Table{
		Schema:      "public",
		Name:        "courses",
		RLSEnabled:  true,
		RLSForced:   true,
		Columns:     []objects.Column{{Schema: "public", Table: "courses", Name: "id", DataType: "uuid", IsNullable: false}},
		PrimaryKeys: []objects.PrimaryKey{{Name: "id"}},
	}

	stmt, err := BuildCreateTableQuery(table)
	assert.NoError(t, err)
	assert.Contains(t, stmt, "CREATE TABLE IF NOT EXISTS \"public\".\"courses\"")
	assert.Contains(t, stmt, "PRIMARY KEY (\"id\")")
	assert.Contains(t, stmt, "ENABLE ROW LEVEL SECURITY")
	assert.Contains(t, stmt, "FORCE ROW LEVEL SECURITY")
}

func TestBuildDeleteHelpers(t *testing.T) {
	dropTable := BuildDeleteTableQuery(objects.Table{Schema: "public", Name: "courses"}, true)
	assert.Equal(t, "DROP TABLE \"public\".\"courses\" CASCADE;", dropTable)

	dropColumn := BuildDeleteColumnQuery(objects.Column{Schema: "public", Table: "courses", Name: "archived"})
	assert.Equal(t, "ALTER TABLE \"public\".\"courses\" DROP COLUMN \"archived\";", dropColumn)
}

func TestBuildUpdateTableQueryPrimaryKey(t *testing.T) {
	newTable := objects.Table{Schema: "public", Name: "courses", PrimaryKeys: []objects.PrimaryKey{{Name: "id"}}}
	oldTable := objects.Table{Schema: "public", Name: "courses", ID: 10, PrimaryKeys: []objects.PrimaryKey{{Name: "old_id"}}}
	stmt := BuildUpdateTableQuery(newTable, objects.UpdateTableParam{
		OldData:     oldTable,
		ChangeItems: []objects.UpdateTableType{objects.UpdateTablePrimaryKey},
	})

	assert.Contains(t, stmt, "DROP CONSTRAINT")
	assert.Contains(t, stmt, "ADD PRIMARY KEY (\"id\")")
}

func TestBuildUpdateTableQueryRename(t *testing.T) {
	newTable := objects.Table{Schema: "app", Name: "lessons", RLSEnabled: false}
	oldTable := objects.Table{Schema: "public", Name: "courses", RLSEnabled: true}
	stmt := BuildUpdateTableQuery(newTable, objects.UpdateTableParam{
		OldData: oldTable,
		ChangeItems: []objects.UpdateTableType{
			objects.UpdateTableSchema,
			objects.UpdateTableName,
			objects.UpdateTableRlsEnable,
		},
	})

	assert.Contains(t, stmt, "SET SCHEMA \"app\"")
	assert.Contains(t, stmt, "RENAME TO \"lessons\"")
	assert.Contains(t, stmt, "DISABLE ROW LEVEL SECURITY")
}

func TestBuildUpdateColumnQueryCoversBranches(t *testing.T) {
	oldCol := objects.Column{Schema: "public", Table: "courses", Name: "counter", DataType: "integer"}
	newCol := objects.Column{
		Schema:             "public",
		Table:              "courses",
		Name:               "counter",
		DataType:           "text",
		IsUnique:           true,
		IsNullable:         true,
		DefaultValue:       "5",
		IdentityGeneration: "ALWAYS",
		IsIdentity:         true,
	}

	stmt := BuildUpdateColumnQuery(oldCol, newCol, objects.UpdateColumnItem{
		UpdateItems: []objects.UpdateColumnType{
			objects.UpdateColumnDataType,
			objects.UpdateColumnUnique,
			objects.UpdateColumnNullable,
			objects.UpdateColumnDefaultValue,
			objects.UpdateColumnIdentity,
		},
	})

	assert.Contains(t, stmt, "SET DATA TYPE text")
	assert.Contains(t, stmt, "ADD CONSTRAINT")
	assert.Contains(t, stmt, "DROP NOT NULL")
	assert.Contains(t, stmt, "SET DEFAULT 5")
	assert.Contains(t, stmt, "ADD GENERATED ALWAYS AS IDENTITY")
}

func TestBuildUpdateColumnQueryDropsFlags(t *testing.T) {
	oldCol := objects.Column{Schema: "public", Table: "courses", Name: "counter", DataType: "text", IsIdentity: true}
	newCol := objects.Column{Schema: "public", Table: "courses", Name: "counter", DataType: "text", IsUnique: false, IsNullable: false, DefaultValue: nil, IsIdentity: false}

	stmt := BuildUpdateColumnQuery(oldCol, newCol, objects.UpdateColumnItem{
		UpdateItems: []objects.UpdateColumnType{
			objects.UpdateColumnUnique,
			objects.UpdateColumnNullable,
			objects.UpdateColumnDefaultValue,
			objects.UpdateColumnIdentity,
		},
	})

	assert.Contains(t, stmt, "DROP CONSTRAINT")
	assert.Contains(t, stmt, "SET NOT NULL")
	assert.Contains(t, stmt, "DROP DEFAULT")
	assert.Contains(t, stmt, "DROP IDENTITY IF EXISTS")
}

func TestBuildCreateColumnQueryIdentityConflict(t *testing.T) {
	_, err := BuildCreateColumnQuery(objects.Column{Schema: "public", Table: "courses", Name: "id", DataType: "int8", IsIdentity: true, DefaultValue: "1"}, false)
	assert.Error(t, err)
}

func TestBuildCreateTableQueryDefaultHandling(t *testing.T) {
	table := objects.Table{
		Schema: "public",
		Name:   "defaults",
		Columns: []objects.Column{
			{Schema: "public", Table: "defaults", Name: "flag", DataType: "boolean", DefaultValue: "true"},
			{Schema: "public", Table: "defaults", Name: "uid", DataType: "uuid", DefaultValue: "uuid_generate_v4()"},
			func() objects.Column {
				str := "hello"
				return objects.Column{Schema: "public", Table: "defaults", Name: "greeting", DataType: "text", DefaultValue: &str}
			}(),
		},
	}

	stmt, err := BuildCreateTableQuery(table)
	assert.NoError(t, err)
	assert.Contains(t, stmt, "DEFAULT true")
	assert.Contains(t, stmt, "DEFAULT uuid_generate_v4()")
	assert.Contains(t, stmt, "DEFAULT 'hello'")
}
