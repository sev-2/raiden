package raiden

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file with valid content
	file, err := os.CreateTemp("", "config*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
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
TRACE_ENABLE: true
TRACE_COLLECTOR: "collector"
TRACE_COLLECTOR_ENDPOINT: "endpoint"
VERSION: "2.0.0"
`
	if _, err := file.WriteString(configContent); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	file.Close()

	path := file.Name()
	config, err := LoadConfig(&path)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if config.ServerHost != "127.0.0.1" {
		t.Errorf("expected server host to be '127.0.0.1', got %s", config.ServerHost)
	}

	if config.ServerPort != "8080" {
		t.Errorf("expected server port to be '8080', got %s", config.ServerPort)
	}
}

func TestLoadConfig_Defaults(t *testing.T) {
	// Create a temporary config file with minimal content
	file, err := os.CreateTemp("", "config*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(file.Name())

	configContent := `
ACCESS_TOKEN: "test-access-token"
`
	if _, err := file.WriteString(configContent); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	file.Close()

	path := file.Name()
	config, err := LoadConfig(&path)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if config.ServerHost != "127.0.0.1" {
		t.Errorf("expected default server host to be '127.0.0.1', got %s", config.ServerHost)
	}

	if config.ServerPort != "8002" {
		t.Errorf("expected default server port to be '8002', got %s", config.ServerPort)
	}

	if config.Version != "1.0.0" {
		t.Errorf("expected default version to be '1.0.0', got %s", config.Version)
	}

	if config.Environment != "development" {
		t.Errorf("expected default environment to be 'development', got %s", config.Environment)
	}

	if config.ScheduleStatus != "off" {
		t.Errorf("expected default schedule status to be 'off', got %s", config.ScheduleStatus)
	}
}

func TestLoadConfig_InvalidFile(t *testing.T) {
	invalidPath := "non_existent_file.yaml"
	_, err := LoadConfig(&invalidPath)
	if err == nil {
		t.Fatalf("expected an error, got nil")
	}
}

func TestLoadConfig_InvalidContent(t *testing.T) {
	// Create a temporary config file with invalid content
	file, err := os.CreateTemp("", "config.yaml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(file.Name())

	invalidContent := `invalid content`
	if _, err := file.WriteString(invalidContent); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	file.Close()

	path := file.Name()
	_, err = LoadConfig(&path)
	if err == nil {
		t.Fatalf("expected an error, got nil")
	}
}
