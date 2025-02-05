package types_test

import (
	"fmt"
	"testing"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/resource/rpc"
	"github.com/sev-2/raiden/pkg/state"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/stretchr/testify/assert"
)

func TestBuildMigrateData(t *testing.T) {
	extractedLocalData := state.ExtractRpcResult{
		New: []objects.Function{
			{Name: "function4"},
		},
		Existing: []objects.Function{
			{Name: "function2"},
			{Name: "function3"},
		},
		Delete: []objects.Function{
			{Name: "function1"},
		},
	}

	supabaseRpcs := []objects.Function{
		{Name: "function1"},
		{Name: "function2"},
		{Name: "function3"},
	}

	migrateData, err := rpc.BuildMigrateData(extractedLocalData, supabaseRpcs)
	assert.NoError(t, err)
	assert.Equal(t, 4, len(migrateData))
}

func TestBuildMigrateItem(t *testing.T) {
	localRpcs := []objects.Function{
		{Name: "function1"},
		{Name: "function2"},
	}

	supabaseRpcs := []objects.Function{
		{Name: "function1"},
		{Name: "function3"},
	}

	migrateData, err := rpc.BuildMigrateItem(supabaseRpcs, localRpcs)
	assert.NoError(t, err)
	fmt.Println(migrateData)
}

func TestMigrate(t *testing.T) {
	config := &raiden.Config{}
	stateChan := make(chan any)
	defer close(stateChan)

	migrateItems := []rpc.MigrateItem{
		{
			Type:    "create",
			NewData: objects.Function{Name: "function1"},
		},
	}

	errors := rpc.Migrate(config, migrateItems, stateChan, rpc.ActionFunc)
	assert.Equal(t, 1, len(errors))
}
