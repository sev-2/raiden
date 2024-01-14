package handlers

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/labstack/echo/v4"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/raiden/utils"
)

func RegisterProxyHandler(
	e *echo.Echo,
	path string,
	upstreamUrl *url.URL,
) {
	e.GET(path, proxyHandler(upstreamUrl))
}

func proxyHandler(url *url.URL) echo.HandlerFunc {
	return func(c echo.Context) error {
		proxy := httputil.NewSingleHostReverseProxy(url)
		proxy.Director = func(req *http.Request) {
			req.URL.Scheme = url.Scheme
			req.URL.Host = url.Host
			logger.NewLogger().Infof("Proxying to : %s %s\n", utils.GetColoredHttpMethod(req.Method), req.URL.String())
		}

		proxy.ServeHTTP(c.Response(), c.Request())
		return nil
	}
}
