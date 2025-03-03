package bar

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
	Payload *FooRequest
	Result  FooResponse
}

func (c *FooController) Post(ctx raiden.Context) error {
	c.Result.Message = fmt.Sprintf("foo : %s", c.Payload.Name)
	return ctx.SendJson(c.Result)
}

func (c *FooController) Patch(ctx raiden.Context) error {
	c.Result.Message = fmt.Sprintf("foo : %s", c.Payload.Name)
	return ctx.SendJson(c.Result)
}

func (c *FooController) Put(ctx raiden.Context) error {
	c.Result.Message = fmt.Sprintf("foo : %s", c.Payload.Name)
	return ctx.SendJson(c.Result)
}

func (c *FooController) Delete(ctx raiden.Context) error {
	c.Result.Data = map[string]any{
		"deleted": "true",
	}
	return ctx.SendJson(c.Result)
}

func (c *FooController) Head(ctx raiden.Context) error {
	c.Result.Data = map[string]any{
		"x-header": "raiden",
	}
	return ctx.SendJson(c.Result)
}

func (c *FooController) Options(ctx raiden.Context) error {
	c.Result.Data = map[string]any{
		"x-header": "raiden",
	}
	return ctx.SendJson(c.Result)
}

func (c *FooController) Get(ctx raiden.Context) error {
	c.Result.Data = map[string]any{
		"foo": c.Payload.Name,
	}
	return ctx.SendJson(c.Result)
}
