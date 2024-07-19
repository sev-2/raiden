package supabase_test

import (
	"errors"
	"testing"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/mock"
	"github.com/sev-2/raiden/pkg/supabase"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/stretchr/testify/assert"
)

func loadCloudConfig() *raiden.Config {
	return &raiden.Config{
		DeploymentTarget:    raiden.DeploymentTargetCloud,
		ProjectId:           "test-project-id",
		ProjectName:         "My Great Project",
		SupabaseApiBasePath: "/v1",
		SupabaseApiUrl:      "http://supabase.cloud.com",
	}
}

func loadSelfHostedConfig() *raiden.Config {
	return &raiden.Config{
		DeploymentTarget:    raiden.DeploymentTargetSelfHosted,
		ProjectId:           "test-project-local-id",
		SupabaseApiBasePath: "/v1",
		SupabaseApiUrl:      "http://supabase.local.com",
	}
}

func TestGetPolicyName(t *testing.T) {

	expectedPolicyName := "enable test-policy access for some-resource some-action"
	assert.Equal(t, expectedPolicyName, supabase.GetPolicyName("test-policy", "some-resource", "some-action"))
}

func TestFindProject_Cloud(t *testing.T) {
	cfg := loadCloudConfig()

	_, err0 := supabase.GetTables(cfg, []string{"test-schema"})
	assert.Error(t, err0)

	project := objects.Project{
		Id:   "test-project-id",
		Name: "My Great Project",
	}

	mock := mock.MockSupabase{Cfg: cfg}
	mock.Activate()
	defer mock.Deactivate()

	err := mock.MockFindProjectWithExpectedResponse(200, project)
	assert.NoError(t, err)

	project, err1 := supabase.FindProject(cfg)
	assert.NoError(t, err1)
	assert.Equal(t, cfg.ProjectId, project.Id)
}

func TestFindProject_SelfHosted(t *testing.T) {
	cfg := loadSelfHostedConfig()

	expectedError := errors.New("FindProject not implemented for self hosted")
	project, err := supabase.FindProject(cfg)
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Equal(t, objects.Project{}, project)
}

func TestGetTables_Cloud(t *testing.T) {
	cfg := loadCloudConfig()

	_, err0 := supabase.GetTables(cfg, []string{"test-schema"})
	assert.Error(t, err0)

	remoteTables := []objects.Table{
		{
			ID:   1,
			Name: "some-table",
		},
		{
			ID:   2,
			Name: "another-table",
		},
	}

	mock := mock.MockSupabase{Cfg: cfg}
	mock.Activate()
	defer mock.Deactivate()

	err := mock.MockGetTablesWithExpectedResponse(200, remoteTables)
	assert.NoError(t, err)

	tables, err1 := supabase.GetTables(cfg, []string{"test-schema"})
	assert.NoError(t, err1)
	assert.Equal(t, len(remoteTables), len(tables))
}

func TestGetTables_SelfHosted(t *testing.T) {
	cfg := loadCloudConfig()

	_, err0 := supabase.GetTables(cfg, []string{"test-schema"})
	assert.Error(t, err0)

	remoteTables := []objects.Table{
		{
			ID:   1,
			Name: "some-table",
		},
		{
			ID:   2,
			Name: "another-table",
		},
	}

	mock := mock.MockSupabase{Cfg: cfg}
	mock.Activate()
	defer mock.Deactivate()

	err := mock.MockGetTablesWithExpectedResponse(200, remoteTables)
	assert.NoError(t, err)

	tables, err1 := supabase.GetTables(cfg, []string{"test-schema"})
	assert.NoError(t, err1)
	assert.Equal(t, len(remoteTables), len(tables))
}

