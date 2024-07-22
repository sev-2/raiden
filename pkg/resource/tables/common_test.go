package tables_test

import (
	"testing"

	"github.com/sev-2/raiden/pkg/resource/tables"
	"github.com/sev-2/raiden/pkg/state"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/stretchr/testify/assert"
)

func TestGetNewCountData(t *testing.T) {
	supabaseTables := []objects.Table{
		{Name: "table1"},
		{Name: "table2"},
		{Name: "table3"},
	}

	extractResult := state.ExtractTableResult{
		Delete: []state.ExtractTableItem{
			{Table: objects.Table{Name: "table1"}},
			{Table: objects.Table{Name: "table2"}},
		},
	}

	count := tables.GetNewCountData(supabaseTables, extractResult)
	assert.Equal(t, 2, count)
}

func TestGetNewCountDataNoMatch(t *testing.T) {
	supabaseTables := []objects.Table{
		{Name: "table1"},
		{Name: "table2"},
	}

	extractResult := state.ExtractTableResult{
		Delete: []state.ExtractTableItem{
			{Table: objects.Table{Name: "table3"}},
			{Table: objects.Table{Name: "table4"}},
		},
	}

	count := tables.GetNewCountData(supabaseTables, extractResult)
	assert.Equal(t, 0, count)
}

func TestGetNewCountDataEmpty(t *testing.T) {
	supabaseTables := []objects.Table{}

	extractResult := state.ExtractTableResult{
		Delete: state.ExtractTableItems{},
	}

	count := tables.GetNewCountData(supabaseTables, extractResult)
	assert.Equal(t, 0, count)
}

func TestGetNewCountDataPartialMatch(t *testing.T) {
	supabaseTables := []objects.Table{
		{Name: "table1"},
		{Name: "table2"},
		{Name: "table3"},
	}

	extractResult := state.ExtractTableResult{
		Delete: []state.ExtractTableItem{
			{Table: objects.Table{Name: "table1"}},
			{Table: objects.Table{Name: "table4"}},
		},
	}

	count := tables.GetNewCountData(supabaseTables, extractResult)
	assert.Equal(t, 1, count)
}
