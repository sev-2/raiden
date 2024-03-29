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

func (h *Controller) Get(ctx raiden.Context) error {
	h.Result.Message = fmt.Sprintf("%s - hello %s", h.Payload.Type, h.Payload.Name)
	return ctx.SendJson(h.Result)
}

func TestControllerHandler(t *testing.T) {
	var testResult any
	mockContext := mock.MockContext{
		SendJsonFn: func(data any) error {
			testResult = data
			return nil
		},
		SendErrorWithCodeFn: func(statusCode int, err error) error {
			return nil
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

	err = controller.Get(&mockContext)
	assert.NoError(t, err)
	result, isResult := testResult.(Result)
	assert.Equal(t, true, isResult)
	assert.Equal(t, fmt.Sprintf("%s - hello %s", "greeting", "raiden"), result.Message)
}
