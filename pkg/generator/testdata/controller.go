package testdata

import (
	"fmt"

	"github.com/sev-2/raiden"
)

type FooRequest struct {
	Name string `path:"name"`
}

type FooResponse struct {
	Message string `json:"message"`
	Data    any    `json:"data"`
}

type FooController struct {
	raiden.ControllerBase
	Http    string `path:"/foo/{name}" type:"custom"`
	Payload *FooRequest
	Result  FooResponse
}

func (c *FooController) Post(ctx raiden.Context) raiden.Presenter {
	c.Result.Message = fmt.Sprintf("foo : %s", c.Payload.Name)
	return ctx.SendJson(c.Result)
}

func (c *FooController) Get(ctx raiden.Context) raiden.Presenter {
	c.Result.Data = map[string]any{
		"foo": c.Payload.Name,
	}
	return ctx.SendJson(c.Result)
}

type BarRequest struct {
}

type BarResponse struct {
	Message string `json:"message"`
}

type BarController struct {
	raiden.ControllerBase
	Http    string `path:"/bar" type:"function"`
	Payload *BarRequest
	Result  BarResponse
}

func (c *BarController) Get(ctx raiden.Context) raiden.Presenter {
	c.Result.Message = "bar message"
	return ctx.SendJson(c.Result)
}
