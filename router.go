package raiden

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"reflect"
	"strings"

	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/utils"
	"github.com/valyala/fasthttp"
	"go.opentelemetry.io/otel/trace"

	fs_router "github.com/fasthttp/router"
)

var RouterLogger = logger.HcLog().Named("raiden.router")

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
		Storage    Bucket
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
	jobChan     chan JobParams
	pubSub      PubSub
}

func (r *router) SetJobChan(jobChan chan JobParams) {
	r.jobChan = jobChan
}

func (r *router) SetTracer(tracer trace.Tracer) {
	r.tracer = tracer
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
		if len(route.Methods) == 0 && route.Type != RouteTypeRest && route.Type != RouteTypeStorage {
			RouterLogger.Error("unknown method in route path", "path", route.Path)
			os.Exit(1)
		}

		if r.config.Mode == SvcMode && route.Type != RouteTypeCustom {
			RouterLogger.Error("only custom routes are allowed in service mode", "path", route.Path)
			os.Exit(1)
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
				RouterLogger.Error("invalid route, model must be define", "route", route.Path)
			}
			r.registerRestHandler(route)
		case RouteTypeStorage:
			if route.Storage == nil {
				RouterLogger.Error("invalid route,  storage must be define", "route", route.Path)
			}
			r.registerStorageHandler(route)
		case RouteTypeRealtime:
			RouterLogger.Error(fmt.Sprintf("register route type %v is not implemented, wait for update :) ", route.Type))
		}
	}

	// Proxy auth url
	if r.config.Mode == BffMode {
		chain := NewChain()
		if r.config.TraceEnable {
			chain = chain.Append(TraceMiddleware)
		}

		// if len(r.middlewares) > 0 {
		// 	chain = r.buildAppMiddleware(chain)
		// }

		u, err := url.Parse(r.config.SupabasePublicUrl)
		if err == nil {
			r.engine.ANY("/auth/v1/{path:*}", AuthProxy(r.config, chain, nil, nil))
			r.engine.GET("/realtime/v1/websocket", func(ctx *fasthttp.RequestCtx) {
				WebSocketHandler(ctx, u)
			})

			r.engine.POST("/realtime/v1/api/broadcast", func(ctx *fasthttp.RequestCtx) {
				RealtimeBroadcastHandler(ctx, u)
			})
		}
	}

	if r.pubSub != nil {
		pushSubscriptionHandlers := r.pubSub.Handlers()
		if len(pushSubscriptionHandlers) > 0 {
			pubSubGroup := r.engine.Group("/" + SubscriptionPrefixEndpoint)
			for _, pushSubscription := range r.pubSub.Handlers() {
				if pushSubscription.SubscriptionType() == SubscriptionTypePull {
					continue
				}

				pHandler, err := r.pubSub.Serve(pushSubscription)
				if err != nil {
					RouterLogger.Error("serve push subscription", "message", err)
					os.Exit(1)
				}

				var endpoint = pushSubscription.PushEndpoint()
				if !strings.HasPrefix(endpoint, "/") {
					endpoint = "/" + endpoint
				}

				pubSubGroup.POST(endpoint, pHandler)
			}
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
		chain = chain.Append(m)
	}
	return chain
}

func (r *router) findRouteGroup(routeType RouteType) *fs_router.Group {
	return r.groups[routeType]
}

func (r *router) bindRoute(chain Chain, route *Route) {
	for _, m := range route.Methods {
		switch strings.ToUpper(m) {
		case fasthttp.MethodGet:
			r.engine.GET(route.Path, chain.Then(r.config, r.tracer, r.jobChan, r.pubSub, m, route.Type, route.Controller))
		case fasthttp.MethodPost:
			r.engine.POST(route.Path, chain.Then(r.config, r.tracer, r.jobChan, r.pubSub, m, route.Type, route.Controller))
		case fasthttp.MethodPut:
			r.engine.PUT(route.Path, chain.Then(r.config, r.tracer, r.jobChan, r.pubSub, m, route.Type, route.Controller))
		case fasthttp.MethodPatch:
			r.engine.PATCH(route.Path, chain.Then(r.config, r.tracer, r.jobChan, r.pubSub, m, route.Type, route.Controller))
		case fasthttp.MethodDelete:
			r.engine.DELETE(route.Path, chain.Then(r.config, r.tracer, r.jobChan, r.pubSub, m, route.Type, route.Controller))
		case fasthttp.MethodOptions:
			r.engine.OPTIONS(route.Path, chain.Then(r.config, r.tracer, r.jobChan, r.pubSub, m, route.Type, route.Controller))
		case fasthttp.MethodHead:
			r.engine.HEAD(route.Path, chain.Then(r.config, r.tracer, r.jobChan, r.pubSub, m, route.Type, route.Controller))
		}
	}
}

