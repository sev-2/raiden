package raiden

import (
	"strings"

	"github.com/sev-2/raiden/pkg/utils"
	"github.com/valyala/fasthttp"

	fs_router "github.com/fasthttp/router"
)

type RouteHandlerFn func(ctx Context) Presenter
type MiddlewareFn func(next RouteHandlerFn) RouteHandlerFn

// Build router
func NewRouter(appCtx *context) *router {
	fsRouter := fs_router.New()
	return &router{
		fsRouter:      fsRouter,
		appCtx:        appCtx,
		functionRoute: fsRouter.Group("/functions/v1"),
		modelRoute:    fsRouter.Group("/rest/v1"),
		rpcRoute:      fsRouter.Group("/rest/v1/rpc"),
		realtimeRoute: fsRouter.Group("/realtime/v1"),
		storageRoute:  fsRouter.Group("/storage/v1"),
	}
}

type router struct {
	appCtx        *context
	middlewares   []MiddlewareFn
	fsRouter      *fs_router.Router
	functionRoute *fs_router.Group
	rpcRoute      *fs_router.Group
	modelRoute    *fs_router.Group
	realtimeRoute *fs_router.Group
	storageRoute  *fs_router.Group
}

func (r *router) RegisterControllers(controllers []Controller) {
	for i := range controllers {
		c := controllers[i]
		h := c.Handler

		for _, m := range r.middlewares {
			h = m(h)
		}

		switch c.Options.Type {
		case ControllerTypeFunction:
			r.functionRoute.POST(c.Options.Path, WrapRouteHandler(*r.appCtx, h))
		case ControllerTypeHttpHandler:
			switch strings.ToUpper(c.Options.Method) {
			case fasthttp.MethodGet:
				r.fsRouter.GET(c.Options.Path, WrapRouteHandler(*r.appCtx, h))
			case fasthttp.MethodPost:
				r.fsRouter.POST(c.Options.Path, WrapRouteHandler(*r.appCtx, h))
			case fasthttp.MethodPut:
				r.fsRouter.PUT(c.Options.Path, WrapRouteHandler(*r.appCtx, h))
			case fasthttp.MethodPatch:
				r.fsRouter.PATCH(c.Options.Path, WrapRouteHandler(*r.appCtx, h))
			case fasthttp.MethodDelete:
				r.fsRouter.DELETE(c.Options.Path, WrapRouteHandler(*r.appCtx, h))
			}
		}

	}
}

func (r *router) GetHandler() fasthttp.RequestHandler {
	return r.fsRouter.Handler
}

func (r *router) GetRegisteredRoutes() map[string][]string {
	return r.fsRouter.List()
}

func (r *router) RegisterMiddlewares(middlewares []MiddlewareFn) {
	r.middlewares = append(r.middlewares, middlewares...)
}

func (r *router) PrintRegisteredRoute() {
	registeredRoutes := r.fsRouter.List()
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
func WrapMiddleware(handler RouteHandlerFn, middleware MiddlewareFn) RouteHandlerFn {
	return middleware(handler)
}

func WrapRouteHandler(appContext context, next RouteHandlerFn) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		appContext.RequestCtx = ctx
		presenter := next(&appContext)
		presenter.Write()
	}
}
