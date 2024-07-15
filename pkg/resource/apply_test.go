package resource_test

import (
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/resource"
	"github.com/sev-2/raiden/pkg/resource/migrator"
	"github.com/sev-2/raiden/pkg/resource/policies"
	"github.com/sev-2/raiden/pkg/resource/roles"
	"github.com/sev-2/raiden/pkg/resource/rpc"
	"github.com/sev-2/raiden/pkg/resource/storages"
	"github.com/sev-2/raiden/pkg/resource/tables"
	"github.com/sev-2/raiden/pkg/state"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/stretchr/testify/assert"
)

func loadConfig() *raiden.Config {
	// Create a temporary config file with valid content
	file, err := os.CreateTemp("", "config*.yaml")
	if err != nil {
		return nil
	}
	defer os.Remove(file.Name())

	configContent := `
ACCESS_TOKEN: "test-access-token"
ANON_KEY: "test-anon-key"
BREAKER_ENABLE: true
CORS_ALLOWED_ORIGINS: "*"
CORS_ALLOWED_METHODS: "GET,POST"
CORS_ALLOWED_HEADERS: "Content-Type"
CORS_ALLOWED_CREDENTIALS: true
DEPLOYMENT_TARGET: "cloud"
ENVIRONMENT: "production"
PROJECT_ID: "test-project-id"
PROJECT_NAME: "test-project"
SERVICE_KEY: "test-service-key"
SERVER_HOST: "127.0.0.1"
SERVER_PORT: "8080"
SUPABASE_API_URL: "http://test-supabase-api-url"
SUPABASE_API_BASE_PATH: "/api"
SUPABASE_PUBLIC_URL: "http://test-supabase-public-url"
SCHEDULE_STATUS: "on"
TRACE_ENABLE: false
TRACE_COLLECTOR: "zipkin"
TRACE_COLLECTOR_ENDPOINT: "endpoint"
VERSION: "2.0.0"
`
	if _, err := file.WriteString(configContent); err != nil {
		return nil
	}
	file.Close()

	path := file.Name()
	config, err := raiden.LoadConfig(&path)
	if err != nil {
		return nil
	}
	return config
}

func TestApply(t *testing.T) {
	if os.Getenv("TEST_RUN") == "1" {
		flags := &resource.Flags{
			DryRun:        true,
			AllowedSchema: "public",
		}
		config := loadConfig()

		err := resource.Apply(flags, config)
		assert.NoError(t, err)
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestApply")
	cmd.Env = append(os.Environ(), "TEST_RUN=1")
	err := cmd.Start()
	assert.NoError(t, err)

	time.Sleep(1 * time.Second)
	err1 := cmd.Process.Signal(syscall.SIGTERM)
	assert.NoError(t, err1)
}

func TestMigrate(t *testing.T) {
	config := &raiden.Config{}
	importState := &state.LocalState{}
	projectPath := "/path/to/project"
	resources := &resource.MigrateData{
		Tables:   []tables.MigrateItem{},
		Roles:    []roles.MigrateItem{},
		Rpc:      []rpc.MigrateItem{},
		Policies: []policies.MigrateItem{},
		Storages: []storages.MigrateItem{},
	}

	errs := resource.Migrate(config, importState, projectPath, resources)
	assert.Empty(t, errs)
}

func TestUpdateLocalStateFromApply(t *testing.T) {
	projectPath := "/path/to/project"
	localState := &state.LocalState{}
	stateChan := make(chan any)
	done := resource.UpdateLocalStateFromApply(projectPath, localState, stateChan)

	go func() {
		defer close(stateChan)
		stateChan <- &tables.MigrateItem{
			Type:    migrator.MigrateTypeCreate,
			NewData: objects.Table{Name: "test_table"},
		}
	}()

	select {
	case err := <-done:
		assert.NoError(t, err)
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for UpdateLocalStateFromApply to complete")
	}
}

func TestPrintApplyChangeReport(t *testing.T) {
	migrateData := resource.MigrateData{
		Tables: []tables.MigrateItem{
			{Type: migrator.MigrateTypeCreate, NewData: objects.Table{Name: "test_table"}},
		},
	}

	resource.PrintApplyChangeReport(migrateData)
}