func (r *router) registerRpcAndFunctionHandler(route *Route) {
	var routeType, routePath string
	if route.Type == RouteTypeFunction {
		routeType = "function "
		routePath = strings.TrimPrefix(route.Path, "/functions/v1")
	} else {
		routeType = "rpc"
		routePath = strings.TrimPrefix(route.Path, "/rest/v1/rpc")
	}

	if len(route.Methods) > 1 {
		RouterLogger.Error(`only allowed set 1 method and only allowed post method`, "type", routeType, "path", route.Path)
		os.Exit(1)
	}

	if route.Methods[0] != fasthttp.MethodPost {
		RouterLogger.Error("only allowed setup with Post method", "type", routeType, "path", route.Path)
		os.Exit(1)
	}

	chain := NewChain()
	if group := r.findRouteGroup(route.Type); group != nil {
		chain = r.buildNativeMiddleware(route, chain)
		if len(r.middlewares) > 0 {
			chain = r.buildAppMiddleware(chain)
		}

		group.POST(routePath, chain.Then(r.config, r.tracer, r.jobChan, r.pubSub, fasthttp.MethodPost, route.Type, route.Controller))
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

		restController := RestController{
			Controller: route.Controller,
			Model:      route.Model,
			TableName:  GetTableName(route.Model),
		}

		path := strings.TrimPrefix(route.Path, "/rest/v1")
		group.GET(path, chain.Then(r.config, r.tracer, r.jobChan, r.pubSub, fasthttp.MethodGet, route.Type, restController))
		group.POST(path, chain.Then(r.config, r.tracer, r.jobChan, r.pubSub, fasthttp.MethodPost, route.Type, restController))
		group.PUT(path, chain.Then(r.config, r.tracer, r.jobChan, r.pubSub, fasthttp.MethodPut, route.Type, restController))
		group.PATCH(path, chain.Then(r.config, r.tracer, r.jobChan, r.pubSub, fasthttp.MethodPatch, route.Type, restController))
		group.DELETE(path, chain.Then(r.config, r.tracer, r.jobChan, r.pubSub, fasthttp.MethodDelete, route.Type, restController))
	}
}

