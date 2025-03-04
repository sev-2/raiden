package bar

import (
	"github.com/sev-2/raiden"
)

type BarRequest struct {
}

type BarResponse struct {
	Message string `json:"message"`
}

type BarController struct {
	raiden.ControllerBase
	Payload *BarRequest
	Result  BarResponse
}

func (c *BarController) Post(ctx raiden.Context) error {
	c.Result.Message = "bar message"
	return ctx.SendJson(c.Result)
}
