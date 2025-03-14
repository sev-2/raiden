package raiden_test

import (
	"context"
	"errors"
	"testing"

	"github.com/sev-2/raiden"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
	"go.opentelemetry.io/otel/trace"
)

// Mock data for testing
type SomeParams struct {
	Name string `json:"name" column:"name:s_name;type:varchar"`
}
type GetItem struct {
	Id    int64  `json:"id" column:"name:id;type:integer"`
	SName string `json:"s_name" column:"name:s_name;type:varchar"`
}

type GetResult []GetItem

type SomeRpc struct {
	raiden.Rpc
	Params *SomeParams `json:"-"`
	Return GetResult   `json:"-"`
}

type SomeLib struct {
	raiden.BaseLibrary
	config *raiden.Config
}

var (
	mockSpan    = trace.SpanFromContext(context.Background())
	mockDataKey = "key"
	mockDataVal = "value"
	mockCtx     = context.Background()
)

// Helper function to create a new Ctx instance
func newTestCtx() *raiden.Ctx {
	req := fasthttp.RequestCtx{}
	req.SetUserValue("name", "name_xxxx")
	req.URI().QueryArgs().Add("status", "approved_xxx")
	ctx := &raiden.Ctx{
		Context:    mockCtx,
		RequestCtx: &req,
	}

	ctx.SetSpan(mockSpan)

	return ctx
}

func TestCtx_RequestContext(t *testing.T) {
	ctx := newTestCtx()
	assert.Equal(t, ctx.RequestCtx, ctx.RequestContext())
}

func TestCtx_Span(t *testing.T) {
	ctx := newTestCtx()
	assert.Equal(t, mockSpan, ctx.Span())
}

func TestCtx_SetSpan(t *testing.T) {
	ctx := newTestCtx()
	newSpan := trace.SpanFromContext(context.Background())
	ctx.SetSpan(newSpan)
	assert.Equal(t, newSpan, ctx.Span())
}

func TestCtx_Ctx(t *testing.T) {
	ctx := newTestCtx()
	assert.Equal(t, mockCtx, ctx.Ctx())
}

func TestCtx_SetCtx(t *testing.T) {
	ctx := newTestCtx()
	newCtx := context.WithValue(context.Background(), mockDataKey, mockDataVal)
	ctx.SetCtx(newCtx)
	assert.Equal(t, newCtx, ctx.Ctx())
}

func TestCtx_GetSet(t *testing.T) {
	ctx := newTestCtx()
	ctx.Set(mockDataKey, mockDataVal)
	assert.Equal(t, mockDataVal, ctx.Get(mockDataKey))
}

func TestCtx_GetParams(t *testing.T) {
	ctx := newTestCtx()
	param := ctx.GetParam("name")

	paramStr, isString := param.(string)
	assert.True(t, isString)
	assert.Equal(t, "name_xxxx", paramStr)
}

func TestCtx_GetQuery(t *testing.T) {
	ctx := newTestCtx()
	status := ctx.GetQuery("status")
	assert.Equal(t, "approved_xxx", status)
}

func TestCtx_SendJson(t *testing.T) {
	ctx := newTestCtx()
	data := map[string]string{"key": "value"}

	err := ctx.SendJson(data)
	assert.NoError(t, err)
	assert.Equal(t, "application/json", string(ctx.Response.Header.ContentType()))
}

func TestCtx_SendError(t *testing.T) {
	ctx := newTestCtx()
	err := ctx.SendError("error message")

	assert.Error(t, err)
	assert.Equal(t, fasthttp.StatusInternalServerError, err.(*raiden.ErrorResponse).StatusCode)
}

func TestCtx_SendErrorWithCode(t *testing.T) {
	ctx := newTestCtx()
	err := ctx.SendErrorWithCode(fasthttp.StatusBadRequest, errors.New("bad request"))

	assert.Error(t, err)
	assert.Equal(t, fasthttp.StatusBadRequest, err.(*raiden.ErrorResponse).StatusCode)
}

func TestCtx_WriteError(t *testing.T) {
	ctx := newTestCtx()
	err := &raiden.ErrorResponse{
		Message:    "error message",
		StatusCode: fasthttp.StatusInternalServerError,
	}

	ctx.WriteError(err)
	assert.Equal(t, "application/json", string(ctx.Response.Header.ContentType()))
}

func TestCtx_Write(t *testing.T) {
	ctx := newTestCtx()
	data := []byte("response data")

	ctx.Write(data)
	assert.Equal(t, fasthttp.StatusOK, ctx.Response.StatusCode())
	assert.Equal(t, data, ctx.Response.Body())
}

func TestCtx_Job(t *testing.T) {
	jChan := make(chan raiden.JobParams)
	ctx := newTestCtx()
	ctx.SetJobChan(jChan)

	c, e := ctx.NewJobCtx()
	assert.NoError(t, e)
	assert.NotNil(t, c)
}

func TestCtx_JobNil(t *testing.T) {
	ctx := newTestCtx()
	c, e := ctx.NewJobCtx()
	assert.Error(t, e)
	assert.Nil(t, c)
}

func TestCtx_GetLib_Ptr(t *testing.T) {
	ctx := newTestCtx()
	err := ctx.ResolveLibrary(SomeLib{})
	assert.Error(t, err)
}

func TestCtx_GetLib_Nil(t *testing.T) {
	ctx := newTestCtx()
	err := ctx.ResolveLibrary(&SomeLib{})
	assert.Error(t, err)
}

func TestCtx_SetLib(t *testing.T) {
	lib := SomeLib{}
	tLib := map[string]any{"SomeLib": lib}
	ctx := newTestCtx()
	ctx.RegisterLibraries(tLib)
	err := ctx.ResolveLibrary(&lib)
	assert.NoError(t, err)
}
