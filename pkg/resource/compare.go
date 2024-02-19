package resource

import (
	"encoding/json"
	"strings"

	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/utils"
)

type (
	CompareMode         string
	CompareDiffType     string
	CompareDiffCategory string

	CompareDiffResult struct {
		Name           string
		Category       CompareDiffCategory
		SourceResource any
		TargetResource any
	}
)

const (
	CompareModeImport CompareMode = "import"
	CompareModeApply  CompareMode = "apply"

	CompareDiffCategoryConflict      CompareDiffCategory = "conflict"
	CompareDiffCategoryCloudNofFound CompareDiffCategory = "cloud-not-found"
	CompareDiffCategoryAppNotFound   CompareDiffCategory = "app-not-found"
)

func CompareRoles(sourceRole []objects.Role, targetRole []objects.Role) (diffResult []CompareDiffResult, err error) {
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
				diffResult = append(diffResult, CompareDiffResult{
					Name:           r.Name,
					Category:       CompareDiffCategoryConflict,
					SourceResource: r,
					TargetResource: targetRole,
				})
			}
		}
	}

	return
}

func CompareTables(sourceTable []objects.Table, targetTable []objects.Table) (diffResult []CompareDiffResult, err error) {
	mapTargetTable := make(map[int]objects.Table)
	for i := range targetTable {
		t := targetTable[i]
		mapTargetTable[t.ID] = t
	}

	for i := range sourceTable {
		t := sourceTable[i]

		targetTable, isExist := mapTargetTable[t.ID]
		if isExist {
			scByte, err := json.Marshal(t)
			if err != nil {
				return diffResult, err
			}
			scHash := utils.HashByte(scByte)

			// make sure set default to empty array
			// because default value from response is empty array
			if targetTable.Relationships == nil {
				targetTable.Relationships = make([]objects.TablesRelationship, 0)
			}

			targetByte, err := json.Marshal(targetTable)
			if err != nil {
				return diffResult, err
			}
			targetHash := utils.HashByte(targetByte)

			if scHash != targetHash {
				diffResult = append(diffResult, CompareDiffResult{
					Name:           t.Name,
					Category:       CompareDiffCategoryConflict,
					SourceResource: t,
					TargetResource: targetTable,
				})
			}
		}
	}

	return
}

func CompareRpcFunctions(sourceFn []objects.Function, targetFn []objects.Function) (diffResult []CompareDiffResult, err error) {
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
			logger.Info("source : ", sFn.CompleteStatement)
			diffResult = append(diffResult, CompareDiffResult{
				Name:           sFn.Name,
				Category:       CompareDiffCategoryConflict,
				SourceResource: sFn,
				TargetResource: tFn,
			})
		}
	}

	return
}
