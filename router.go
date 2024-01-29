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

// ----- define route type, constant and variable -----
type (
	RouteHandlerFn func(ctx Context) Presenter
	RouteType      string

	Route struct {
		Type       RouteType
		Method     string
		Path       string
		Controller Controller
	}
)

const (
	RouteTypeHttpHandler RouteType = "http-handler"
	RouteTypeFunction    RouteType = "function"
	RouteTypeRest        RouteType = "rest"
	RouteTypeRpc         RouteType = "rpc"
	RouteTypeRealtime    RouteType = "realtime"
	RouteTypeStorage     RouteType = "storage"
)

// ----- Route functionality -----

func NewRouter(config *Config) *router {
	engine := fs_router.New()
	groups := createRouteGroups(engine)

	var tracer trace.Tracer
	if config.TraceEnable {
		tracer = otel.Tracer(fmt.Sprintf("%s tracer", config.ProjectName))
	}

	defaultRoutes := []*Route{
		{
			Type:       RouteTypeHttpHandler,
			Method:     fasthttp.MethodGet,
			Path:       "/health",
			Controller: &HealthController{},
		},
	}

	return &router{
		engine: engine,
		config: config,
		groups: groups,
		tracer: tracer,
		routes: defaultRoutes,
	}
}

type router struct {
	config      *Config
	engine      *fs_router.Router
	groups      map[RouteType]*fs_router.Group
	middlewares []MiddlewareFn
	routes      []*Route
	tracer      trace.Tracer
}

func (r *router) RegisterMiddlewares(middlewares []MiddlewareFn) *router {
	r.middlewares = append(r.middlewares, middlewares...)
	return r
}

func (r *router) Register(routes []*Route) *router {
	r.routes = append(r.routes, routes...)
	return r
}

func (r *router) BuildHandler() {
	for _, route := range r.routes {
		if route == nil {
			continue
		}

		switch route.Type {
		case RouteTypeFunction:
			r.bindFunctionRoute(route)
		case RouteTypeHttpHandler:
			r.bindHttpHandlerRoute(route)
		case RouteTypeRest, RouteTypeRpc, RouteTypeRealtime, RouteTypeStorage:
			Infof("register route type %v is not implemented, wait for update :) ", route.Type)
		}
	}
}

func (r *router) buildNativeMiddleware(route *Route, chain Chain) Chain {
	if r.config.TraceEnable {
		chain = chain.Append(TraceMiddleware)
	}

	if r.config.BreakerEnable {
		chain = chain.Append(BreakerMiddleware(route.Method, route.Path))
	}

	return chain
}

func (r *router) buildAppMiddleware(chain Chain) Chain {
	for _, m := range r.middlewares {
		chain.Append(m)
	}
	return chain
}

func (r *router) findRouteGroup(routeType RouteType) *fs_router.Group {
	return r.groups[routeType]
}

func (r *router) bindFunctionRoute(route *Route) {
	chain := NewChain()
	if group := r.findRouteGroup(route.Type); group != nil {
		chain = r.buildNativeMiddleware(route, chain)
		if len(r.middlewares) > 0 {
			chain = r.buildAppMiddleware(chain)
		}

		handler := chain.Then(route.Controller)
		group.POST(route.Path, wrapRouteHandler(r.config, r.tracer, handler))
	}
}

func (r *router) bindHttpHandlerRoute(route *Route) {
	chain := NewChain()
	chain = r.buildNativeMiddleware(route, chain)
	if len(r.middlewares) > 0 {
		chain = r.buildAppMiddleware(chain)
	}

	handler := chain.Then(route.Controller)

	switch strings.ToUpper(route.Method) {
	case fasthttp.MethodGet:
		r.engine.GET(route.Path, wrapRouteHandler(r.config, r.tracer, handler))
	case fasthttp.MethodPost:
		r.engine.POST(route.Path, wrapRouteHandler(r.config, r.tracer, handler))
	case fasthttp.MethodPut:
		r.engine.PUT(route.Path, wrapRouteHandler(r.config, r.tracer, handler))
	case fasthttp.MethodPatch:
		r.engine.PATCH(route.Path, wrapRouteHandler(r.config, r.tracer, handler))
	case fasthttp.MethodDelete:
		r.engine.DELETE(route.Path, wrapRouteHandler(r.config, r.tracer, handler))
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

func wrapRouteHandler(config *Config, tracer trace.Tracer, handler RouteHandlerFn) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		appContext := &context{
			RequestCtx: ctx,
			config:     config,
			tracer:     tracer,
		}
		// get and execute handler
		presenter := handler(appContext)
		presenter.Write()
	}
}

// create http handler from controller
// inside http handler will running auto inject and validate payload
// and running lifecycle
func buildHandler(c Controller) RouteHandlerFn {
	return func(ctx Context) Presenter {
		// marshall and validate http request data
		// mush return error if `Payload` field is not define in controller
		if err := MarshallAndValidate(ctx.RequestContext(), c); err != nil {
			return ctx.SendJsonError(err)
		}

		// running before middleware
		if err := c.Before(ctx); err != nil {
			return ctx.SendJsonError(err)
		}

		// running http handler
		presenter := c.Handler(ctx)

		// running after middleware
		if err := c.After(ctx); err != nil {
			Error(err)
		}

		// return presenter
		return presenter
	}

}