func TestCreateTable_Cloud(t *testing.T) {
	cfg := loadCloudConfig()

	_, err := supabase.CreateTable(cfg, objects.Table{})
	assert.Error(t, err)

	localTable := objects.Table{
		Name: "some-table",
	}

	mock := mock.MockSupabase{Cfg: cfg}
	mock.Activate()
	defer mock.Deactivate()

	err0 := mock.MockGetTableByNameWithExpectedResponse(200, localTable)
	assert.NoError(t, err0)

	err1 := mock.MockCreateTableWithExpectedResponse(200, localTable)
	assert.NoError(t, err1)

	createdTable, err2 := supabase.CreateTable(cfg, localTable)
	assert.NoError(t, err2)
	assert.Equal(t, localTable.Name, createdTable.Name)
}

func TestCreateTable_SelfHosted(t *testing.T) {
	cfg := loadSelfHostedConfig()

	_, err := supabase.CreateTable(cfg, objects.Table{})
	assert.Error(t, err)

	localTable := objects.Table{
		Name: "some-table",
	}

	mock := mock.MockSupabase{Cfg: cfg}
	mock.Activate()
	defer mock.Deactivate()

	err0 := mock.MockGetTableByNameWithExpectedResponse(200, localTable)
	assert.NoError(t, err0)

	err1 := mock.MockCreateTableWithExpectedResponse(200, localTable)
	assert.NoError(t, err1)

	createdTable, err2 := supabase.CreateTable(cfg, localTable)
	assert.NoError(t, err2)
	assert.Equal(t, localTable.Name, createdTable.Name)
}

