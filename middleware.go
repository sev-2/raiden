package raiden

import (
	"github.com/sev-2/raiden/pkg/tracer"
	"github.com/valyala/fasthttp"
	"go.opentelemetry.io/otel/codes"
)

// ----- define core middleware ----

// extract trace id and span id from incoming request
// and create new trace context and span context,
// inject trace context and span context to app context
// capture presenter and set span status
func TraceMiddleware(next RouteHandlerFn) RouteHandlerFn {
	return RouteHandlerFn(func(ctx Context) Presenter {
		r := &ctx.FastHttpRequestContext().Request
		traceCtx, span := tracer.Extract(ctx.Context(), ctx.Tracer(), r)
		defer span.End()

		ctx.SetContext(traceCtx)
		ctx.SetSpan(span)

		presenter := next(ctx)

		resStatusCode := ctx.FastHttpRequestContext().Response.StatusCode()
		if resStatusCode < 200 || resStatusCode > 299 {
			span.SetStatus(codes.Error, fasthttp.StatusMessage(resStatusCode))
			span.RecordError(presenter.GetError())
		} else {
			span.SetStatus(codes.Ok, "request ok")
		}
		return presenter
	})

}
