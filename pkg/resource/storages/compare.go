package storages

import "github.com/sev-2/raiden/pkg/supabase/objects"

type CompareDiffResult struct {
	Name           string
	SourceResource objects.Bucket
	TargetResource objects.Bucket
	DiffItems      objects.UpdateBucketParam
	IsConflict     bool
}

func Compare(source []objects.Bucket, target []objects.Bucket) error {
	diffResult, err := CompareList(source, target)
	if err != nil {
		return err
	}
	return PrintDiffResult(diffResult)
}

func CompareList(sourceStorage, targetStorage []objects.Bucket) (diffResult []CompareDiffResult, err error) {
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
		diffResult = append(diffResult, CompareItem(s, ts))
	}

	return
}

func CompareItem(source, target objects.Bucket) (diffResult CompareDiffResult) {
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