func TestUpdateTable_Cloud(t *testing.T) {
	cfg := loadCloudConfig()

	err := supabase.UpdateTable(cfg, objects.Table{}, objects.UpdateTableParam{})
	assert.Error(t, err)

	localTable := objects.Table{
		Name: "some-table",
		Columns: []objects.Column{
			{
				Name: "some-column",
			},
			{
				Name: "another-column",
			},
		},
		Relationships: []objects.TablesRelationship{
			{
				ConstraintName:    "some-constraint",
				SourceSchema:      "some-schema",
				SourceColumnName:  "some-column",
				TargetTableSchema: "other-schema",
			},
		},
	}

	updateParam := objects.UpdateTableParam{
		OldData: localTable,
		ChangeColumnItems: []objects.UpdateColumnItem{
			{
				Name: "some-column",
				UpdateItems: []objects.UpdateColumnType{
					objects.UpdateColumnName,
				},
			},
		},
		ChangeItems: []objects.UpdateTableType{
			objects.UpdateTableName,
		},
		ChangeRelationItems: []objects.UpdateRelationItem{
			{
				Data: objects.TablesRelationship{
					ConstraintName:    "",
					SourceSchema:      "some-schema",
					SourceColumnName:  "some-column",
					TargetTableSchema: "other-schema",
				},
				Type: objects.UpdateRelationCreate,
			},
		},
		ForceCreateRelation: true,
	}

	updateParam1 := objects.UpdateTableParam{
		OldData: localTable,
		ChangeColumnItems: []objects.UpdateColumnItem{
			{
				Name: "some-column",
				UpdateItems: []objects.UpdateColumnType{
					objects.UpdateColumnName,
				},
			},
		},
		ChangeItems: []objects.UpdateTableType{
			objects.UpdateTableName,
		},
		ChangeRelationItems: []objects.UpdateRelationItem{
			{
				Data: objects.TablesRelationship{
					ConstraintName:    "some-constraint",
					SourceSchema:      "some-schema",
					SourceColumnName:  "some-column",
					TargetTableSchema: "other-schema",
				},
				Type: objects.UpdateRelationCreate,
			},
		},
		ForceCreateRelation: false,
	}

	updateParam1NoConstraint := objects.UpdateTableParam{
		OldData: localTable,
		ChangeColumnItems: []objects.UpdateColumnItem{
			{
				Name: "some-column",
				UpdateItems: []objects.UpdateColumnType{
					objects.UpdateColumnName,
				},
			},
		},
		ChangeItems: []objects.UpdateTableType{
			objects.UpdateTableName,
		},
		ChangeRelationItems: []objects.UpdateRelationItem{
			{
				Data: objects.TablesRelationship{
					ConstraintName:    "",
					SourceSchema:      "some-schema",
					SourceColumnName:  "some-column",
					TargetTableSchema: "other-schema",
				},
				Type: objects.UpdateRelationCreate,
			},
		},
		ForceCreateRelation: false,
	}

	updateParam2 := objects.UpdateTableParam{
		OldData: objects.Table{
			Name: "some-table",
			Columns: []objects.Column{
				{
					Name: "old-column",
				},
				{
					Name: "another-old-column",
				},
			},
		},
		ChangeColumnItems: []objects.UpdateColumnItem{
			{
				Name: "some-column",
				UpdateItems: []objects.UpdateColumnType{
					objects.UpdateColumnDelete,
				},
			},
		},
		ChangeItems: []objects.UpdateTableType{
			objects.UpdateTableName,
		},
		ChangeRelationItems: []objects.UpdateRelationItem{
			{
				Data: objects.TablesRelationship{
					ConstraintName:    "some-constraint",
					SourceSchema:      "some-schema",
					SourceColumnName:  "some-column",
					TargetTableSchema: "other-schema",
				},
				Type: objects.UpdateRelationUpdate,
			},
		},
		ForceCreateRelation: false,
	}

	updateParam2NoConstraint := objects.UpdateTableParam{
		OldData: objects.Table{
			Name: "some-table",
			Columns: []objects.Column{
				{
					Name: "old-column",
				},
				{
					Name: "another-old-column",
				},
			},
		},
		ChangeColumnItems: []objects.UpdateColumnItem{
			{
				Name: "some-column",
				UpdateItems: []objects.UpdateColumnType{
					objects.UpdateColumnDelete,
				},
			},
		},
		ChangeItems: []objects.UpdateTableType{
			objects.UpdateTableName,
		},
		ChangeRelationItems: []objects.UpdateRelationItem{
			{
				Data: objects.TablesRelationship{
					ConstraintName:    "",
					SourceSchema:      "some-schema",
					SourceColumnName:  "some-column",
					TargetTableSchema: "other-schema",
				},
				Type: objects.UpdateRelationUpdate,
			},
		},
		ForceCreateRelation: false,
	}

	updateParam3 := objects.UpdateTableParam{
		OldData: localTable,
		ChangeColumnItems: []objects.UpdateColumnItem{
			{
				Name: "some-column",
				UpdateItems: []objects.UpdateColumnType{
					objects.UpdateColumnNew,
				},
			},
		},
		ChangeItems: []objects.UpdateTableType{
			objects.UpdateTableName,
		},
		ChangeRelationItems: []objects.UpdateRelationItem{
			{
				Data: objects.TablesRelationship{
					ConstraintName:    "",
					SourceSchema:      "some-schema",
					SourceColumnName:  "some-column",
					TargetTableSchema: "other-schema",
				},
				Type: objects.UpdateRelationDelete,
			},
		},
		ForceCreateRelation: false,
	}

	mock := mock.MockSupabase{Cfg: cfg}
	mock.Activate()
	defer mock.Deactivate()

	err0 := mock.MockUpdateTableWithExpectedResponse(200)
	assert.NoError(t, err0)

	err1 := supabase.UpdateTable(cfg, localTable, updateParam)
	assert.NoError(t, err1)

	err2 := supabase.UpdateTable(cfg, localTable, updateParam1)
	assert.NoError(t, err2)

	err2NoC := supabase.UpdateTable(cfg, localTable, updateParam1NoConstraint)
	assert.NoError(t, err2NoC)

	err3 := supabase.UpdateTable(cfg, localTable, updateParam2)
	assert.NoError(t, err3)

	err3NoC := supabase.UpdateTable(cfg, localTable, updateParam2NoConstraint)
	assert.NoError(t, err3NoC)

	err4 := supabase.UpdateTable(cfg, localTable, updateParam3)
	assert.NoError(t, err4)
}

