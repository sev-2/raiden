package tables

import (
	"fmt"
	"strings"

	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/utils"
)

type CompareDiffResult struct {
	Name           string
	SourceResource objects.Table
	TargetResource objects.Table
	DiffItems      objects.UpdateTableParam
	IsConflict     bool
}

func Compare(source []objects.Table, target []objects.Table) error {
	diffResult, err := CompareList(source, target)
	if err != nil {
		return err
	}

	sMapRelation := buildGenerateMapRelations(tableToMap(source))
	tMapRelation := buildGenerateMapRelations(tableToMap(target))

	return PrintDiffResult(diffResult, sMapRelation, tMapRelation)
}

func CompareList(sourceTable, targetTable []objects.Table) (diffResult []CompareDiffResult, err error) {
	mapTargetTable := make(map[int]objects.Table)
	for i := range targetTable {
		t := targetTable[i]
		mapTargetTable[t.ID] = t
	}

	for i := range sourceTable {
		t := sourceTable[i]

		tt, isExist := mapTargetTable[t.ID]
		if !isExist {
			continue
		}

		df := CompareItem(t, tt)
		if !df.IsConflict {
			continue
		}
		diffResult = append(diffResult, df)
	}

	return
}

func CompareItem(source, target objects.Table) (diffResult CompareDiffResult) {
	var updateItem objects.UpdateTableParam

	// create map pk for compare target pk with source pk
	mapTargetPrimaryKey := make(map[string]bool)
	for i := range target.PrimaryKeys {
		pk := target.PrimaryKeys[i]
		key := fmt.Sprintf("%s.%s.%s", pk.Schema, pk.TableName, pk.Name)
		mapTargetPrimaryKey[key] = true
	}

	// assign diff result object
	diffResult.SourceResource = source
	diffResult.TargetResource = target

	// table config compare
	if source.Name != target.Name {
		updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdateTableName)
	}

	if source.Schema != target.Schema {
		updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdateTableSchema)
	}

	if source.RLSEnabled != target.RLSEnabled {
		updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdateTableRlsEnable)
	}

	if source.RLSForced != target.RLSForced {
		updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdateTableRlsForced)
	}

	for i := range source.PrimaryKeys {
		pk := source.PrimaryKeys[i]
		key := fmt.Sprintf("%s.%s.%s", pk.Schema, pk.TableName, pk.Name)
		if _, exist := mapTargetPrimaryKey[key]; !exist {
			updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdateTablePrimaryKey)
			break
		}
	}

	// compare columns
	updateItem.ChangeColumnItems = compareColumns(source.Columns, target.Columns)

	// compare relations
	updateItem.ChangeRelationItems = compareRelations(&source, source.Relationships, target.Relationships)

	if len(updateItem.ChangeItems) == 0 && len(updateItem.ChangeColumnItems) == 0 && len(updateItem.ChangeRelationItems) == 0 {
		diffResult.IsConflict = false
	} else {
		Logger.Debug("changeItems", "conflict-check", updateItem.ChangeItems)
		Logger.Debug("changeColumnItems", "conflict-check", updateItem.ChangeColumnItems)
		Logger.Debug("changeRelationItems", "conflict-check", updateItem.ChangeRelationItems)
		diffResult.IsConflict = true
	}

	diffResult.DiffItems = updateItem
	return
}

func compareColumns(source, target []objects.Column) (updateItems []objects.UpdateColumnItem) {
	mapTargetColumn := make(map[string]objects.Column)
	for i := range target {
		c := target[i]
		mapTargetColumn[c.Name] = c
	}

	for i := range source {
		sc := source[i]

		tc, exist := mapTargetColumn[sc.Name]
		if !exist {
			updateItems = append(updateItems, objects.UpdateColumnItem{
				Name:        sc.Name,
				UpdateItems: []objects.UpdateColumnType{objects.UpdateColumnNew},
			})
			continue
		}

		var updateColumnItems []objects.UpdateColumnType

		if sc.Name != tc.Name {
			updateColumnItems = append(updateColumnItems, objects.UpdateColumnName)
		}

		var sourceDefault, targetDefault *string
		switch d := sc.DefaultValue.(type) {
		case string:
			sourceDefault = &d
		case *string:
			if d != nil {
				sourceDefault = d
			}
		case nil:
			sourceDefault = nil
		}

		switch d := tc.DefaultValue.(type) {
		case string:
			targetDefault = &d
		case *string:
			if d != nil {
				targetDefault = d
			}
		case nil:
			targetDefault = nil
		}

		if (sourceDefault != nil && targetDefault == nil) ||
			(sourceDefault == nil && targetDefault != nil) ||
			(sourceDefault != nil && targetDefault != nil && utils.CleanDoubleColonPattern(*sourceDefault) != utils.CleanDoubleColonPattern(*targetDefault)) {
			updateColumnItems = append(updateColumnItems, objects.UpdateColumnDefaultValue)
		}

		// updateColumnItems = append(updateColumnItems, objects.UpdateColumnDefaultValue)
		if sc.DataType != tc.DataType {
			fmt.Println("[sc.DataType] : ", sc.DataType)
			fmt.Println("[tc.DataType] : ", tc.DataType)
			updateColumnItems = append(updateColumnItems, objects.UpdateColumnDataType)
		}

		if sc.IsUnique != tc.IsUnique {
			updateColumnItems = append(updateColumnItems, objects.UpdateColumnUnique)
		}

		if sc.IsNullable != tc.IsNullable {
			updateColumnItems = append(updateColumnItems, objects.UpdateColumnNullable)
		}

		if sc.IsIdentity != tc.IsIdentity {
			updateColumnItems = append(updateColumnItems, objects.UpdateColumnIdentity)
		}

		if len(updateColumnItems) == 0 {
			delete(mapTargetColumn, sc.Name)
			continue
		}

		updateItems = append(updateItems, objects.UpdateColumnItem{
			Name:        sc.Name,
			UpdateItems: updateColumnItems,
		})
		delete(mapTargetColumn, sc.Name)
	}

	if len(mapTargetColumn) > 0 {
		for _, c := range mapTargetColumn {
			updateItems = append(updateItems, objects.UpdateColumnItem{
				Name:        c.Name,
				UpdateItems: []objects.UpdateColumnType{objects.UpdateColumnDelete},
			})
		}
	}
	return
}

