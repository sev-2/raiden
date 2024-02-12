package state

import (
	"strings"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/supabase"
	"github.com/sev-2/raiden/pkg/utils"
)

func ToRpc(rpcState []RpcState, appRpc []raiden.Rpc) (listRpc []supabase.Function, err error) {
	mapRpcState := map[string]RpcState{}
	for i := range rpcState {
		r := rpcState[i]
		mapRpcState[r.Function.Name] = r
	}

	for _, r := range appRpc {
		fn, e := createRpcFunction(mapRpcState, r)
		if e != nil {
			err = e
			return
		}

		if fn.ID > 0 {
			listRpc = append(listRpc, fn)
		}
	}

	return
}

func createRpcFunction(mapRpcState map[string]RpcState, rpc raiden.Rpc) (fn supabase.Function, err error) {
	if err = raiden.BuildRpc(rpc); err != nil {
		return
	}

	state, isStateExist := mapRpcState[rpc.GetName()]
	if !isStateExist {
		return
	}
	fn = state.Function
	fn.CompleteStatement = rpc.GetCompleteStmt()
	return
}

func CompareRpcFunctions(supabaseFn []supabase.Function, appFn []supabase.Function) (diffResult []CompareDiffResult, err error) {
	mapAppFn := make(map[int]supabase.Function)
	for i := range appFn {
		f := appFn[i]
		mapAppFn[f.ID] = f
	}

	for i := range supabaseFn {
		sf := supabaseFn[i]

		af, isExist := mapAppFn[sf.ID]
		if !isExist {
			continue
		}

		dFields := strings.Fields(utils.CleanUpString(sf.CompleteStatement))
		for i := range dFields {
			d := dFields[i]
			if strings.HasSuffix(d, ";") && strings.ToLower(d) != "end;" {
				dFields[i] = strings.ReplaceAll(d, ";", " ;")
			}
		}
		sf.CompleteStatement = strings.ToLower(strings.Join(dFields, " "))

		if af.CompleteStatement != sf.CompleteStatement {
			diffResult = append(diffResult, CompareDiffResult{
				Name:             sf.Name,
				Category:         CompareDiffCategoryConflict,
				SupabaseResource: sf,
				AppResource:      af,
			})
		}
	}

	return
}