func TestUpdateTable_SelfHosted(t *testing.T) {
	cfg := loadSelfHostedConfig()

	err := supabase.UpdateTable(cfg, objects.Table{}, objects.UpdateTableParam{})
	assert.Error(t, err)

	localTable := objects.Table{
		Name: "some-table",
		Columns: []objects.Column{
			{
				Name: "some-column",
			},
			{
				Name: "another-column",
			},
		},
		Relationships: []objects.TablesRelationship{
			{
				ConstraintName:    "some-constraint",
				SourceSchema:      "some-schema",
				SourceColumnName:  "some-column",
				TargetTableSchema: "other-schema",
			},
		},
	}

	updateParam := objects.UpdateTableParam{
		OldData: localTable,
		ChangeColumnItems: []objects.UpdateColumnItem{
			{
				Name: "some-column",
				UpdateItems: []objects.UpdateColumnType{
					objects.UpdateColumnName,
				},
			},
		},
		ChangeItems: []objects.UpdateTableType{
			objects.UpdateTableName,
		},
		ChangeRelationItems: []objects.UpdateRelationItem{
			{
				Data: objects.TablesRelationship{
					ConstraintName:    "",
					SourceSchema:      "some-schema",
					SourceColumnName:  "some-column",
					TargetTableSchema: "other-schema",
				},
				Type: objects.UpdateRelationCreate,
			},
		},
		ForceCreateRelation: true,
	}

	updateParam1 := objects.UpdateTableParam{
		OldData: localTable,
		ChangeColumnItems: []objects.UpdateColumnItem{
			{
				Name: "some-column",
				UpdateItems: []objects.UpdateColumnType{
					objects.UpdateColumnName,
				},
			},
		},
		ChangeItems: []objects.UpdateTableType{
			objects.UpdateTableName,
		},
		ChangeRelationItems: []objects.UpdateRelationItem{
			{
				Data: objects.TablesRelationship{
					ConstraintName:    "some-constraint",
					SourceSchema:      "some-schema",
					SourceColumnName:  "some-column",
					TargetTableSchema: "other-schema",
				},
				Type: objects.UpdateRelationCreate,
			},
		},
		ForceCreateRelation: false,
	}

	updateParam1NoConstraint := objects.UpdateTableParam{
		OldData: localTable,
		ChangeColumnItems: []objects.UpdateColumnItem{
			{
				Name: "some-column",
				UpdateItems: []objects.UpdateColumnType{
					objects.UpdateColumnName,
				},
			},
		},
		ChangeItems: []objects.UpdateTableType{
			objects.UpdateTableName,
		},
		ChangeRelationItems: []objects.UpdateRelationItem{
			{
				Data: objects.TablesRelationship{
					ConstraintName:    "",
					SourceSchema:      "some-schema",
					SourceColumnName:  "some-column",
					TargetTableSchema: "other-schema",
				},
				Type: objects.UpdateRelationCreate,
			},
		},
		ForceCreateRelation: false,
	}

	updateParam2 := objects.UpdateTableParam{
		OldData: objects.Table{
			Name: "some-table",
			Columns: []objects.Column{
				{
					Name: "old-column",
				},
				{
					Name: "another-old-column",
				},
			},
		},
		ChangeColumnItems: []objects.UpdateColumnItem{
			{
				Name: "some-column",
				UpdateItems: []objects.UpdateColumnType{
					objects.UpdateColumnDelete,
				},
			},
		},
		ChangeItems: []objects.UpdateTableType{
			objects.UpdateTableName,
		},
		ChangeRelationItems: []objects.UpdateRelationItem{
			{
				Data: objects.TablesRelationship{
					ConstraintName:    "some-constraint",
					SourceSchema:      "some-schema",
					SourceColumnName:  "some-column",
					TargetTableSchema: "other-schema",
				},
				Type: objects.UpdateRelationUpdate,
			},
		},
		ForceCreateRelation: false,
	}

	updateParam2NoConstraint := objects.UpdateTableParam{
		OldData: objects.Table{
			Name: "some-table",
			Columns: []objects.Column{
				{
					Name: "old-column",
				},
				{
					Name: "another-old-column",
				},
			},
		},
		ChangeColumnItems: []objects.UpdateColumnItem{
			{
				Name: "some-column",
				UpdateItems: []objects.UpdateColumnType{
					objects.UpdateColumnDelete,
				},
			},
		},
		ChangeItems: []objects.UpdateTableType{
			objects.UpdateTableName,
		},
		ChangeRelationItems: []objects.UpdateRelationItem{
			{
				Data: objects.TablesRelationship{
					ConstraintName:    "",
					SourceSchema:      "some-schema",
					SourceColumnName:  "some-column",
					TargetTableSchema: "other-schema",
				},
				Type: objects.UpdateRelationUpdate,
			},
		},
		ForceCreateRelation: false,
	}

	updateParam3 := objects.UpdateTableParam{
		OldData: localTable,
		ChangeColumnItems: []objects.UpdateColumnItem{
			{
				Name: "some-column",
				UpdateItems: []objects.UpdateColumnType{
					objects.UpdateColumnNew,
				},
			},
		},
		ChangeItems: []objects.UpdateTableType{
			objects.UpdateTableName,
		},
		ChangeRelationItems: []objects.UpdateRelationItem{
			{
				Data: objects.TablesRelationship{
					ConstraintName:    "",
					SourceSchema:      "some-schema",
					SourceColumnName:  "some-column",
					TargetTableSchema: "other-schema",
				},
				Type: objects.UpdateRelationDelete,
			},
		},
		ForceCreateRelation: false,
	}

	mock := mock.MockSupabase{Cfg: cfg}
	mock.Activate()
	defer mock.Deactivate()

	err0 := mock.MockUpdateTableWithExpectedResponse(200)
	assert.NoError(t, err0)

	err1 := supabase.UpdateTable(cfg, localTable, updateParam)
	assert.NoError(t, err1)

	err2 := supabase.UpdateTable(cfg, localTable, updateParam1)
	assert.NoError(t, err2)

	err2NoC := supabase.UpdateTable(cfg, localTable, updateParam1NoConstraint)
	assert.NoError(t, err2NoC)

	err3 := supabase.UpdateTable(cfg, localTable, updateParam2)
	assert.NoError(t, err3)

	err3NoC := supabase.UpdateTable(cfg, localTable, updateParam2NoConstraint)
	assert.NoError(t, err3NoC)

	err4 := supabase.UpdateTable(cfg, localTable, updateParam3)
	assert.NoError(t, err4)
}

