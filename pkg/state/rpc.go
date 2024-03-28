package state

import (
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/supabase/objects"
)

type ExtractRpcResult struct {
	Existing []objects.Function
	New      []objects.Function
	Delete   []objects.Function
}

func ExtractRpc(rpcState []RpcState, appRpc []raiden.Rpc) (result ExtractRpcResult, err error) {
	mapRpcState := map[string]RpcState{}
	for i := range rpcState {
		r := rpcState[i]
		mapRpcState[r.Function.Name] = r
	}

	for _, r := range appRpc {
		state, isStateExist := mapRpcState[r.GetName()]
		if !isStateExist {
			fn := objects.Function{}
			if err := BindRpcFunction(r, &fn); err != nil {
				return result, err
			}
			result.New = append(result.New, fn)
			continue
		}

		fn := state.Function
		if err := BindRpcFunction(r, &fn); err != nil {
			return result, err
		}

		if fn.CompleteStatement != "" {
			result.Existing = append(result.Existing, fn)
		}
		delete(mapRpcState, r.GetName())
	}

	for _, state := range mapRpcState {
		result.Delete = append(result.Delete, state.Function)
	}

	return
}

func BindRpcFunction(rpc raiden.Rpc, fn *objects.Function) (err error) {
	if err = raiden.BuildRpc(rpc); err != nil {
		return
	}

	fn.Name = rpc.GetName()
	fn.Schema = rpc.GetSchema()
	fn.CompleteStatement = rpc.GetCompleteStmt()
	return
}
