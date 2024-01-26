package mock

import (
	"context"

	"github.com/sev-2/raiden"
	"github.com/valyala/fasthttp"
	"go.opentelemetry.io/otel/trace"
)

type MockContext struct {
	ContextFn               func() context.Context
	SetContextFn            func(ctx context.Context)
	ConfigFn                func() *raiden.Config
	SendJsonFn              func(data any) raiden.Presenter
	SendJsonErrorFn         func(err error) raiden.Presenter
	SendJsonErrorWithCodeFn func(statusCode int, err error) raiden.Presenter
	RequestContextFn        func() *fasthttp.RequestCtx
	SpanFn                  func() trace.Span
	SetSpanFn               func(span trace.Span)
	TracerFn                func() trace.Tracer
}

func (c *MockContext) Context() context.Context {
	return c.ContextFn()
}

func (c *MockContext) SetContext(ctx context.Context) {
	c.SetContextFn(ctx)
}

func (c *MockContext) Config() *raiden.Config {
	return c.ConfigFn()
}

func (c *MockContext) SendJson(data any) raiden.Presenter {
	return c.SendJsonFn(data)
}

func (c *MockContext) SendJsonError(err error) raiden.Presenter {
	return c.SendJsonErrorFn(err)
}

func (c *MockContext) SendJsonErrorWithCode(statusCode int, err error) raiden.Presenter {
	return c.SendJsonErrorWithCodeFn(statusCode, err)
}

func (c *MockContext) RequestContext() *fasthttp.RequestCtx {
	return c.RequestContextFn()
}

func (c *MockContext) Span() trace.Span {
	return c.SpanFn()
}

func (c *MockContext) SetSpan(span trace.Span) {
	c.SetSpanFn(span)
}

func (c *MockContext) Tracer() trace.Tracer {
	return c.TracerFn()
}
