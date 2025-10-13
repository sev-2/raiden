package supabase_test

import (
	"errors"
	"testing"
	"time"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/mock"
	"github.com/sev-2/raiden/pkg/supabase"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/stretchr/testify/assert"
)

var (
	relationAction = objects.TablesRelationshipAction{
		ConstraintName: "constraint1",
		UpdateAction:   "c",
		DeletionAction: "c",
	}
	sampleUpdateNewTable = objects.Table{
		Schema: "some-schema",
		Name:   "some-table",
		Columns: []objects.Column{
			{
				Name:         "some-column",
				DataType:     "json",
				IsNullable:   true,
				DefaultValue: "[{\"key\": \"value\"}]",
			},
			{
				Name:               "another-column",
				DataType:           "bool",
				IsNullable:         false,
				DefaultValue:       "true",
				IsIdentity:         true,
				IdentityGeneration: "BY DEFAULT",
			},
		},
		Relationships: []objects.TablesRelationship{
			{
				ConstraintName:    "some-constraint",
				SourceSchema:      "some-schema",
				SourceColumnName:  "some-column",
				TargetTableSchema: "other-schema",
				Action:            &relationAction,
				Index:             &objects.Index{Schema: "public", Table: "table1", Name: "index1", Definition: "index1"},
			},
		},
		RLSEnabled: true,
		RLSForced:  true,
		PrimaryKeys: []objects.PrimaryKey{
			{
				Name: "some-pk",
			},
		},
	}

	sampleUpdateOldTable = objects.Table{
		Name: "some-table",
		Columns: []objects.Column{
			{
				Name:     "some-column",
				DataType: "text",
			},
			{
				Name:     "another-column",
				DataType: "text",
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
		PrimaryKeys: []objects.PrimaryKey{
			{
				Name: "old-pk",
			},
		},
	}

	checkPolicy = "some-check"

	localPolicy = objects.Policy{
		Name:       "some-policy",
		Definition: "SOME DEFINITION",
		Check:      &checkPolicy,
		Roles:      []string{"some-role"},
		Schema:     "some-schema",
		Table:      "some-table",
		Command:    "ALL",
	}
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
		DeploymentTarget:     raiden.DeploymentTargetSelfHosted,
		ProjectId:            "test-project-local-id",
		SupabaseApiBasePath:  "/v1",
		SupabaseApiUrl:       "http://supabase.local.com",
		SupabaseApiToken:     "qwe3qe12e23e32r42ear",
		SupabaseApiTokenType: "basic",
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

	updateParam := objects.UpdateTableParam{
		OldData: sampleUpdateOldTable,
		ChangeColumnItems: []objects.UpdateColumnItem{
			{
				Name: "some-column",
				UpdateItems: []objects.UpdateColumnType{
					objects.UpdateColumnName,
				},
			},
		},
		ChangeItems: []objects.UpdateTableType{
			objects.UpdateTableSchema,
			objects.UpdateTableName,
			objects.UpdateTableRlsEnable,
			objects.UpdateTableRlsForced,
			objects.UpdateTablePrimaryKey,
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
		OldData: sampleUpdateOldTable,
		ChangeColumnItems: []objects.UpdateColumnItem{
			{
				Name: "some-column",
				UpdateItems: []objects.UpdateColumnType{
					objects.UpdateColumnName,
					objects.UpdateColumnDataType,
					objects.UpdateColumnUnique,
					objects.UpdateColumnNullable,
					objects.UpdateColumnDefaultValue,
					objects.UpdateColumnIdentity,
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
		OldData: sampleUpdateOldTable,
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
		OldData: sampleUpdateOldTable,
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

	relationAction := objects.TablesRelationshipAction{
		ConstraintName: "constraint1",
		UpdateAction:   "c",
		DeletionAction: "c",
	}
	updateParam4 := objects.UpdateTableParam{
		OldData: sampleUpdateOldTable,
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
			{
				Type: objects.UpdateRelationCreateIndex,
				Data: objects.TablesRelationship{
					ConstraintName:    "constraint1",
					SourceSchema:      "public",
					SourceTableName:   "table1",
					SourceColumnName:  "id",
					TargetTableSchema: "public",
					TargetTableName:   "table2",
					TargetColumnName:  "id",
					Index:             &objects.Index{Schema: "public", Table: "table1", Name: "index1", Definition: "index1"},
				},
			},
			{
				Type: objects.UpdateRelationActionOnUpdate,
				Data: objects.TablesRelationship{
					ConstraintName:    "constraint1",
					SourceSchema:      "public",
					SourceTableName:   "table1",
					SourceColumnName:  "id",
					TargetTableSchema: "public",
					TargetTableName:   "table2",
					TargetColumnName:  "id",
					Action:            &relationAction,
				},
			},
			{
				Type: objects.UpdateRelationActionOnDelete,
				Data: objects.TablesRelationship{
					ConstraintName:    "constraint1",
					SourceSchema:      "public",
					SourceTableName:   "table1",
					SourceColumnName:  "id",
					TargetTableSchema: "public",
					TargetTableName:   "table2",
					TargetColumnName:  "id",
					Action:            &relationAction,
				},
			},
		},
		ForceCreateRelation: false,
	}

	mock := mock.MockSupabase{Cfg: cfg}
	mock.Activate()
	defer mock.Deactivate()

	err0 := mock.MockUpdateTableWithExpectedResponse(200)
	assert.NoError(t, err0)

	err1 := supabase.UpdateTable(cfg, sampleUpdateNewTable, updateParam)
	assert.NoError(t, err1)

	err2 := supabase.UpdateTable(cfg, sampleUpdateOldTable, updateParam1)
	assert.NoError(t, err2)

	err2NoC := supabase.UpdateTable(cfg, sampleUpdateOldTable, updateParam1NoConstraint)
	assert.NoError(t, err2NoC)

	err3 := supabase.UpdateTable(cfg, sampleUpdateOldTable, updateParam2)
	assert.NoError(t, err3)

	err3NoC := supabase.UpdateTable(cfg, sampleUpdateOldTable, updateParam2NoConstraint)
	assert.NoError(t, err3NoC)

	err4 := supabase.UpdateTable(cfg, sampleUpdateOldTable, updateParam3)
	assert.NoError(t, err4)

	err5 := supabase.UpdateTable(cfg, sampleUpdateOldTable, updateParam4)
	assert.NoError(t, err5)
}

func TestUpdateTable_SelfHosted(t *testing.T) {
	cfg := loadSelfHostedConfig()

	err := supabase.UpdateTable(cfg, objects.Table{}, objects.UpdateTableParam{})
	assert.Error(t, err)

	updateParam := objects.UpdateTableParam{
		OldData: sampleUpdateOldTable,
		ChangeColumnItems: []objects.UpdateColumnItem{
			{
				Name: "some-column",
				UpdateItems: []objects.UpdateColumnType{
					objects.UpdateColumnName,
				},
			},
		},
		ChangeItems: []objects.UpdateTableType{
			objects.UpdateTableSchema,
			objects.UpdateTableName,
			objects.UpdateTableRlsEnable,
			objects.UpdateTableRlsForced,
			objects.UpdateTablePrimaryKey,
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
		OldData: sampleUpdateOldTable,
		ChangeColumnItems: []objects.UpdateColumnItem{
			{
				Name: "some-column",
				UpdateItems: []objects.UpdateColumnType{
					objects.UpdateColumnName,
					objects.UpdateColumnDataType,
					objects.UpdateColumnUnique,
					objects.UpdateColumnNullable,
					objects.UpdateColumnDefaultValue,
					objects.UpdateColumnIdentity,
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
		OldData: sampleUpdateOldTable,
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
		OldData: sampleUpdateOldTable,
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

	err1 := supabase.UpdateTable(cfg, sampleUpdateNewTable, updateParam)
	assert.NoError(t, err1)

	err2 := supabase.UpdateTable(cfg, sampleUpdateOldTable, updateParam1)
	assert.NoError(t, err2)

	err2NoC := supabase.UpdateTable(cfg, sampleUpdateOldTable, updateParam1NoConstraint)
	assert.NoError(t, err2NoC)

	err3 := supabase.UpdateTable(cfg, sampleUpdateOldTable, updateParam2)
	assert.NoError(t, err3)

	err3NoC := supabase.UpdateTable(cfg, sampleUpdateOldTable, updateParam2NoConstraint)
	assert.NoError(t, err3NoC)

	err4 := supabase.UpdateTable(cfg, sampleUpdateOldTable, updateParam3)
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
			Config: map[string]interface{}{
				"somekey":  "somevalue",
				"otherkey": "othervalue",
			},
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
			Config: map[string]interface{}{
				"somekey":  "somevalue",
				"otherkey": "othervalue",
			},
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

func TestGetRoleMemberships_Cloud(t *testing.T) {
	cfg := loadCloudConfig()

	_, err := supabase.GetRoleMemberships(cfg, nil)
	assert.Error(t, err)

	remoteMemberships := []objects.RoleMembership{
		{ParentID: 1, ParentRole: "student", InheritID: 2, InheritRole: "admin"},
	}

	mock := mock.MockSupabase{Cfg: cfg}
	mock.Activate()
	defer mock.Deactivate()

	err0 := mock.MockGetRoleMembershipsWithExpectedResponse(200, remoteMemberships)
	assert.NoError(t, err0)

	memberships, err1 := supabase.GetRoleMemberships(cfg, nil)
	assert.NoError(t, err1)
	assert.Equal(t, len(remoteMemberships), len(memberships))
}

func TestGetRoleMemberships_SelfHosted(t *testing.T) {
	cfg := loadSelfHostedConfig()

	_, err := supabase.GetRoleMemberships(cfg, nil)
	assert.Error(t, err)

	remoteMemberships := []objects.RoleMembership{
		{ParentID: 1, ParentRole: "student", InheritID: 2, InheritRole: "admin"},
	}

	mock := mock.MockSupabase{Cfg: cfg}
	mock.Activate()
	defer mock.Deactivate()

	err0 := mock.MockGetRoleMembershipsWithExpectedResponse(200, remoteMemberships)
	assert.NoError(t, err0)

	memberships, err1 := supabase.GetRoleMemberships(cfg, nil)
	assert.NoError(t, err1)
	assert.Equal(t, len(remoteMemberships), len(memberships))
}

func TestGetRoleByName_Cloud(t *testing.T) {
	cfg := loadCloudConfig()

	_, err := supabase.GetRoleByName(cfg, "some-role")
	assert.Error(t, err)

	remoteRole := objects.Role{
		ID:   1,
		Name: "some-role",
	}

	mock := mock.MockSupabase{Cfg: cfg}
	mock.Activate()
	defer mock.Deactivate()

	err0 := mock.MockGetRoleByNameWithExpectedResponse(200, remoteRole)
	assert.NoError(t, err0)

	role, err1 := supabase.GetRoleByName(cfg, "some-role")
	assert.NoError(t, err1)
	assert.Equal(t, remoteRole.Name, role.Name)
}

func TestGetRoleByName_SelfHosted(t *testing.T) {
	cfg := loadSelfHostedConfig()

	_, err := supabase.GetRoleByName(cfg, "some-role")
	assert.Error(t, err)

	remoteRole := objects.Role{
		ID:   1,
		Name: "some-role",
	}

	mock := mock.MockSupabase{Cfg: cfg}
	mock.Activate()
	defer mock.Deactivate()

	err0 := mock.MockGetRoleByNameWithExpectedResponse(200, remoteRole)
	assert.NoError(t, err0)

	role, err1 := supabase.GetRoleByName(cfg, "some-role")
	assert.NoError(t, err1)
	assert.Equal(t, remoteRole.Name, role.Name)
}

func TestCreateRole_Cloud(t *testing.T) {
	cfg := loadCloudConfig()

	_, err := supabase.CreateRole(cfg, objects.Role{})
	assert.Error(t, err)

	localRole := objects.Role{
		Name: "some-role",
	}

	mock := mock.MockSupabase{Cfg: cfg}
	mock.Activate()
	defer mock.Deactivate()

	err0 := mock.MockCreateRoleWithExpectedResponse(200, localRole)
	assert.NoError(t, err0)

	createdRole, err1 := supabase.CreateRole(cfg, localRole)
	assert.NoError(t, err1)
	assert.Equal(t, localRole.Name, createdRole.Name)
}

func TestCreateRole_SelfHosted(t *testing.T) {
	cfg := loadSelfHostedConfig()

	_, err := supabase.CreateRole(cfg, objects.Role{})
	assert.Error(t, err)

	localRole := objects.Role{
		Name: "some-role",
	}

	mock := mock.MockSupabase{Cfg: cfg}
	mock.Activate()
	defer mock.Deactivate()

	err0 := mock.MockCreateRoleWithExpectedResponse(200, localRole)
	assert.NoError(t, err0)

	createdRole, err1 := supabase.CreateRole(cfg, localRole)
	assert.NoError(t, err1)
	assert.Equal(t, localRole.Name, createdRole.Name)
}

func TestUpdateRole_Cloud(t *testing.T) {
	cfg := loadCloudConfig()

	err := supabase.UpdateRole(cfg, objects.Role{}, objects.UpdateRoleParam{})
	assert.Error(t, err)

	var validUntil = objects.NewSupabaseTime(time.Now())

	_, errT := validUntil.MarshalJSON()
	assert.NoError(t, errT)

	localRole := objects.Role{
		Name:            "some-role",
		CanLogin:        true,
		IsSuperuser:     true,
		ValidUntil:      validUntil,
		ConnectionLimit: 11,
		Config: map[string]interface{}{
			"somekey":  "somevalue",
			"otherkey": "othervalue",
		},
	}

	updateParam := objects.UpdateRoleParam{
		OldData: localRole,
		ChangeItems: []objects.UpdateRoleType{
			objects.UpdateConnectionLimit,
			objects.UpdateRoleName,
			objects.UpdateRoleIsReplication,
			objects.UpdateRoleIsSuperUser,
			objects.UpdateRoleInheritRole,
			objects.UpdateRoleCanBypassRls,
			objects.UpdateRoleCanCreateRole,
			objects.UpdateRoleCanCreateDb,
			objects.UpdateRoleCanLogin,
			objects.UpdateRoleValidUntil,
			objects.UpdateRoleConfig,
		},
	}

	mock := mock.MockSupabase{Cfg: cfg}
	mock.Activate()
	defer mock.Deactivate()

	err0 := mock.MockUpdateRoleWithExpectedResponse(200)
	assert.NoError(t, err0)

	err1 := supabase.UpdateRole(cfg, localRole, updateParam)
	assert.NoError(t, err1)
}

func TestUpdateRole_SelfHosted(t *testing.T) {
	cfg := loadSelfHostedConfig()

	err := supabase.UpdateRole(cfg, objects.Role{}, objects.UpdateRoleParam{})
	assert.Error(t, err)

	var validUntil = objects.NewSupabaseTime(time.Now())

	_, errT := validUntil.MarshalJSON()
	assert.NoError(t, errT)

	localRole := objects.Role{
		Name:            "some-role",
		CanLogin:        true,
		IsSuperuser:     true,
		ValidUntil:      validUntil,
		ConnectionLimit: 11,
		Config: map[string]interface{}{
			"somekey":  "somevalue",
			"otherkey": "othervalue",
		},
	}

	updateParam := objects.UpdateRoleParam{
		OldData: localRole,
		ChangeItems: []objects.UpdateRoleType{
			objects.UpdateConnectionLimit,
			objects.UpdateRoleName,
			objects.UpdateRoleIsReplication,
			objects.UpdateRoleIsSuperUser,
			objects.UpdateRoleInheritRole,
			objects.UpdateRoleCanBypassRls,
			objects.UpdateRoleCanCreateRole,
			objects.UpdateRoleCanCreateDb,
			objects.UpdateRoleCanLogin,
			objects.UpdateRoleValidUntil,
			objects.UpdateRoleConfig,
		},
	}

	mock := mock.MockSupabase{Cfg: cfg}
	mock.Activate()
	defer mock.Deactivate()

	err0 := mock.MockUpdateRoleWithExpectedResponse(200)
	assert.NoError(t, err0)

	err1 := supabase.UpdateRole(cfg, localRole, updateParam)
	assert.NoError(t, err1)
}

func TestDeleteRole_Cloud(t *testing.T) {
	cfg := loadCloudConfig()

	err := supabase.DeleteRole(cfg, objects.Role{})
	assert.Error(t, err)

	localRole := objects.Role{
		Name: "some-role",
	}

	mock := mock.MockSupabase{Cfg: cfg}
	mock.Activate()
	defer mock.Deactivate()

	err0 := mock.MockDeleteRoleWithExpectedResponse(200)
	assert.NoError(t, err0)

	err1 := supabase.DeleteRole(cfg, localRole)
	assert.NoError(t, err1)
}

func TestDeleteRole_SelfHosted(t *testing.T) {
	cfg := loadSelfHostedConfig()

	err := supabase.DeleteRole(cfg, objects.Role{})
	assert.Error(t, err)

	localRole := objects.Role{
		Name: "some-role",
	}

	mock := mock.MockSupabase{Cfg: cfg}
	mock.Activate()
	defer mock.Deactivate()

	err0 := mock.MockDeleteRoleWithExpectedResponse(200)
	assert.NoError(t, err0)

	err1 := supabase.DeleteRole(cfg, localRole)
	assert.NoError(t, err1)
}

func TestGetPolicies_Cloud(t *testing.T) {
	cfg := loadCloudConfig()

	_, err := supabase.GetPolicies(cfg)
	assert.Error(t, err)

	remotePolicies := []objects.Policy{
		{
			ID:   1,
			Name: "some-policy",
		},
		{
			ID:   2,
			Name: "another-policy",
		},
	}

	mock := mock.MockSupabase{Cfg: cfg}
	mock.Activate()
	defer mock.Deactivate()

	err0 := mock.MockGetPoliciesWithExpectedResponse(200, remotePolicies)
	assert.NoError(t, err0)

	policies, err1 := supabase.GetPolicies(cfg)
	assert.NoError(t, err1)
	assert.Equal(t, len(remotePolicies), len(policies))
}

func TestGetPolicies_SelfHosted(t *testing.T) {
	cfg := loadSelfHostedConfig()

	_, err := supabase.GetPolicies(cfg)
	assert.Error(t, err)

	remotePolicies := []objects.Policy{
		{
			ID:   1,
			Name: "some-policy",
		},
		{
			ID:   2,
			Name: "another-policy",
		},
	}

	mock := mock.MockSupabase{Cfg: cfg}
	mock.Activate()
	defer mock.Deactivate()

	err0 := mock.MockGetPoliciesWithExpectedResponse(200, remotePolicies)
	assert.NoError(t, err0)

	policies, err1 := supabase.GetPolicies(cfg)
	assert.NoError(t, err1)
	assert.Equal(t, len(remotePolicies), len(policies))
}

func TestGetPolicyByName_Cloud(t *testing.T) {
	cfg := loadCloudConfig()

	_, err := supabase.GetPolicyByName(cfg, "some-policy")
	assert.Error(t, err)

	remotePolicy := objects.Policy{
		ID:   1,
		Name: "some-policy",
	}

	mock := mock.MockSupabase{Cfg: cfg}
	mock.Activate()
	defer mock.Deactivate()

	err0 := mock.MockGetPolicyByNameWithExpectedResponse(200, remotePolicy)
	assert.NoError(t, err0)

	policy, err1 := supabase.GetPolicyByName(cfg, "some-policy")
	assert.NoError(t, err1)
	assert.Equal(t, remotePolicy.Name, policy.Name)
}

func TestGetPolicyByName_SelfHosted(t *testing.T) {
	cfg := loadSelfHostedConfig()

	_, err := supabase.GetPolicyByName(cfg, "some-policy")
	assert.Error(t, err)

	remotePolicy := objects.Policy{
		ID:   1,
		Name: "some-policy",
	}

	mock := mock.MockSupabase{Cfg: cfg}
	mock.Activate()
	defer mock.Deactivate()

	err0 := mock.MockGetPolicyByNameWithExpectedResponse(200, remotePolicy)
	assert.NoError(t, err0)

	policy, err1 := supabase.GetPolicyByName(cfg, "some-policy")
	assert.NoError(t, err1)
	assert.Equal(t, remotePolicy.Name, policy.Name)
}

func TestCreatePolicy_Cloud(t *testing.T) {
	cfg := loadCloudConfig()

	_, err := supabase.CreatePolicy(cfg, objects.Policy{})
	assert.Error(t, err)

	mock := mock.MockSupabase{Cfg: cfg}
	mock.Activate()
	defer mock.Deactivate()

	err0 := mock.MockCreatePolicyWithExpectedResponse(200, localPolicy)
	assert.NoError(t, err0)
}

func TestCreatePolicy_SelfHosted(t *testing.T) {
	cfg := loadSelfHostedConfig()

	_, err := supabase.CreatePolicy(cfg, objects.Policy{})
	assert.Error(t, err)

	mock := mock.MockSupabase{Cfg: cfg}
	mock.Activate()
	defer mock.Deactivate()

	err0 := mock.MockCreatePolicyWithExpectedResponse(200, localPolicy)
	assert.NoError(t, err0)
}

func TestUpdatePolicy_Cloud(t *testing.T) {
	cfg := loadCloudConfig()

	err := supabase.UpdatePolicy(cfg, objects.Policy{}, objects.UpdatePolicyParam{})
	assert.Error(t, err)

	updateParam := objects.UpdatePolicyParam{
		Name: "some-policy",
		ChangeItems: []objects.UpdatePolicyType{
			objects.UpdatePolicyName,
			objects.UpdatePolicyCheck,
			objects.UpdatePolicyDefinition,
			objects.UpdatePolicyRoles,
		},
	}

	mock := mock.MockSupabase{Cfg: cfg}
	mock.Activate()
	defer mock.Deactivate()

	err0 := mock.MockUpdatePolicyWithExpectedResponse(200)
	assert.NoError(t, err0)

	err1 := supabase.UpdatePolicy(cfg, localPolicy, updateParam)
	assert.NoError(t, err1)
}

func TestUpdatePolicy_SelfHosted(t *testing.T) {
	cfg := loadSelfHostedConfig()

	err := supabase.UpdatePolicy(cfg, objects.Policy{}, objects.UpdatePolicyParam{})
	assert.Error(t, err)

	updateParam := objects.UpdatePolicyParam{
		Name: "some-policy",
		ChangeItems: []objects.UpdatePolicyType{
			objects.UpdatePolicyName,
			objects.UpdatePolicyCheck,
			objects.UpdatePolicyDefinition,
			objects.UpdatePolicyRoles,
		},
	}

	mock := mock.MockSupabase{Cfg: cfg}
	mock.Activate()
	defer mock.Deactivate()

	err0 := mock.MockUpdatePolicyWithExpectedResponse(200)
	assert.NoError(t, err0)

	err1 := supabase.UpdatePolicy(cfg, localPolicy, updateParam)
	assert.NoError(t, err1)
}

func TestDeletePolicy_Cloud(t *testing.T) {
	cfg := loadCloudConfig()

	err := supabase.DeletePolicy(cfg, objects.Policy{})
	assert.Error(t, err)

	localPolicy := objects.Policy{
		Name: "some-policy",
	}

	mock := mock.MockSupabase{Cfg: cfg}
	mock.Activate()
	defer mock.Deactivate()

	err0 := mock.MockDeletePolicyWithExpectedResponse(200)
	assert.NoError(t, err0)

	err1 := supabase.DeletePolicy(cfg, localPolicy)
	assert.NoError(t, err1)
}

func TestDeletePolicy_SelfHosted(t *testing.T) {
	cfg := loadSelfHostedConfig()

	err := supabase.DeletePolicy(cfg, objects.Policy{})
	assert.Error(t, err)

	localPolicy := objects.Policy{
		Name: "some-policy",
	}

	mock := mock.MockSupabase{Cfg: cfg}
	mock.Activate()
	defer mock.Deactivate()

	err0 := mock.MockDeletePolicyWithExpectedResponse(200)
	assert.NoError(t, err0)

	err1 := supabase.DeletePolicy(cfg, localPolicy)
	assert.NoError(t, err1)
}

func TestGetFunctions_Cloud(t *testing.T) {
	cfg := loadCloudConfig()

	_, err := supabase.GetFunctions(cfg)
	assert.Error(t, err)

	remoteFunctions := []objects.Function{
		{
			Name: "some-function",
		},
		{
			Name: "another-function",
		},
	}

	mock := mock.MockSupabase{Cfg: cfg}
	mock.Activate()
	defer mock.Deactivate()

	err0 := mock.MockGetFunctionsWithExpectedResponse(200, remoteFunctions)
	assert.NoError(t, err0)

	functions, err1 := supabase.GetFunctions(cfg)
	assert.NoError(t, err1)
	assert.Equal(t, len(remoteFunctions), len(functions))
}

func TestGetFunctions_SelfHosted(t *testing.T) {
	cfg := loadSelfHostedConfig()

	_, err := supabase.GetFunctions(cfg)
	assert.Error(t, err)

	remoteFunctions := []objects.Function{
		{
			Name: "some-function",
		},
		{
			Name: "another-function",
		},
	}

	mock := mock.MockSupabase{Cfg: cfg}
	mock.Activate()
	defer mock.Deactivate()

	err0 := mock.MockGetFunctionsWithExpectedResponse(200, remoteFunctions)
	assert.NoError(t, err0)

	functions, err1 := supabase.GetFunctions(cfg)
	assert.NoError(t, err1)
	assert.Equal(t, len(remoteFunctions), len(functions))
}

func TestGetFunctionByName_Cloud(t *testing.T) {
	cfg := loadCloudConfig()

	_, err := supabase.GetFunctions(cfg)
	assert.Error(t, err)

	remoteFunction := objects.Function{
		Name: "some-function",
	}

	mock := mock.MockSupabase{Cfg: cfg}
	mock.Activate()
	defer mock.Deactivate()

	err0 := mock.MockGetFunctionByNameWithExpectedResponse(200, remoteFunction)
	assert.NoError(t, err0)

	function, err1 := supabase.GetFunctionByName(cfg, "some-schema", "some-function")
	assert.NoError(t, err1)
	assert.Equal(t, remoteFunction.Name, function.Name)
}

func TestGetFunctionByName_SelfHosted(t *testing.T) {
	cfg := loadSelfHostedConfig()

	_, err := supabase.GetFunctions(cfg)
	assert.Error(t, err)

	remoteFunction := objects.Function{
		Name: "some-function",
	}

	mock := mock.MockSupabase{Cfg: cfg}
	mock.Activate()
	defer mock.Deactivate()

	err0 := mock.MockGetFunctionByNameWithExpectedResponse(200, remoteFunction)
	assert.NoError(t, err0)

	function, err1 := supabase.GetFunctionByName(cfg, "some-schema", "some-function")
	assert.NoError(t, err1)
	assert.Equal(t, remoteFunction.Name, function.Name)
}

func TestCreateFunction_Cloud(t *testing.T) {
	cfg := loadCloudConfig()

	_, err := supabase.CreateFunction(cfg, objects.Function{})
	assert.Error(t, err)

	localFunction := objects.Function{
		Name: "some-function",
	}

	mock := mock.MockSupabase{Cfg: cfg}
	mock.Activate()
	defer mock.Deactivate()

	err0 := mock.MockCreateFunctionWithExpectedResponse(200, localFunction)
	assert.NoError(t, err0)

	createdFunction, err1 := supabase.CreateFunction(cfg, localFunction)
	assert.NoError(t, err1)
	assert.Equal(t, localFunction.Name, createdFunction.Name)
}

func TestCreateFunction_SelfHosted(t *testing.T) {
	cfg := loadSelfHostedConfig()

	_, err := supabase.CreateFunction(cfg, objects.Function{})
	assert.Error(t, err)

	localFunction := objects.Function{
		Name: "some-function",
	}

	mock := mock.MockSupabase{Cfg: cfg}
	mock.Activate()
	defer mock.Deactivate()

	err0 := mock.MockCreateFunctionWithExpectedResponse(200, localFunction)
	assert.NoError(t, err0)

	createdFunction, err1 := supabase.CreateFunction(cfg, localFunction)
	assert.NoError(t, err1)
	assert.Equal(t, localFunction.Name, createdFunction.Name)
}

func TestUpdateFunction_Cloud(t *testing.T) {
	cfg := loadCloudConfig()

	err := supabase.UpdateFunction(cfg, objects.Function{})
	assert.Error(t, err)

	localFunction := objects.Function{
		Name: "some-function",
	}

	mock := mock.MockSupabase{Cfg: cfg}
	mock.Activate()
	defer mock.Deactivate()

	err0 := mock.MockUpdateFunctionWithExpectedResponse(200)
	assert.NoError(t, err0)

	err1 := supabase.UpdateFunction(cfg, localFunction)
	assert.NoError(t, err1)
}

func TestUpdateFunction_SelfHosted(t *testing.T) {
	cfg := loadSelfHostedConfig()

	err := supabase.UpdateFunction(cfg, objects.Function{})
	assert.Error(t, err)

	localFunction := objects.Function{
		Name: "some-function",
	}

	mock := mock.MockSupabase{Cfg: cfg}
	mock.Activate()
	defer mock.Deactivate()

	err0 := mock.MockUpdateFunctionWithExpectedResponse(200)
	assert.NoError(t, err0)

	err1 := supabase.UpdateFunction(cfg, localFunction)
	assert.NoError(t, err1)
}

func TestDeleteFunction_Cloud(t *testing.T) {
	cfg := loadCloudConfig()

	err := supabase.DeleteFunction(cfg, objects.Function{})
	assert.Error(t, err)

	localFunction := objects.Function{
		Name: "some-function",
	}

	mock := mock.MockSupabase{Cfg: cfg}
	mock.Activate()
	defer mock.Deactivate()

	err0 := mock.MockDeleteFunctionWithExpectedResponse(200)
	assert.NoError(t, err0)

	err1 := supabase.DeleteFunction(cfg, localFunction)
	assert.NoError(t, err1)
}

func TestDeleteFunction_SelfHosted(t *testing.T) {
	cfg := loadSelfHostedConfig()

	err := supabase.DeleteFunction(cfg, objects.Function{})
	assert.Error(t, err)

	localFunction := objects.Function{
		Name: "some-function",
	}

	mock := mock.MockSupabase{Cfg: cfg}
	mock.Activate()
	defer mock.Deactivate()

	err0 := mock.MockDeleteFunctionWithExpectedResponse(200)
	assert.NoError(t, err0)

	err1 := supabase.DeleteFunction(cfg, localFunction)
	assert.NoError(t, err1)
}

func TestGetIndexes_Cloud(t *testing.T) {
	cfg := loadCloudConfig()

	_, err0 := supabase.GetIndexes(cfg, "")
	assert.Error(t, err0)

	_, err1 := supabase.GetIndexes(cfg, "public")
	assert.Error(t, err1)
}

func TestGetIndexes_SelfHosted(t *testing.T) {
	cfg := loadSelfHostedConfig()

	_, err0 := supabase.GetIndexes(cfg, "")
	assert.Error(t, err0)

	_, err1 := supabase.GetIndexes(cfg, "public")
	assert.Error(t, err1)
}

func TestGetActions_Cloud(t *testing.T) {
	cfg := loadCloudConfig()

	_, err0 := supabase.GetTableRelationshipActions(cfg, "")
	assert.Error(t, err0)

	_, err1 := supabase.GetTableRelationshipActions(cfg, "public")
	assert.Error(t, err1)
}

func TestGetActions_SelfHosted(t *testing.T) {
	cfg := loadSelfHostedConfig()

	_, err0 := supabase.GetTableRelationshipActions(cfg, "")
	assert.Error(t, err0)

	_, err1 := supabase.GetTableRelationshipActions(cfg, "public")
	assert.Error(t, err1)
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

func TestGetTypes_Cloud(t *testing.T) {
	cfg := loadCloudConfig()

	_, err := supabase.GetTypes(cfg, []string{"public"})
	assert.Error(t, err)

	remoteTypes := []objects.Type{
		{
			ID:   1,
			Name: "some-type",
		},
		{
			ID:   2,
			Name: "another-type",
		},
	}

	mock := mock.MockSupabase{Cfg: cfg}
	mock.Activate()
	defer mock.Deactivate()

	err0 := mock.MockGetTypesWithExpectedResponse(200, remoteTypes)
	assert.NoError(t, err0)

	types, err1 := supabase.GetTypes(cfg, []string{"public"})
	assert.NoError(t, err1)
	assert.Equal(t, len(remoteTypes), len(types))
}

func TestGetTypes_SelfHosted(t *testing.T) {
	cfg := loadSelfHostedConfig()

	_, err := supabase.GetTypes(cfg, []string{"public"})
	assert.Error(t, err)

	remoteTypes := []objects.Type{
		{
			ID:   1,
			Name: "some-type",
		},
		{
			ID:   2,
			Name: "another-type",
		},
	}

	mock := mock.MockSupabase{Cfg: cfg}
	mock.Activate()
	defer mock.Deactivate()

	err0 := mock.MockGetTypesWithExpectedResponse(200, remoteTypes)
	assert.NoError(t, err0)

	types, err1 := supabase.GetTypes(cfg, []string{"public"})
	assert.NoError(t, err1)
	assert.Equal(t, len(remoteTypes), len(types))
}

func TestGetTypeByName_Cloud(t *testing.T) {
	cfg := loadCloudConfig()

	_, err := supabase.GetTypeByName(cfg, []string{"public"}, "some-type")
	assert.Error(t, err)

	remoteType := objects.Type{
		ID:   1,
		Name: "some-type",
	}

	mock := mock.MockSupabase{Cfg: cfg}
	mock.Activate()
	defer mock.Deactivate()

	err0 := mock.MockGetTypeByNameWithExpectedResponse(200, remoteType)
	assert.NoError(t, err0)

	dataType, err1 := supabase.GetTypeByName(cfg, []string{"public"}, "some-type")
	assert.NoError(t, err1)
	assert.Equal(t, remoteType.Name, dataType.Name)
}

func TestGetTypeByName_SelfHosted(t *testing.T) {
	cfg := loadSelfHostedConfig()

	_, err := supabase.GetTypeByName(cfg, []string{"public"}, "some-type")
	assert.Error(t, err)

	remoteType := objects.Type{
		ID:   1,
		Name: "some-type",
	}

	mock := mock.MockSupabase{Cfg: cfg}
	mock.Activate()
	defer mock.Deactivate()

	err0 := mock.MockGetTypeByNameWithExpectedResponse(200, remoteType)
	assert.NoError(t, err0)

	dataType, err1 := supabase.GetTypeByName(cfg, []string{"public"}, "some-type")
	assert.NoError(t, err1)
	assert.Equal(t, remoteType.Name, dataType.Name)
}

func TestCreateType_Cloud(t *testing.T) {
	cfg := loadCloudConfig()

	_, err := supabase.CreateType(cfg, objects.Type{})
	assert.Error(t, err)

	localType := objects.Type{
		Name: "some-type",
	}

	mock := mock.MockSupabase{Cfg: cfg}
	mock.Activate()
	defer mock.Deactivate()

	err0 := mock.MockCreateTypeWithExpectedResponse(200, localType)
	assert.NoError(t, err0)

	createdType, err1 := supabase.CreateType(cfg, localType)
	assert.NoError(t, err1)
	assert.Equal(t, localType.Name, createdType.Name)
}

func TestCreateType_SelfHosted(t *testing.T) {
	cfg := loadSelfHostedConfig()

	_, err := supabase.CreateType(cfg, objects.Type{})
	assert.Error(t, err)

	localType := objects.Type{
		Name: "some-type",
	}

	mock := mock.MockSupabase{Cfg: cfg}
	mock.Activate()
	defer mock.Deactivate()

	err0 := mock.MockCreateTypeWithExpectedResponse(200, localType)
	assert.NoError(t, err0)

	createdType, err1 := supabase.CreateType(cfg, localType)
	assert.NoError(t, err1)
	assert.Equal(t, localType.Name, createdType.Name)
}

func TestUpdateType_Cloud(t *testing.T) {
	cfg := loadCloudConfig()

	err := supabase.UpdateType(cfg, objects.Type{})
	assert.Error(t, err)

	var validUntil = objects.NewSupabaseTime(time.Now())

	_, errT := validUntil.MarshalJSON()
	assert.NoError(t, errT)

	localType := objects.Type{
		Name:       "some-type",
		Schema:     "public",
		Format:     "",
		Enums:      []string{"test_1"},
		Attributes: []objects.TypeAttribute{},
		Comment:    nil,
	}

	mock := mock.MockSupabase{Cfg: cfg}
	mock.Activate()
	defer mock.Deactivate()

	err0 := mock.MockUpdateTypeWithExpectedResponse(200)
	assert.NoError(t, err0)

	err1 := supabase.UpdateType(cfg, localType)
	assert.NoError(t, err1)
}

// Type Test

func TestUpdateType_SelfHosted(t *testing.T) {
	cfg := loadSelfHostedConfig()

	err := supabase.UpdateType(cfg, objects.Type{})
	assert.Error(t, err)

	var validUntil = objects.NewSupabaseTime(time.Now())

	_, errT := validUntil.MarshalJSON()
	assert.NoError(t, errT)

	localType := objects.Type{
		Name:       "some-type",
		Schema:     "public",
		Format:     "",
		Enums:      []string{"test_1"},
		Attributes: []objects.TypeAttribute{},
		Comment:    nil,
	}

	mock := mock.MockSupabase{Cfg: cfg}
	mock.Activate()
	defer mock.Deactivate()

	err0 := mock.MockUpdateTypeWithExpectedResponse(200)
	assert.NoError(t, err0)

	err1 := supabase.UpdateType(cfg, localType)
	assert.NoError(t, err1)
}

func TestDeleteType_Cloud(t *testing.T) {
	cfg := loadCloudConfig()

	err := supabase.DeleteType(cfg, objects.Type{})
	assert.Error(t, err)

	localType := objects.Type{
		Name: "some-type",
	}

	mock := mock.MockSupabase{Cfg: cfg}
	mock.Activate()
	defer mock.Deactivate()

	err0 := mock.MockDeleteTypeWithExpectedResponse(200)
	assert.NoError(t, err0)

	err1 := supabase.DeleteType(cfg, localType)
	assert.NoError(t, err1)
}

func TestDeleteType_SelfHosted(t *testing.T) {
	cfg := loadSelfHostedConfig()

	err := supabase.DeleteType(cfg, objects.Type{})
	assert.Error(t, err)

	localType := objects.Type{
		Name: "some-type",
	}

	mock := mock.MockSupabase{Cfg: cfg}
	mock.Activate()
	defer mock.Deactivate()

	err0 := mock.MockDeleteTypeWithExpectedResponse(200)
	assert.NoError(t, err0)

	err1 := supabase.DeleteType(cfg, localType)
	assert.NoError(t, err1)
}

func TestGetTableByName_Cloud(t *testing.T) {
	cfg := loadCloudConfig()

	_, err := supabase.GetTableByName(cfg, "some-table", "some-schema", true)
	assert.Error(t, err)

	remoteTable := objects.Table{
		ID:   1,
		Name: "some-table",
	}

	mock := mock.MockSupabase{Cfg: cfg}
	mock.Activate()
	defer mock.Deactivate()

	err0 := mock.MockGetTableByNameWithExpectedResponse(200, remoteTable)
	assert.NoError(t, err0)

	table, err1 := supabase.GetTableByName(cfg, "some-table", "some-schema", true)
	assert.NoError(t, err1)
	assert.Equal(t, remoteTable.Name, table.Name)
}

func TestGetTableByName_Cloud_NotFound(t *testing.T) {
	cfg := loadCloudConfig()

	// Test when table is not found (no table in response)
	mock := mock.MockSupabase{Cfg: cfg}
	mock.Activate()
	defer mock.Deactivate()

	// Mocking empty response to trigger "not found" error
	err0 := mock.MockGetTablesWithExpectedResponse(200, []objects.Table{}) // GetTableByName calls GetTables internally
	assert.NoError(t, err0)

	_, err1 := supabase.GetTableByName(cfg, "nonexistent-table", "some-schema", true)
	assert.Error(t, err1)
	assert.Contains(t, err1.Error(), "is not found")
}

func TestUpdateBucket_NoChanges(t *testing.T) {
	cfg := loadCloudConfig()

	bucket := objects.Bucket{
		ID: "test-bucket",
	}
	updateParam := objects.UpdateBucketParam{
		ChangeItems: []objects.UpdateBucketType{}, // Empty - should cause early return
	}

	err := supabase.UpdateBucket(cfg, bucket, updateParam)
	assert.NoError(t, err) // Should return without error due to early return
}

func TestGetTables_Cloud_ErrorHandling(t *testing.T) {
	cfg := loadCloudConfig()

	// Test error scenario
	_, err := supabase.GetTables(cfg, []string{"test-schema"})
	assert.Error(t, err)

	mock := mock.MockSupabase{Cfg: cfg}
	mock.Activate()
	defer mock.Deactivate()

	// Mock error response
	err0 := mock.MockGetTablesWithExpectedResponse(500, []objects.Table{})
	assert.NoError(t, err0)

	_, err1 := supabase.GetTables(cfg, []string{"test-schema"})
	assert.Error(t, err1)
}

func TestGetTables_SelfHosted_ErrorHandling(t *testing.T) {
	cfg := loadSelfHostedConfig()

	// Test error scenario
	_, err := supabase.GetTables(cfg, []string{"test-schema"})
	assert.Error(t, err)

	mock := mock.MockSupabase{Cfg: cfg}
	mock.Activate()
	defer mock.Deactivate()

	// Mock error response
	err0 := mock.MockGetTablesWithExpectedResponse(500, []objects.Table{})
	assert.NoError(t, err0)

	_, err1 := supabase.GetTables(cfg, []string{"test-schema"})
	assert.Error(t, err1)
}

func TestUpdateRole_WithInheritChanges(t *testing.T) {
	cfg := loadCloudConfig()

	var validUntil = objects.NewSupabaseTime(time.Now())

	_, errT := validUntil.MarshalJSON()
	assert.NoError(t, errT)

	localRole := objects.Role{
		Name:            "some-role",
		CanLogin:        true,
		IsSuperuser:     true,
		ValidUntil:      validUntil,
		ConnectionLimit: 11,
		Config: map[string]interface{}{
			"somekey":  "somevalue",
			"otherkey": "othervalue",
		},
	}

	// Create update param with inherit items only
	updateParamInheritOnly := objects.UpdateRoleParam{
		OldData: localRole,
		ChangeInheritItems: []objects.UpdateRoleInheritItem{
			{
				Role: objects.Role{Name: "parent-role"},
				Type: objects.UpdateRoleInheritGrant,
			},
			{
				Role: objects.Role{Name: "another-parent-role"},
				Type: objects.UpdateRoleInheritRevoke,
			},
		},
	}

	mock := mock.MockSupabase{Cfg: cfg}
	mock.Activate()
	defer mock.Deactivate()

	err0 := mock.MockUpdateRoleWithExpectedResponse(200)
	assert.NoError(t, err0)

	// This should test the path where only inherit changes are made
	err1 := supabase.UpdateRole(cfg, localRole, updateParamInheritOnly)
	assert.NoError(t, err1)

	// Test with both normal changes and inherit changes
	updateParamBoth := objects.UpdateRoleParam{
		OldData: localRole,
		ChangeItems: []objects.UpdateRoleType{
			objects.UpdateRoleName,
			objects.UpdateConnectionLimit,
		},
		ChangeInheritItems: []objects.UpdateRoleInheritItem{
			{
				Role: objects.Role{Name: "parent-role-2"},
				Type: objects.UpdateRoleInheritGrant,
			},
		},
	}

	err2 := supabase.UpdateRole(cfg, localRole, updateParamBoth)
	assert.NoError(t, err2)
}

func TestUpdateRole_WithInheritChanges_SelfHosted(t *testing.T) {
	cfg := loadSelfHostedConfig()

	var validUntil = objects.NewSupabaseTime(time.Now())

	_, errT := validUntil.MarshalJSON()
	assert.NoError(t, errT)

	localRole := objects.Role{
		Name:            "some-role-selfhosted",
		CanLogin:        true,
		ValidUntil:      validUntil,
		ConnectionLimit: 11,
	}

	// Create update param with inherit items only
	updateParamInheritOnly := objects.UpdateRoleParam{
		OldData: localRole,
		ChangeInheritItems: []objects.UpdateRoleInheritItem{
			{
				Role: objects.Role{Name: "parent-role"},
				Type: objects.UpdateRoleInheritGrant,
			},
		},
	}

	mock := mock.MockSupabase{Cfg: cfg}
	mock.Activate()
	defer mock.Deactivate()

	err0 := mock.MockUpdateRoleWithExpectedResponse(200)
	assert.NoError(t, err0)

	err1 := supabase.UpdateRole(cfg, localRole, updateParamInheritOnly)
	assert.NoError(t, err1)
}
