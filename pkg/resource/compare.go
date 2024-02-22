package resource

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/utils"
)

type (
	CompareMode string

	CompareDiffResult[T any, D any] struct {
		Name           string
		SourceResource T
		TargetResource T
		DiffItems      D
		IsConflict     bool
	}
)

const (
	CompareModeImport CompareMode = "import"
	CompareModeApply  CompareMode = "apply"
)

func CompareRoles(sourceRole []objects.Role, targetRole []objects.Role) (diffResult []CompareDiffResult[objects.Role, any], err error) {
	mapTargetRoles := make(map[int]objects.Role)
	for i := range targetRole {
		r := targetRole[i]
		mapTargetRoles[r.ID] = r
	}

	for i := range sourceRole {
		r := sourceRole[i]

		targetRole, isExist := mapTargetRoles[r.ID]
		if isExist {
			scByte, err := json.Marshal(r)
			if err != nil {
				return diffResult, err
			}
			scHash := utils.HashByte(scByte)

			targetByte, err := json.Marshal(targetRole)
			if err != nil {
				return diffResult, err
			}
			targetHash := utils.HashByte(targetByte)

			if scHash != targetHash {
				diffResult = append(diffResult, CompareDiffResult[objects.Role, any]{
					Name:           r.Name,
					SourceResource: r,
					TargetResource: targetRole,
				})
			}
		}
	}

	return
}

// CompareTables is function for find different table configuration between supabase and app
//
// for CompareModeImport source table is supabase table and target table is app table
// for CompareModeApply source table is app table and target table is supabase table
func CompareTables(sourceTable []objects.Table, targetTable []objects.Table, mode CompareMode) (diffResult []CompareDiffResult[objects.Table, objects.UpdateTableItem], err error) {
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

		if mode == CompareModeImport {
			// make sure set default to empty array
			// because default value from response is empty array
			if tt.Relationships == nil {
				tt.Relationships = make([]objects.TablesRelationship, 0)
			}

			diff, err := compareTableImportMode(t, tt)
			if err != nil {
				return diffResult, err
			}

			diffResult = append(diffResult, diff)
			continue
		}

		if mode == CompareModeApply {
			diffResult = append(diffResult, compareTableApplyMode(t, tt))
			continue
		}
	}

	return
}

func compareTableImportMode(source, target objects.Table) (diffResult CompareDiffResult[objects.Table, objects.UpdateTableItem], err error) {
	scByte, err := json.Marshal(source)
	if err != nil {
		return diffResult, err
	}
	scHash := utils.HashByte(scByte)

	targetByte, err := json.Marshal(target)
	if err != nil {
		return diffResult, err
	}

	targetHash := utils.HashByte(targetByte)

	if scHash != targetHash {
		diffResult = CompareDiffResult[objects.Table, objects.UpdateTableItem]{
			Name:           source.Name,
			SourceResource: source,
			TargetResource: target,
		}
	}

	return
}

func compareTableApplyMode(source, target objects.Table) (diffResult CompareDiffResult[objects.Table, objects.UpdateTableItem]) {
	var updateTableItem objects.UpdateTableItem

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
		updateTableItem.UpdateItems = append(updateTableItem.UpdateItems, objects.UpdateTableName)
	}

	if source.Schema != target.Schema {
		updateTableItem.UpdateItems = append(updateTableItem.UpdateItems, objects.UpdateTableSchema)
	}

	if source.RLSEnabled != target.RLSEnabled {
		updateTableItem.UpdateItems = append(updateTableItem.UpdateItems, objects.UpdateTableRlsEnable)
	}

	if source.RLSForced != target.RLSForced {
		updateTableItem.UpdateItems = append(updateTableItem.UpdateItems, objects.UpdateTableRlsForced)
	}

	for i := range source.PrimaryKeys {
		pk := target.PrimaryKeys[i]
		key := fmt.Sprintf("%s.%s.%s", pk.Schema, pk.TableName, pk.Name)
		if _, exist := mapTargetPrimaryKey[key]; !exist {
			updateTableItem.UpdateItems = append(updateTableItem.UpdateItems, objects.UpdateTablePrimaryKey)
			break
		}
	}

	updateTableItem.Column = compareColumns(source.Columns, target.Columns)

	if len(updateTableItem.UpdateItems) == 0 && len(updateTableItem.Column) == 0 {
		diffResult.IsConflict = false
	} else {
		diffResult.IsConflict = true
	}

	diffResult.DiffItems = updateTableItem
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

		switch d := sc.DefaultValue.(type) {
		case string:
			if d != tc.DefaultValue {
				updateColumnItems = append(updateColumnItems, objects.UpdateColumnDefaultValue)
			}
		case *string:
			if d != nil {
				if *d != tc.DefaultValue {
					updateColumnItems = append(updateColumnItems, objects.UpdateColumnDefaultValue)
				}
			} else if d == nil && tc.DefaultValue != nil {
				updateColumnItems = append(updateColumnItems, objects.UpdateColumnDefaultValue)
			}
		case nil:
			if tc.DefaultValue != nil {
				updateColumnItems = append(updateColumnItems, objects.UpdateColumnDefaultValue)
			}
		}

		if sc.DataType != tc.DataType {
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

func CompareRpcFunctions(sourceFn []objects.Function, targetFn []objects.Function) (diffResult []CompareDiffResult[objects.Function, any], err error) {
	mapTargetFn := make(map[int]objects.Function)
	for i := range targetFn {
		f := targetFn[i]
		mapTargetFn[f.ID] = f
	}

	for i := range sourceFn {
		sFn := sourceFn[i]

		tFn, isExist := mapTargetFn[sFn.ID]
		if !isExist {
			continue
		}

		dFields := strings.Fields(utils.CleanUpString(sFn.CompleteStatement))
		for i := range dFields {
			d := dFields[i]
			if strings.HasSuffix(d, ";") && strings.ToLower(d) != "end;" {
				dFields[i] = strings.ReplaceAll(d, ";", " ;")
			}

			if strings.Contains(strings.ToLower(d), "end;$") {
				dFields[i] = strings.ReplaceAll(d, ";", "; ")
			}

		}
		sFn.CompleteStatement = strings.ToLower(strings.Join(dFields, " "))

		if sFn.CompleteStatement != tFn.CompleteStatement {
			diffResult = append(diffResult, CompareDiffResult[objects.Function, any]{
				Name:           sFn.Name,
				SourceResource: sFn,
				TargetResource: tFn,
			})
		}
	}

	return
}
