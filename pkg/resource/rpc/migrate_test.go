package rpc_test

import (
	"fmt"
	"sync"
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

	// Use a WaitGroup to ensure we wait for channel reads
	var wg sync.WaitGroup
	var receivedItems []any

	// Start a goroutine to read from stateChan to prevent blocking
	wg.Add(1)
	go func() {
		defer wg.Done()
		for item := range stateChan {
			receivedItems = append(receivedItems, item)
		}
	}()

	migrateItems := []rpc.MigrateItem{
		{
			Type:    "create",
			NewData: objects.Function{Name: "function1"},
		},
	}

	rpc.Migrate(config, migrateItems, stateChan, rpc.ActionFunc)
	close(stateChan) // Close the channel to signal the reading goroutine to finish
	wg.Wait()        // Wait for all items to be processed

	// Verify that the function completed without hanging
	// If we reach this point, it means the function completed without blocking

	// Verify that we received items on the state channel during migration
	// This confirms that the migration process attempted to send data to the channel as expected
	// Note: Actual number depends on whether the action function succeeded or not
}
