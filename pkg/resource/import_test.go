package resource_test

import (
	"testing"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/generator"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/resource"
	"github.com/sev-2/raiden/pkg/state"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/stretchr/testify/assert"
)

func TestImport(t *testing.T) {
	flags := &resource.Flags{
		ProjectPath:   "test_project",
		AllowedSchema: "public",
	}
	config := &raiden.Config{}
	resource.ImportLogger = logger.HcLog().Named("import")

	err := resource.Import(flags, config)
	assert.Error(t, err)
}

func TestUpdateLocalStateFromImport(t *testing.T) {
	localState := &state.LocalState{}
	stateChan := make(chan any)
	done := resource.UpdateLocalStateFromImport(localState, stateChan)

	close(stateChan)
	err := <-done
	assert.NoError(t, err)
}

func TestPrintImportReport(t *testing.T) {
	report := resource.ImportReport{
		Table:   1,
		Role:    2,
		Rpc:     3,
		Storage: 4,
	}

	resource.PrintImportReport(report, false)
	resource.PrintImportReport(report, true)

	report = resource.ImportReport{}
	resource.PrintImportReport(report, false)
	resource.PrintImportReport(report, true)
}

func TestFindImportResource(t *testing.T) {
	data := []objects.Role{
		{Name: "role1"},
		{Name: "role2"},
	}
	findFunc := func(item objects.Role, input generator.GenerateInput) bool {
		if i, ok := input.BindData.(generator.GenerateRoleData); ok {
			return i.Name == item.Name
		}
		return false
	}

	input := generator.GenerateInput{
		BindData: generator.GenerateRoleData{Name: "role1"},
	}

	item, found := resource.FindImportResource(data, input, findFunc)
	assert.True(t, found)
	assert.Equal(t, "role1", item.Name)

	input = generator.GenerateInput{
		BindData: generator.GenerateRoleData{Name: "role3"},
	}

	item, found = resource.FindImportResource(data, input, findFunc)
	assert.False(t, found)
}
