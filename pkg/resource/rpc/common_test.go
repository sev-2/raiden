package rpc_test

import (
	"testing"

	"github.com/sev-2/raiden/pkg/resource/rpc"
	"github.com/sev-2/raiden/pkg/state"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/stretchr/testify/assert"
)

func TestGetNewCountData(t *testing.T) {
	supabaseFunctions := []objects.Function{
		{Name: "function1"},
		{Name: "function2"},
		{Name: "function3"},
	}

	extractResult := state.ExtractRpcResult{
		Delete: []objects.Function{
			{Name: "function1"},
			{Name: "function2"},
		},
	}

	count := rpc.GetNewCountData(supabaseFunctions, extractResult)
	assert.Equal(t, 2, count)
}

func TestGetNewCountDataNoMatch(t *testing.T) {
	supabaseFunctions := []objects.Function{
		{Name: "function1"},
		{Name: "function2"},
	}

	extractResult := state.ExtractRpcResult{
		Delete: []objects.Function{
			{Name: "function3"},
			{Name: "function4"},
		},
	}

	count := rpc.GetNewCountData(supabaseFunctions, extractResult)
	assert.Equal(t, 0, count)
}

func TestGetNewCountDataEmpty(t *testing.T) {
	supabaseFunctions := []objects.Function{}

	extractResult := state.ExtractRpcResult{
		Delete: []objects.Function{},
	}

	count := rpc.GetNewCountData(supabaseFunctions, extractResult)
	assert.Equal(t, 0, count)
}

func TestGetNewCountDataPartialMatch(t *testing.T) {
	supabaseFunctions := []objects.Function{
		{Name: "function1"},
		{Name: "function2"},
		{Name: "function3"},
	}

	extractResult := state.ExtractRpcResult{
		Delete: []objects.Function{
			{Name: "function1"},
			{Name: "function4"},
		},
	}

	count := rpc.GetNewCountData(supabaseFunctions, extractResult)
	assert.Equal(t, 1, count)
}
