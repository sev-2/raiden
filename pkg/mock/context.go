package mock

import (
	"context"

	"github.com/sev-2/raiden"
	"github.com/valyala/fasthttp"
	"go.opentelemetry.io/otel/trace"
)

type MockContext struct {
	CtxFn                  func() context.Context
	SetCtxFn               func(ctx context.Context)
	ConfigFn               func() *raiden.Config
	ExecuteRpcFn           func(rpc *raiden.Rpc, data any) error
	ExecuteRpcWithParamsFn func(rpc *raiden.Rpc, params map[string]any, data any) error
	SendJsonFn             func(data any) error
	SendErrorFn            func(err string) error
	SendErrorWithCodeFn    func(statusCode int, err error) error
	RequestContextFn       func() *fasthttp.RequestCtx
	SpanFn                 func() trace.Span
	SetSpanFn              func(span trace.Span)
	TracerFn               func() trace.Tracer
	WriteFn                func(data []byte)
	WriteErrorFn           func(err error)
}

func (c *MockContext) Ctx() context.Context {
	return c.CtxFn()
}

func (c *MockContext) SetCtx(ctx context.Context) {
	c.SetCtxFn(ctx)
}

func (c *MockContext) Config() *raiden.Config {
	return c.ConfigFn()
}

func (c *MockContext) ExecuteRpc(rpc *raiden.Rpc, data any) error {
	return c.ExecuteRpcFn(rpc, data)
}

func (c *MockContext) ExecuteRpcWithParams(rpc *raiden.Rpc, params map[string]any, data any) error {
	return c.ExecuteRpcWithParamsFn(rpc, params, data)
}

func (c *MockContext) SendJson(data any) error {
	return c.SendJsonFn(data)
}

func (c *MockContext) SendError(message string) error {
	return c.SendError(message)
}

func (c *MockContext) SendErrorWithCode(statusCode int, err error) error {
	return c.SendErrorWithCodeFn(statusCode, err)
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

func (c *MockContext) Write(data []byte) {
	c.WriteFn(data)
}

func (c *MockContext) WriteError(err error) {
	c.WriteErrorFn(err)
}
