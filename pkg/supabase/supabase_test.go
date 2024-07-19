package supabase_test

import (
	"errors"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"

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
	if os.Getenv("TEST_RUN") == "1" {
		cfg := loadCloudConfig()

		expectedProject := objects.Project{Id: "test-project-id", Name: "Test Project"}

		project, err := supabase.FindProject(cfg)
		assert.NoError(t, err)
		assert.Equal(t, expectedProject, project)
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestFindProject_Cloud")
	cmd.Env = append(os.Environ(), "TEST_RUN=1")
	err := cmd.Start()
	assert.NoError(t, err)

	time.Sleep(1 * time.Second)
	err1 := cmd.Process.Signal(syscall.SIGTERM)
	assert.NoError(t, err1)
}

func TestFindProject_SelfHosted(t *testing.T) {
	if os.Getenv("TEST_RUN") == "1" {
		cfg := loadSelfHostedConfig()

		expectedError := errors.New("FindProject not implemented for self hosted")
		project, err := supabase.FindProject(cfg)
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.Equal(t, objects.Project{}, project)
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestFindProject_SelfHosted")
	cmd.Env = append(os.Environ(), "TEST_RUN=1")
	err := cmd.Start()
	assert.NoError(t, err)

	time.Sleep(1 * time.Second)
	err1 := cmd.Process.Signal(syscall.SIGTERM)
	assert.NoError(t, err1)
}
