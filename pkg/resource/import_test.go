package resource_test

import (
	"os"
	"testing"
	"time"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/generator"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/mock"
	"github.com/sev-2/raiden/pkg/postgres"
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

func (m MockOtherRole) CanBypassRls() bool {
	return true
}

func (r MockOtherRole) ValidUntil() *objects.SupabaseTime {
	return objects.NewSupabaseTime(time.Now())
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
	return "mock_get_vote_by"
}

func (m *MockGetVoteBy) GetReturnType() raiden.RpcReturnDataType {
	return "json"
}

func (m *MockGetVoteBy) GetRawDefinition() string {
	return "SELECT * FROM some_table;end $function$"
}

func TestImport(t *testing.T) {
	flags := &resource.Flags{
		ProjectPath: "test_project",
		DryRun:      true,
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
					Name: "mock_other_role",
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
					Name: "mock_get_vote_by",
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

	err0 := mock.MockGetBucketsWithExpectedResponse(200, []objects.Bucket{
		{Name: "some_bucket"},
		{Name: "other_bucket"},
	})
	assert.NoError(t, err0)

	err1 := mock.MockGetTablesWithExpectedResponse(200, []objects.Table{
		{ID: 1, Name: "some_table", Schema: "public"},
		{ID: 2, Name: "other_table", Schema: "public"},
		{ID: 3, Name: "other_table_again", Schema: "public"},
		{ID: 4, Name: "completely_new_table", Schema: "public"},
	})
	assert.NoError(t, err1)

	err2 := mock.MockGetFunctionsWithExpectedResponse(200, []objects.Function{
		{ID: 1, Schema: "public", Name: "some_function", Definition: "SELECT * FROM some_table;end $function$", ReturnType: "json"},
		{ID: 2, Schema: "public", Name: "other_function", Definition: "SELECT * FROM other_table", ReturnType: "json"},
		{ID: 3, Schema: "public", Name: "completely_new_function", Definition: "SELECT * FROM completely_new_table", ReturnType: "json"},
	})
	assert.NoError(t, err2)

	err3 := mock.MockGetRolesWithExpectedResponse(200, []objects.Role{
		{
			ID:              1,
			ConnectionLimit: 10,
			Name:            "mock_other_role",
			InheritRole:     true,
			CanLogin:        true,
			CanCreateDB:     true,
			CanCreateRole:   true,
			CanBypassRLS:    true,
		},
		{
			ID:              2,
			ConnectionLimit: 10,
			Name:            "other_role",
			InheritRole:     true,
			CanLogin:        true,
			CanCreateDB:     true,
			CanCreateRole:   true,
			CanBypassRLS:    true,
		},
	})
	assert.NoError(t, err3)

	errImport2 := resource.Import(flags, config)
	assert.NoError(t, errImport2)

	flags.DryRun = false
	errImport3 := resource.Import(flags, config)
	assert.NoError(t, errImport3)

	assert.Equal(t, true, utils.IsFolderExists(dir+"/internal/roles"))
	assert.Equal(t, true, utils.IsFolderExists(dir+"/internal/storages"))

	defer os.RemoveAll(dir)

	errReset := state.Save(&state.State{})
	assert.NoError(t, errReset)
}

func TestImportRpcOnly(t *testing.T) {
	flags := &resource.Flags{
		ProjectPath: "test_project",
		DryRun:      false,
		RpcOnly:     true,
	}

	config := loadConfig()
	resource.ImportLogger = logger.HcLog().Named("import")

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
		Rpc: []state.RpcState{
			{
				Function: objects.Function{
					Name: "mock_get_vote_by",
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
	resource.RegisterRpc(&MockGetVoteBy{})

	errSaveState := state.Save(&testState)
	assert.NoError(t, errSaveState)

	err1 := mock.MockGetTablesWithExpectedResponse(200, []objects.Table{
		{ID: 1, Name: "some_table", Schema: "public"},
		{ID: 2, Name: "other_table", Schema: "public"},
		{ID: 3, Name: "other_table_again", Schema: "public"},
		{ID: 4, Name: "completely_new_table", Schema: "public"},
	})
	assert.NoError(t, err1)

	err2 := mock.MockGetFunctionsWithExpectedResponse(200, []objects.Function{
		{ID: 1, Schema: "public", Name: "some_function", Definition: "SELECT * FROM some_table;end $function$", ReturnType: "json"},
		{ID: 2, Schema: "public", Name: "other_function", Definition: "SELECT * FROM other_table", ReturnType: "json"},
		{ID: 3, Schema: "public", Name: "completely_new_function", Definition: "SELECT * FROM completely_new_table", ReturnType: "json"},
	})
	assert.NoError(t, err2)

	err := resource.Import(flags, config)
	assert.NoError(t, err)

	assert.Equal(t, true, utils.IsFolderExists(dir+"/internal/rpc"))

	defer os.RemoveAll(dir)

	errReset := state.Save(&state.State{})
	assert.NoError(t, errReset)
}

type MockType struct {
	raiden.TypeBase
}

func (*MockType) Name() string {
	return "test_status"
}

func (*MockType) Format() string {
	return "test_status"
}

func (*MockType) Schema() string {
	return "public"
}

func TestImportModelsOnly(t *testing.T) {
	flags := &resource.Flags{
		ProjectPath: "test_project",
		DryRun:      false,
		ModelsOnly:  true,
	}

	config := loadConfig()
	config.AllowedTables = "test_local_table"
	resource.ImportLogger = logger.HcLog().Named("import")

	mock := &mock.MockSupabase{Cfg: config}
	mock.Activate()
	defer mock.Deactivate()

	dir, errDir := os.MkdirTemp("", "import")
	assert.NoError(t, errDir)
	flags.ProjectPath = dir

	resource.RegisterTypes(&MockType{})

	testState := state.State{
		Tables: []state.TableState{
			{
				Table: objects.Table{
					Name:        "test_local_table",
					Schema:      "public",
					PrimaryKeys: []objects.PrimaryKey{{Name: "id"}},
					Columns: []objects.Column{
						{Name: "id", DataType: "uuid"},
						{Name: "name", DataType: "text"},
						{Name: "status", DataType: string(postgres.UserDefined), Format: "test_status", IsNullable: true},
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
		},
		Types: []state.TypeState{
			{
				Type: objects.Type{
					ID:     1,
					Name:   "test_status",
					Format: "test_status",
					Schema: "public",
				},
				Name: "test_status",
			},
			{
				Type: objects.Type{
					ID:     2,
					Name:   "delete_test_status",
					Format: "delete_test_status",
					Schema: "public",
				},
				Name: "test_status",
			},
		},
	}

	resource.RegisterModels(MockNewTable{}, MockOtherTable{})

	errSaveState := state.Save(&testState)
	assert.NoError(t, errSaveState)

	err1 := mock.MockGetTablesWithExpectedResponse(200, []objects.Table{
		{ID: 1, Name: "test_local_table", Schema: "public"},
		{ID: 2, Name: "other_table", Schema: "public"},
		{ID: 3, Name: "other_table_again", Schema: "public"},
		{ID: 4, Name: "completely_new_table", Schema: "public"},
	})
	assert.NoError(t, err1)

	err2 := mock.MockGetTypesWithExpectedResponse(200, []objects.Type{
		{
			ID:     1,
			Name:   "test_status",
			Format: "test_status",
			Schema: "public",
		},
	})
	assert.NoError(t, err2)

	err := resource.Import(flags, config)
	assert.NoError(t, err)

	assert.Equal(t, true, utils.IsFolderExists(dir+"/internal/models"))

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

func TestImportUpdateStateOnly(t *testing.T) {
	flags := &resource.Flags{
		ProjectPath:     "test_project",
		DryRun:          false,
		UpdateStateOnly: true,
	}
	config := loadConfig()
	config.AllowedTables = "some_table"

	resource.ImportLogger = logger.HcLog().Named("import")

	mock := &mock.MockSupabase{Cfg: config}
	mock.Activate()
	defer mock.Deactivate()

	dir, errDir := os.MkdirTemp("", "import")
	assert.NoError(t, errDir)
	flags.ProjectPath = dir

	resource.RegisterModels(MockNewTable{}, MockOtherTable{})
	resource.RegisterRole(MockNewRole{}, MockOtherRole{})
	resource.RegisterStorages(MockOtherBucket{})
	resource.RegisterRpc(&MockGetVoteBy{})

	err0 := mock.MockGetBucketsWithExpectedResponse(200, []objects.Bucket{
		{Name: "some_bucket"},
		{Name: "other_bucket"},
	})
	assert.NoError(t, err0)

	err1 := mock.MockGetTablesWithExpectedResponse(200, []objects.Table{
		{ID: 1, Name: "some_table", Schema: "public"},
	})
	assert.NoError(t, err1)

	err2 := mock.MockGetFunctionsWithExpectedResponse(200, []objects.Function{
		{ID: 1, Schema: "public", Name: "some_function", Definition: "SELECT * FROM some_table;end $function$", ReturnType: "json"},
	})
	assert.NoError(t, err2)

	err3 := mock.MockGetRolesWithExpectedResponse(200, []objects.Role{
		{
			ID:              1,
			ConnectionLimit: 10,
			Name:            "mock_other_role",
			InheritRole:     true,
			CanLogin:        true,
			CanCreateDB:     true,
			CanCreateRole:   true,
			CanBypassRLS:    true,
		},
	})
	assert.NoError(t, err3)

	// reset state
	err := state.Save(&state.State{})
	assert.NoError(t, err)

	err = resource.Import(flags, config)
	assert.NoError(t, err)

	defer os.RemoveAll(dir)

	// validate state
	localState, err := state.Load()
	assert.NoError(t, err)
	assert.Len(t, localState.Tables, 1)
	assert.Len(t, localState.Roles, 1)
	assert.Len(t, localState.Rpc, 1)
	assert.Len(t, localState.Storage, 2)

}

func TestImportAllTableStateOnly(t *testing.T) {
	flags := &resource.Flags{
		ProjectPath:     "test_project",
		DryRun:          false,
		UpdateStateOnly: true,
	}
	config := loadConfig()
	config.AllowedTables = "some_table"

	resource.ImportLogger = logger.HcLog().Named("import")

	mock := &mock.MockSupabase{Cfg: config}
	mock.Activate()
	defer mock.Deactivate()

	dir, errDir := os.MkdirTemp("", "import")
	assert.NoError(t, errDir)
	flags.ProjectPath = dir

	resource.RegisterModels(MockNewTable{}, MockOtherTable{})
	resource.RegisterRole(MockNewRole{}, MockOtherRole{})
	resource.RegisterStorages(MockOtherBucket{})
	resource.RegisterRpc(&MockGetVoteBy{})

	err0 := mock.MockGetBucketsWithExpectedResponse(200, []objects.Bucket{
		{Name: "some_bucket"},
		{Name: "other_bucket"},
	})
	assert.NoError(t, err0)

	err1 := mock.MockGetTablesWithExpectedResponse(200, []objects.Table{
		{ID: 1, Name: "some_table", Schema: "public"},
	})
	assert.NoError(t, err1)

	err2 := mock.MockGetFunctionsWithExpectedResponse(200, []objects.Function{
		{ID: 1, Schema: "public", Name: "some_function", Definition: "SELECT * FROM some_table;end $function$", ReturnType: "json"},
	})
	assert.NoError(t, err2)

	err3 := mock.MockGetRolesWithExpectedResponse(200, []objects.Role{
		{
			ID:              1,
			ConnectionLimit: 10,
			Name:            "mock_other_role",
			InheritRole:     true,
			CanLogin:        true,
			CanCreateDB:     true,
			CanCreateRole:   true,
			CanBypassRLS:    true,
		},
	})
	assert.NoError(t, err3)

	// reset state
	err := state.Save(&state.State{})
	assert.NoError(t, err)

	err = resource.Import(flags, config)
	assert.NoError(t, err)

	defer os.RemoveAll(dir)

	// validate state
	localState, err := state.Load()
	assert.NoError(t, err)
	assert.Len(t, localState.Tables, 1)
	assert.Len(t, localState.Roles, 1)
	assert.Len(t, localState.Rpc, 1)
	assert.Len(t, localState.Storage, 2)

}

func TestImportSvc(t *testing.T) {
	flags := &resource.Flags{
		ProjectPath: "test_project",
		DryRun:      true,
	}
	config := loadConfig()
	config.Mode = raiden.SvcMode
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
					Name: "mock_other_role",
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
					Name: "mock_get_vote_by",
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

	errSaveState := state.Save(&testState)
	assert.NoError(t, errSaveState)

	err1 := mock.MockGetTablesWithExpectedResponse(200, []objects.Table{
		{ID: 1, Name: "some_table", Schema: "public"},
		{ID: 2, Name: "other_table", Schema: "public"},
		{ID: 3, Name: "other_table_again", Schema: "public"},
		{ID: 4, Name: "completely_new_table", Schema: "public"},
	})
	assert.NoError(t, err1)

	defer os.RemoveAll(dir)

	errReset := state.Save(&state.State{})
	assert.NoError(t, errReset)
}
