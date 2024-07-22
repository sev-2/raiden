package storages_test

import (
	"fmt"
	"testing"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/resource/storages"
	"github.com/sev-2/raiden/pkg/state"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/stretchr/testify/assert"
)

func TestBuildMigrateData(t *testing.T) {
	extractedLocalData := state.ExtractStorageResult{
		New: []state.ExtractStorageItem{
			{Storage: objects.Bucket{Name: "bucket4"}},
		},
		Existing: []state.ExtractStorageItem{
			{Storage: objects.Bucket{Name: "bucket2"}},
			{Storage: objects.Bucket{Name: "bucket3"}},
		},
		Delete: []state.ExtractStorageItem{
			{Storage: objects.Bucket{Name: "bucket1"}},
		},
	}

	supabaseStorages := []objects.Bucket{
		{Name: "bucket1"},
		{Name: "bucket2"},
		{Name: "bucket3"},
	}

	migrateData, err := storages.BuildMigrateData(extractedLocalData, supabaseStorages)
	assert.NoError(t, err)
	assert.Equal(t, 4, len(migrateData))
}

func TestBuildMigrateItem(t *testing.T) {
	localStorages := []objects.Bucket{
		{Name: "bucket1"},
		{Name: "bucket2"},
	}

	supabaseStorages := []objects.Bucket{
		{Name: "bucket1"},
		{Name: "bucket3"},
	}

	migrateData, err := storages.BuildMigrateItem(supabaseStorages, localStorages)
	assert.NoError(t, err)
	fmt.Println(migrateData)
}

func TestMigrate(t *testing.T) {
	config := &raiden.Config{}
	stateChan := make(chan any)
	defer close(stateChan)

	migrateItems := []storages.MigrateItem{
		{
			Type:    "create",
			NewData: objects.Bucket{Name: "bucket1"},
		},
	}

	errors := storages.Migrate(config, migrateItems, stateChan, storages.ActionFunc)
	assert.Equal(t, 1, len(errors))
}
