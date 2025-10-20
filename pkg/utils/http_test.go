package utils

import (
	"errors"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttputil"
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
		result := GetColoredHttpMethod(test.method)
		require.Contains(t, result, test.expected)
	}
}

func withInmemoryClient(t *testing.T, handler fasthttp.RequestHandler) string {
	ln := fasthttputil.NewInmemoryListener()
	server := &fasthttp.Server{Handler: handler}
	go func() {
		_ = server.Serve(ln)
	}()

	t.Cleanup(func() {
		_ = ln.Close()
		_ = server.Shutdown()
		httpClient = nil
	})

	httpClient = &fasthttp.Client{
		Dial: func(addr string) (net.Conn, error) {
			return ln.Dial()
		},
	}

	return "http://inmemory"
}

func TestSendRequest_Success(t *testing.T) {
	url := withInmemoryClient(t, func(ctx *fasthttp.RequestCtx) {
		ctx.SetStatusCode(http.StatusOK)
		ctx.SetBody([]byte("OK"))
	})

	body, err := SendRequest(http.MethodGet, url, nil, nil)
	require.NoError(t, err)
	require.Equal(t, []byte("OK"), body)

	time.Sleep(10 * time.Millisecond)
}

func TestSendRequest_Error(t *testing.T) {
	_, err := SendRequest(http.MethodGet, "http://invalid-url", nil, nil)
	require.Error(t, err)
}

func TestSendRequest_NonOKStatus(t *testing.T) {
	url := withInmemoryClient(t, func(ctx *fasthttp.RequestCtx) {
		ctx.SetStatusCode(http.StatusBadRequest)
		ctx.SetBody([]byte("Bad Request"))
	})

	body, err := SendRequest(http.MethodGet, url, nil, nil)
	require.Error(t, err)
	require.Nil(t, body)

	var sendErr SendRequestError
	require.True(t, errors.As(err, &sendErr))
	require.Equal(t, []byte("Bad Request"), sendErr.Body)
}
