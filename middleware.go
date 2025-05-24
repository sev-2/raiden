package raiden

import (
	"context"
	"fmt"
	"strings"

	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/tracer"
	"github.com/valyala/fasthttp"
	"github.com/zeromicro/go-zero/core/breaker"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var MiddlewareLogger = logger.HcLog().Named("raiden.middleware")

// --- define type and constant ----
type (
	MiddlewareFn func(next RouteHandlerFn) RouteHandlerFn
	// Chain defines a chain of middleware.
	Chain interface {
		Append(middlewares ...MiddlewareFn) Chain
		Prepend(middlewares ...MiddlewareFn) Chain
		Then(route *Route, config *Config, tracer trace.Tracer, jobChan chan JobParams, pubSub PubSub, httpMethod string, routeType RouteType, lib map[string]any) fasthttp.RequestHandler
		ServeFsHandle(cfg *Config, fsHandle fasthttp.RequestHandler) fasthttp.RequestHandler
	}

	// chain acts as a list of application middlewares.
	// chain is effectively immutable:
	// once created, it will always hold
	// the same set of middlewares in the same order.
	chain struct {
		middlewares []MiddlewareFn
	}
)

// ----- define chain middleware ----

// This is a modified version of https://github.com/zeromicro/go-zero/blob/master/rest/chain/chain.go
// New creates a new Chain, memorizing the given list of middleware middlewares.
// New serves no other function, middlewares are only called upon a call to Then() or ThenFunc().
func NewChain(middlewares ...MiddlewareFn) Chain {
	return chain{middlewares: append(([]MiddlewareFn)(nil), middlewares...)}
}

// Append extends a chain, adding the specified middlewares as the last ones in the request flow.
//
//	c := chain.New(m1, m2)
//	c.Append(m3, m4)
//	// requests in c go m1 -> m2 -> m3 -> m4
func (c chain) Append(middlewares ...MiddlewareFn) Chain {
	return chain{middlewares: join(c.middlewares, middlewares)}
}

// Prepend extends a chain by adding the specified chain as the first one in the request flow.
//
//	c := chain.New(m3, m4)
//	c1 := chain.New(m1, m2)
//	c.Prepend(c1)
//	// requests in c go m1 -> m2 -> m3 -> m4
func (c chain) Prepend(middlewares ...MiddlewareFn) Chain {
	return chain{middlewares: join(middlewares, c.middlewares)}
}

// Then chains the middleware and returns the final http.Handler.
//
//	New(m1, m2, m3).Then(h)
//
// is equivalent to:
//
//	m1(m2(m3(h)))
//
// When the request comes in, it will be passed to m1, then m2, then m3
// and finally, the given handler
// (assuming every middleware calls the following one).
func (c chain) Then(route *Route, config *Config, tracer trace.Tracer, jobChan chan JobParams, pubSub PubSub, httpMethod string, routeType RouteType, lib map[string]any) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		handler := createHandleFunc(httpMethod, route)
		for i := range c.middlewares {
			handler = c.middlewares[len(c.middlewares)-1-i](handler)
		}

		appContext := &Ctx{
			Context:         context.Background(),
			RequestCtx:      ctx,
			config:          config,
			tracer:          tracer,
			jobChan:         jobChan,
			pubSub:          pubSub,
			libraryRegistry: lib,
		}

		// execute actual handler from controller
		if err := handler(appContext); err != nil {
			appContext.WriteError(err)
		}
	}
}

func (c chain) ServeFsHandle(cfg *Config, next fasthttp.RequestHandler) fasthttp.RequestHandler {
	handler := func(c Context) error {
		next(c.RequestContext())
		return nil
	}

	return func(fsCtx *fasthttp.RequestCtx) {
		appContext := &Ctx{
			config:     cfg,
			RequestCtx: fsCtx,
		}
		for i := range c.middlewares {
			handler = c.middlewares[len(c.middlewares)-1-i](handler)
		}
		if err := handler(appContext); err != nil {
			appContext.WriteError(err)
		}
	}
}

