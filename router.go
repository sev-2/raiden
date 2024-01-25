package raiden

import (
	"fmt"
	"strings"

	"github.com/sev-2/raiden/pkg/utils"
	"github.com/valyala/fasthttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	fs_router "github.com/fasthttp/router"
)

type (
	RouteHandlerFn func(ctx Context) Presenter
	MiddlewareFn   func(next RouteHandlerFn) RouteHandlerFn
	RouteType      string
)

const (
	RouteTypeHttpHandler RouteType = "http-handler"
	RouteTypeFunction    RouteType = "function"
	RouteTypeRest        RouteType = "rest"
	RouteTypeRpc         RouteType = "rpc"
	RouteTypeRealtime    RouteType = "realtime"
	RouteTypeStorage     RouteType = "storage"
)

// Build router
func NewRouter(config *Config) *router {
	engine := fs_router.New()
	groups := createRouteGroups(engine)

	var tracer trace.Tracer
	if config.TraceEnable {
		tracer = otel.Tracer(fmt.Sprintf("%s tracer", config.ProjectName))
	}

	return &router{
		engine: engine,
		config: config,
		groups: groups,
		tracer: tracer,
	}
}

type router struct {
	config      *Config
	engine      *fs_router.Router
	groups      map[RouteType]*fs_router.Group
	middlewares []MiddlewareFn
	controllers []*Controller
	tracer      trace.Tracer
}

func (r *router) RegisterMiddlewares(middlewares []MiddlewareFn) *router {
	r.middlewares = append(r.middlewares, middlewares...)
	return r
}

func (r *router) RegisterControllers(controllers []*Controller) *router {
	r.controllers = append(r.controllers, controllers...)
	return r
}

func (r *router) BuildHandler() {
	for _, c := range r.controllers {
		if c == nil {
			continue
		}

		switch c.Options.Type {
		case RouteTypeFunction:
			r.bindFunctionRoute(c)
		case RouteTypeHttpHandler:
			r.bindHttpHandlerRoute(c)
		case RouteTypeRest, RouteTypeRpc, RouteTypeRealtime, RouteTypeStorage:
			Infof("register route type %v is not implemented, wait for update :) ", c.Options.Type)
		}
	}
}

func (r *router) buildNativeMiddleware(c *Controller) RouteHandlerFn {
	handler := c.Handler
	if r.config.TraceEnable {
		handler = TraceMiddleware(handler)
	}

	if r.config.BreakerEnable {
		handler = BreakerMiddleware(c.Options.Method, c.Options.Path, handler)
	}

	return handler
}

func (r *router) buildAppMiddleware(handler RouteHandlerFn) RouteHandlerFn {
	for _, m := range r.middlewares {
		handler = m(handler)
	}
	return handler
}

func (r *router) findRouteGroup(routeType RouteType) *fs_router.Group {
	return r.groups[routeType]
}

func (r *router) bindFunctionRoute(c *Controller) {
	if group := r.findRouteGroup(c.Options.Type); group != nil {
		handler := r.buildNativeMiddleware(c)
		if len(r.middlewares) > 0 {
			handler = r.buildAppMiddleware(handler)
		}

		group.POST(c.Options.Path, WrapRouteHandler(r.config, r.tracer, handler))
	}
}

func (r *router) bindHttpHandlerRoute(c *Controller) {
	handler := r.buildNativeMiddleware(c)
	if len(r.middlewares) > 0 {
		handler = r.buildAppMiddleware(handler)
	}

	switch strings.ToUpper(c.Options.Method) {
	case fasthttp.MethodGet:
		r.engine.GET(c.Options.Path, WrapRouteHandler(r.config, r.tracer, handler))
	case fasthttp.MethodPost:
		r.engine.POST(c.Options.Path, WrapRouteHandler(r.config, r.tracer, handler))
	case fasthttp.MethodPut:
		r.engine.PUT(c.Options.Path, WrapRouteHandler(r.config, r.tracer, handler))
	case fasthttp.MethodPatch:
		r.engine.PATCH(c.Options.Path, WrapRouteHandler(r.config, r.tracer, handler))
	case fasthttp.MethodDelete:
		r.engine.DELETE(c.Options.Path, WrapRouteHandler(r.config, r.tracer, handler))
	}
}

func (r *router) GetHandler() fasthttp.RequestHandler {
	return r.engine.Handler
}

func (r *router) GetRegisteredRoutes() map[string][]string {
	return r.engine.List()
}

func (r *router) PrintRegisteredRoute() {
	registeredRoutes := r.engine.List()
	Infof("%s Registered Route %s ", strings.Repeat("=", 11), strings.Repeat("=", 11))
	for method, routes := range registeredRoutes {
		Infof("%s", utils.GetColoredHttpMethod(method))
		for _, route := range routes {
			Infof("- %s", route)
		}
	}
	Info(strings.Repeat("=", 40))
}

// ----- helper function -----
func createRouteGroups(engine *fs_router.Router) map[RouteType]*fs_router.Group {
	return map[RouteType]*fs_router.Group{
		RouteTypeFunction: engine.Group("/functions/v1"),
		RouteTypeRest:     engine.Group("/rest/v1"),
		RouteTypeRpc:      engine.Group("/rest/v1/rpc"),
		RouteTypeRealtime: engine.Group("/realtime/v1"),
		RouteTypeStorage:  engine.Group("/storage/v1"),
	}
}

func WrapMiddleware(handler RouteHandlerFn, middleware MiddlewareFn) RouteHandlerFn {
	return middleware(handler)
}

func WrapRouteHandler(config *Config, tracer trace.Tracer, next RouteHandlerFn) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		appContext := &context{
			RequestCtx: ctx,
			config:     config,
			tracer:     tracer,
		}
		presenter := next(appContext)
		presenter.Write()
	}
}
