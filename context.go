package raiden

import (
	go_ctx "context"

	"github.com/valyala/fasthttp"
	"go.opentelemetry.io/otel/trace"
)

type (
	Context interface {
		Context() go_ctx.Context
		SetContext(ctx go_ctx.Context)

		Config() *Config

		SendJson(data any) Presenter
		SendJsonError(err error) Presenter
		SendJsonErrorWithCode(statusCode int, err error) Presenter

		RequestContext() *fasthttp.RequestCtx

		Span() trace.Span
		SetSpan(span trace.Span)

		Tracer() trace.Tracer
	}

	context struct {
		*fasthttp.RequestCtx
		config   *Config
		rContext go_ctx.Context
		span     trace.Span
		tracer   trace.Tracer
	}
)

func (c *context) Config() *Config {
	return c.config
}

func (c *context) RequestContext() *fasthttp.RequestCtx {
	return c.RequestCtx
}

func (c *context) Span() trace.Span {
	return c.span
}

func (c *context) SetSpan(span trace.Span) {
	c.span = span
}

func (c *context) Tracer() trace.Tracer {
	return c.tracer
}

func (c *context) Context() go_ctx.Context {
	return c.rContext
}

func (c *context) SetContext(ctx go_ctx.Context) {
	c.rContext = ctx
}

func (c *context) SendJson(data any) Presenter {
	presenter := NewJsonPresenter(&c.Response)
	presenter.SetData(data)
	return presenter
}

func (c *context) SendJsonError(err error) Presenter {
	presenter := NewJsonPresenter(&c.Response)
	if _, ok := err.(*ErrorResponse); !ok {
		errResponse := &ErrorResponse{
			Message:    err.Error(),
			StatusCode: fasthttp.StatusInternalServerError,
		}
		err = errResponse
	}

	presenter.SetError(err)
	return presenter
}

func (c *context) SendJsonErrorWithCode(statusCode int, err error) Presenter {
	presenter := NewJsonPresenter(&c.Response)
	if errResponse, ok := err.(*ErrorResponse); ok {
		errResponse.StatusCode = statusCode
		err = errResponse
	} else {
		errResponse := &ErrorResponse{
			Message:    err.Error(),
			StatusCode: statusCode,
			Code:       fasthttp.StatusMessage(statusCode),
		}
		err = errResponse
	}

	presenter.SetError(err)
	return presenter
}
