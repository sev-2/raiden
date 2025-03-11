package raiden_test

import (
	"log"
	"net/http"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/mock"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

type SomeModel struct {
	raiden.ModelBase
	Metadata string `tableName:"some_model"`
}

type SomeBucket struct {
	raiden.BucketBase
}

func (m *SomeBucket) Name() string {
	return "some_bucket"
}

type HelloWorldRequest struct {
}

type HelloWorldResponse struct {
	Message string `json:"message"`
}

type HelloWorldController struct {
	raiden.ControllerBase
	Http    string `path:"/hello" type:"custom"`
	Payload *HelloWorldRequest
	Result  HelloWorldResponse
	Model   SomeModel
}

func (c *HelloWorldController) Get(ctx raiden.Context) error {
	ctx.Set("method_get", true)
	c.Result.Message = "success get data"
	return ctx.SendJson(c.Result)
}

func (c *HelloWorldController) Post(ctx raiden.Context) error {
	ctx.Set("method_post", true)
	c.Result.Message = "success post data"
	return ctx.SendJson(c.Result)
}

func (c *HelloWorldController) Patch(ctx raiden.Context) error {
	ctx.Set("method_path", true)
	c.Result.Message = "success patch data"
	return ctx.SendJson(c.Result)
}

func (c *HelloWorldController) Put(ctx raiden.Context) error {
	ctx.Set("method_put", true)
	c.Result.Message = "success put data"
	return ctx.SendJson(c.Result)
}

func (c *HelloWorldController) Delete(ctx raiden.Context) error {
	ctx.Set("method_delete", true)
	c.Result.Message = "success delete data"
	return ctx.SendJson(c.Result)
}

type UnimplementedController struct {
	raiden.ControllerBase
	Http string `path:"/unimplemented" type:"custom"`
}

func (*UnimplementedController) AfterGet(ctx raiden.Context) error {
	return http.ErrNotSupported
}

func (*UnimplementedController) BeforePost(ctx raiden.Context) error {
	return http.ErrNotSupported
}

func (*UnimplementedController) AfterPost(ctx raiden.Context) error {
	return http.ErrNotSupported
}

func (*UnimplementedController) BeforePut(ctx raiden.Context) error {
	return http.ErrNotSupported
}

func (*UnimplementedController) AfterPut(ctx raiden.Context) error {
	return http.ErrNotSupported
}

func (*UnimplementedController) BeforePatch(ctx raiden.Context) error {
	return http.ErrNotSupported
}

func (*UnimplementedController) AfterPatch(ctx raiden.Context) error {
	return http.ErrNotSupported
}

func (*UnimplementedController) BeforeDelete(ctx raiden.Context) error {
	return http.ErrNotSupported
}

func (*UnimplementedController) AfterDelete(ctx raiden.Context) error {
	return http.ErrNotSupported
}

func (*UnimplementedController) BeforeOptions(ctx raiden.Context) error {
	return http.ErrNotSupported
}

func (*UnimplementedController) AfterOptions(ctx raiden.Context) error {
	return http.ErrNotSupported
}

func (*UnimplementedController) BeforeHead(ctx raiden.Context) error {
	return http.ErrNotSupported
}

func (*UnimplementedController) AfterHead(ctx raiden.Context) error {
	return http.ErrNotSupported
}

type StorageController struct {
	raiden.ControllerBase
	Http    string `path:"/assets" type:"storage"`
	Storage *SomeBucket
}

func loadConfig() *raiden.Config {
	return &raiden.Config{
		DeploymentTarget:    raiden.DeploymentTargetCloud,
		ProjectId:           "test-project-id",
		ProjectName:         "My Great Project",
		SupabaseApiBasePath: "/v1",
		SupabaseApiUrl:      "http://supabase.cloud.com",
		CorsAllowedOrigins:  "*",
		CorsAllowedMethods:  "GET, POST, PUT, DELETE, OPTIONS",
		CorsAllowedHeaders:  "X-Requested-With, Content-Type, Authorization",
		Mode:                raiden.BffMode,
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
	route := raiden.Route{Methods: []string{"GET", "POST", "PATCH", "DELETE"}, Path: "/"}
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

	storageRoute := raiden.Route{
		Type:       raiden.RouteTypeStorage,
		Path:       "/some_bucket/",
		Methods:    []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD"},
		Controller: &HelloWorldController{},
		Storage:    &SomeBucket{},
	}

	customRoute := raiden.Route{
		Type:    raiden.RouteTypeCustom,
		Path:    "/some_custom/",
		Methods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD"},
	}

	routes := []*raiden.Route{
		&rpcRoute,
		&restRoute,
		&storageRoute,
		&customRoute,
	}
	router.Register(routes)

	router.BuildHandler()

	registeredRoutes := router.GetRegisteredRoutes()
	log.Printf("REGISTERED_ROUTES: %v", registeredRoutes)
	assert.NotNil(t, registeredRoutes)
}

func TestRouter_NewRouteFromCustomController(t *testing.T) {
	methods := []string{fasthttp.MethodGet}
	r := raiden.NewRouteFromController(&HelloWorldController{}, methods)

	assert.Equal(t, raiden.RouteTypeCustom, r.Type)
	assert.Equal(t, "/hello", r.Path)
	assert.Implements(t, (*raiden.Controller)(nil), r.Controller)
	assert.IsType(t, SomeModel{}, r.Model)

	for _, v := range r.Methods {
		assert.Equal(t, methods[0], v)
	}
}

func TestRouter_NewRouteFromStorageController(t *testing.T) {
	methods := []string{fasthttp.MethodGet, fasthttp.MethodPost}
	r := raiden.NewRouteFromController(&StorageController{}, methods)

	assert.Equal(t, raiden.RouteTypeStorage, r.Type)
	assert.Equal(t, "/assets", r.Path)
	assert.Implements(t, (*raiden.Controller)(nil), r.Controller)
	assert.Implements(t, (*raiden.Bucket)(nil), r.Storage)

	assert.Len(t, r.Methods, 2)

	assert.Equal(t, methods[0], r.Methods[0])
	assert.Equal(t, methods[1], r.Methods[1])
}

func TestRoute_BuildAppMiddleware(t *testing.T) {
	conf := loadConfig()
	conf.TraceEnable = true
	router := raiden.NewRouter(conf)

	rpcRoute := raiden.Route{
		Type:    raiden.RouteTypeRpc,
		Path:    "/some_rpc/",
		Methods: []string{"POST"},
	}

	router.RegisterMiddlewares([]raiden.MiddlewareFn{
		func(next raiden.RouteHandlerFn) raiden.RouteHandlerFn {
			return next
		},
	})
	router.Register([]*raiden.Route{&rpcRoute})
	router.BuildHandler()

	mapRoute := router.GetRegisteredRoutes()
	assert.NotNil(t, mapRoute)
}

func Test_Route(t *testing.T) {
	a := raiden.NewChain()
	controller := &HelloWorldController{}

	fsCtx := fasthttp.RequestCtx{}

	mockCtx := &mock.MockContext{
		TracerFn: func() trace.Tracer {
			noopProvider := noop.NewTracerProvider()
			tracer := noopProvider.Tracer("test")
			return tracer
		},
		ConfigFn: func() *raiden.Config {
			return &raiden.Config{
				DeploymentTarget:    raiden.DeploymentTargetCloud,
				ProjectId:           "test-project-id",
				ProjectName:         "My Great Project",
				SupabaseApiBasePath: "/v1",
				SupabaseApiUrl:      "http://supabase.cloud.com",
				SupabasePublicUrl:   "http://supabase.cloud.com",
				CorsAllowedOrigins:  "*",
				CorsAllowedMethods:  "GET, POST, PUT, DELETE, OPTIONS",
				CorsAllowedHeaders:  "X-Requested-With, Content-Type, Authorization",
			}
		},
	}

	fn := a.Then(mockCtx.ConfigFn(), mockCtx.TracerFn(), nil, nil, "GET", raiden.RouteTypeCustom, controller)
	fn(&fsCtx)
	assert.Equal(t, "{\"message\":\"success get data\"}", string(fsCtx.Response.Body()))
	fsCtx.Response.SetBody(nil)

	fn = a.Then(mockCtx.ConfigFn(), mockCtx.TracerFn(), nil, nil, "POST", raiden.RouteTypeCustom, controller)
	fn(&fsCtx)
	assert.Equal(t, "{\"message\":\"success post data\"}", string(fsCtx.Response.Body()))
	fsCtx.Response.SetBody(nil)

	fn = a.Then(mockCtx.ConfigFn(), mockCtx.TracerFn(), nil, nil, "PUT", raiden.RouteTypeCustom, controller)
	fn(&fsCtx)
	assert.Equal(t, "{\"message\":\"success put data\"}", string(fsCtx.Response.Body()))
	fsCtx.Response.SetBody(nil)

	fn = a.Then(mockCtx.ConfigFn(), mockCtx.TracerFn(), nil, nil, "PATCH", raiden.RouteTypeCustom, controller)
	fn(&fsCtx)
	assert.Equal(t, "{\"message\":\"success patch data\"}", string(fsCtx.Response.Body()))
	fsCtx.Response.SetBody(nil)

	fn = a.Then(mockCtx.ConfigFn(), mockCtx.TracerFn(), nil, nil, "DELETE", raiden.RouteTypeCustom, controller)
	fn(&fsCtx)
	assert.Equal(t, "{\"message\":\"success delete data\"}", string(fsCtx.Response.Body()))
	fsCtx.Response.SetBody(nil)

	fn = a.Then(mockCtx.ConfigFn(), mockCtx.TracerFn(), nil, nil, "OPTIONS", raiden.RouteTypeCustom, controller)
	fn(&fsCtx)
	assert.Equal(t, "{\"code\":\"Not Implemented\",\"message\":\"handler not implemented\"}", string(fsCtx.Response.Body()))
	fsCtx.Response.SetBody(nil)

	fn = a.Then(mockCtx.ConfigFn(), mockCtx.TracerFn(), nil, nil, "HEAD", raiden.RouteTypeCustom, controller)
	fn(&fsCtx)
	assert.Equal(t, "{\"code\":\"Not Implemented\",\"message\":\"handler not implemented\"}", string(fsCtx.Response.Body()))
	fsCtx.Response.SetBody(nil)

}

func Test_RouteUnimplemented(t *testing.T) {
	a := raiden.NewChain()
	controller := &UnimplementedController{}
	fsCtx := fasthttp.RequestCtx{}
	mockCtx := &mock.MockContext{
		TracerFn: func() trace.Tracer {
			noopProvider := noop.NewTracerProvider()
			tracer := noopProvider.Tracer("test")
			return tracer
		},
		ConfigFn: func() *raiden.Config {
			return &raiden.Config{
				DeploymentTarget: raiden.DeploymentTargetCloud,
				ProjectId:        "test-project-id",
				ProjectName:      "My Great Project"}
		},
	}

	fn := a.Then(mockCtx.ConfigFn(), mockCtx.TracerFn(), nil, nil, "GET", raiden.RouteTypeCustom, controller)
	fn(&fsCtx)
	assert.Equal(t, "field Payload is not exist in UnimplementedController", string(fsCtx.Response.Body()))
	fsCtx.Response.SetBody(nil)

	fn = a.Then(mockCtx.ConfigFn(), mockCtx.TracerFn(), nil, nil, "POST", raiden.RouteTypeCustom, controller)
	fn(&fsCtx)
	assert.Equal(t, "field Payload is not exist in UnimplementedController", string(fsCtx.Response.Body()))
	fsCtx.Response.SetBody(nil)

	fn = a.Then(mockCtx.ConfigFn(), mockCtx.TracerFn(), nil, nil, "PUT", raiden.RouteTypeCustom, controller)
	fn(&fsCtx)
	assert.Equal(t, "field Payload is not exist in UnimplementedController", string(fsCtx.Response.Body()))
	fsCtx.Response.SetBody(nil)

	fn = a.Then(mockCtx.ConfigFn(), mockCtx.TracerFn(), nil, nil, "PATCH", raiden.RouteTypeCustom, controller)
	fn(&fsCtx)
	assert.Equal(t, "field Payload is not exist in UnimplementedController", string(fsCtx.Response.Body()))
	fsCtx.Response.SetBody(nil)

	fn = a.Then(mockCtx.ConfigFn(), mockCtx.TracerFn(), nil, nil, "DELETE", raiden.RouteTypeCustom, controller)
	fn(&fsCtx)
	assert.Equal(t, "field Payload is not exist in UnimplementedController", string(fsCtx.Response.Body()))
	fsCtx.Response.SetBody(nil)

	fn = a.Then(mockCtx.ConfigFn(), mockCtx.TracerFn(), nil, nil, "OPTIONS", raiden.RouteTypeCustom, controller)
	fn(&fsCtx)
	assert.Equal(t, "field Payload is not exist in UnimplementedController", string(fsCtx.Response.Body()))
	fsCtx.Response.SetBody(nil)

	fn = a.Then(mockCtx.ConfigFn(), mockCtx.TracerFn(), nil, nil, "HEAD", raiden.RouteTypeCustom, controller)
	fn(&fsCtx)
	assert.Equal(t, "field Payload is not exist in UnimplementedController", string(fsCtx.Response.Body()))
	fsCtx.Response.SetBody(nil)

}