func TestDeleteTable_Cloud(t *testing.T) {
	cfg := loadCloudConfig()

	err := supabase.DeleteTable(cfg, objects.Table{}, true)
	assert.Error(t, err)

	localTable := objects.Table{
		Name: "some-table",
	}

	mock := mock.MockSupabase{Cfg: cfg}
	mock.Activate()
	defer mock.Deactivate()

	err0 := mock.MockDeleteTableWithExpectedResponse(200)
	assert.NoError(t, err0)

	err1 := supabase.DeleteTable(cfg, localTable, true)
	assert.NoError(t, err1)
}

func TestDeleteTable_SelfHosted(t *testing.T) {
	cfg := loadSelfHostedConfig()

	err := supabase.DeleteTable(cfg, objects.Table{}, true)
	assert.Error(t, err)

	localTable := objects.Table{
		Name: "some-table",
	}

	mock := mock.MockSupabase{Cfg: cfg}
	mock.Activate()
	defer mock.Deactivate()

	err0 := mock.MockDeleteTableWithExpectedResponse(200)
	assert.NoError(t, err0)

	err1 := supabase.DeleteTable(cfg, localTable, true)
	assert.NoError(t, err1)
}

func TestGetRoles_Cloud(t *testing.T) {
	cfg := loadCloudConfig()

	_, err := supabase.GetRoles(cfg)
	assert.Error(t, err)

	remoteRoles := []objects.Role{
		{
			ID:   1,
			Name: "some-role",
		},
		{
			ID:   2,
			Name: "another-role",
		},
	}

	mock := mock.MockSupabase{Cfg: cfg}
	mock.Activate()
	defer mock.Deactivate()

	err0 := mock.MockGetRolesWithExpectedResponse(200, remoteRoles)
	assert.NoError(t, err0)

	roles, err1 := supabase.GetRoles(cfg)
	assert.NoError(t, err1)
	assert.Equal(t, len(remoteRoles), len(roles))
}

