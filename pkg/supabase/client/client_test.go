package client_test

import (
	"testing"
	"time"

	"github.com/sev-2/raiden/pkg/supabase/client"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

type MockClient struct {
	DoFn        func(*fasthttp.Request, *fasthttp.Response) error
	DoTimeoutFn func(*fasthttp.Request, *fasthttp.Response, time.Duration) error
}

func (m *MockClient) Do(req *fasthttp.Request, rest *fasthttp.Response) error {
	return m.DoFn(req, rest)
}

func (m *MockClient) DoTimeout(req *fasthttp.Request, rest *fasthttp.Response, timeout time.Duration) error {
	return m.DoTimeoutFn(req, rest, timeout)
}

func TestSendRequest(t *testing.T) {
	t.Parallel()

	c := MockClient{
		DoTimeoutFn: func(r1 *fasthttp.Request, r2 *fasthttp.Response, d time.Duration) error {
			r2.SetBodyRaw([]byte("{ \"message\": \"hello from home\" }"))
			return nil
		},
	}

	client.SetClient(&c)

	type Response struct {
		Message string `json:"message"`
	}

	rs, _ := client.Get[Response]("http://localhost:8080", 1000000, nil, nil)
	assert.Equal(t, "", rs.Message)
}
