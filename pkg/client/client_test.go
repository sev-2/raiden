package client_test

import (
	"testing"
	"time"

	"github.com/sev-2/raiden/pkg/client"
	"github.com/sev-2/raiden/pkg/mock"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func TestSendRequest(t *testing.T) {
	t.Parallel()

	c := mock.MockClient{
		DoTimeoutFn: func(r1 *fasthttp.Request, r2 *fasthttp.Response, d time.Duration) error {
			r2.SetBodyRaw([]byte("{ \"message\": \"hello from home\" }"))
			return nil
		},
	}

	client.GetClientFn = func() client.Client {
		return &c
	}

	type Response struct {
		Message string `json:"message"`
	}

	rs, _ := client.Get[Response]("http://localhost:8080", 1000000, nil, nil)
	assert.Equal(t, "hello from home", rs.Message)
}
