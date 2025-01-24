package tables_test

import (
	"fmt"
	"testing"

	"github.com/jarcoal/httpmock"
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

	migrateData, err := tables.BuildMigrateData(extractedLocalData, supabaseTables, []string{"table1", "table3", "table4"})
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

func TestActionFunc_Create(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// Mock endpoint for pgmeta.CreateTable
	httpmock.RegisterResponder("POST", "https://pgmeta.test.com/query",
		httpmock.NewJsonResponderOrPanic(200, []objects.Table{
			{
				Name:    "test_table",
				Schema:  "public",
				Comment: "Test table creation",
			},
		}))

	// Mock endpoint for supabase.CreateTable
	httpmock.RegisterResponder("POST", "https://api.test.com/query",
		httpmock.NewJsonResponderOrPanic(200, []objects.Table{
			{
				Name:    "test_table",
				Schema:  "public",
				Comment: "Test table creation",
			},
		}))

	// Test for `pgmeta.CreateTable` when cfg.Mode == SvcMode
	cfg := &raiden.Config{
		Mode:           raiden.SvcMode,
		PgMetaUrl:      "https://pgmeta.test.com",
		SupabaseApiUrl: "https://api.test.com",
	}

	param := objects.Table{
		Name:   "test_table",
		Schema: "public",
	}

	response, err := tables.ActionFunc.CreateFunc(cfg, param)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if response.Name != "test_table" || response.Schema != "public" {
		t.Errorf("unexpected response: %+v", response)
	}

	// Test for `supabase.CreateTable` when cfg.Mode != SvcMode
	cfg.Mode = raiden.BffMode

	response, err = tables.ActionFunc.CreateFunc(cfg, param)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if response.Name != "test_table" || response.Schema != "public" {
		t.Errorf("unexpected response: %+v", response)
	}

	// Debug mock call counts
	t.Log(httpmock.GetCallCountInfo())
}

func TestActionFunc_Update(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// Mock endpoint for pgmeta.CreateTable
	httpmock.RegisterResponder("POST", "https://pgmeta.test.com/query",
		httpmock.NewJsonResponderOrPanic(200, []objects.Table{
			{
				Name:    "test_table",
				Schema:  "public",
				Comment: "Test table creation",
			},
		}))

	// Mock endpoint for supabase.CreateTable
	httpmock.RegisterResponder("POST", "https://api.test.com/query",
		httpmock.NewJsonResponderOrPanic(200, []objects.Table{
			{
				Name:    "test_table",
				Schema:  "public",
				Comment: "Test table creation",
			},
		}))

	// Test for `pgmeta.CreateTable` when cfg.Mode == SvcMode
	cfg := &raiden.Config{
		Mode:           raiden.SvcMode,
		PgMetaUrl:      "https://pgmeta.test.com",
		SupabaseApiUrl: "https://api.test.com",
	}

	param := objects.Table{
		Name:   "test_table",
		Schema: "public",
	}

	items := objects.UpdateTableParam{}

	err := tables.ActionFunc.UpdateFunc(cfg, param, items)
	assert.NoError(t, err)

	// Test for `supabase.CreateTable` when cfg.Mode != SvcMode
	cfg.Mode = raiden.BffMode

	err = tables.ActionFunc.UpdateFunc(cfg, param, items)
	assert.NoError(t, err)

	// Debug mock call counts
	t.Log(httpmock.GetCallCountInfo())
}

func TestActionFunc_Delete(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// Mock endpoint for pgmeta.CreateTable
	httpmock.RegisterResponder("POST", "https://pgmeta.test.com/query",
		httpmock.NewJsonResponderOrPanic(200, map[string]any{
			"message": "success",
		}),
	)

	// Mock endpoint for supabase.CreateTable
	httpmock.RegisterResponder("POST", "https://api.test.com/query",
		httpmock.NewJsonResponderOrPanic(200, map[string]any{
			"message": "success",
		}),
	)

	// Test for `pgmeta.CreateTable` when cfg.Mode == SvcMode
	cfg := &raiden.Config{
		Mode:           raiden.SvcMode,
		PgMetaUrl:      "https://pgmeta.test.com",
		SupabaseApiUrl: "https://api.test.com",
	}

	param := objects.Table{
		Name:   "test_table",
		Schema: "public",
	}

	err := tables.ActionFunc.DeleteFunc(cfg, param)
	assert.NoError(t, err)

	// Test for `supabase.CreateTable` when cfg.Mode != SvcMode
	cfg.Mode = raiden.BffMode

	err = tables.ActionFunc.DeleteFunc(cfg, param)
	assert.NoError(t, err)

	// Debug mock call counts
	t.Log(httpmock.GetCallCountInfo())
}
