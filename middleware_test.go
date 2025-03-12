package raiden_test

import (
	"context"
	"testing"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/mock"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
	"go.opentelemetry.io/otel/trace"
)

func TestNewChain(t *testing.T) {
	c := raiden.NewChain(m1, m2)

	assert.NotNil(t, c)
}

func TestChain_Append(t *testing.T) {
	c := raiden.NewChain(m1, m2)

	assert.NotNil(t, c.Append(m3, m4))
}

func TestChain_Prepend(t *testing.T) {
	c := raiden.NewChain(m3, m4)

	assert.NotNil(t, c.Prepend(m1, m2))
}

func TestChain_Then(t *testing.T) {
	// Create a new chain with two middlewares
	c := raiden.NewChain(m1, m2)

	// setup required data
	controller := &HelloWorldController{}

	// Call Then
	fn := c.Then(nil, nil, nil, nil, "GET", raiden.RouteTypeCustom, controller, nil)

	// Test the returned function
	assert.NotNil(t, fn)
}

func Test_Tracer(t *testing.T) {
	a := raiden.NewChain(m1, m2)

	breakerMiddleware := raiden.BreakerMiddleware("/some-path")

	a.Append(breakerMiddleware)

	controller := &HelloWorldController{}

	fn := a.Then(nil, nil, nil, nil, "GET", raiden.RouteTypeCustom, controller, nil)
	assert.NotNil(t, fn)

	mockCtx := &mock.MockContext{
		CtxFn: func() context.Context {
			return context.Background()
		},
		SetCtxFn:  func(ctx context.Context) {},
		SetSpanFn: func(span trace.Span) {},
		SendErrorWithCodeFn: func(statusCode int, err error) error {
			return nil
		},
		TracerFn: func() trace.Tracer {
			noopProvider := trace.NewNoopTracerProvider()
			tracer := noopProvider.Tracer("test")
			return tracer
		},
		ConfigFn: func() *raiden.Config {
			return &raiden.Config{
				DeploymentTarget:    raiden.DeploymentTargetCloud,
				ProjectId:           "test-project-id",
				ProjectName:         "My Great Project",
				SupabaseApiBasePath: "/v1",
				SupabaseApiUrl:      "http://supabase.cloud.com",
				SupabasePublicUrl:   "http://supabase.cloud.com",
				CorsAllowedOrigins:  "*",
				CorsAllowedMethods:  "GET, POST, PUT, DELETE, OPTIONS",
				CorsAllowedHeaders:  "X-Requested-With, Content-Type, Authorization",
			}
		},
		RequestContextFn: func() *fasthttp.RequestCtx {
			return &fasthttp.RequestCtx{}
		},
		SendJsonFn: func(data any) error {
			return nil
		},
	}

	tracedChain := raiden.TraceMiddleware(func(ctx raiden.Context) error {
		return nil
	})
	assert.NotNil(t, tracedChain)

	res := tracedChain(mockCtx)
	assert.Nil(t, res)

	corsFn := raiden.CorsMiddleware(mockCtx.Config())
	corsFn(&fasthttp.RequestCtx{
		Request: fasthttp.Request{
			Header: fasthttp.RequestHeader{},
		},
	})
}

func m1(next raiden.RouteHandlerFn) raiden.RouteHandlerFn {
	return func(ctx raiden.Context) error {
		return next(ctx)
	}
}

func m2(next raiden.RouteHandlerFn) raiden.RouteHandlerFn {
	return func(ctx raiden.Context) error {
		return next(ctx)
	}
}

func m3(next raiden.RouteHandlerFn) raiden.RouteHandlerFn {
	return func(ctx raiden.Context) error {
		return next(ctx)
	}
}

func m4(next raiden.RouteHandlerFn) raiden.RouteHandlerFn {
	return func(ctx raiden.Context) error {
		return next(ctx)
	}
}
