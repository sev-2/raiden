package raiden

import (
	"context"
	"fmt"
	"net/url"
	"reflect"
	"strings"

	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/utils"
	"github.com/valyala/fasthttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	fs_router "github.com/fasthttp/router"
)

// ----- define route type, constant and variable -----
type (
	RouteHandlerFn func(ctx Context) error
	RouteType      string

	Route struct {
		Type       RouteType
		Methods    []string
		Path       string
		Controller Controller
		Model      any
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
		if len(route.Methods) == 0 && route.Type != RouteTypeRest {
			Panicf("Unknown method in route path %s", route.Path)
		}

		if route == nil {
			continue
		}

		switch route.Type {
		case RouteTypeFunction, RouteTypeRpc:
			r.registerRpcAndFunctionHandler(route)
		case RouteTypeCustom:
			r.registerHttpHandler(route)
		case RouteTypeRest:
			if route.Model == nil {
				logger.Errorf("[Build Route] invalid route %s, model must be define", route.Path)
			}
			r.registerRestHandler(route)
		case RouteTypeRealtime, RouteTypeStorage:
			Panicf("register route type %v is not implemented, wait for update :) ", route.Type)
		}
	}

	// Proxy auth url
	u, err := url.Parse(r.config.SupabasePublicUrl)
	if err == nil {
		r.engine.ANY("/auth/v1/{path:*}", ProxyHandler(u, "/auth/v1", nil, nil))
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
		handler := chain.Then(m, route.Type, route.Controller)
		switch strings.ToUpper(m) {
		case fasthttp.MethodGet:
			group.GET(route.Path, buildHandler(r.config, r.tracer, handler))
		case fasthttp.MethodPost:
			group.POST(route.Path, buildHandler(r.config, r.tracer, handler))
		case fasthttp.MethodPut:
			group.PUT(route.Path, buildHandler(r.config, r.tracer, handler))
		case fasthttp.MethodPatch:
			group.PATCH(route.Path, buildHandler(r.config, r.tracer, handler))
		case fasthttp.MethodDelete:
			group.DELETE(route.Path, buildHandler(r.config, r.tracer, handler))
		case fasthttp.MethodOptions:
			group.OPTIONS(route.Path, buildHandler(r.config, r.tracer, handler))
		case fasthttp.MethodHead:
			group.HEAD(route.Path, buildHandler(r.config, r.tracer, handler))
		}
	}
}

func (r *router) bindRoute(chain Chain, route *Route) {
	for _, m := range route.Methods {
		handler := chain.Then(m, route.Type, route.Controller)
		switch strings.ToUpper(m) {
		case fasthttp.MethodGet:
			r.engine.GET(route.Path, buildHandler(r.config, r.tracer, handler))
		case fasthttp.MethodPost:
			r.engine.POST(route.Path, buildHandler(r.config, r.tracer, handler))
		case fasthttp.MethodPut:
			r.engine.PUT(route.Path, buildHandler(r.config, r.tracer, handler))
		case fasthttp.MethodPatch:
			r.engine.PATCH(route.Path, buildHandler(r.config, r.tracer, handler))
		case fasthttp.MethodDelete:
			r.engine.DELETE(route.Path, buildHandler(r.config, r.tracer, handler))
		case fasthttp.MethodOptions:
			r.engine.OPTIONS(route.Path, buildHandler(r.config, r.tracer, handler))
		case fasthttp.MethodHead:
			r.engine.HEAD(route.Path, buildHandler(r.config, r.tracer, handler))
		}
	}
}

func (r *router) registerHandler(route *Route) {
	chain := NewChain()
	if group := r.findRouteGroup(route.Type); group != nil {
		chain = r.buildNativeMiddleware(route, chain)
		if len(r.middlewares) > 0 {
			chain = r.buildAppMiddleware(chain)
		}
		r.bindRouteGroup(group, chain, route)
	}
}

func (r *router) registerRpcAndFunctionHandler(route *Route) {
	var routeType string
	if route.Type == RouteTypeFunction {
		routeType = "function "
	} else {
		routeType = "rpc"
	}

	if len(route.Methods) > 1 {
		Panicf(`route %s with type %s,only allowed set 1 method and only allowed post method`, route.Path, routeType)
	}

	if route.Methods[0] != fasthttp.MethodPost {
		Panicf("route %s with type function,only allowed setup with Post method", route.Path)
	}

	chain := NewChain()
	if group := r.findRouteGroup(route.Type); group != nil {
		chain = r.buildNativeMiddleware(route, chain)
		if len(r.middlewares) > 0 {
			chain = r.buildAppMiddleware(chain)
		}
		r.engine.POST(route.Path, buildHandler(
			r.config, r.tracer, chain.Then(fasthttp.MethodPost, route.Type, route.Controller),
		))
	}
}

func (r *router) registerHttpHandler(route *Route) {
	chain := NewChain()
	chain = r.buildNativeMiddleware(route, chain)
	if len(r.middlewares) > 0 {
		chain = r.buildAppMiddleware(chain)
	}

	r.bindRoute(chain, route)
}

