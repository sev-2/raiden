package raiden_test

import (
	"testing"

	"github.com/sev-2/raiden"
	"github.com/stretchr/testify/assert"
)

func TestNewChain(t *testing.T) {
	c := raiden.NewChain(m1, m2)

	assert.NotNil(t, c)
}

func TestChain_Append(t *testing.T) {
	c := raiden.NewChain(m1, m2)

	assert.NotNil(t, c.Append(m3, m4))
}

func TestChain_Prepend(t *testing.T) {
	c := raiden.NewChain(m3, m4)

	assert.NotNil(t, c.Prepend(m1, m2))
}

func TestChain_Then(t *testing.T) {
	// Create a new chain with two middlewares
	c := raiden.NewChain(m1, m2)

	// setup required data
	controller := &Controller{}

	// Call Then
	fn := c.Then("GET", raiden.RouteTypeCustom, controller)

	// Test the returned function
	assert.NotNil(t, fn)
}

func m1(next raiden.RouteHandlerFn) raiden.RouteHandlerFn {
	return func(ctx raiden.Context) error {
		return next(ctx)
	}
}

func m2(next raiden.RouteHandlerFn) raiden.RouteHandlerFn {
	return func(ctx raiden.Context) error {
		return next(ctx)
	}
}

func m3(next raiden.RouteHandlerFn) raiden.RouteHandlerFn {
	return func(ctx raiden.Context) error {
		return next(ctx)
	}
}

func m4(next raiden.RouteHandlerFn) raiden.RouteHandlerFn {
	return func(ctx raiden.Context) error {
		return next(ctx)
	}
}
