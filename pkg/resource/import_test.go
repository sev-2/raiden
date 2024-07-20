package resource_test

import (
	"os"
	"testing"

	"github.com/sev-2/raiden/pkg/generator"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/mock"
	"github.com/sev-2/raiden/pkg/resource"
	"github.com/sev-2/raiden/pkg/state"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestImport(t *testing.T) {
	flags := &resource.Flags{
		ProjectPath:   "test_project",
		AllowedSchema: "public",
	}
	config := loadConfig()
	resource.ImportLogger = logger.HcLog().Named("import")

	err := resource.Import(flags, config)
	assert.Error(t, err)

	mock := &mock.MockSupabase{Cfg: config}
	mock.Activate()
	defer mock.Deactivate()

	err0 := mock.MockFindProjectWithExpectedResponse(200, objects.Project{})
	assert.NoError(t, err0)

	err1 := mock.MockGetTablesWithExpectedResponse(200, []objects.Table{
		{Name: "some_table"},
		{Name: "other_table"},
	})
	assert.NoError(t, err1)

	err2 := mock.MockGetFunctionsWithExpectedResponse(200, []objects.Function{
		{Name: "some_function"},
		{Name: "other_function"},
	})
	assert.NoError(t, err2)

	err3 := mock.MockGetRolesWithExpectedResponse(200, []objects.Role{
		{Name: "some_role"},
		{Name: "other_role"},
	})
	assert.NoError(t, err3)

	err4 := mock.MockGetBucketsWithExpectedResponse(200, []objects.Bucket{
		{Name: "some_bucket"},
		{Name: "other_bucket"},
	})
	assert.NoError(t, err4)

	dir, errDir := os.MkdirTemp("", "import")
	assert.NoError(t, errDir)
	flags.ProjectPath = dir

	resource.RegisterModels(MockNewTable{})
	resource.RegisterModels(MockOtherTable{})
	resource.RegisterRole(MockNewRole{})

	errFinal := resource.Import(flags, config)
	assert.NoError(t, errFinal)

	assert.Equal(t, true, utils.IsFolderExists(dir+"/internal/roles"))
	assert.Equal(t, true, utils.IsFolderExists(dir+"/internal/storages"))

	defer os.RemoveAll(dir)
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
