package utils_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/sev-2/raiden/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func TestGetColoredHttpMethod(t *testing.T) {
	tests := []struct {
		method   string
		expected string
	}{
		{http.MethodGet, " GET "},
		{http.MethodPost, " POST "},
		{http.MethodPatch, " PATCH "},
		{http.MethodPut, " PUT "},
		{http.MethodDelete, " DELETE "},
		{"UNKNOWN", " UNKNOWN "},
	}

	for _, test := range tests {
		result := utils.GetColoredHttpMethod(test.method)
		assert.Contains(t, result, test.expected)
	}
}

func TestSendRequest_Success(t *testing.T) {
	// Mock server
	server := fasthttp.Server{
		Handler: func(ctx *fasthttp.RequestCtx) {
			ctx.SetStatusCode(http.StatusOK)
			ctx.SetBody([]byte("OK"))
		},
	}
	go func() {
		err0 := server.ListenAndServe(":9090")
		assert.NoError(t, err0)
	}()
	time.Sleep(100 * time.Millisecond)

	body, err := utils.SendRequest(http.MethodGet, "http://localhost:9090", nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, []byte("OK"), body)
}

func TestSendRequest_Error(t *testing.T) {
	_, err := utils.SendRequest(http.MethodGet, "http://invalid-url", nil, nil)
	assert.Error(t, err)
}

func TestSendRequest_NonOKStatus(t *testing.T) {
	// Mock server
	server := fasthttp.Server{
		Handler: func(ctx *fasthttp.RequestCtx) {
			ctx.SetStatusCode(http.StatusBadRequest)
			ctx.SetBody([]byte("Bad Request"))
		},
	}
	go func() {
		err0 := server.ListenAndServe(":8081")
		assert.NoError(t, err0)
	}()
	time.Sleep(100 * time.Millisecond)

	body, err := utils.SendRequest(http.MethodGet, "http://localhost:8081", nil, nil)
	assert.Error(t, err)
	assert.Nil(t, body)
	assert.IsType(t, utils.SendRequestError{}, err)
}
