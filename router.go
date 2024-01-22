package raiden

import (
	"strings"

	"github.com/sev-2/raiden/pkg/utils"
	"github.com/valyala/fasthttp"

	fs_router "github.com/fasthttp/router"
)

type RouteHandler func(ctx *Context) Presenter

// Build router
func NewRouter(config *Config) *router {
	Info("Setup New Router")
	r := &router{}
	r.fsRouter = fs_router.New()
	r.config = config

	// set reserved group
	r.functionRoute = r.fsRouter.Group("/functions/v1")
	r.modelRoute = r.fsRouter.Group("/rest/v1")
	r.rpcRoute = r.fsRouter.Group("/rest/v1/rpc")
	r.realtimeRoute = r.fsRouter.Group("/realtime/v1")
	r.storageRoute = r.fsRouter.Group("/storage/v1")

	return r
}

type router struct {
	fsRouter *fs_router.Router
	config   *Config

	functionRoute *fs_router.Group
	rpcRoute      *fs_router.Group
	modelRoute    *fs_router.Group
	realtimeRoute *fs_router.Group
	storageRoute  *fs_router.Group
}

func (r *router) RegisterControllers(controllers []Controller) {
	for i := range controllers {
		c := controllers[i]
		switch c.Options.Type {
		case ControllerTypeFunction:
			r.functionRoute.POST(c.Options.Path, wrapRouteHandler(r.config, &c))
		case ControllerTypeHttpHandler:
			switch strings.ToUpper(c.Options.Method) {
			case fasthttp.MethodGet:
				r.fsRouter.GET(c.Options.Path, wrapRouteHandler(r.config, &c))
			case fasthttp.MethodPost:
				r.fsRouter.POST(c.Options.Path, wrapRouteHandler(r.config, &c))
			case fasthttp.MethodPut:
				r.fsRouter.PUT(c.Options.Path, wrapRouteHandler(r.config, &c))
			case fasthttp.MethodPatch:
				r.fsRouter.PATCH(c.Options.Path, wrapRouteHandler(r.config, &c))
			case fasthttp.MethodDelete:
				r.fsRouter.DELETE(c.Options.Path, wrapRouteHandler(r.config, &c))
			}
		}

	}
}

func (r *router) GetHandler() fasthttp.RequestHandler {
	return r.fsRouter.Handler
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

func wrapRouteHandler(config *Config, controller *Controller) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		appCtx := &Context{
			RequestCtx: ctx,
			config:     config,
		}
		presenter := controller.Handler(appCtx)
		presenter.Write()
	}
}
