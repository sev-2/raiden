package mock

import (
	"time"

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
