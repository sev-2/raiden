package state

import (
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/supabase/objects"
)

func ExtractRpc(rpcState []RpcState, appRpc []raiden.Rpc) (listRpc []objects.Function, err error) {
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

func createRpcFunction(mapRpcState map[string]RpcState, rpc raiden.Rpc) (fn objects.Function, err error) {
	if err = raiden.BuildRpc(rpc); err != nil {
		return
	}

	state, isStateExist := mapRpcState[rpc.GetName()]
	if !isStateExist {
		// TODO : handler new rpc
		return
	}
	fn = state.Function
	fn.CompleteStatement = rpc.GetCompleteStmt()
	return
}
