package storages

import (
	"github.com/sev-2/raiden/pkg/state"
	"github.com/sev-2/raiden/pkg/supabase/objects"
)

func GetNewCountData(cloudData []objects.Bucket, localData state.ExtractStorageResult) int {
	var newCount int

	mapData := localData.ToDeleteFlatMap()
	for i := range cloudData {
		r := cloudData[i]

		if _, exist := mapData[r.Name]; exist {
			newCount++
		}
	}

	return newCount
}
