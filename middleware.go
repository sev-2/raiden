package raiden

import (
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
		Then(fn Controller) RouteHandlerFn
	}

	// chain acts as a list of http.Handler middlewares.
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
func (c chain) Then(controller Controller) RouteHandlerFn {
	handler := buildHandler(controller)
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
// inject trace context and span context to app context
// capture presenter and set span status
func TraceMiddleware(next RouteHandlerFn) RouteHandlerFn {
	return func(ctx Context) Presenter {
		if ctx.Tracer() == nil {
			return next(ctx)
		}

		r := &ctx.RequestContext().Request
		traceCtx, span := tracer.Extract(ctx.Context(), ctx.Tracer(), r)
		defer span.End()

		ctx.SetContext(traceCtx)
		ctx.SetSpan(span)

		presenter := next(ctx)

		resStatusCode := ctx.RequestContext().Response.StatusCode()
		if resStatusCode < 200 || resStatusCode > 399 {
			span.SetStatus(codes.Error, fasthttp.StatusMessage(resStatusCode))
			span.RecordError(presenter.GetError())
		} else {
			span.SetStatus(codes.Ok, "request ok")
		}

		if resStatusCode == fasthttp.StatusServiceUnavailable {
			Error(presenter.GetError())
		}
		return presenter
	}
}

// this middleware is modified version from go-zero (https://github.com/zeromicro/go-zero)
const breakerSeparator = "://"

// handler forward to handler base on request error throttle
func BreakerMiddleware(method string, path string) MiddlewareFn {
	brk := breaker.NewBreaker(breaker.WithName(strings.Join([]string{method, path}, breakerSeparator)))
	return func(next RouteHandlerFn) RouteHandlerFn {
		return func(ctx Context) Presenter {
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

				presenter := NewJsonPresenter(&ctx.RequestContext().Response)
				presenter.SetError(&err)
				return presenter
			}

			presenter := next(ctx)
			resStatusCode := ctx.RequestContext().Response.StatusCode()
			if resStatusCode < fasthttp.StatusInternalServerError {
				promise.Accept()
			} else {
				reason := fmt.Sprintf("%d %s", resStatusCode, fasthttp.StatusMessage(resStatusCode))
				promise.Reject(reason)
			}

			return presenter
		}
	}
}