func join(a, b []MiddlewareFn) []MiddlewareFn {
	mids := make([]MiddlewareFn, 0, len(a)+len(b))
	mids = append(mids, a...)
	mids = append(mids, b...)
	return mids
}

// ----- define native middleware ----

// extract trace id and span id from incoming request
// and create new trace context and span context,
// inject trace and span context to request context, and set span status
func TraceMiddleware(next RouteHandlerFn) RouteHandlerFn {
	return func(ctx Context) error {
		if ctx.Tracer() == nil {
			return next(ctx)
		}

		r := &ctx.RequestContext().Request
		traceCtx, span := tracer.Extract(ctx.Ctx(), ctx.Tracer(), r)
		defer span.End()

		ctx.SetCtx(traceCtx)
		ctx.SetSpan(span)

		err := next(ctx)

		resStatusCode := ctx.RequestContext().Response.StatusCode()
		if resStatusCode < 200 || resStatusCode > 399 {
			span.SetStatus(codes.Error, fasthttp.StatusMessage(resStatusCode))
			span.RecordError(err)
		} else {
			span.SetStatus(codes.Ok, "request ok")
		}

		if resStatusCode == fasthttp.StatusServiceUnavailable {
			MiddlewareLogger.Error("status code", "msg", err)
		}

		return err
	}
}

// this middleware is modified version from go-zero (https://github.com/zeromicro/go-zero)
const breakerSeparator = "://"

var breakerMiddleware = logger.HcLog().Named("raiden.middleware.breaker")

// Handler open / close circuit breaker base on request error throttle
func BreakerMiddleware(path string) MiddlewareFn {
	brk := breaker.NewBreaker(breaker.WithName(strings.Join([]string{path}, breakerSeparator)))
	return func(next RouteHandlerFn) RouteHandlerFn {
		return func(ctx Context) error {
			promise, err := brk.Allow()
			if err != nil {
				breakerMiddleware.
					With("uri", string(ctx.RequestContext().RequestURI())).
					With("addr", ctx.RequestContext().RemoteAddr().String()).
					With("user-agent", string(ctx.RequestContext().UserAgent())).
					Error("dropped request")

				err := ErrorResponse{
					StatusCode: fasthttp.StatusServiceUnavailable,
					Code:       "Server Unhealthy",
					Hint:       "circuit breaker open",
					Details:    fmt.Sprintf("open breaker for %s", ctx.RequestContext().RequestURI()),
					Message:    err.Error(),
				}

				return &err
			}

			err = next(ctx)
			resStatusCode := ctx.RequestContext().Response.StatusCode()
			if resStatusCode < fasthttp.StatusInternalServerError {
				promise.Accept()
			} else {
				reason := fmt.Sprintf("%d %s", resStatusCode, fasthttp.StatusMessage(resStatusCode))
				promise.Reject(reason)
			}

			return err
		}
	}
}

// Handle cors
type CorsOptions struct {
	AllowedOrigins     []string
	AllowedMethods     []string
	AllowedHeaders     []string
	AllowCredentials   bool
	OptionsPassthrough bool
}

