package types_test

import (
	"fmt"
	"testing"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/resource/types"
	"github.com/sev-2/raiden/pkg/state"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/stretchr/testify/assert"
)

func TestBuildMigrateData(t *testing.T) {
	extractedLocalData := state.ExtractTypeResult{
		New: []objects.Type{
			{Name: "type_1"},
		},
		Existing: []objects.Type{
			{Name: "type_2"},
			{Name: "type_3"},
		},
		Delete: []objects.Type{
			{Name: "type_4"},
		},
	}

	supabaseTypes := []objects.Type{
		{Name: "type_1"},
		{Name: "type_2"},
		{Name: "type_4"},
	}

	migrateData, err := types.BuildMigrateData(extractedLocalData, supabaseTypes)

	fmt.Printf("%+v", migrateData)
	assert.NoError(t, err)
	assert.Equal(t, 4, len(migrateData))
}

func TestBuildMigrateItem(t *testing.T) {
	localTypes := []objects.Type{
		{Name: "type_1"},
		{Name: "type_2"},
	}

	supabaseTypes := []objects.Type{
		{Name: "type_1"},
		{Name: "type_3"},
	}

	migrateData, err := types.BuildMigrateItem(supabaseTypes, localTypes)
	assert.NoError(t, err)
	fmt.Println(migrateData)
}

func TestMigrate(t *testing.T) {
	config := &raiden.Config{}
	stateChan := make(chan any)
	defer close(stateChan)

	migrateItems := []types.MigrateItem{
		{
			Type:    "create",
			NewData: objects.Type{Name: "type_1"},
		},
	}

	errors := types.Migrate(config, migrateItems, stateChan, types.ActionFunc)
	assert.Equal(t, 1, len(errors))
}
