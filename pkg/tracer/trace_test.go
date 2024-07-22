package tracer_test

import (
	"context"
	"testing"

	raidenTracer "github.com/sev-2/raiden/pkg/tracer"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
	trc "go.opentelemetry.io/otel/trace"
)

func TestExtract_NewSpan(t *testing.T) {
	noopProvider := trc.NewNoopTracerProvider()
	tracer := noopProvider.Tracer("test")
	ctx := context.Background()
	req := fasthttp.Request{}

	traceID, _ := trc.TraceIDFromHex("01020304050607080102040810203040")
	spanID, _ := trc.SpanIDFromHex("0102040810203040")

	req.Header.Set(raidenTracer.TraceIdHeaderKey, traceID.String())
	req.Header.Set(raidenTracer.SpanIdHeaderKey, spanID.String())
	req.Header.SetMethod("GET")
	req.SetRequestURI("/test")

	newCtx, span := raidenTracer.Extract(ctx, tracer, &req)
	defer span.End()

	spanContext := trc.SpanContextFromContext(newCtx)
	assert.True(t, spanContext.IsValid(), "Span context should be valid")
	assert.NotEmpty(t, spanContext.TraceID().String(), "Trace ID should not be empty")
	assert.NotEmpty(t, spanContext.SpanID().String(), "Span ID should not be empty")
}

func TestExtract_ExistingSpan(t *testing.T) {
	noopProvider := trc.NewNoopTracerProvider()
	tracer := noopProvider.Tracer("test")
	ctx := context.Background()
	req := fasthttp.Request{}

	traceID, _ := trc.TraceIDFromHex("01020304050607080102040810203040")
	spanID, _ := trc.SpanIDFromHex("0102040810203040")

	req.Header.Set(raidenTracer.TraceIdHeaderKey, traceID.String())
	req.Header.Set(raidenTracer.SpanIdHeaderKey, spanID.String())

	newCtx, span := raidenTracer.Extract(ctx, tracer, &req)
	defer span.End()

	spanContext := trc.SpanContextFromContext(newCtx)
	assert.True(t, spanContext.IsValid(), "Span context should be valid")
	assert.Equal(t, traceID, spanContext.TraceID(), "Trace ID should match the provided Trace ID")
	assert.Equal(t, spanID, spanContext.SpanID(), "Span ID should match the provided Span ID")
}

func TestInject(t *testing.T) {
	noopProvider := trc.NewNoopTracerProvider()
	tracer := noopProvider.Tracer("test")
	ctx := context.Background()
	req := fasthttp.Request{}

	traceID, _ := trc.TraceIDFromHex("01020304050607080102040810203040")
	spanID, _ := trc.SpanIDFromHex("0102040810203040")

	req.Header.Set(raidenTracer.TraceIdHeaderKey, traceID.String())
	req.Header.Set(raidenTracer.SpanIdHeaderKey, spanID.String())

	ctx, span := tracer.Start(ctx, "test-span")
	defer span.End()

	_, newSpan := raidenTracer.Inject(ctx, tracer, &req)
	defer newSpan.End()

	injectedTraceID := req.Header.Peek(raidenTracer.TraceIdHeaderKey)
	injectedSpanID := req.Header.Peek(raidenTracer.SpanIdHeaderKey)

	assert.NotEmpty(t, injectedTraceID, "Trace ID should be injected into request header")
	assert.NotEmpty(t, injectedSpanID, "Span ID should be injected into request header")
}