func CorsMiddleware(config *Config) fasthttp.RequestHandler {
	defaultAllowedMathods := []string{
		fasthttp.MethodGet,
		fasthttp.MethodPost,
		fasthttp.MethodPut,
		fasthttp.MethodPatch,
		fasthttp.MethodDelete,
		fasthttp.MethodOptions,
	}

	corsOptions := CorsOptions{
		AllowedOrigins:     []string{"*"},
		AllowedMethods:     defaultAllowedMathods,
		AllowedHeaders:     []string{},
		AllowCredentials:   false,
		OptionsPassthrough: false,
	}

	if len(config.CorsAllowedOrigins) > 0 {
		allowedOrigin := strings.Split(config.CorsAllowedOrigins, ",")
		if len(allowedOrigin) > 0 {
			corsOptions.AllowedOrigins = allowedOrigin
		}
	}

	if len(config.CorsAllowedMethods) > 0 {
		allowedMethods := strings.Split(config.CorsAllowedMethods, ",")
		if len(allowedMethods) > 0 {
			mapAllowedMethod := map[string]bool{}
			for _, m := range defaultAllowedMathods {
				mapAllowedMethod[m] = true
			}

			corsOptions.AllowedMethods = make([]string, 0)
			for _, m := range allowedMethods {
				if _, ok := mapAllowedMethod[strings.ToUpper(m)]; ok {
					corsOptions.AllowedMethods = append(corsOptions.AllowedMethods, strings.ToUpper(m))
				}
			}
		}
	}

	if len(config.CorsAllowedHeaders) > 0 {
		allowedHeaders := strings.Split(config.CorsAllowedHeaders, ",")
		if len(allowedHeaders) > 0 {
			for _, h := range allowedHeaders {
				canonicalHeader := getCanonicalHeaderKey(h)
				corsOptions.AllowedHeaders = append(corsOptions.AllowedHeaders, canonicalHeader)
			}
		}
	}
	corsOptions.AllowCredentials = config.CorsAllowCredentials

	return func(ctx *fasthttp.RequestCtx) {
		origin := string(ctx.Request.Header.Peek("Origin"))
		if !isValidOrigin(origin, corsOptions.AllowedOrigins) {
			// Return a custom error response for CORS errors
			Info("CORS origin not allowed")
			ctx.Error("CORS origin not allowed", fasthttp.StatusForbidden)
			return
		}

		method := string(ctx.Request.Header.Method())
		if !isValidMethod(method, corsOptions.AllowedMethods) {
			// Return a custom error response for CORS errors
			Info("CORS method not allowed")
			ctx.Error("CORS method not allowed", fasthttp.StatusMethodNotAllowed)
			return
		}

		requestHeaders := string(ctx.Request.Header.Peek("Access-Control-Request-Headers"))
		if !isValidHeaders(requestHeaders, corsOptions.AllowedHeaders) {
			// Return a custom error response for CORS errors
			Info("CORS header not allowed")
			ctx.Error("CORS header not allowed", fasthttp.StatusForbidden)
			return
		}

		responseAllowedOrigin := "*"
		responseAllowedHeader := "*"
		responseAllowedMethod := strings.Join(corsOptions.AllowedMethods, ",")

		if len(config.CorsAllowedOrigins) > 0 {
			responseAllowedOrigin = strings.Join(corsOptions.AllowedOrigins, ",")
		}

		if len(config.CorsAllowedHeaders) > 0 {
			responseAllowedHeader = strings.Join(corsOptions.AllowedHeaders, ",")
		}

		ctx.Response.Header.Set("Access-Control-Allow-Origin", responseAllowedOrigin)
		ctx.Response.Header.Set("Access-Control-Allow-Methods", responseAllowedMethod)
		ctx.Response.Header.Set("Access-Control-Allow-Headers", responseAllowedHeader)
		ctx.Response.Header.Set("Access-Control-Allow-Credentials", fmt.Sprintf("%t", corsOptions.AllowCredentials))

		ctx.Response.Header.Set("Access-Control-Max-Age", "86400")
		ctx.Response.Header.SetStatusCode(fasthttp.StatusNoContent)
	}
}

func isValidOrigin(origin string, allowedOrigins []string) bool {
	for _, allowedOrigin := range allowedOrigins {
		if origin == allowedOrigin || allowedOrigin == "*" {
			return true
		}
	}
	return false
}

func isValidMethod(method string, allowedMethods []string) bool {
	for _, allowedMethod := range allowedMethods {
		if method == allowedMethod {
			return true
		}
	}
	return false
}

func isValidHeaders(requestHeaders string, allowedHeaders []string) bool {
	if len(allowedHeaders) == 0 {
		return true
	}

	headers := strings.Split(requestHeaders, ", ")
	for _, header := range headers {
		if !contains(allowedHeaders, header) {
			return false
		}
	}
	return true
}

func contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}

func getCanonicalHeaderKey(input string) string {
	return strings.ReplaceAll(strings.ToLower(input), " ", "_")
}
