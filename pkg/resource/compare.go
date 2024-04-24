package resource

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/go-hclog"
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/utils"
)

var CompareLogger hclog.Logger = logger.HcLog().Named("import.compare")

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

func CompareRoles(sourceRole, targetRole []objects.Role) (diffResult []CompareDiffResult[objects.Role, objects.UpdateRoleParam], err error) {
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
		// diff, err := CompareRoleImportMode(r, tr)
		// if err != nil {
		// 	return diffResult, err
		// }

		// if !diff.IsConflict {
		// 	continue
		// }

		// diffResult = append(diffResult, diff)
		// continue
		diffResult = append(diffResult, compareRoleApplyMode(r, tr))
		continue
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
func CompareTables(sourceTable, targetTable []objects.Table) (diffResult []CompareDiffResult[objects.Table, objects.UpdateTableParam], err error) {
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

		df := compareTable(t, tt)
		if !df.IsConflict {
			continue
		}
		diffResult = append(diffResult, df)
	}

	return
}

func compareTable(source, target objects.Table) (diffResult CompareDiffResult[objects.Table, objects.UpdateTableParam]) {
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
			(sourceDefault != nil && targetDefault != nil && *sourceDefault != *targetDefault) {
			updateColumnItems = append(updateColumnItems, objects.UpdateColumnDefaultValue)
		}

		// updateColumnItems = append(updateColumnItems, objects.UpdateColumnDefaultValue)
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
			CompareLogger.Debug("update relation, source not match", "constrain-name", sc.ConstraintName)
			continue
		}

		if (sc.TargetTableSchema != t.TargetTableSchema) || (sc.TargetTableName != t.TargetTableName) || (sc.TargetColumnName != t.TargetColumnName) {
			updateItems = append(updateItems, objects.UpdateRelationItem{
				Data: sc,
				Type: objects.UpdateRelationUpdate,
			})
			CompareLogger.Debug("update relation, target not match", "constrain-name", sc.ConstraintName)
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

		sFn.CompleteStatement = strings.ToLower(utils.CleanUpString(sFn.CompleteStatement))
		tFn.CompleteStatement = strings.ToLower(utils.CleanUpString(tFn.CompleteStatement))

		sourceCompare := strings.ReplaceAll(sFn.CompleteStatement, " ", "")
		targetCompare := strings.ReplaceAll(tFn.CompleteStatement, " ", "")
		if sourceCompare != targetCompare {
			diffResult = append(diffResult, CompareDiffResult[objects.Function, any]{
				Name:           sFn.Name,
				SourceResource: sFn,
				TargetResource: tFn,
				IsConflict:     true,
			})
		}
	}

	return
}

func ComparePolicies(sourcePolicies, targetPolicies []objects.Policy) (diffResult []CompareDiffResult[objects.Policy, objects.UpdatePolicyParam]) {
	mapTargetPolicies := make(map[string]objects.Policy)
	for i := range targetPolicies {
		r := targetPolicies[i]
		mapTargetPolicies[r.Name] = r
	}
	for i := range sourcePolicies {
		p := sourcePolicies[i]

		tp, isExist := mapTargetPolicies[p.Name]
		if !isExist {
			continue
		}
		diffResult = append(diffResult, comparePolicy(p, tp))
	}

	return
}

func comparePolicy(source, target objects.Policy) (diffResult CompareDiffResult[objects.Policy, objects.UpdatePolicyParam]) {
	var updateItem objects.UpdatePolicyParam

	// assign diff result object
	diffResult.SourceResource = source
	diffResult.TargetResource = target
	updateItem.Name = source.Name

	sourceName := strings.ToLower(source.Name)
	targetName := strings.ToLower(target.Name)
	if sourceName != targetName {
		updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdatePolicyName)
	}

	if source.Definition != target.Definition {
		updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdatePolicyDefinition)
	}

	if (source.Check == nil && target.Check != nil) || (source.Check != nil && target.Check == nil) || (source.Check != nil && target.Check != nil && *source.Check != *target.Check) {
		updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdatePolicyCheck)
	}

	if len(source.Roles) != len(target.Roles) {
		updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdatePolicyRoles)
	} else {
		for sr := range source.Roles {
			isFound := false
			for tr := range target.Roles {
				if sr == tr {
					isFound = true
					break
				}
			}

			if !isFound {
				updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdatePolicyRoles)
				break
			}
		}
	}

	diffResult.IsConflict = len(updateItem.ChangeItems) > 0
	diffResult.DiffItems = updateItem
	return
}

// Compare Storage
func CompareStorage(sourceStorage, targetStorage []objects.Bucket) (diffResult []CompareDiffResult[objects.Bucket, objects.UpdateBucketParam], err error) {
	mapTargetStorage := make(map[string]objects.Bucket)
	for i := range targetStorage {
		s := targetStorage[i]
		mapTargetStorage[s.ID] = s
	}

	for i := range sourceStorage {
		s := sourceStorage[i]

		ts, isExist := mapTargetStorage[s.ID]
		if !isExist {
			continue
		}
		diffResult = append(diffResult, compareStorage(s, ts))
	}

	return
}

func compareStorage(source, target objects.Bucket) (diffResult CompareDiffResult[objects.Bucket, objects.UpdateBucketParam]) {
	var updateItem objects.UpdateBucketParam

	// assign diff result object
	diffResult.SourceResource = source
	diffResult.TargetResource = target

	if source.Public != target.Public {
		updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdateBucketIsPublic)
	}

	if len(source.AllowedMimeTypes) != len(target.AllowedMimeTypes) {
		updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdateBucketAllowedMimeTypes)
	} else {
		mapAllowed := make(map[string]bool)
		for _, amt := range target.AllowedMimeTypes {
			mapAllowed[amt] = true
		}

		for _, samt := range source.AllowedMimeTypes {
			if _, exist := mapAllowed[samt]; !exist {
				updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdateBucketAllowedMimeTypes)
				break
			}
		}
	}

	if (source.FileSizeLimit != nil && target.FileSizeLimit == nil) ||
		(source.FileSizeLimit == nil && target.FileSizeLimit != nil) ||
		(source.FileSizeLimit != nil && target.FileSizeLimit != nil && *source.FileSizeLimit != *target.FileSizeLimit) {
		updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdateBucketFileSizeLimit)
	}

	diffResult.IsConflict = len(updateItem.ChangeItems) > 0
	diffResult.DiffItems = updateItem
	return
}
