package supabase_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/supabase"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/stretchr/testify/assert"

	"github.com/jarcoal/httpmock"
)

func loadCloudConfig() *raiden.Config {
	return &raiden.Config{
		DeploymentTarget:    "cloud",
		ProjectId:           "test-project-id",
		ProjectName:         "My Great Project",
		SupabaseApiBasePath: "/v1",
		SupabaseApiUrl:      "http://supabase.cloud.com",
	}
}

func loadSelfHostedConfig() *raiden.Config {
	return &raiden.Config{
		DeploymentTarget:    "self-hosted",
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
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "http://supabase.cloud.com/v1/projects",
		httpmock.NewStringResponder(200, `[{"id": "test-project-id", "name": "My Great Project"}]`))

	cfg := loadCloudConfig()

	project, err := supabase.FindProject(cfg)
	assert.NoError(t, err)
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

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	url := fmt.Sprintf("%s/v1/projects/%s/database/query", cfg.SupabaseApiUrl, cfg.ProjectId)
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

	data, err := json.Marshal(remoteTables)
	assert.NoError(t, err)

	httpmock.RegisterResponder("POST", url,
		httpmock.NewStringResponder(200, string(data)))

	tables, err1 := supabase.GetTables(cfg, []string{"test-schema"})
	assert.NoError(t, err1)
	assert.Equal(t, len(remoteTables), len(tables))
}

func TestGetTables_SelfHosted(t *testing.T) {
	cfg := loadSelfHostedConfig()

	_, err0 := supabase.GetTables(cfg, []string{"test-schema"})
	assert.Error(t, err0)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	url := fmt.Sprintf("%s%s/tables", cfg.SupabaseApiUrl,cfg.SupabaseApiBasePath)
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

	data, err := json.Marshal(remoteTables)
	assert.NoError(t, err)

	httpmock.RegisterResponder("GET", url,
		httpmock.NewStringResponder(200, string(data)))

	tables, err1 := supabase.GetTables(cfg, []string{"test-schema"})
	assert.NoError(t, err1)
	assert.Equal(t, len(remoteTables), len(tables))
}

func TestCreateTable_Cloud(t *testing.T) {
	cfg := loadCloudConfig()

	_, err := supabase.CreateTable(cfg, objects.Table{})
	assert.Error(t, err)
}

func TestCreateTable_SelfHosted(t *testing.T) {
	cfg := loadSelfHostedConfig()

	_, err := supabase.CreateTable(cfg, objects.Table{})
	assert.Error(t, err)
}

func TestUpdateTable_Cloud(t *testing.T) {
	cfg := loadCloudConfig()

	err := supabase.UpdateTable(cfg, objects.Table{}, objects.UpdateTableParam{})
	assert.Error(t, err)
}

func TestUpdateTable_SelfHosted(t *testing.T) {
	cfg := loadSelfHostedConfig()

	err := supabase.UpdateTable(cfg, objects.Table{}, objects.UpdateTableParam{})
	assert.Error(t, err)
}

func TestDeleteTable_Cloud(t *testing.T) {
	cfg := loadCloudConfig()

	err := supabase.DeleteTable(cfg, objects.Table{}, true)
	assert.Error(t, err)
}

func TestDeleteTable_SelfHosted(t *testing.T) {
	cfg := loadSelfHostedConfig()

	err := supabase.DeleteTable(cfg, objects.Table{}, true)
	assert.Error(t, err)
}

func TestGetRoles_Cloud(t *testing.T) {
	cfg := loadCloudConfig()

	_, err := supabase.GetRoles(cfg)
	assert.Error(t, err)
}

func TestGetRoles_SelfHosted(t *testing.T) {
	cfg := loadSelfHostedConfig()

	_, err := supabase.GetRoles(cfg)
	assert.Error(t, err)
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