func (r *router) registerStorageHandler(route *Route) {
	chain := NewChain()
	if group := r.findRouteGroup(route.Type); group != nil {
		chain = r.buildNativeMiddleware(route, chain)
		if len(r.middlewares) > 0 {
			chain = r.buildAppMiddleware(chain)
		}

		restController := StorageController{
			Controller: route.Controller,
			BucketName: route.Storage.Name(),
			RoutePath:  route.Path,
		}

		path := strings.ReplaceAll(route.Path, "/storage/v1", "/storage/v1/object")
		group.GET(path+"/{path:*}", chain.Then(r.config, r.tracer, r.jobChan, r.pubSub, fasthttp.MethodGet, route.Type, restController))
		group.POST(path+"/{path:*}", chain.Then(r.config, r.tracer, r.jobChan, r.pubSub, fasthttp.MethodPost, route.Type, restController))
		group.PUT(path+"/{path:*}", chain.Then(r.config, r.tracer, r.jobChan, r.pubSub, fasthttp.MethodPut, route.Type, restController))
		group.PATCH(path+"/{path:*}", chain.Then(r.config, r.tracer, r.jobChan, r.pubSub, fasthttp.MethodPatch, route.Type, restController))
		group.DELETE(path+"/{path:*}", chain.Then(r.config, r.tracer, r.jobChan, r.pubSub, fasthttp.MethodDelete, route.Type, restController))
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
	RouterLogger.Info(fmt.Sprintf("%s Registered Route %s ", strings.Repeat("=", 11), strings.Repeat("=", 11)))
	for method, routes := range registeredRoutes {
		RouterLogger.Info(utils.GetColoredHttpMethod(method))
		for _, route := range routes {
			RouterLogger.Info(fmt.Sprintf("- %s", route))
		}
	}
	RouterLogger.Info(strings.Repeat("=", 40))
}

// The function creates and returns a map of route groups based on different route types.
func createRouteGroups(engine *fs_router.Router) map[RouteType]*fs_router.Group {
	return map[RouteType]*fs_router.Group{ // available type custom
		RouteTypeFunction: engine.Group("/functions/v1"),      // type function
		RouteTypeRest:     engine.Group("/rest/v1"),           // type rest
		RouteTypeRpc:      engine.Group("/rest/v1/rpc"),       // type rpc
		RouteTypeRealtime: engine.Group("/realtime/v1"),       // type realtime
		RouteTypeStorage:  engine.Group("/storage/v1/object"), // type storage
	}
}

// The function "buildHandler" creates a fasthttp.RequestHandler that executes a given RouteHandlerFn
// with a provided Config and trace.Tracer.
func buildHandler(config *Config, tracer trace.Tracer, jobChan chan JobParams, pubSub PubSub, handler RouteHandlerFn) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		appContext := &Ctx{
			RequestCtx: ctx,
			config:     config,
			tracer:     tracer,
			Context:    context.Background(),
			jobChan:    jobChan,
			pubSub:     pubSub,
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
		if routeType != RouteTypeRest && routeType != RouteTypeStorage {
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
				logger.HcLog().Error("middleware.after.get", "msg", err.Error())
			}
		case fasthttp.MethodPost:
			if err := c.BeforePost(ctx); err != nil {
				return err
			}

			if err := c.Post(ctx); err != nil {
				return err
			}

			if err := c.AfterPost(ctx); err != nil {
				logger.HcLog().Error("middleware.after.post", "msg", err.Error())
			}
		case fasthttp.MethodPut:
			if err := c.BeforePut(ctx); err != nil {
				return err
			}

			if err := c.Put(ctx); err != nil {
				return err
			}

			if err := c.AfterPut(ctx); err != nil {
				logger.HcLog().Error("middleware.after.put", "msg", err.Error())
			}
		case fasthttp.MethodPatch:
			if err := c.BeforePatch(ctx); err != nil {
				return err
			}

			if err := c.Patch(ctx); err != nil {
				return err
			}

			if err := c.AfterPatch(ctx); err != nil {
				logger.HcLog().Error("middleware.after.patch", "msg", err.Error())
			}
		case fasthttp.MethodDelete:
			if err := c.BeforeDelete(ctx); err != nil {
				return err
			}

			if err := c.Delete(ctx); err != nil {
				return err
			}

			if err := c.AfterDelete(ctx); err != nil {
				logger.HcLog().Error("middleware.after.delete", "msg", err.Error())
			}
		case fasthttp.MethodOptions:
			if err := c.BeforeOptions(ctx); err != nil {
				return err
			}

			if err := c.Options(ctx); err != nil {
				return err
			}

			if err := c.AfterOptions(ctx); err != nil {
				logger.HcLog().Error("middleware.after.options", "msg", err.Error())
			}
		case fasthttp.MethodHead:
			if err := c.BeforeHead(ctx); err != nil {
				return err
			}

			if err := c.Head(ctx); err != nil {
				return err
			}

			if err := c.AfterHead(ctx); err != nil {
				logger.HcLog().Error("middleware.after.head", "msg", err.Error())
			}
		}

		if err := c.AfterAll(ctx); err != nil {
			logger.HcLog().Error("middleware.after.all", "msg", err.Error())
		}

		return nil
	}
}

func NewRouteFromController(controller Controller, methods []string) *Route {
	r := &Route{Controller: controller, Methods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD"}}
	rv := reflect.ValueOf(controller)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	if len(methods) > 0 {
		r.Methods = methods
	}

	// check route type and path
	if sf, exist := rv.Type().FieldByName("Http"); exist {
		// find and assign route type
		rType := sf.Tag.Get("type")
		switch rType {
		case string(RouteTypeFunction):
			r.Type = RouteTypeFunction
		case string(RouteTypeCustom):
			r.Type = RouteTypeCustom
		case string(RouteTypeRpc):
			r.Type = RouteTypeRpc
		case string(RouteTypeRest):
			r.Type = RouteTypeRest
		case string(RouteTypeRealtime):
			r.Type = RouteTypeRealtime
		case string(RouteTypeStorage):
			r.Type = RouteTypeStorage
		}

		// find and assign tag name
		if rPath := sf.Tag.Get("path"); rPath != "" {
			r.Path = rPath
		}
	}

	// // find and assign model
	if modelField, ok := rv.Type().FieldByName("Model"); ok {
		modelType := modelField.Type
		newModel := reflect.New(modelType).Elem()

		r.Model = newModel.Interface()
	}

	// find and assign storage
	if storageField, ok := rv.Type().FieldByName("Storage"); ok {
		storageType := storageField.Type

		if reflect.TypeOf(storageType).Kind() == reflect.Pointer {
			storageType = storageType.Elem()
		}
		newStorage := reflect.New(storageType)
		bucketInterface := reflect.TypeOf((*Bucket)(nil)).Elem()
		if newStorage.Type().Implements(bucketInterface) {
			r.Storage = newStorage.Interface().(Bucket)
		}
	}

	return r
}
