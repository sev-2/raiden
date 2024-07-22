package testdata

import (
	"fmt"
	"time"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/postgres/roles"
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

func (c *BarController) Post(ctx raiden.Context) error {
	c.Result.Message = "bar message"
	return ctx.SendJson(c.Result)
}

type Sc struct {
	raiden.ModelBase
	Id          int64      `json:"id,omitempty" column:"name:id;type:bigint;primaryKey;autoIncrement;nullable:false"`
	Name        string     `json:"name,omitempty" column:"name:name;type:varchar;nullable:false"`
	Description *string    `json:"description,omitempty" column:"name:description;type:text;nullable"`
	IsActive    *bool      `json:"is_active,omitempty" column:"name:is_active;type:boolean;nullable;default:true"`
	CreatedAt   *time.Time `json:"created_at,omitempty" column:"name:created_at;type:timestampz;nullable;default:now()"`

	// Table information
	Metadata string `json:"-" schema:"public" rlsEnable:"true" rlsForced:"false"`

	// Access control
	Acl string `json:"-" read:"authenticated,sc" write:"authenticated,sc" readUsing:"true" writeCheck:"true" writeUsing:"true"`
}

type ScController struct {
	raiden.ControllerBase
	Http  string `path:"/sc-list" type:"rest"`
	Model roles.Anon
}

func (sc *ScController) BeforeGet(ctx raiden.Context) error {
	raiden.Info("[ScController] before get fire")
	return nil
}

func (sc *ScController) AfterGet(ctx raiden.Context) error {
	raiden.Info("[ScController] after get fire")
	return nil
}