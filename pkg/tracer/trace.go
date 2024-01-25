package tracer

import (
	"context"
	"fmt"

	"github.com/valyala/fasthttp"
	trc "go.opentelemetry.io/otel/trace"
)

const (
	TraceIdHeaderKey string = "trace-id"
	SpanIdHeaderKey  string = "span-id"
)

// ---- Trace helper function ----

// extract trace and span id from header, if trace id not exist
// create new span context for start new tracer and return trace span
func Extract(ctx context.Context, tracer trc.Tracer, request *fasthttp.Request) (rCtx context.Context, span trc.Span) {
	traceIdByte := request.Header.Peek(TraceIdHeaderKey)
	spanIdByte := request.Header.Peek(SpanIdHeaderKey)
	spanName := fmt.Sprintf("%s - %s", request.Header.Method(), request.URI().String())

	traceId, spanId := string(traceIdByte), string(spanIdByte)
	if traceId == "" {
		return tracer.Start(ctx, spanName)
	}

	var spanContextConfig trc.SpanContextConfig
	spanContextConfig.TraceID, _ = trc.TraceIDFromHex(traceId)
	spanContextConfig.SpanID, _ = trc.SpanIDFromHex(spanId)
	spanContextConfig.TraceFlags = 01
	spanContextConfig.Remote = true

	spanContext := trc.NewSpanContext(spanContextConfig)
	traceCtx := trc.ContextWithSpanContext(ctx, spanContext)
	return tracer.Start(traceCtx, spanName)
}

// inject trace id and span id to request header and start new tracer
func Inject(ctx context.Context, tracer trc.Tracer, request *fasthttp.Request) (context.Context, trc.Span) {
	spanCtx := trc.SpanContextFromContext(ctx)
	if spanCtx.HasTraceID() {
		request.Header.Add(TraceIdHeaderKey, spanCtx.TraceID().String())
	}

	if spanCtx.HasSpanID() {
		request.Header.Add(SpanIdHeaderKey, spanCtx.SpanID().String())
	}

	spanName := fmt.Sprintf("%s - %s", request.Header.Method(), request.URI().String())
	return tracer.Start(ctx, spanName)
}
