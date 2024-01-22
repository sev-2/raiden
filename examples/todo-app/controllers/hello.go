package controllers

import (
	"github.com/sev-2/raiden"
)

// @type function
// @route /hello-word
func HelloWordHandler(ctx *raiden.Context) raiden.Presenter {
	response := map[string]any{
		"message": "hello word",
	}
	return ctx.SendData(response)
}

// @type function
// @route /greeting
func GreetingHandler(ctx *raiden.Context) raiden.Presenter {
	response := map[string]any{
		"message": "hello raiden",
	}
	return ctx.SendData(response)
}