func TestGetRoles_SelfHosted(t *testing.T) {
	cfg := loadSelfHostedConfig()

	_, err := supabase.GetRoles(cfg)
	assert.Error(t, err)

	remoteRoles := []objects.Role{
		{
			ID:   1,
			Name: "some-role",
		},
		{
			ID:   2,
			Name: "another-role",
		},
	}

	mock := mock.MockSupabase{Cfg: cfg}
	mock.Activate()
	defer mock.Deactivate()

	err0 := mock.MockGetRolesWithExpectedResponse(200, remoteRoles)
	assert.NoError(t, err0)

	roles, err1 := supabase.GetRoles(cfg)
	assert.NoError(t, err1)
	assert.Equal(t, len(remoteRoles), len(roles))
}

func TestCreateRole_Cloud(t *testing.T) {
	cfg := loadCloudConfig()

	_, err := supabase.CreateRole(cfg, objects.Role{})
	assert.Error(t, err)
}

func TestCreateRole_SelfHosted(t *testing.T) {
	cfg := loadSelfHostedConfig()

	_, err := supabase.CreateRole(cfg, objects.Role{})
	assert.Error(t, err)
}

func TestUpdateRole_Cloud(t *testing.T) {
	cfg := loadCloudConfig()

	err := supabase.UpdateRole(cfg, objects.Role{}, objects.UpdateRoleParam{})
	assert.Error(t, err)
}

func TestUpdateRole_SelfHosted(t *testing.T) {
	cfg := loadSelfHostedConfig()

	err := supabase.UpdateRole(cfg, objects.Role{}, objects.UpdateRoleParam{})
	assert.Error(t, err)
}

func TestDeleteRole_Cloud(t *testing.T) {
	cfg := loadCloudConfig()

	err := supabase.DeleteRole(cfg, objects.Role{})
	assert.Error(t, err)
}

func TestDeleteRole_SelfHosted(t *testing.T) {
	cfg := loadSelfHostedConfig()

	err := supabase.DeleteRole(cfg, objects.Role{})
	assert.Error(t, err)
}

func TestGetPolicies_Cloud(t *testing.T) {
	cfg := loadCloudConfig()

	_, err := supabase.GetPolicies(cfg)
	assert.Error(t, err)
}

func TestGetPolicies_SelfHosted(t *testing.T) {
	cfg := loadSelfHostedConfig()

	_, err := supabase.GetPolicies(cfg)
	assert.Error(t, err)
}

func TestCreatePolicy_Cloud(t *testing.T) {
	cfg := loadCloudConfig()

	_, err := supabase.CreatePolicy(cfg, objects.Policy{})
	assert.Error(t, err)
}

func TestCreatePolicy_SelfHosted(t *testing.T) {
	cfg := loadSelfHostedConfig()

	_, err := supabase.CreatePolicy(cfg, objects.Policy{})
	assert.Error(t, err)
}

func TestUpdatePolicy_Cloud(t *testing.T) {
	cfg := loadCloudConfig()

	err := supabase.UpdatePolicy(cfg, objects.Policy{}, objects.UpdatePolicyParam{})
	assert.Error(t, err)
}

