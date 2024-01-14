package handlers

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/labstack/echo/v4"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/utils"
)

func RegisterProxyHandler(
	e *echo.Echo,
	path string,
	upstreamUrl *url.URL,
	requestInterceptor func(req *http.Request),
	responseInterceptor func(r *http.Response) error,
) {
	e.GET(path, proxyHandler(upstreamUrl, requestInterceptor, responseInterceptor))
}

func proxyHandler(
	url *url.URL,
	requestInterceptor func(req *http.Request),
	responseInterceptor func(r *http.Response) error,
) echo.HandlerFunc {
	return func(c echo.Context) error {
		proxy := httputil.NewSingleHostReverseProxy(url)
		proxy.Director = func(req *http.Request) {
			req.URL.Scheme = url.Scheme
			req.URL.Host = url.Host
			logger.NewLogger().Infof("Proxying to : %s %s\n", utils.GetColoredHttpMethod(req.Method), req.URL.String())

			// intercept request
			if requestInterceptor != nil {
				requestInterceptor(req)
			}
		}

		if responseInterceptor != nil {
			proxy.ModifyResponse = responseInterceptor
		}

		proxy.ServeHTTP(c.Response(), c.Request())
		return nil
	}
}
