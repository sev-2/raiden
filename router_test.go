package raiden_test

import (
	"log"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"

	"github.com/sev-2/raiden"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/trace"
)

type SomeModel struct {
	raiden.ModelBase
	Metadata string `tableName:"some_model"`
}

type HelloWorldRequest struct {
}

type HelloWorldResponse struct {
	Message string `json:"message"`
}

type HelloWorldController struct {
	raiden.ControllerBase
	Http    string `path:"/hello/{name}" type:"rest"`
	Payload *HelloWorldRequest
	Result  HelloWorldResponse
	Model   SomeModel
}

func loadConfig() *raiden.Config {
	return &raiden.Config{
		DeploymentTarget:    raiden.DeploymentTargetCloud,
		ProjectId:           "test-project-id",
		ProjectName:         "My Great Project",
		SupabaseApiBasePath: "/v1",
		SupabaseApiUrl:      "http://supabase.cloud.com",
	}
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

func TestRouter_BuildHandler(t *testing.T) {
	conf := loadConfig()
	router := raiden.NewRouter(conf)

	rpcRoute := raiden.Route{
		Type:    raiden.RouteTypeRpc,
		Path:    "/some_rpc/",
		Methods: []string{"POST"},
	}

	restRoute := raiden.Route{
		Type:       raiden.RouteTypeRest,
		Path:       "/some_rest/",
		Methods:    []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD"},
		Controller: &HelloWorldController{},
		Model:      &SomeModel{},
	}

	routes := []*raiden.Route{
		&rpcRoute,
		&restRoute,
	}
	router.Register(routes)

	router.BuildHandler()

	registeredRoutes := router.GetRegisteredRoutes()
	log.Printf("REGISTERED_ROUTES: %v", registeredRoutes)
	assert.NotNil(t, registeredRoutes)
}
