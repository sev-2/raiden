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

func newMockCtx() *raiden.Ctx {
	ctx := &raiden.Ctx{
		Context:    mockCtx,
		RequestCtx: &fasthttp.RequestCtx{},
	}

	ctx.SetSpan(mockSpan)

	return ctx
}

// Test ControllerBase methods
func TestControllerBase(t *testing.T) {
	ctx := newMockCtx()
	base := &raiden.ControllerBase{}

	// Test all ControllerBase methods
	assert.NoError(t, base.BeforeAll(ctx))
	assert.NoError(t, base.AfterAll(ctx))

	assert.NoError(t, base.BeforeGet(ctx))
	assert.Error(t, base.Get(ctx))
	assert.NoError(t, base.AfterGet(ctx))

	assert.NoError(t, base.BeforePost(ctx))
	assert.Error(t, base.Post(ctx))
	assert.NoError(t, base.AfterPost(ctx))

	assert.NoError(t, base.BeforePut(ctx))
	assert.Error(t, base.Put(ctx))
	assert.NoError(t, base.AfterPut(ctx))

	assert.NoError(t, base.BeforePatch(ctx))
	assert.Error(t, base.Patch(ctx))
	assert.NoError(t, base.AfterPatch(ctx))

	assert.NoError(t, base.BeforeDelete(ctx))
	assert.Error(t, base.Delete(ctx))
	assert.NoError(t, base.AfterDelete(ctx))

	assert.NoError(t, base.BeforeOptions(ctx))
	assert.Error(t, base.Options(ctx))
	assert.NoError(t, base.AfterOptions(ctx))

	assert.NoError(t, base.BeforeHead(ctx))
	assert.Error(t, base.Head(ctx))
	assert.NoError(t, base.AfterHead(ctx))
}

// Test RestController methods
func TestRestController(t *testing.T) {
	ctx := newMockCtx()
	rest := raiden.RestController{Controller: &raiden.ControllerBase{}, TableName: "test_table"}

	// Test RestController methods
	assert.NoError(t, rest.BeforeAll(ctx))
	assert.NoError(t, rest.AfterAll(ctx))

	assert.NoError(t, rest.BeforeGet(ctx))
	assert.NoError(t, rest.AfterGet(ctx))

	assert.NoError(t, rest.BeforePost(ctx))
	assert.NoError(t, rest.AfterPost(ctx))

	assert.NoError(t, rest.BeforePut(ctx))
	assert.NoError(t, rest.AfterPut(ctx))

	assert.NoError(t, rest.BeforePatch(ctx))
	assert.NoError(t, rest.AfterPatch(ctx))

	assert.NoError(t, rest.BeforeDelete(ctx))
	assert.NoError(t, rest.AfterDelete(ctx))

	assert.NoError(t, rest.BeforeOptions(ctx))
	assert.NoError(t, rest.AfterOptions(ctx))

	assert.NoError(t, rest.BeforeHead(ctx))
	assert.NoError(t, rest.AfterHead(ctx))
}

// Test StorageController methods
func TestStorageController(t *testing.T) {
	ctx := newMockCtx()
	storage := raiden.StorageController{Controller: &raiden.ControllerBase{}, BucketName: "test_bucket"}

	// Test StorageController methods
	assert.NoError(t, storage.BeforeAll(ctx))
	assert.NoError(t, storage.AfterAll(ctx))

	assert.NoError(t, storage.BeforeGet(ctx))
	assert.Error(t, storage.Get(ctx))
	assert.NoError(t, storage.AfterGet(ctx))

	assert.NoError(t, storage.BeforePost(ctx))
	assert.Error(t, storage.Post(ctx))
	assert.NoError(t, storage.AfterPost(ctx))

	assert.NoError(t, storage.BeforePut(ctx))
	assert.Error(t, storage.Put(ctx))
	assert.NoError(t, storage.AfterPut(ctx))

	assert.NoError(t, storage.BeforePatch(ctx))
	assert.Error(t, storage.Patch(ctx))
	assert.NoError(t, storage.AfterPatch(ctx))

	assert.NoError(t, storage.BeforeDelete(ctx))
	assert.Error(t, storage.Delete(ctx))
	assert.NoError(t, storage.AfterDelete(ctx))

	assert.NoError(t, storage.BeforeOptions(ctx))
	assert.NoError(t, storage.AfterOptions(ctx))

	assert.NoError(t, storage.BeforeHead(ctx))
	assert.NoError(t, storage.AfterHead(ctx))
}

