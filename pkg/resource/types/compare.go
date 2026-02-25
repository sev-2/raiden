package types

import (
	"github.com/sev-2/raiden/pkg/supabase/objects"
)

type CompareDiffResult struct {
	Name           string
	SourceResource objects.Type
	TargetResource objects.Type
	DiffItems      objects.UpdateTypeParam
	IsConflict     bool
}

func Compare(source []objects.Type, target []objects.Type) error {
	diffResult, err := CompareList(source, target)
	if err != nil {
		return err
	}
	return PrintDiffResult(diffResult)
}

func CompareList(sourceType, targetType []objects.Type) (diffResult []CompareDiffResult, err error) {
	mapTargetTypes := make(map[int]objects.Type)
	for i := range targetType {
		r := targetType[i]
		mapTargetTypes[r.ID] = r
	}

	for i := range sourceType {
		r := sourceType[i]

		tr, isExist := mapTargetTypes[r.ID]
		if !isExist {
			continue
		}

		diffResult = append(diffResult, CompareItem(r, tr))
	}

	return
}

func CompareItem(source, target objects.Type) (diffResult CompareDiffResult) {

	var updateItem objects.UpdateTypeParam

	// assign diff result object
	diffResult.SourceResource = source
	diffResult.TargetResource = target

	if source.Name != target.Name {
		updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdateTypeName)
	}

	if (source.Comment != nil && target.Comment == nil) || (source.Comment == nil && target.Comment != nil) {
		updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdateTypeComment)
	} else if source.Comment != nil && target.Comment != nil {
		if *source.Comment != *target.Comment {
			updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdateTypeComment)
		}
	}

	if source.Format != target.Format {
		updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdateTypeFormat)
	}

	if source.Schema != target.Schema {
		updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdateTypeSchema)
	}

	if len(source.Enums) != len(target.Enums) {
		updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdateTypeEnums)
	} else {
		for _, se := range source.Enums {
			isFound := false
			for _, te := range target.Enums {
				if se == te {
					isFound = true
					break
				}
			}

			if !isFound {
				updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdateTypeEnums)
				break
			}
		}
	}

	if len(source.Attributes) != len(target.Attributes) {
		updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdateTypeAttributes)
	} else {
		for _, sa := range source.Attributes {
			isFound := false
			for _, ta := range target.Attributes {
				if sa.Name == ta.Name && sa.TypeID == ta.TypeID {
					isFound = true
					break
				}
			}

			if !isFound {
				updateItem.ChangeItems = append(updateItem.ChangeItems, objects.UpdateTypeAttributes)
				break
			}
		}
	}

	diffResult.IsConflict = len(updateItem.ChangeItems) > 0
	diffResult.DiffItems = updateItem

	return
}
