package state

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

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
		// Preserve the state's CompleteStatement (captured from pg_get_functiondef
		// during the last import). BindRpcFunction rebuilds it via BuildRpc() which
		// differs in formatting (param prefix, default quoting, search_path),
		// causing false update detections even when no code was changed.
		stateCompleteStatement := fn.CompleteStatement
		if err := BindRpcFunction(r, &fn); err != nil {
			return result, err
		}
		if stateCompleteStatement != "" {
			fn.CompleteStatement = stateCompleteStatement
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
	fn.Language = rpc.GetLanguage()
	fn.CompleteStatement = rpc.GetCompleteStmt()

	// validate definition query
	cleanStatement := strings.ReplaceAll(fn.CompleteStatement, "::", "")
	matches := regexp.MustCompile(`:\w+`).FindAllString(cleanStatement, -1)
	if len(matches) > 0 {
		var errMsg string
		if len(matches) > 1 {
			errMsg = fmt.Sprintf("rpc %q is invalid, There are %q keys that are not mapped with any parameters or models.", rpc.GetName(), strings.Join(matches, ","))
		} else {
			errMsg = fmt.Sprintf("rpc %q is invalid, There is %q key that is not mapped with any parameters or models.", rpc.GetName(), matches[0])
		}
		return errors.New(errMsg)
	}

	return
}

func (er ExtractRpcResult) ToDeleteFlatMap() map[string]*objects.Function {
	mapData := make(map[string]*objects.Function)

	if len(er.Delete) > 0 {
		for i := range er.Delete {
			r := er.Delete[i]
			mapData[r.Name] = &r
		}
	}

	return mapData
}
