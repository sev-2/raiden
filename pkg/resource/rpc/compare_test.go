package rpc_test

import (
	"testing"

	"github.com/sev-2/raiden/pkg/resource/rpc"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/stretchr/testify/assert"
)

func TestCompare(t *testing.T) {
	source := []objects.Function{
		{Name: "function1"},
		{Name: "function2"},
	}

	target := []objects.Function{
		{Name: "function1"},
		{Name: "function2"},
	}

	err := rpc.Compare(source, target)
	assert.NoError(t, err)
}

func TestCompareList(t *testing.T) {
	source := []objects.Function{
		{Name: "function1"},
		{Name: "function2"},
	}

	target := []objects.Function{
		{Name: "function1_updated"},
		{Name: "function2"},
	}

	diffResult, err := rpc.CompareList(source, target)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(diffResult))
	assert.Equal(t, "function2", diffResult[0].SourceResource.Name)
	assert.Equal(t, "function2", diffResult[0].TargetResource.Name)
}

func TestCompareItem(t *testing.T) {
	source := objects.Function{
		Name: "function1",
	}

	target := objects.Function{
		Name: "function1_updated",
	}

	diffResult := rpc.CompareItem(source, target)
	assert.False(t, diffResult.IsConflict)
	assert.Equal(t, "function1", diffResult.SourceResource.Name)
	assert.Equal(t, "function1_updated", diffResult.TargetResource.Name)
}
