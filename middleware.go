package raiden

import (
	"errors"
	"fmt"
	"strings"

	"github.com/sev-2/raiden/pkg/tracer"
	"github.com/valyala/fasthttp"
	"github.com/zeromicro/go-zero/core/breaker"
	"go.opentelemetry.io/otel/codes"
)

// --- define type and constant ----
type (
	MiddlewareFn func(next RouteHandlerFn) RouteHandlerFn
	// Chain defines a chain of middleware.
	Chain interface {
		Append(middlewares ...MiddlewareFn) Chain
		Prepend(middlewares ...MiddlewareFn) Chain
		Then(httpMethod string, routeType RouteType, fn Controller) RouteHandlerFn
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
func (c chain) Then(httpMethod string, routeType RouteType, controller Controller) RouteHandlerFn {
	handler := createHandleFunc(httpMethod, routeType, controller)
	for i := range c.middlewares {
		handler = c.middlewares[len(c.middlewares)-1-i](handler)
	}
	return handler
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
			Error(err)
		}

		return err
	}
}

// this middleware is modified version from go-zero (https://github.com/zeromicro/go-zero)
const breakerSeparator = "://"

// Handler open / close circuit breaker base on request error throttle
func BreakerMiddleware(path string) MiddlewareFn {
	brk := breaker.NewBreaker(breaker.WithName(strings.Join([]string{path}, breakerSeparator)))
	return func(next RouteHandlerFn) RouteHandlerFn {
		return func(ctx Context) error {
			promise, err := brk.Allow()
			if err != nil {
				Errorf("[http] dropped, %s - %s - %s",
					string(ctx.RequestContext().RequestURI()),
					ctx.RequestContext().RemoteAddr().String(),
					string(ctx.RequestContext().UserAgent()),
				)
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

func CorsMiddleware(next RouteHandlerFn) RouteHandlerFn {
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

	return func(ctx Context) error {

		if len(ctx.Config().CorsAllowedOrigins) > 0 {
			allowedOrigin := strings.Split(ctx.Config().CorsAllowedOrigins, ",")
			if len(allowedOrigin) > 0 {
				corsOptions.AllowedOrigins = allowedOrigin
			}
		}

		if len(ctx.Config().CorsAllowedMethods) > 0 {
			allowedMethods := strings.Split(ctx.Config().CorsAllowedMethods, ",")
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

		if len(ctx.Config().CorsAllowedHeaders) > 0 {
			allowedHeaders := strings.Split(ctx.Config().CorsAllowedHeaders, ",")
			if len(allowedHeaders) > 0 {
				for _, h := range allowedHeaders {
					canonicalHeader := getCanonicalHeaderKey(h)
					corsOptions.AllowedHeaders = append(corsOptions.AllowedHeaders, canonicalHeader)
				}
			}
		}

		corsOptions.AllowCredentials = ctx.Config().CorsAllowCredentials

		origin := string(ctx.RequestContext().Request.Header.Peek("Origin"))
		if !isValidOrigin(origin, corsOptions.AllowedOrigins) {
			return ctx.SendErrorWithCode(fasthttp.StatusForbidden, errors.New("invalid origin"))
		}

		method := string(ctx.RequestContext().Request.Header.Method())
		if !isValidMethod(method, corsOptions.AllowedMethods) {
			return ctx.SendErrorWithCode(fasthttp.StatusForbidden, errors.New("invalid method"))
		}

		requestHeaders := string(ctx.RequestContext().Request.Header.Peek("Access-Control-Request-Headers"))
		if !isValidHeaders(requestHeaders, corsOptions.AllowedHeaders) {
			return ctx.SendErrorWithCode(fasthttp.StatusForbidden, errors.New("invalid headers"))
		}

		ctx.RequestContext().Response.Header.Set("Access-Control-Allow-Origin", origin)
		ctx.RequestContext().Response.Header.Set("Access-Control-Allow-Methods", strings.Join(corsOptions.AllowedMethods, ", "))
		ctx.RequestContext().Response.Header.Set("Access-Control-Allow-Headers", strings.Join(corsOptions.AllowedHeaders, ", "))
		ctx.RequestContext().Response.Header.Set("Access-Control-Allow-Credentials", fmt.Sprintf("%t", corsOptions.AllowCredentials))

		return next(ctx)
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
