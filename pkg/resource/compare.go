package resource

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/logger"
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

func CompareRoles(sourceRole, targetRole []objects.Role, mode CompareMode) (diffResult []CompareDiffResult[objects.Role, objects.UpdateRoleParam], err error) {
	mapTargetRoles := make(map[int]objects.Role)
	for i := range targetRole {
		r := targetRole[i]
		mapTargetRoles[r.ID] = r
	}

	for i := range sourceRole {
		r := sourceRole[i]

		tr, isExist := mapTargetRoles[r.ID]
		if !isExist {
			continue
		}

		if mode == CompareModeImport {
			diff, err := CompareRoleImportMode(r, tr)
			if err != nil {
				return diffResult, err
			}

			diffResult = append(diffResult, diff)
			continue
		}

		if mode == CompareModeApply {
			diffResult = append(diffResult, compareRoleApplyMode(r, tr))
			continue
		}

	}

	return
}

func CompareRoleImportMode(source, target objects.Role) (diffResult CompareDiffResult[objects.Role, objects.UpdateRoleParam], err error) {
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
		diffResult = CompareDiffResult[objects.Role, objects.UpdateRoleParam]{
			Name:           source.Name,
			SourceResource: source,
			TargetResource: target,
		}
	}

	return
}

func compareRoleApplyMode(source, target objects.Role) (diffResult CompareDiffResult[objects.Role, objects.UpdateRoleParam]) {

	var updateItem objects.UpdateRoleParam

	// assign diff result object
	diffResult.SourceResource = source
	diffResult.TargetResource = target

	if source.Name != target.Name {
		updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdateRoleName)
	}

	if source.ConnectionLimit != target.ConnectionLimit {
		updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdateConnectionLimit)
	}

	if source.CanBypassRLS != target.CanBypassRLS {
		updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdateRoleCanBypassRls)
	}

	if source.CanCreateDB != target.CanCreateDB {
		updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdateRoleCanCreateDb)
	}

	if source.CanCreateRole != target.CanCreateRole {
		updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdateRoleCanCreateRole)
	}

	if source.CanLogin != target.CanLogin {
		updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdateRoleCanLogin)
	}

	if (source.Config == nil && target.Config != nil) || (source.Config != nil && target.Config == nil) || (len(source.Config) != len(target.Config)) {
		updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdateRoleConfig)
	} else if source.Config != nil && target.Config != nil {
		for k, v := range source.Config {
			tv, exist := target.Config[k]
			if !exist {
				updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdateRoleConfig)
				break
			}

			svValue := reflect.ValueOf(v)
			tvValue := reflect.ValueOf(tv)

			if svValue.Kind() != tvValue.Kind() {
				updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdateRoleConfig)
				break
			}

			if svValue.Kind() == reflect.Ptr {
				if svValue.IsNil() && !tvValue.IsNil() {
					updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdateRoleConfig)
					break
				}

				if svValue.Interface() != tvValue.Interface() {
					updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdateRoleConfig)
					break
				}
			} else {
				if svValue.Interface() != tvValue.Interface() {
					updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdateRoleConfig)
					break
				}
			}
		}
	}

	if source.InheritRole != target.InheritRole {
		updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdateRoleInheritRole)
	}

	// Unsupported now, because need superuser role
	// if source.IsReplicationRole != target.IsReplicationRole {
	// 	updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdateRoleIsReplication)
	// }

	// if source.IsSuperuser != target.IsSuperuser {
	// 	updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdateRoleIsSuperUser)
	// }

	if (source.ValidUntil != nil && target.ValidUntil == nil) || (source.ValidUntil == nil && target.ValidUntil != nil) {
		updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdateRoleValidUntil)
	} else if source.ValidUntil != nil && target.ValidUntil != nil {
		sv := source.ValidUntil.Format(raiden.DefaultRoleValidUntilLayout)
		tv := target.ValidUntil.Format(raiden.DefaultRoleValidUntilLayout)

		if sv != tv {
			updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdateRoleValidUntil)
		}
	}

	diffResult.IsConflict = len(updateItem.ChangeItems) > 0
	diffResult.DiffItems = updateItem

	return
}

// CompareTables is function for find different table configuration between supabase and app
//
// for CompareModeImport source table is supabase table and target table is app table
// for CompareModeApply source table is app table and target table is supabase table
func CompareTables(sourceTable, targetTable []objects.Table, mode CompareMode) (diffResult []CompareDiffResult[objects.Table, objects.UpdateTableParam], err error) {
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

func compareTableImportMode(source, target objects.Table) (diffResult CompareDiffResult[objects.Table, objects.UpdateTableParam], err error) {
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
		diffResult = CompareDiffResult[objects.Table, objects.UpdateTableParam]{
			Name:           source.Name,
			SourceResource: source,
			TargetResource: target,
		}
	}

	return
}

func compareTableApplyMode(source, target objects.Table) (diffResult CompareDiffResult[objects.Table, objects.UpdateTableParam]) {
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
			} else if tc.DefaultValue != nil {
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

func compareRelations(table *objects.Table, source, target []objects.TablesRelationship) (updateItems []objects.UpdateRelationItem) {
	mapTargetRelation := make(map[string]objects.TablesRelationship)
	for i := range target {
		c := target[i]
		mapTargetRelation[c.ConstraintName] = c
	}

	for i := range source {
		sc := source[i]

		if sc.SourceTableName != table.Name {
			continue
		}

		t, exist := mapTargetRelation[sc.ConstraintName]
		if !exist {
			updateItems = append(updateItems, objects.UpdateRelationItem{
				Data: sc,
				Type: objects.UpdateRelationCreate,
			})
			continue
		}

		delete(mapTargetRelation, sc.ConstraintName)

		if (sc.SourceSchema != t.SourceSchema) || (sc.SourceTableName != t.SourceTableName) || (sc.SourceColumnName != t.SourceColumnName) {
			updateItems = append(updateItems, objects.UpdateRelationItem{
				Data: sc,
				Type: objects.UpdateRelationUpdate,
			})
			logger.Debugf("update %s relation, source not match", sc.ConstraintName)
			continue
		}

		if (sc.TargetTableSchema != t.TargetTableSchema) || (sc.TargetTableName != t.TargetTableName) || (sc.TargetColumnName != t.TargetColumnName) {
			updateItems = append(updateItems, objects.UpdateRelationItem{
				Data: sc,
				Type: objects.UpdateRelationUpdate,
			})
			logger.Debug("update %s relation, target not match", sc.ConstraintName)
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
