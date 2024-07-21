package resource_test

import (
	"os"
	"testing"
	"time"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/generator"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/mock"
	"github.com/sev-2/raiden/pkg/resource"
	"github.com/sev-2/raiden/pkg/state"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/sev-2/raiden/pkg/utils"
	"github.com/stretchr/testify/assert"
)

type MockOtherRole struct {
	raiden.Role
}

func (m MockOtherRole) Name() string {
	return "mock_other_role"
}

func (m MockOtherRole) CanLogin() bool {
	return true
}

func (m MockOtherRole) CanCreateDB() bool {
	return true
}

func (m MockOtherRole) CanCreateRole() bool {
	return true
}

func (m MockOtherRole) InheritRole() bool {
	return true
}

func (m MockOtherRole) ConnectionLimit() int {
	return 10
}

func (m MockOtherRole) ValidUntil() *objects.SupabaseTime {
	return objects.NewSupabaseTime(time.Now().AddDate(0, 1, 0))
}

func (m MockOtherRole) CanBypassRls() bool {
	return true
}

type MockOtherBucket struct {
	raiden.BucketBase
}

func (b MockOtherBucket) Name() string {
	return "test_other_bucket"
}

func (b MockOtherBucket) Public() bool {
	return false
}

func (b MockOtherBucket) AllowedMimeTypes() []string {
	return nil
}

func (b MockOtherBucket) FileSizeLimit() int {
	return 0
}

func (b MockOtherBucket) AvifAutoDetection() bool {
	return false
}

type MockGetVoteByParams struct{}

type MockGetVoteByResult any

type MockGetVoteBy struct {
	raiden.RpcBase
	Params *MockGetVoteByParams `json:"-"`
	Return MockGetVoteByResult  `json:"-"`
}

func (m *MockGetVoteBy) Name() string {
	return "some_function"
}

func TestImport(t *testing.T) {
	flags := &resource.Flags{
		ProjectPath: "test_project",
		DryRun:      false,
	}
	config := loadConfig()
	resource.ImportLogger = logger.HcLog().Named("import")

	err := resource.Import(flags, config)
	assert.Error(t, err)

	mock := &mock.MockSupabase{Cfg: config}
	mock.Activate()
	defer mock.Deactivate()

	dir, errDir := os.MkdirTemp("", "import")
	assert.NoError(t, errDir)
	flags.ProjectPath = dir

	testState := state.State{
		Tables: []state.TableState{
			{
				Table: objects.Table{
					Name:        "test_local_table",
					PrimaryKeys: []objects.PrimaryKey{{Name: "id"}},
					Columns: []objects.Column{
						{Name: "id", DataType: "uuid"},
						{Name: "name", DataType: "text"},
					},
					Relationships: []objects.TablesRelationship{
						{
							ConstraintName:    "test_local_constraint",
							SourceSchema:      "public",
							SourceTableName:   "test_local_table",
							SourceColumnName:  "id",
							TargetTableSchema: "public",
							TargetTableName:   "test_table",
							TargetColumnName:  "id",
						},
					},
				},
			},
			{
				Table: objects.Table{
					Name:        "test_table",
					PrimaryKeys: []objects.PrimaryKey{{Name: "id"}},
					Columns: []objects.Column{
						{Name: "id", DataType: "uuid"},
					},
					Relationships: []objects.TablesRelationship{
						{
							ConstraintName:    "test_constraint",
							SourceSchema:      "public",
							SourceTableName:   "test_table",
							SourceColumnName:  "id",
							TargetTableSchema: "public",
							TargetTableName:   "test_other_table",
							TargetColumnName:  "id",
						},
					},
				},
			},
		},
		Storage: []state.StorageState{
			{
				Storage: objects.Bucket{
					Name:   "test_bucket_policy",
					Public: true,
				},
			},
			{
				Storage: objects.Bucket{
					Name:   "test_other_bucket",
					Public: true,
				},
			},
		},
		Roles: []state.RoleState{
			{
				Role: objects.Role{
					Name: "test_role_local",
				},
			},
			{
				Role: objects.Role{
					Name: "test_other_role",
				},
			},
		},
		Rpc: []state.RpcState{
			{
				Function: objects.Function{
					Name: "some_function",
				},
			},
			{
				Function: objects.Function{
					Name: "test_other_rpc",
				},
			},
		},
	}

	resource.RegisterModels(MockNewTable{}, MockOtherTable{})
	resource.RegisterRole(MockNewRole{}, MockOtherRole{})
	resource.RegisterStorages(MockOtherBucket{})
	resource.RegisterRpc(&MockGetVoteBy{})

	errSaveState := state.Save(&testState)
	assert.NoError(t, errSaveState)

	// err0 := mock.MockFindProjectWithExpectedResponse(200, objects.Project{})
	// assert.NoError(t, err0)

	err0 := mock.MockGetBucketsWithExpectedResponse(200, []objects.Bucket{
		{Name: "some_bucket"},
		{Name: "other_bucket"},
	})
	assert.NoError(t, err0)

	err1 := mock.MockGetTablesWithExpectedResponse(200, []objects.Table{
		{Name: "some_table", Schema: "public"},
		{Name: "other_table", Schema: "private"},
		{Name: "other_table_again", Schema: "public"},
	})
	assert.NoError(t, err1)

	err2 := mock.MockGetFunctionsWithExpectedResponse(200, []objects.Function{
		{Name: "some_function"},
		{Name: "other_function"},
	})
	assert.NoError(t, err2)

	errImport1 := resource.Import(flags, config)
	assert.NoError(t, errImport1)

	err3 := mock.MockGetRolesWithExpectedResponse(200, []objects.Role{
		{Name: "some_role"},
		{Name: "other_role"},
	})
	assert.NoError(t, err3)

	errImport3 := resource.Import(flags, config)
	assert.NoError(t, errImport3)

	errImport4 := resource.Import(flags, config)
	assert.NoError(t, errImport4)

	assert.Equal(t, true, utils.IsFolderExists(dir+"/internal/roles"))
	assert.Equal(t, false, utils.IsFolderExists(dir+"/internal/models"))
	assert.Equal(t, true, utils.IsFolderExists(dir+"/internal/storages"))

	defer os.RemoveAll(dir)

	errReset := state.Save(&state.State{})
	assert.NoError(t, errReset)
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
