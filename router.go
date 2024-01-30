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
		Methods    []string
		Path       string
		Controller Controller
	}
)

const (
	RouteTypeCustom   RouteType = "custom"
	RouteTypeFunction RouteType = "function"
	RouteTypeRest     RouteType = "rest"
	RouteTypeRpc      RouteType = "rpc"
	RouteTypeRealtime RouteType = "realtime"
	RouteTypeStorage  RouteType = "storage"
)

// ----- Route functionality -----

func NewRouter(config *Config) *router {
	engine := fs_router.New()
	groups := createRouteGroups(engine)

	var tracer trace.Tracer
	if config.TraceEnable {
		tracer = otel.Tracer(fmt.Sprintf("%s tracer", config.ProjectName))
	}

	// register native controller
	defaultRoutes := []*Route{
		{
			Type:       RouteTypeCustom,
			Methods:    []string{fasthttp.MethodGet},
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
		case RouteTypeCustom:
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
		chain = chain.Append(BreakerMiddleware(route.Path))
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

func (r *router) bindRouteGroup(group *fs_router.Group, chain Chain, route *Route) {
	for _, m := range route.Methods {
		handler := chain.Then(m, route.Controller)
		switch strings.ToUpper(m) {
		case fasthttp.MethodGet:
			group.GET(route.Path, wrapHandler(r.config, r.tracer, handler))
		case fasthttp.MethodPost:
			group.POST(route.Path, wrapHandler(r.config, r.tracer, handler))
		case fasthttp.MethodPut:
			group.PUT(route.Path, wrapHandler(r.config, r.tracer, handler))
		case fasthttp.MethodPatch:
			group.PATCH(route.Path, wrapHandler(r.config, r.tracer, handler))
		case fasthttp.MethodDelete:
			group.DELETE(route.Path, wrapHandler(r.config, r.tracer, handler))
		case fasthttp.MethodOptions:
			group.OPTIONS(route.Path, wrapHandler(r.config, r.tracer, handler))
		case fasthttp.MethodHead:
			group.HEAD(route.Path, wrapHandler(r.config, r.tracer, handler))
		}
	}
}

func (r *router) bindRoute(chain Chain, route *Route) {
	for _, m := range route.Methods {
		handler := chain.Then(m, route.Controller)
		switch strings.ToUpper(m) {
		case fasthttp.MethodGet:
			r.engine.GET(route.Path, wrapHandler(r.config, r.tracer, handler))
		case fasthttp.MethodPost:
			r.engine.POST(route.Path, wrapHandler(r.config, r.tracer, handler))
		case fasthttp.MethodPut:
			r.engine.PUT(route.Path, wrapHandler(r.config, r.tracer, handler))
		case fasthttp.MethodPatch:
			r.engine.PATCH(route.Path, wrapHandler(r.config, r.tracer, handler))
		case fasthttp.MethodDelete:
			r.engine.DELETE(route.Path, wrapHandler(r.config, r.tracer, handler))
		case fasthttp.MethodOptions:
			r.engine.OPTIONS(route.Path, wrapHandler(r.config, r.tracer, handler))
		case fasthttp.MethodHead:
			r.engine.HEAD(route.Path, wrapHandler(r.config, r.tracer, handler))
		}
	}
}

func (r *router) bindFunctionRoute(route *Route) {
	chain := NewChain()
	if group := r.findRouteGroup(route.Type); group != nil {
		chain = r.buildNativeMiddleware(route, chain)
		if len(r.middlewares) > 0 {
			chain = r.buildAppMiddleware(chain)
		}
		r.bindRouteGroup(group, chain, route)
	}
}

func (r *router) bindHttpHandlerRoute(route *Route) {
	chain := NewChain()
	chain = r.buildNativeMiddleware(route, chain)
	if len(r.middlewares) > 0 {
		chain = r.buildAppMiddleware(chain)
	}

	r.bindRoute(chain, route)
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

func createRouteGroups(engine *fs_router.Router) map[RouteType]*fs_router.Group {
	return map[RouteType]*fs_router.Group{ // available type custom
		RouteTypeFunction: engine.Group("/functions/v1"), // type function
		RouteTypeRest:     engine.Group("/rest/v1"),      // type rest
		RouteTypeRpc:      engine.Group("/rest/v1/rpc"),  // type rpc
		RouteTypeRealtime: engine.Group("/realtime/v1"),  // type realtime
		RouteTypeStorage:  engine.Group("/storage/v1"),   // type storage
	}
}

func wrapHandler(config *Config, tracer trace.Tracer, handler RouteHandlerFn) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		appContext := &context{
			RequestCtx: ctx,
			config:     config,
			tracer:     tracer,
		}
		// execute handler actual handler from controller
		presenter := handler(appContext)
		presenter.Write()
	}
}

// create http handler from controller
// inside http handler will running auto inject and validate payload
// and running lifecycle
func buildHandler(httpMethod string, c Controller) RouteHandlerFn {
	return func(ctx Context) Presenter {
		// marshall and validate http request data
		// mush return error if `Payload` field is not define in controller
		if err := MarshallAndValidate(ctx.RequestContext(), c); err != nil {
			return ctx.SendJsonError(err)
		}

		var presenter Presenter

		switch httpMethod {
		case fasthttp.MethodGet:
			if err := c.BeforeGet(ctx); err != nil {
				return ctx.SendJsonError(err)
			}

			presenter = c.Get(ctx)

			if err := c.AfterGet(ctx); err != nil {
				Error(err)
			}
		case fasthttp.MethodPost:
			if err := c.BeforePost(ctx); err != nil {
				return ctx.SendJsonError(err)
			}

			presenter = c.Post(ctx)

			if err := c.AfterPost(ctx); err != nil {
				Error(err)
			}
		case fasthttp.MethodPut:
			if err := c.BeforePut(ctx); err != nil {
				return ctx.SendJsonError(err)
			}

			presenter = c.Put(ctx)

			if err := c.AfterPut(ctx); err != nil {
				Error(err)
			}
		case fasthttp.MethodPatch:
			if err := c.BeforePatch(ctx); err != nil {
				return ctx.SendJsonError(err)
			}

			presenter = c.Patch(ctx)

			if err := c.AfterPatch(ctx); err != nil {
				Error(err)
			}
		case fasthttp.MethodDelete:
			if err := c.BeforeDelete(ctx); err != nil {
				return ctx.SendJsonError(err)
			}

			presenter = c.Delete(ctx)

			if err := c.AfterDelete(ctx); err != nil {
				Error(err)
			}
		case fasthttp.MethodOptions:
			if err := c.BeforeOptions(ctx); err != nil {
				return ctx.SendJsonError(err)
			}

			presenter = c.Options(ctx)

			if err := c.AfterOptions(ctx); err != nil {
				Error(err)
			}
		case fasthttp.MethodHead:
			if err := c.BeforeHead(ctx); err != nil {
				return ctx.SendJsonError(err)
			}

			presenter = c.Head(ctx)

			if err := c.AfterHead(ctx); err != nil {
				Error(err)
			}
		}

		// return presenter
		return presenter
	}

}
