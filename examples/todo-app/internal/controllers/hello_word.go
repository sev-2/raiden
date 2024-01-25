package controllers

import (
	"fmt"

	"github.com/sev-2/raiden"
)

// type definitions
type HelloWordRequest struct {
	Name    string `path:"name" validate:"required"`
	Type    string `query:"type" validate:"required"`
	Message string `json:"message" validate:"required"`
}

type HelloWordResponse struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// @type function
// @route /hello/{name}
func HelloWordController(ctx raiden.Context) raiden.Presenter {
	payload, err := raiden.UnmarshalRequestAndValidate[HelloWordRequest](ctx)
	if err != nil {
		return ctx.SendJsonError(err)
	}

	response := HelloWordResponse{
		Message: fmt.Sprintf("hello %s, %s", payload.Name, payload.Message),
		Type:    payload.Type,
	}

	return ctx.SendJson(response)
}