// Test AuthProxy
func TestAuthProxy(t *testing.T) {
	ctx := &mock.MockContext{
		RequestCtx: &fasthttp.RequestCtx{},
		ConfigFn: func() *raiden.Config {
			return &raiden.Config{SupabasePublicUrl: "/"}
		},
	}

	handler := raiden.AuthProxy(
		ctx.Config(),
		func(req *fasthttp.Request) {},
		func(resp *fasthttp.Response) error { return nil },
	)

	ctx.Request.SetRequestURI("/")
	handler(ctx.RequestCtx)
	assert.Equal(t, fasthttp.StatusOK, ctx.Response.StatusCode())
}

// Test AuthProxy
func TestAuthProxy_ActualCallWithoutMock(t *testing.T) {
	ctx := &mock.MockContext{
		RequestCtx: &fasthttp.RequestCtx{},
		ConfigFn: func() *raiden.Config {
			return &raiden.Config{SupabasePublicUrl: "/"}
		},
	}

	handler := raiden.AuthProxy(
		ctx.Config(),
		func(req *fasthttp.Request) {},
		func(resp *fasthttp.Response) error { return nil },
	)

	ctx.Request.SetRequestURI("/auth/v1/signup")
	handler(ctx.RequestCtx)
	assert.Equal(t, fasthttp.StatusInternalServerError, ctx.Response.StatusCode())
}

func TestAuthProxy_NotAllowedPath(t *testing.T) {
	ctx := &mock.MockContext{
		RequestCtx: &fasthttp.RequestCtx{},
		ConfigFn: func() *raiden.Config {
			return &raiden.Config{SupabasePublicUrl: "/"}
		},
	}

	handler := raiden.AuthProxy(
		ctx.Config(),
		func(req *fasthttp.Request) {},
		func(resp *fasthttp.Response) error { return nil },
	)

	ctx.Request.SetRequestURI("/auth/v1/ %gh&%ij")
	handler(ctx.RequestCtx)
	assert.Equal(t, fasthttp.StatusBadRequest, ctx.Response.StatusCode())

	ctx.Request.SetRequestURI("/auth/v1/anymore")
	handler(ctx.RequestCtx)
	assert.Equal(t, fasthttp.StatusNotFound, ctx.Response.StatusCode())
}

func TestAuthProxy_AllowedWithSpecificPath(t *testing.T) {
	ctx := &mock.MockContext{
		RequestCtx: &fasthttp.RequestCtx{},
		ConfigFn: func() *raiden.Config {
			return &raiden.Config{SupabasePublicUrl: "/"}
		},
	}

	handler := raiden.AuthProxy(
		ctx.Config(),
		func(req *fasthttp.Request) {},
		func(resp *fasthttp.Response) error { return nil },
	)

	ctx.Request.SetRequestURI("/auth/v1/saml/metadata")
	handler(ctx.RequestCtx)
	assert.Equal(t, fasthttp.StatusInternalServerError, ctx.Response.StatusCode())
}

// Test MarshallAndValidate
func TestMarshallAndValidate(t *testing.T) {
	ctx := newMockCtx()
	type Request struct {
		Search   string `query:"q"`
		Resource string `path:"resource" validate:"required"`
	}
	type Controller struct {
		raiden.ControllerBase
		Payload *Request
	}
	controller := &Controller{}

	ctx.QueryArgs().Set("q", "search_value")
	ctx.SetUserValue("resource", "resource_value")

	err := raiden.MarshallAndValidate(ctx.RequestContext(), controller)
	assert.NoError(t, err)
	assert.Equal(t, "search_value", controller.Payload.Search)
	assert.Equal(t, "resource_value", controller.Payload.Resource)

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
