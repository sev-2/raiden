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
