package raiden_test

import (
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"

	"github.com/sev-2/raiden"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/trace"
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

func TestNewRouter(t *testing.T) {
	conf := loadConfig()
	router := raiden.NewRouter(conf)
	assert.NotNil(t, router)
}

func TestRouter_SetJobChan_NotNil(t *testing.T) {
	conf := loadConfig()
	router := raiden.NewRouter(conf)
	router.SetJobChan(make(chan raiden.JobParams))
}

func TestRouter_SetTracer_NotNil(t *testing.T) {
	conf := loadConfig()
	router := raiden.NewRouter(conf)
	router.SetTracer(trace.NewNoopTracerProvider().Tracer("raiden"))
}

func TestRouter_RegisterMiddlewares(t *testing.T) {
	conf := loadConfig()
	router := raiden.NewRouter(conf)
	router.RegisterMiddlewares([]raiden.MiddlewareFn{})
}

func TestRouter_AddRoute(t *testing.T) {
	conf := loadConfig()
	router := raiden.NewRouter(conf)
	route := raiden.Route{Methods: []string{"GET", "POST"}, Path: "/"}
	routes := []*raiden.Route{&route}
	router.Register(routes)
}

func TestRouter_GetRegisteredRoutes(t *testing.T) {
	conf := loadConfig()
	router := raiden.NewRouter(conf)
	route := raiden.Route{Methods: []string{"GET", "POST"}, Path: "/"}
	routes := []*raiden.Route{&route}
	router.Register(routes)
	allRegisteredRoutes := router.GetRegisteredRoutes()
	assert.NotNil(t, allRegisteredRoutes)
}

func TestRouter_PrintRoutes(t *testing.T) {
	if os.Getenv("TEST_RUN") == "1" {
		conf := loadConfig()
		router := raiden.NewRouter(conf)
		route := raiden.Route{Methods: []string{"GET", "POST"}, Path: "/"}
		routes := []*raiden.Route{&route}
		router.Register(routes)
		router.PrintRegisteredRoute()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestRouter_PrintRoutes")
	cmd.Env = append(os.Environ(), "TEST_RUN=1")
	err := cmd.Start()
	assert.NoError(t, err)

	time.Sleep(1 * time.Second)
	err1 := cmd.Process.Signal(syscall.SIGTERM)
	assert.NoError(t, err1)
}

func TestRouter_GetHandler(t *testing.T) {
	conf := loadConfig()
	router := raiden.NewRouter(conf)
	route := raiden.Route{Methods: []string{"GET", "POST"}, Path: "/"}
	routes := []*raiden.Route{&route}
	router.Register(routes)
	handler := router.GetHandler()
	assert.NotNil(t, handler)
}
