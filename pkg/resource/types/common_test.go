package types_test

import (
	"testing"

	"github.com/sev-2/raiden/pkg/resource/types"
	"github.com/sev-2/raiden/pkg/state"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/stretchr/testify/assert"
)

func TestGetNewCountData(t *testing.T) {
	supabaseTypes := []objects.Type{
		{Name: "type1"},
		{Name: "type2"},
		{Name: "type3"},
	}

	extractResult := state.ExtractTypeResult{
		Delete: []objects.Type{
			{Name: "type1"},
			{Name: "type2"},
		},
	}

	count := types.GetNewCountData(supabaseTypes, extractResult)
	assert.Equal(t, 2, count)
}

func TestGetNewCountDataNoMatch(t *testing.T) {
	supabaseTypes := []objects.Type{
		{Name: "type1"},
		{Name: "type2"},
	}

	extractResult := state.ExtractTypeResult{
		Delete: []objects.Type{
			{Name: "type3"},
			{Name: "type4"},
		},
	}

	count := types.GetNewCountData(supabaseTypes, extractResult)
	assert.Equal(t, 0, count)
}

func TestGetNewCountDataEmpty(t *testing.T) {
	supabaseTypes := []objects.Type{}

	extractResult := state.ExtractTypeResult{
		Delete: []objects.Type{},
	}

	count := types.GetNewCountData(supabaseTypes, extractResult)
	assert.Equal(t, 0, count)
}

func TestGetNewCountDataPartialMatch(t *testing.T) {
	supabaseTypes := []objects.Type{
		{Name: "type1"},
		{Name: "type2"},
		{Name: "type3"},
	}

	extractResult := state.ExtractTypeResult{
		Delete: []objects.Type{
			{Name: "type1"},
			{Name: "type4"},
		},
	}

	count := types.GetNewCountData(supabaseTypes, extractResult)
	assert.Equal(t, 1, count)
}
