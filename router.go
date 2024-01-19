package raiden

import (
	"net/http"
	"strings"

	"github.com/fatih/color"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sev-2/raiden/pkg/utils"
)

// Build router
func NewRouter(config *Config) *router {
	Info("Setup New Router")
	r := &router{}
	r.echo = echo.New()
	r.config = config

	// Setup App Context
	applyAppContext(r.echo, config)
	applyDefaultMiddleware(r.echo)

	// register health check
	r.echo.GET("/health", HealthHandler())

	// set reserved group
	r.functionRoute = r.echo.Group("/functions/v1")
	r.modelRoute = r.echo.Group("/rest/v1")
	r.rpcRoute = r.echo.Group("/rest/v1/rpc")
	r.realtimeRoute = r.echo.Group("/realtime/v1")
	r.storageRoute = r.echo.Group("/storage/v1")

	return r
}

type router struct {
	echo   *echo.Echo
	config *Config

	functionRoute *echo.Group
	rpcRoute      *echo.Group
	modelRoute    *echo.Group
	realtimeRoute *echo.Group
	storageRoute  *echo.Group
}

func (r *router) GetHandler() http.Handler {
	return r.echo
}

func (r *router) PrintRegisteredRoute() {
	registeredRoutes := r.echo.Routes()
	Infof("%s Registered Route %s ", strings.Repeat("=", 11), strings.Repeat("=", 11))
	for _, route := range registeredRoutes {
		Infof(" %s - %s", strings.ToUpper(route.Method), route.Path)
	}
	Info(strings.Repeat("=", 40))
}

func applyDefaultMiddleware(e *echo.Echo) {
	// apply log middleware
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus: true,
		LogURI:    true,
		BeforeNextFunc: func(c echo.Context) {
		},
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			requestInformation := color.HiBlackString("host : %v - uri : %v - status : %v ", c.Request().RemoteAddr, v.URI, v.Status)
			Infof("%s %s", utils.GetColoredHttpMethod(c.Request().Method), requestInformation)
			return nil
		},
	}))

	// apply recover middleware
	e.Use(middleware.Recover())
}

func applyAppContext(e *echo.Echo, config *Config) {
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			context := &Context{
				Context: c,
				config:  config,
			}
			return next(context)
		}
	})
}
