package rpc

import (
	"github.com/sev-2/raiden/pkg/state"
	"github.com/sev-2/raiden/pkg/supabase/objects"
)

func GetNewCountData(supabaseData []objects.Function, localData state.ExtractRpcResult) int {
	var newCount int

	mapData := localData.ToDeleteFlatMap()
	for i := range supabaseData {
		r := supabaseData[i]

		if _, exist := mapData[r.Name]; exist {
			newCount++
		}
	}

	return newCount
}
