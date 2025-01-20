package tables_test

import (
	"fmt"
	"testing"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/resource/tables"
	"github.com/sev-2/raiden/pkg/state"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/stretchr/testify/assert"
)

func TestBuildMigrateData(t *testing.T) {
	extractedLocalData := state.ExtractTableResult{
		New: state.ExtractTableItems{
			{Table: objects.Table{ID: 4, Name: "table4"}},
		},
		Existing: state.ExtractTableItems{
			{Table: objects.Table{ID: 2, Name: "table2"}},
			{Table: objects.Table{ID: 3, Name: "table3"}},
		},
		Delete: state.ExtractTableItems{
			{Table: objects.Table{ID: 1, Name: "table1"}},
		},
	}

	supabaseTables := []objects.Table{
		{ID: 1, Name: "table1"},
		{ID: 2, Name: "table2"},
		{ID: 3, Name: "table3"},
	}

	migrateData, err := tables.BuildMigrateData(extractedLocalData, supabaseTables, []string{"table1", "table2", "table3", "table4"})
	assert.NoError(t, err)
	assert.Equal(t, 2, len(migrateData))
}

func TestBuildMigrateDataAll(t *testing.T) {
	extractedLocalData := state.ExtractTableResult{
		New: state.ExtractTableItems{
			{Table: objects.Table{ID: 4, Name: "table4"}},
		},
		Existing: state.ExtractTableItems{
			{Table: objects.Table{ID: 2, Name: "table2"}},
			{Table: objects.Table{ID: 3, Name: "table3"}},
		},
		Delete: state.ExtractTableItems{
			{Table: objects.Table{ID: 1, Name: "table1"}},
		},
	}

	supabaseTables := []objects.Table{
		{ID: 1, Name: "table1"},
		{ID: 2, Name: "table2"},
		{ID: 3, Name: "table3"},
	}

	migrateData, err := tables.BuildMigrateData(extractedLocalData, supabaseTables, []string{})
	assert.NoError(t, err)
	assert.Equal(t, 2, len(migrateData))
}

func TestBuildMigrateDataNewTableNotAllowed(t *testing.T) {
	extractedLocalData := state.ExtractTableResult{
		New: state.ExtractTableItems{
			{Table: objects.Table{ID: 4, Name: "table4"}},
		},
		Existing: state.ExtractTableItems{
			{Table: objects.Table{ID: 2, Name: "table2"}},
			{Table: objects.Table{ID: 3, Name: "table3"}},
		},
		Delete: state.ExtractTableItems{
			{Table: objects.Table{ID: 1, Name: "table1"}},
		},
	}

	supabaseTables := []objects.Table{
		{ID: 1, Name: "table1"},
		{ID: 2, Name: "table2"},
		{ID: 3, Name: "table3"},
	}

	migrateData, err := tables.BuildMigrateData(extractedLocalData, supabaseTables, []string{"table1", "table2", "table3"})
	assert.Error(t, err)
	assert.Equal(t, 0, len(migrateData))
}

func TestBuildMigrateDataUpdateTableNotAllowed(t *testing.T) {
	extractedLocalData := state.ExtractTableResult{
		New: state.ExtractTableItems{
			{Table: objects.Table{ID: 4, Name: "table4"}},
		},
		Existing: state.ExtractTableItems{
			{Table: objects.Table{ID: 2, Name: "table2"}},
			{Table: objects.Table{ID: 3, Name: "table3"}},
		},
		Delete: state.ExtractTableItems{
			{Table: objects.Table{ID: 1, Name: "table1"}},
		},
	}

	supabaseTables := []objects.Table{
		{ID: 1, Name: "table1"},
		{ID: 2, Name: "table2"},
		{ID: 3, Name: "table3"},
	}

	migrateData, err := tables.BuildMigrateData(extractedLocalData, supabaseTables, []string{"table2", "table3", "table4"})
	assert.Error(t, err)
	assert.Equal(t, 0, len(migrateData))
}

func TestBuildMigrateDataDeleteTableNotAllowed(t *testing.T) {
	extractedLocalData := state.ExtractTableResult{
		New: state.ExtractTableItems{
			{Table: objects.Table{ID: 4, Name: "table4"}},
		},
		Existing: state.ExtractTableItems{
			{Table: objects.Table{ID: 2, Name: "table2"}},
			{Table: objects.Table{ID: 3, Name: "table3"}},
		},
		Delete: state.ExtractTableItems{
			{Table: objects.Table{ID: 1, Name: "table1"}},
		},
	}

	supabaseTables := []objects.Table{
		{ID: 1, Name: "table1"},
		{ID: 2, Name: "table2"},
		{ID: 3, Name: "table3"},
	}

	migrateData, err := tables.BuildMigrateData(extractedLocalData, supabaseTables, []string{"table1"})
	assert.Error(t, err)
	assert.Equal(t, 0, len(migrateData))
}

func TestBuildMigrateItem(t *testing.T) {
	localTables := []objects.Table{
		{ID: 1, Name: "table1"},
		{ID: 2, Name: "table2"},
	}

	supabaseTables := []objects.Table{
		{ID: 1, Name: "table1"},
		{ID: 3, Name: "table3"},
	}

	migrateData, err := tables.BuildMigrateItem(supabaseTables, localTables)
	assert.NoError(t, err)
	fmt.Println(migrateData)
}

func TestMigrate(t *testing.T) {
	config := &raiden.Config{}
	stateChan := make(chan any)
	defer close(stateChan)

	migrateItems := []tables.MigrateItem{
		{
			Type:    "create",
			NewData: objects.Table{ID: 1, Name: "table1"},
		},
	}

	errors := tables.Migrate(config, migrateItems, stateChan, tables.ActionFunc)
	assert.Equal(t, 1, len(errors))
}