func compareRelations(table *objects.Table, source, target []objects.TablesRelationship) (updateItems []objects.UpdateRelationItem) {
	mapTargetRelation := make(map[string]objects.TablesRelationship)
	for i := range target {
		c := target[i]
		if !strings.HasPrefix(c.ConstraintName, fmt.Sprintf("%s_", c.SourceSchema)) {
			c.ConstraintName = fmt.Sprintf("%s_%s", c.SourceSchema, c.ConstraintName)
		}
		mapTargetRelation[c.ConstraintName] = c
	}

	for i := range source {
		sc := source[i]

		if sc.SourceTableName != table.Name {
			continue
		}

		if !strings.HasPrefix(sc.ConstraintName, fmt.Sprintf("%s_", sc.SourceSchema)) {
			Logger.Debug("constrain name update", "from", sc.ConstraintName, "to", fmt.Sprintf("%s_%s", sc.SourceSchema, sc.ConstraintName))
			sc.ConstraintName = fmt.Sprintf("%s_%s", sc.SourceSchema, sc.ConstraintName)
		}

		t, exist := mapTargetRelation[sc.ConstraintName]
		if !exist {
			updateItems = append(updateItems, objects.UpdateRelationItem{
				Data: sc,
				Type: objects.UpdateRelationCreate,
			})
			continue
		}

		if t.Index == nil && sc.Index == nil {
			updateItems = append(updateItems, objects.UpdateRelationItem{
				Data: sc,
				Type: objects.UpdateRelationCreateIndex,
			})
			Logger.Debug("create new index", "constrain-name", sc.ConstraintName)
		}

		if t.Action != nil && sc.Action != nil {
			if t.Action.UpdateAction != sc.Action.UpdateAction {
				updateItems = append(updateItems, objects.UpdateRelationItem{
					Data: sc,
					Type: objects.UpdateRelationActionOnUpdate,
				})
				Logger.Debug("check on update", "t-on-update", t.Action.UpdateAction, "sc-on-delete", sc.Action.UpdateAction, "same", t.Action.UpdateAction == sc.Action.UpdateAction)
			}

			if t.Action.DeletionAction != sc.Action.DeletionAction {
				updateItems = append(updateItems, objects.UpdateRelationItem{
					Data: sc,
					Type: objects.UpdateRelationActionOnDelete,
				})
				Logger.Debug("check on delete", "t-on-delete", t.Action.DeletionAction, "sc-on-delete", sc.Action.DeletionAction, "same", t.Action.DeletionAction == sc.Action.DeletionAction)
			}
		} else if t.Action != nil && sc.Action == nil {
			updateItems = append(updateItems, objects.UpdateRelationItem{
				Data: sc,
				Type: objects.UpdateRelationActionOnUpdate,
			})

			updateItems = append(updateItems, objects.UpdateRelationItem{
				Data: sc,
				Type: objects.UpdateRelationActionOnDelete,
			})
			Logger.Debug("create relation new action", "on-update", t.Action.UpdateAction, "on-delete", t.Action.DeletionAction)
		}

		delete(mapTargetRelation, sc.ConstraintName)

		if (sc.SourceSchema != t.SourceSchema) || (sc.SourceTableName != t.SourceTableName) || (sc.SourceColumnName != t.SourceColumnName) {
			updateItems = append(updateItems, objects.UpdateRelationItem{
				Data: sc,
				Type: objects.UpdateRelationUpdate,
			})
			Logger.Debug("update relation, source not match", "constrain-name", sc.ConstraintName)
			continue
		}

		if (sc.TargetTableSchema != t.TargetTableSchema) || (sc.TargetTableName != t.TargetTableName) || (sc.TargetColumnName != t.TargetColumnName) {
			updateItems = append(updateItems, objects.UpdateRelationItem{
				Data: sc,
				Type: objects.UpdateRelationUpdate,
			})
			Logger.Debug("update relation, target not match", "constrain-name", sc.ConstraintName)
			continue
		}
	}

	if len(mapTargetRelation) > 0 {
		for _, r := range mapTargetRelation {
			if r.SourceTableName != table.Name {
				continue
			}

			updateItems = append(updateItems, objects.UpdateRelationItem{
				Data: r,
				Type: objects.UpdateRelationDelete,
			})
		}
	}

	return
}
