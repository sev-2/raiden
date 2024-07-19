package raiden_test

import (
	"encoding/json"
	"errors"
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

	// Reserved Field
	Payload *Payload
	Result  Result
}

func (h *Controller) Get(ctx raiden.Context) error {
	h.Result.Message = fmt.Sprintf("%s - hello %s", h.Payload.Type, h.Payload.Name)
	return ctx.SendJson(h.Result)
}

func (h *Controller) Post(ctx raiden.Context) error {
	h.Result.Message = fmt.Sprintf("Data: %v", ctx.RequestContext().Request.String())
	return ctx.SendJson(h.Result)
}

type DataController struct {
	raiden.ControllerBase

	// Reserved Field
	Payload map[string]any
	Result  map[string]any
}

func (h *DataController) BeforeGet(ctx raiden.Context) error {
	ctx.Set("message", "from before get middleware")
	return nil
}

func (h *DataController) Get(ctx raiden.Context) error {
	msg := ctx.Get("message")
	msg, isString := msg.(string)
	if !isString {
		return ctx.SendErrorWithCode(500, errors.New("get - invalid message type"))
	}

	if msg != "from before get middleware" {
		return ctx.SendErrorWithCode(500, errors.New("get - invalid message value"))
	}

	h.Result = map[string]any{
		"message": msg,
	}
	ctx.Set("message", "from get handler")
	return ctx.SendJson(h.Result)
}

func (h *DataController) AfterGet(ctx raiden.Context) error {
	msg := ctx.Get("message")
	msg, isString := msg.(string)
	if !isString {
		return ctx.SendErrorWithCode(500, errors.New("after get - invalid message type"))
	}

	if msg != "from get handler" {
		return ctx.SendErrorWithCode(500, errors.New("after get - invalid message value"))
	}

	return nil
}

func (h *DataController) BeforePost(ctx raiden.Context) error {
	ctx.Set("message", "from before post middleware")
	return nil
}

func (h *DataController) Post(ctx raiden.Context) error {
	msg := ctx.Get("message")
	msg, isString := msg.(string)
	if !isString {
		return ctx.SendErrorWithCode(500, errors.New("post - invalid message type"))
	}

	if msg != "from before post middleware" {
		return ctx.SendErrorWithCode(500, errors.New("post - invalid message value"))
	}

	h.Result = map[string]any{
		"message": msg,
	}
	ctx.Set("message", "from post handler")
	return ctx.SendJson(h.Result)
}

func (h *DataController) AfterPost(ctx raiden.Context) error {
	msg := ctx.Get("message")
	msg, isString := msg.(string)
	if !isString {
		return ctx.SendErrorWithCode(500, errors.New("after post - invalid message type"))
	}

	if msg != "from post handler" {
		return ctx.SendErrorWithCode(500, errors.New("after post - invalid message value"))
	}

	return nil
}

func TestController_Handler(t *testing.T) {
	var testResult any
	mockContext := mock.MockContext{
		SendJsonFn: func(data any) error {
			testResult = data
			return nil
		},
		SendErrorWithCodeFn: func(statusCode int, err error) error {
			return err
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

func TestController_PassData(t *testing.T) {
	var reqCtx fasthttp.RequestCtx
	context := raiden.Ctx{
		RequestCtx: &reqCtx,
	}

	// setup required data
	controller := &DataController{}

	err := controller.BeforeGet(&context)
	assert.NoError(t, err)

	err = controller.Get(&context)
	assert.NoError(t, err)

	err = controller.AfterGet(&context)
	assert.NoError(t, err)

	response := reqCtx.Response.Body()
	assert.NotNil(t, response)

	mapData := make(map[string]any)
	err = json.Unmarshal(response, &mapData)
	assert.NoError(t, err)

	message, isMessageExist := mapData["message"]
	assert.True(t, isMessageExist)
	assert.Equal(t, "from before get middleware", message)
}

func TestController_PassDataRestPost(t *testing.T) {
	var reqCtx fasthttp.RequestCtx
	context := raiden.Ctx{
		RequestCtx: &reqCtx,
	}

	// setup required data
	controller := &DataController{}

	err := controller.BeforePost(&context)
	assert.NoError(t, err)

	err = controller.Post(&context)
	assert.NoError(t, err)

	err = controller.AfterPost(&context)
	assert.NoError(t, err)

	response := reqCtx.Response.Body()
	assert.NotNil(t, response)

	mapData := make(map[string]any)
	err = json.Unmarshal(response, &mapData)
	assert.NoError(t, err)

	message, isMessageExist := mapData["message"]
	assert.True(t, isMessageExist)
	assert.Equal(t, "from before post middleware", message)
}
