package supabase_test

import (
	"errors"
	"testing"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/supabase"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/stretchr/testify/assert"
)

func loadCloudConfig() *raiden.Config {
	return &raiden.Config{
		DeploymentTarget:    "cloud",
		ProjectId:           "test-project-id",
		SupabaseApiBasePath: "http://localhost:8080",
		SupabaseApiUrl:      "http://localhost:8080",
	}
}

func loadSelfHostedConfig() *raiden.Config {
	return &raiden.Config{
		DeploymentTarget:    "self-hosted",
		ProjectId:           "test-project-local-id",
		SupabaseApiBasePath: "http://localhost:8080",
		SupabaseApiUrl:      "http://localhost:8080",
	}
}

func TestFindProject_Cloud(t *testing.T) {
	cfg := loadCloudConfig()

	_, err := supabase.FindProject(cfg)
	assert.Error(t, err)
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

	_, err := supabase.GetTables(cfg, []string{"test-schema"})
	assert.Error(t, err)
}

func TestGetTables_SelfHosted(t *testing.T) {
	cfg := loadSelfHostedConfig()

	_, err := supabase.GetTables(cfg, []string{"test-schema"})
	assert.Error(t, err)
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
