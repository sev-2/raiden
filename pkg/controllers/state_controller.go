package controllers

import (
	"net/http"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/resource"
)

type StateReadyRequest struct {
}

type StateReadyResponse struct {
	Message string `json:"message"`
}

type StateReadyController struct {
	raiden.ControllerBase
	Http    string `path:"/state/ready" type:"custom"`
	Payload *StateReadyRequest
	Result  StateReadyResponse
}

func (c *StateReadyController) Post(ctx raiden.Context) error {
	if err := resource.Import(&resource.Flags{UpdateStateOnly: true}, ctx.Config()); err != nil {
		return ctx.SendErrorWithCode(http.StatusInternalServerError, err)
	}

	c.Result.Message = "success initiate local state"
	return ctx.SendJson(c.Result)
}
