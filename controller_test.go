package raiden_test

import (
	"fmt"
	"testing"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/mock"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

type Payload struct {
	Type string `query:"type"`
	Name string `json:"name" validate:"required"`
}

type Result struct {
	Message string `json:"message"`
}

type Controller struct {
	raiden.ControllerBase

	// Reserver Field
	Payload *Payload
	Result  Result
}

func (h *Controller) Handler(ctx raiden.Context) raiden.Presenter {
	h.Result.Message = fmt.Sprintf("%s - hello %s", h.Payload.Type, h.Payload.Name)
	return ctx.SendJson(h.Result)
}

func TestControllerHandler(t *testing.T) {
	mockContext := mock.MockContext{
		SendJsonFn: func(data any) raiden.Presenter {
			presenter := raiden.NewJsonPresenter(&fasthttp.Response{})
			presenter.SetData(data)
			return presenter

		},
		SendJsonErrorWithCodeFn: func(statusCode int, err error) raiden.Presenter {
			presenter := raiden.NewJsonPresenter(&fasthttp.Response{})
			presenter.SetError(err)
			return presenter

		},
	}

	// setup required data
	controller := &Controller{}

	// setup request
	jsonBodyString := "{\"name\":\"raiden\"}"
	requestCtx := &fasthttp.RequestCtx{
		Request: fasthttp.Request{},
	}
	requestCtx.Request.SetBodyRaw([]byte(jsonBodyString))
	requestCtx.Request.URI().QueryArgs().Set("type", "greeting")

	// inject request
	err := raiden.MarshallAndValidate(requestCtx, controller)
	assert.NoError(t, err)

	presenter := controller.Handler(&mockContext)
	data := presenter.GetData()
	result, isResult := data.(Result)
	assert.Equal(t, true, isResult)
	assert.Equal(t, fmt.Sprintf("%s - hello %s", "greeting", "raiden"), result.Message)
}
