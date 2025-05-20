package mock

import (
	"context"
	"net/http"
	"time"

	"github.com/sev-2/raiden"
	"github.com/valyala/fasthttp"
	"go.opentelemetry.io/otel/trace"
)

type MockContext struct {
	*fasthttp.RequestCtx
	CtxFn                func() context.Context
	SetCtxFn             func(ctx context.Context)
	ConfigFn             func() *raiden.Config
	SendJsonFn           func(data any) error
	SendErrorFn          func(err string) error
	SendErrorWithCodeFn  func(statusCode int, err error) error
	RequestContextFn     func() *fasthttp.RequestCtx
	SendRpcFn            func(rpc raiden.Rpc) error
	ExecuteRpcFn         func(rpc raiden.Rpc) (any, error)
	SpanFn               func() trace.Span
	SetSpanFn            func(span trace.Span)
	TracerFn             func() trace.Tracer
	NewJobCtxFn          func() (raiden.JobContext, error)
	WriteFn              func(data []byte)
	WriteErrorFn         func(err error)
	SetFn                func(key string, value any)
	GetFn                func(key string) any
	GetPathFn            func(key string) any
	GetQueryFn           func(key string) string
	Data                 map[string]any
	PublishFn            func(ctx context.Context, provider raiden.PubSubProviderType, topic string, message []byte) error
	HttpRequestFn        func(method string, url string, body []byte, headers map[string]string, timeout time.Duration) (*http.Response, error)
	HttpRequestAndBindFn func(method string, url string, body []byte, headers map[string]string, timeout time.Duration, response any) error
	ResolveLibraryFn     func(key any) error
	RegisterLibrariesFn  func(key map[string]any)
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

func (c *MockContext) SendRpc(rpc raiden.Rpc) error {
	return c.SendRpcFn(rpc)
}

func (c *MockContext) ExecuteRpc(rpc raiden.Rpc) (any, error) {
	return c.ExecuteRpcFn(rpc)
}

func (c *MockContext) SendJson(data any) error {
	return c.SendJsonFn(data)
}

func (c *MockContext) SendError(message string) error {
	return c.SendErrorFn(message)
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

func (c *MockContext) NewJobCtx() (raiden.JobContext, error) {
	return c.NewJobCtxFn()
}

func (c *MockContext) Write(data []byte) {
	c.WriteFn(data)
}

func (c *MockContext) WriteError(err error) {
	c.WriteErrorFn(err)
}

func (c *MockContext) Get(key string) any {
	return c.Data[key]
}

func (c *MockContext) Set(key string, value any) {
	c.Data[key] = value
}

func (c *MockContext) GetParam(key string) any {
	return c.GetPathFn(key)
}

func (c *MockContext) GetQuery(key string) string {
	return c.GetQueryFn(key)
}

func (c *MockContext) Publish(ctx context.Context, provider raiden.PubSubProviderType, topic string, message []byte) error {
	return c.PublishFn(ctx, provider, topic, message)
}

func (c *MockContext) HttpRequest(method string, url string, body []byte, headers map[string]string, timeout time.Duration) (*http.Response, error) {
	return c.HttpRequestFn(method, url, body, headers, timeout)
}

func (c *MockContext) HttpRequestAndBind(method string, url string, body []byte, headers map[string]string, timeout time.Duration, response any) error {
	return c.HttpRequestAndBindFn(method, url, body, headers, timeout, response)
}

func (c *MockContext) ResolveLibrary(key any) error {
	return c.ResolveLibraryFn(key)
}

func (c *MockContext) RegisterLibraries(key map[string]any) {
	c.RegisterLibrariesFn(key)
}
