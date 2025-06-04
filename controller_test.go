package raiden_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/mock"
	"github.com/sev-2/raiden/pkg/postgres"
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

	chain := raiden.NewChain()
	handler := raiden.AuthProxy(
		ctx.Config(),
		chain,
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

	chain := raiden.NewChain()
	handler := raiden.AuthProxy(
		ctx.Config(),
		chain,
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

	chain := raiden.NewChain()
	handler := raiden.AuthProxy(
		ctx.Config(),
		chain,
		func(req *fasthttp.Request) {},
		func(resp *fasthttp.Response) error { return nil },
	)

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

	chain := raiden.NewChain()
	handler := raiden.AuthProxy(
		ctx.Config(),
		chain,
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

func TestMarshallAndValidate_ByMethod(t *testing.T) {
	ctx := newMockCtx()
	type Request struct {
		Search   string `query:"q"`
		Resource string `json:"resource" validate:"requiredForMethod=Post"`
	}
	type Controller struct {
		raiden.ControllerBase
		Payload *Request
	}
	controller := &Controller{}

	ctx.Request.Header.SetMethod(fasthttp.MethodGet)
	ctx.SetBodyString("{\"resource\":\"\"}")

	err := raiden.MarshallAndValidate(ctx.RequestContext(), controller)
	assert.NoError(t, err)

	ctx.Request.Header.SetMethod(fasthttp.MethodPost)
	err = raiden.MarshallAndValidate(ctx.RequestContext(), controller)
	assert.Error(t, err)
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

// Test Marshall Query Params

type Status struct {
	raiden.TypeBase
}

func (t *Status) Name() string {
	return "status"
}

func (r *Status) Format() string {
	return "status"
}

func (r *Status) Enums() []string {
	return []string{"success", "failed"}
}

func (r *Status) Comment() *string {
	return nil
}

type QueryParamRequest struct {
	StringValue           string            `query:"string_value"`
	BoolValue             *bool             `query:"bool_value"`
	IntValue              int               `query:"int_value"`
	Int8Value             int8              `query:"int8_value"`
	Int16Value            int16             `query:"int16_value"`
	Int32Value            int32             `query:"int32_value"`
	Int64Value            int64             `query:"int64_value"`
	Int64Ptr              *int64            `query:"int64_ptr"`
	UintValue             uint              `query:"uint_value"`
	Uint8Value            uint8             `query:"uint8_value"`
	Uint16Value           uint16            `query:"uint16_value"`
	Uint32Value           uint32            `query:"uint32_value"`
	Uint64Value           uint64            `query:"uint64_value"`
	Float32Value          *float32          `query:"float32_value"`
	Float64Value          *float64          `query:"float64_value"`
	ByteValue             byte              `query:"byte_value"`               // alias for uint8
	RuneValue             rune              `query:"rune_value"`               // alias for int32
	PostgresDateValue     *postgres.Date    `query:"postgres_date_value"`      // optional custom date
	PostgresDateTimeValue postgres.DateTime `query:"postgres_date_time_value"` // optional custom date
	CustomStatusValue     Status            `query:"custom_status_value"`      // custom type (int/string alias)
	PointValue            postgres.Point    `query:"point_value"`              // custom type (int/string alias)
}

type QueryParamResponse struct{}

type QueryController struct {
	raiden.ControllerBase
	Payload *QueryParamRequest
	Result  QueryParamResponse
}

func TestControllerMarshallAndValidate_QueryParams(t *testing.T) {
	// setup required data
	controller := &QueryController{}

	// setup request
	requestCtx := &fasthttp.RequestCtx{
		Request: fasthttp.Request{},
	}

	// run test 1
	q := requestCtx.Request.URI().QueryArgs()

	// Set values for all fields
	q.Set("string_value", "test string")
	q.Set("bool_value", "true")
	q.Set("int_value", "123")
	q.Set("int8_value", "8")
	q.Set("int16_value", "16")
	q.Set("int32_value", "32")
	q.Set("int64_value", "64")
	q.Set("int64_ptr", "128")
	q.Set("uint_value", "200")
	q.Set("uint8_value", "8")
	q.Set("uint16_value", "16")
	q.Set("uint32_value", "32")
	q.Set("uint64_value", "64")
	q.Set("float32_value", "3.14")
	q.Set("float64_value", "6.28")
	q.Set("byte_value", "65")   // ASCII 'A'
	q.Set("rune_value", "9731") // Unicode snowman ☃
	q.Set("postgres_date_value", "2025-05-23")
	q.Set("postgres_date_time_value", "2025-05-23 15:04:05")
	q.Set("custom_status_value", "success")
	q.Set("point_value", "(144.9631,-37.8136)")

	err := raiden.MarshallAndValidate(requestCtx, controller)
	assert.NoError(t, err)

	assert.Equal(t, "test string", controller.Payload.StringValue)
	assert.NotNil(t, controller.Payload.BoolValue)
	assert.Equal(t, true, *controller.Payload.BoolValue)
	assert.Equal(t, 123, controller.Payload.IntValue)
	assert.Equal(t, int8(8), controller.Payload.Int8Value)
	assert.Equal(t, int16(16), controller.Payload.Int16Value)
	assert.Equal(t, int32(32), controller.Payload.Int32Value)
	assert.Equal(t, int64(64), controller.Payload.Int64Value)
	assert.NotNil(t, controller.Payload.Int64Ptr)
	assert.Equal(t, int64(128), *controller.Payload.Int64Ptr)
	assert.Equal(t, uint(200), controller.Payload.UintValue)
	assert.Equal(t, uint8(8), controller.Payload.Uint8Value)
	assert.Equal(t, uint16(16), controller.Payload.Uint16Value)
	assert.Equal(t, uint32(32), controller.Payload.Uint32Value)
	assert.Equal(t, uint64(64), controller.Payload.Uint64Value)
	assert.NotNil(t, controller.Payload.Float32Value)
	assert.InDelta(t, 3.14, *controller.Payload.Float32Value, 0.001)
	assert.NotNil(t, controller.Payload.Float64Value)
	assert.InDelta(t, 6.28, *controller.Payload.Float64Value, 0.001)
	assert.Equal(t, byte(65), controller.Payload.ByteValue)
	assert.Equal(t, rune(9731), controller.Payload.RuneValue)
	assert.NotNil(t, controller.Payload.PostgresDateValue)
	assert.Equal(t, "2025-05-23", controller.Payload.PostgresDateValue.String())
	assert.Equal(t, "success", controller.Payload.CustomStatusValue.String())
	assert.Equal(t, float64(144.9631), controller.Payload.PointValue.X)
	assert.Equal(t, float64(-37.8136), controller.Payload.PointValue.Y)

	q.Set("bool_value", "invalid")
	err = raiden.MarshallAndValidate(requestCtx, controller)
	assert.EqualError(t, err, "bool: must be boolean value")

	q.Set("bool_value", "true")
	q.Set("int_value", "invalid")
	err = raiden.MarshallAndValidate(requestCtx, controller)
	assert.EqualError(t, err, "int: must be integer value")

	q.Set("int_value", "11")
	q.Set("uint_value", "-1")
	err = raiden.MarshallAndValidate(requestCtx, controller)
	assert.EqualError(t, err, "uint: must be unsigned integer value")

	q.Set("uint_value", "1")
	q.Set("float32_value", "invalid")
	err = raiden.MarshallAndValidate(requestCtx, controller)
	assert.EqualError(t, err, "float32: must be float value")

}

func TestMarshallAndValidate_MultiplatformData(t *testing.T) {
	// success
	ctxForm := newMockCtx()
	type RequestMultiplatform struct {
		Resource string `path:"resource" validate:"required"`
	}
	type ControllerMultiplatform struct {
		raiden.ControllerBase
		Payload *RequestMultiplatform
	}
	controllerMultiplatform := &ControllerMultiplatform{}

	ctxForm.Request.Header.Set(fasthttp.HeaderContentType, "multipart/form-data; boundary=X-BOUNDARY")
	ctxForm.Request.SetBodyString("--X-BOUNDARY\r\nContent-Disposition: form-data; name=\"resource\"\r\n\r\nresource_value\r\n--X-BOUNDARY--\r\n")
	err := raiden.MarshallAndValidate(ctxForm.RequestContext(), controllerMultiplatform)
	assert.NoError(t, err)
	form, errForm := ctxForm.Request.MultipartForm()
	assert.NoError(t, errForm)
	assert.Equal(t, "resource_value", form.Value["resource"][0])

	// invalid payload application json
	type RequestMultiplatformError struct {
		Resource string `path:"resource" validate:"required"`
	}
	type ControllerMultiplatformError struct {
		raiden.ControllerBase
		Payload *RequestMultiplatformError
	}
	controllerMultiplatformError := &ControllerMultiplatformError{}

	ctxForm.Request.Header.Set(fasthttp.HeaderContentType, "application/json")
	ctxForm.Request.SetBodyString("--X-BOUNDARY\r\nContent-Disposition: form-data; name=\"resource\"\r\n\r\nresource_value\r\n--X-BOUNDARY--\r\n")
	err = raiden.MarshallAndValidate(ctxForm.RequestContext(), controllerMultiplatformError)
	assert.Error(t, err)
}
