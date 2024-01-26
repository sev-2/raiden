package raiden

import (
	"fmt"
	"strings"

	"github.com/sev-2/raiden/pkg/tracer"
	"github.com/valyala/fasthttp"
	"github.com/zeromicro/go-zero/core/breaker"
	"go.opentelemetry.io/otel/codes"
)

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