func TestUpdatePolicy_SelfHosted(t *testing.T) {
	cfg := loadSelfHostedConfig()

	err := supabase.UpdatePolicy(cfg, objects.Policy{}, objects.UpdatePolicyParam{})
	assert.Error(t, err)
}

func TestDeletePolicy_Cloud(t *testing.T) {
	cfg := loadCloudConfig()

	err := supabase.DeletePolicy(cfg, objects.Policy{})
	assert.Error(t, err)
}

func TestDeletePolicy_SelfHosted(t *testing.T) {
	cfg := loadSelfHostedConfig()

	err := supabase.DeletePolicy(cfg, objects.Policy{})
	assert.Error(t, err)
}

func TestGetFunctions_Cloud(t *testing.T) {
	cfg := loadCloudConfig()

	_, err := supabase.GetFunctions(cfg)
	assert.Error(t, err)
}

func TestGetFunctions_SelfHosted(t *testing.T) {
	cfg := loadSelfHostedConfig()

	_, err := supabase.GetFunctions(cfg)
	assert.Error(t, err)
}

func TestCreateFunction_Cloud(t *testing.T) {
	cfg := loadCloudConfig()

	_, err := supabase.CreateFunction(cfg, objects.Function{})
	assert.Error(t, err)
}

func TestCreateFunction_SelfHosted(t *testing.T) {
	cfg := loadSelfHostedConfig()

	_, err := supabase.CreateFunction(cfg, objects.Function{})
	assert.Error(t, err)
}

func TestUpdateFunction_Cloud(t *testing.T) {
	cfg := loadCloudConfig()

	err := supabase.UpdateFunction(cfg, objects.Function{})
	assert.Error(t, err)
}

func TestUpdateFunction_SelfHosted(t *testing.T) {
	cfg := loadSelfHostedConfig()

	err := supabase.UpdateFunction(cfg, objects.Function{})
	assert.Error(t, err)
}

func TestDeleteFunction_Cloud(t *testing.T) {
	cfg := loadCloudConfig()

	err := supabase.DeleteFunction(cfg, objects.Function{})
	assert.Error(t, err)
}

func TestDeleteFunction_SelfHosted(t *testing.T) {
	cfg := loadSelfHostedConfig()

	err := supabase.DeleteFunction(cfg, objects.Function{})
	assert.Error(t, err)
}

func TestAdminUpdateUser_Cloud(t *testing.T) {
	cfg := loadCloudConfig()

	_, err := supabase.AdminUpdateUserData(cfg, "some-id", objects.User{})
	assert.Error(t, err)
}

func TestAdminUpdateUser_SelfHosted(t *testing.T) {
	cfg := loadSelfHostedConfig()

	_, err := supabase.AdminUpdateUserData(cfg, "some-id", objects.User{})
	assert.Error(t, err)
}

func TestGetBuckets_All(t *testing.T) {
	cfg := loadCloudConfig()

	_, err := supabase.GetBuckets(cfg)
	assert.Error(t, err)
}

func TestGetBucket_All(t *testing.T) {
	cfg := loadCloudConfig()

	_, err := supabase.GetBucket(cfg, "some-bucket")
	assert.Error(t, err)
}

func TestCreateBucket_All(t *testing.T) {
	cfg := loadCloudConfig()

	_, err := supabase.CreateBucket(cfg, objects.Bucket{})
	assert.Error(t, err)
}

func TestUpdateBucket_All(t *testing.T) {
	cfg := loadCloudConfig()

	err := supabase.UpdateBucket(cfg, objects.Bucket{}, objects.UpdateBucketParam{})
	assert.NoError(t, err)
}

func TestDeleteBucket_All(t *testing.T) {
	cfg := loadCloudConfig()

	err := supabase.DeleteBucket(cfg, objects.Bucket{})
	assert.Error(t, err)
}
