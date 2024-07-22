package db

import (
	"testing"

	"github.com/sev-2/raiden"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func TestCount(t *testing.T) {
	ctx := raiden.Ctx{
		RequestCtx: &fasthttp.RequestCtx{},
	}

	model := Foo{}

	count, _ := NewQuery(&ctx).Model(model).Count()

	assert.IsType(t, 0, count)
}

func TestCountWithOptions(t *testing.T) {
	ctx := raiden.Ctx{
		RequestCtx: &fasthttp.RequestCtx{},
	}

	model := Foo{}

	count, _ := NewQuery(&ctx).Model(model).Count(CountOptions{Count: "planned"})

	assert.IsType(t, 0, count)
}
