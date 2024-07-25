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

	// validate definition query
	allMatches := regexp.MustCompile(`:\w+`).FindAllString(fn.CompleteStatement, -1)
	allExcludes := regexp.MustCompile(`::\w+`).FindAllString(fn.CompleteStatement, -1)

	// Filter out double colon string that work as postgre typecast
	validMatches := []string{}
	excludeSet := make(map[string]bool)
	for _, exclude := range allExcludes {
		excludeSet[exclude] = true
	}

	for _, match := range allMatches {
		if !excludeSet[fmt.Sprintf(":%s", match)] {
			validMatches = append(validMatches, match)
		}
	}

	if len(validMatches) > 0 {
		var errMsg string
		if len(validMatches) > 1 {
			errMsg = fmt.Sprintf("rpc %q is invalid, There are %q keys that are not mapped with any parameters or models.", rpc.GetName(), strings.Join(validMatches, ","))
		} else {
			errMsg = fmt.Sprintf("rpc %q is invalid, There is %q key that is not mapped with any parameters or models.", rpc.GetName(), validMatches[0])
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

// Helper function to check if the index is inside a string literal
func isInsideStringLiteral(s string, index int) bool {
	inSingleQuote := false
	inDoubleQuote := false

	for i := 0; i < index; i++ {
		if s[i] == '\'' && !inDoubleQuote {
			inSingleQuote = !inSingleQuote
		}
		if s[i] == '"' && !inSingleQuote {
			inDoubleQuote = !inDoubleQuote
		}
	}
	return inSingleQuote || inDoubleQuote
}
