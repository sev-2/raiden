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

func TestCompareItem_FallbackMatchByColumn(t *testing.T) {
	// Source has a custom-named FK, target has a default-named FK for the same column.
	// They should match via the column-based fallback.
	action := objects.TablesRelationshipAction{UpdateAction: "c", DeletionAction: "c"}

	source := objects.Table{
		ID: 1, Name: "orders", Schema: "public",
		Relationships: []objects.TablesRelationship{
			{
				ConstraintName:    "fk_custom_name",
				SourceSchema:      "public",
				SourceTableName:   "orders",
				SourceColumnName:  "user_id",
				TargetTableSchema: "public",
				TargetTableName:   "users",
				TargetColumnName:  "id",
				Action:            &action,
			},
		},
	}

	target := objects.Table{
		ID: 1, Name: "orders", Schema: "public",
		Relationships: []objects.TablesRelationship{
			{
				ConstraintName:    "public_orders_user_id_fkey",
				SourceSchema:      "public",
				SourceTableName:   "orders",
				SourceColumnName:  "user_id",
				TargetTableSchema: "public",
				TargetTableName:   "users",
				TargetColumnName:  "id",
				Action:            &action,
			},
		},
	}

	diffResult := tables.CompareItem(tables.CompareModeApply, source, target)
	// Should NOT be a conflict — same FK, different constraint names
	hasRelationCreate := false
	hasRelationDelete := false
	for _, item := range diffResult.DiffItems.ChangeRelationItems {
		if item.Type == objects.UpdateRelationCreate {
			hasRelationCreate = true
		}
		if item.Type == objects.UpdateRelationDelete {
			hasRelationDelete = true
		}
	}
	assert.False(t, hasRelationCreate, "should not create relation — matched via column fallback")
	assert.False(t, hasRelationDelete, "should not delete relation — matched via column fallback")
}

func TestCompareItem_CrossSchemaFKSkip(t *testing.T) {
	// Target has a cross-schema FK (public → auth). It should NOT be flagged as delete.
	source := objects.Table{
		ID: 1, Name: "user_brands", Schema: "public",
		Relationships: []objects.TablesRelationship{},
	}

	target := objects.Table{
		ID: 1, Name: "user_brands", Schema: "public",
		Relationships: []objects.TablesRelationship{
			{
				ConstraintName:    "public_user_brands_user_id_fkey",
				SourceSchema:      "public",
				SourceTableName:   "user_brands",
				SourceColumnName:  "user_id",
				TargetTableSchema: "auth",
				TargetTableName:   "users",
				TargetColumnName:  "id",
			},
		},
	}

	diffResult := tables.CompareItem(tables.CompareModeApply, source, target)
	for _, item := range diffResult.DiffItems.ChangeRelationItems {
		assert.NotEqual(t, objects.UpdateRelationDelete, item.Type,
			"cross-schema FK should NOT be flagged as delete")
	}
}

func TestCompareItem_DuplicateFKSkip(t *testing.T) {
	// Source has one FK, target has two FKs for the same column (custom + default name).
	// The duplicate should NOT be flagged as delete.
	action := objects.TablesRelationshipAction{UpdateAction: "c", DeletionAction: "c"}

	source := objects.Table{
		ID: 1, Name: "creators", Schema: "public",
		Relationships: []objects.TablesRelationship{
			{
				ConstraintName:    "public_creators_div_id_fkey",
				SourceSchema:      "public",
				SourceTableName:   "creators",
				SourceColumnName:  "div_id",
				TargetTableSchema: "public",
				TargetTableName:   "divisions",
				TargetColumnName:  "id",
				Action:            &action,
			},
		},
	}

	target := objects.Table{
		ID: 1, Name: "creators", Schema: "public",
		Relationships: []objects.TablesRelationship{
			{
				ConstraintName:    "public_creators_div_id_fkey",
				SourceSchema:      "public",
				SourceTableName:   "creators",
				SourceColumnName:  "div_id",
				TargetTableSchema: "public",
				TargetTableName:   "divisions",
				TargetColumnName:  "id",
				Action:            &action,
			},
			{
				ConstraintName:    "fk_custom_div",
				SourceSchema:      "public",
				SourceTableName:   "creators",
				SourceColumnName:  "div_id",
				TargetTableSchema: "public",
				TargetTableName:   "divisions",
				TargetColumnName:  "id",
				Action:            &action,
			},
		},
	}

	diffResult := tables.CompareItem(tables.CompareModeApply, source, target)
	for _, item := range diffResult.DiffItems.ChangeRelationItems {
		assert.NotEqual(t, objects.UpdateRelationDelete, item.Type,
			"duplicate FK for matched column should NOT be flagged as delete")
	}
}

func TestCompareItem_NoIndexCreationWhenBothNil(t *testing.T) {
	// Both source and target have nil Index — should NOT propose index creation.
	action := objects.TablesRelationshipAction{UpdateAction: "c", DeletionAction: "c"}

	source := objects.Table{
		ID: 1, Name: "t1", Schema: "public",
		Relationships: []objects.TablesRelationship{
			{
				ConstraintName: "public_t1_col_fkey", SourceSchema: "public",
				SourceTableName: "t1", SourceColumnName: "col",
				TargetTableSchema: "public", TargetTableName: "t2", TargetColumnName: "id",
				Index: nil, Action: &action,
			},
		},
	}

	target := objects.Table{
		ID: 1, Name: "t1", Schema: "public",
		Relationships: []objects.TablesRelationship{
			{
				ConstraintName: "public_t1_col_fkey", SourceSchema: "public",
				SourceTableName: "t1", SourceColumnName: "col",
				TargetTableSchema: "public", TargetTableName: "t2", TargetColumnName: "id",
				Index: nil, Action: &action,
			},
		},
	}

	diffResult := tables.CompareItem(tables.CompareModeApply, source, target)
	for _, item := range diffResult.DiffItems.ChangeRelationItems {
		assert.NotEqual(t, objects.UpdateRelationCreateIndex, item.Type,
			"should NOT create index when both sides have nil Index")
	}
}

