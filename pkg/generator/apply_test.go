package generator_test

import (
	"os"
	"testing"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/generator"
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

func TestGenerateApplyMainFunction(t *testing.T) {
	dir, err := os.MkdirTemp("", "apply")
	assert.NoError(t, err)

	conf := loadConfig()

	err1 := generator.GenerateApplyMainFunction(dir, conf, generator.GenerateFn(generator.Generate))
	assert.NoError(t, err1)
	assert.FileExists(t, dir+"/cmd/apply/main.go")
}
