package rpc

import (
	"strings"

	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/utils"
)

type CompareDiffResult struct {
	Name           string
	SourceResource objects.Function
	TargetResource objects.Function
	IsConflict     bool
}

func Compare(source []objects.Function, target []objects.Function) error {
	diffResult, err := CompareList(source, target)
	if err != nil {
		return err
	}
	return PrintDiffResult(diffResult)
}

func CompareList(sourceFn []objects.Function, targetFn []objects.Function) (diffResult []CompareDiffResult, err error) {
	mapTargetFn := make(map[int]objects.Function)
	for i := range targetFn {
		f := targetFn[i]
		mapTargetFn[f.ID] = f
	}

	for i := range sourceFn {
		s := sourceFn[i]

		t, isExist := mapTargetFn[s.ID]
		if !isExist {
			continue
		}

		diffResult = append(diffResult, CompareItem(s, t))
	}

	return
}

func CompareItem(source, target objects.Function) (diffResult CompareDiffResult) {
	// assign diff result object
	diffResult.SourceResource = source
	diffResult.TargetResource = target
	diffResult.Name = source.Name

	source.CompleteStatement = strings.ToLower(utils.CleanUpString(source.CompleteStatement))
	target.CompleteStatement = strings.ToLower(utils.CleanUpString(target.CompleteStatement))
	sourceCompare := strings.ReplaceAll(source.CompleteStatement, " ", "")
	targetCompare := strings.ReplaceAll(target.CompleteStatement, " ", "")

	diffResult.IsConflict = sourceCompare != targetCompare
	return
}