func TestCompareItem_IndexCreationWhenTargetHasIndex(t *testing.T) {
	// In apply mode: source = local, target = remote (supabase).
	// When source (local) has no index but target (remote) has one,
	// compareRelations checks t.Index (target) != nil && sc.Index (source) == nil.
	action := objects.TablesRelationshipAction{UpdateAction: "c", DeletionAction: "c"}
	idx := &objects.Index{Schema: "public", Table: "t1", Name: "idx1", Definition: "idx1"}

	source := objects.Table{
		ID: 1, Name: "t1", Schema: "public",
		Relationships: []objects.TablesRelationship{
			{
				ConstraintName: "public_t1_col_fkey", SourceSchema: "public",
				SourceTableName: "t1", SourceColumnName: "col",
				TargetTableSchema: "public", TargetTableName: "t2", TargetColumnName: "id",
				Index: nil, Action: &action,
			},
		},
	}

	target := objects.Table{
		ID: 1, Name: "t1", Schema: "public",
		Relationships: []objects.TablesRelationship{
			{
				ConstraintName: "public_t1_col_fkey", SourceSchema: "public",
				SourceTableName: "t1", SourceColumnName: "col",
				TargetTableSchema: "public", TargetTableName: "t2", TargetColumnName: "id",
				Index: idx, Action: &action,
			},
		},
	}

	diffResult := tables.CompareItem(tables.CompareModeApply, source, target)
	hasIndexCreate := false
	for _, item := range diffResult.DiffItems.ChangeRelationItems {
		if item.Type == objects.UpdateRelationCreateIndex {
			hasIndexCreate = true
		}
	}
	assert.True(t, hasIndexCreate, "should create index when target (remote) has it but source (local) doesn't")
}

func TestCompareItem_NilActionApplyMode(t *testing.T) {
	// In apply mode, when target (remote) has Action but source (local) doesn't,
	// compareRelations should create action update items.
	action := objects.TablesRelationshipAction{UpdateAction: "c", DeletionAction: "c"}

	source := objects.Table{
		ID: 1, Name: "t1", Schema: "public",
		Relationships: []objects.TablesRelationship{
			{
				ConstraintName: "public_t1_col_fkey", SourceSchema: "public",
				SourceTableName: "t1", SourceColumnName: "col",
				TargetTableSchema: "public", TargetTableName: "t2", TargetColumnName: "id",
				Action: nil,
			},
		},
	}
	target := objects.Table{
		ID: 1, Name: "t1", Schema: "public",
		Relationships: []objects.TablesRelationship{
			{
				ConstraintName: "public_t1_col_fkey", SourceSchema: "public",
				SourceTableName: "t1", SourceColumnName: "col",
				TargetTableSchema: "public", TargetTableName: "t2", TargetColumnName: "id",
				Action: &action,
			},
		},
	}

	diffResult := tables.CompareItem(tables.CompareModeApply, source, target)
	hasOnUpdate := false
	hasOnDelete := false
	for _, item := range diffResult.DiffItems.ChangeRelationItems {
		if item.Type == objects.UpdateRelationActionOnUpdate {
			hasOnUpdate = true
		}
		if item.Type == objects.UpdateRelationActionOnDelete {
			hasOnDelete = true
		}
	}
	assert.True(t, hasOnUpdate, "should flag action on update diff")
	assert.True(t, hasOnDelete, "should flag action on delete diff")
}

func TestCompareItem_NilActionImportMode(t *testing.T) {
	// In import mode, when target has Action but source doesn't,
	// compareRelations should NOT create action update items (no false positives).
	action := objects.TablesRelationshipAction{UpdateAction: "c", DeletionAction: "c"}

	source := objects.Table{
		ID: 1, Name: "t1", Schema: "public",
		Relationships: []objects.TablesRelationship{
			{
				ConstraintName: "public_t1_col_fkey", SourceSchema: "public",
				SourceTableName: "t1", SourceColumnName: "col",
				TargetTableSchema: "public", TargetTableName: "t2", TargetColumnName: "id",
				Action: nil,
			},
		},
	}
	target := objects.Table{
		ID: 1, Name: "t1", Schema: "public",
		Relationships: []objects.TablesRelationship{
			{
				ConstraintName: "public_t1_col_fkey", SourceSchema: "public",
				SourceTableName: "t1", SourceColumnName: "col",
				TargetTableSchema: "public", TargetTableName: "t2", TargetColumnName: "id",
				Action: &action,
			},
		},
	}

	diffResult := tables.CompareItem(tables.CompareModeImport, source, target)
	for _, item := range diffResult.DiffItems.ChangeRelationItems {
		assert.NotEqual(t, objects.UpdateRelationActionOnUpdate, item.Type, "should not flag action on update in import mode")
		assert.NotEqual(t, objects.UpdateRelationActionOnDelete, item.Type, "should not flag action on delete in import mode")
	}
}