func (r *router) registerRestHandler(route *Route) {
	chain := NewChain()
	if group := r.findRouteGroup(route.Type); group != nil {
		chain = r.buildNativeMiddleware(route, chain)
		if len(r.middlewares) > 0 {
			chain = r.buildAppMiddleware(chain)
		}

		rt := reflect.TypeOf(route.Model)
		if rt.Kind() == reflect.Ptr {
			rt = rt.Elem()
		}

		modelName := rt.Name()

		restController := RestController{
			Controller: route.Controller,
			ModelName:  modelName,
		}

		group.GET(route.Path, buildHandler(r.config, r.tracer, chain.Then(fasthttp.MethodGet, route.Type, restController)))
		group.POST(route.Path, buildHandler(r.config, r.tracer, chain.Then(fasthttp.MethodPost, route.Type, restController)))
		group.PUT(route.Path, buildHandler(r.config, r.tracer, chain.Then(fasthttp.MethodPut, route.Type, restController)))
		group.PATCH(route.Path, buildHandler(r.config, r.tracer, chain.Then(fasthttp.MethodPatch, route.Type, restController)))
		group.DELETE(route.Path, buildHandler(r.config, r.tracer, chain.Then(fasthttp.MethodDelete, route.Type, restController)))
	}
}

func (r *router) GetHandler() fasthttp.RequestHandler {
	r.engine.HandleOPTIONS = true
	r.engine.GlobalOPTIONS = CorsMiddleware(r.config)
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

// The function creates and returns a map of route groups based on different route types.
func createRouteGroups(engine *fs_router.Router) map[RouteType]*fs_router.Group {
	return map[RouteType]*fs_router.Group{ // available type custom
		RouteTypeFunction: engine.Group("/functions/v1"), // type function
		RouteTypeRest:     engine.Group("/rest/v1"),      // type rest
		RouteTypeRpc:      engine.Group("/rest/v1/rpc"),  // type rpc
		RouteTypeRealtime: engine.Group("/realtime/v1"),  // type realtime
		RouteTypeStorage:  engine.Group("/storage/v1"),   // type storage
	}
}

// The function "buildHandler" creates a fasthttp.RequestHandler that executes a given RouteHandlerFn
// with a provided Config and trace.Tracer.
func buildHandler(config *Config, tracer trace.Tracer, handler RouteHandlerFn) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		appContext := &Ctx{
			RequestCtx: ctx,
			config:     config,
			tracer:     tracer,
			Context:    context.Background(),
		}
		// execute actual handler from controller
		if err := handler(appContext); err != nil {
			appContext.WriteError(err)
		}
	}
}

// The `createHandleFunc` function creates a route handler function that handles different HTTP methods
// by calling corresponding methods on a controller object.
func createHandleFunc(httpMethod string, routeType RouteType, c Controller) RouteHandlerFn {
	return func(ctx Context) error {

		// marshall and validate http request data
		// will return error if `Payload` field is not define in controller
		if routeType != RouteTypeRest {
			if err := MarshallAndValidate(ctx.RequestContext(), c); err != nil {
				return err
			}
		}

		if err := c.BeforeAll(ctx); err != nil {
			return err
		}

		switch httpMethod {
		case fasthttp.MethodGet:
			if err := c.BeforeGet(ctx); err != nil {
				return err
			}

			if err := c.Get(ctx); err != nil {
				return err
			}

			if err := c.AfterGet(ctx); err != nil {
				Error(err)
			}
		case fasthttp.MethodPost:
			if err := c.BeforePost(ctx); err != nil {
				return err
			}

			if err := c.Post(ctx); err != nil {
				return err
			}

			if err := c.AfterPost(ctx); err != nil {
				Error(err)
			}
		case fasthttp.MethodPut:
			if err := c.BeforePut(ctx); err != nil {
				return err
			}

			if err := c.Put(ctx); err != nil {
				return err
			}

			if err := c.AfterPut(ctx); err != nil {
				Error(err)
			}
		case fasthttp.MethodPatch:
			if err := c.BeforePatch(ctx); err != nil {
				return err
			}

			if err := c.Patch(ctx); err != nil {
				return err
			}

			if err := c.AfterPatch(ctx); err != nil {
				Error(err)
			}
		case fasthttp.MethodDelete:
			if err := c.BeforeDelete(ctx); err != nil {
				return err
			}

			if err := c.Delete(ctx); err != nil {
				return err
			}

			if err := c.AfterDelete(ctx); err != nil {
				Error(err)
			}
		case fasthttp.MethodOptions:
			if err := c.BeforeOptions(ctx); err != nil {
				return err
			}

			if err := c.Options(ctx); err != nil {
				return err
			}

			if err := c.AfterOptions(ctx); err != nil {
				Error(err)
			}
		case fasthttp.MethodHead:
			if err := c.BeforeHead(ctx); err != nil {
				return err
			}

			if err := c.Head(ctx); err != nil {
				return err
			}

			if err := c.AfterHead(ctx); err != nil {
				Error(err)
			}
		}

		if err := c.AfterAll(ctx); err != nil {
			Error(err)
		}

		return nil
	}
}
