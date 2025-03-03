package generator_test

import (
	"testing"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/generator"
	"github.com/sev-2/raiden/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestRouter(t *testing.T) {
	testPath, err := utils.GetAbsolutePath("/testdata")
	assert.NoError(t, err)

	routes, err := generator.WalkScanControllers(raiden.BffMode, testPath)
	assert.NoError(t, err)

	assert.Equal(t, 6, len(routes))

	var barRoute, fooRoute *generator.GenerateRouteItem
	for i := range routes {
		r := routes[i]
		switch r.Controller {
		case "rest_v1_foo__name.FooController{}":
			fooRoute = &r
		case "function_v1_bar.BarController{}":
			barRoute = &r
		}
	}

	assert.NotNil(t, fooRoute)
	assert.NotNil(t, barRoute)

	assert.Equal(t, "raiden.RouteTypeCustom", fooRoute.Type)
	assert.Equal(t, "\"/internal/controllers/rest/v1/foo/{name}\"", fooRoute.Path)
	assert.Equal(t, "[]string{fasthttp.MethodPost, fasthttp.MethodPatch, fasthttp.MethodPut, fasthttp.MethodDelete, fasthttp.MethodHead, fasthttp.MethodOptions, fasthttp.MethodGet}", fooRoute.Methods)
	assert.Equal(t, "rest_v1_foo__name.FooController{}", fooRoute.Controller)

	assert.Equal(t, "raiden.RouteTypeFunction", barRoute.Type)
	assert.Equal(t, "\"/internal/controllers/function/v1/bar\"", barRoute.Path)
	assert.Equal(t, "[]string{fasthttp.MethodPost}", barRoute.Methods)
	assert.Equal(t, "function_v1_bar.BarController{}", barRoute.Controller)
}

func TestCreateRouteInput(t *testing.T) {
	projectName := "myproject"
	routePath := "/app/routes"
	routes := []generator.GenerateRouteItem{
		{
			Import: struct {
				Alias string
				Path  string
			}{Alias: "myController", Path: "/users"},
			Type:       "rpc",
			Path:       "/users",
			Methods:    "GET",
			Controller: "UserController",
			Model:      "UserModel",
			Storage:    "UserStorage",
		},
		{
			Import: struct {
				Alias string
				Path  string
			}{Alias: "orderController", Path: "/orders"},
			Type:       "function",
			Path:       "/orders",
			Methods:    "POST",
			Controller: "OrderController",
			Model:      "OrderModel",
			Storage:    "OrderStorage",
		},
	}

	input, err := generator.CreateRouteInput(projectName, routePath, routes)
	assert.NoError(t, err)

	// Validate output file path
	assert.Contains(t, input.OutputPath, routePath)

	// Validate template info
	assert.Equal(t, "routerTemplate", input.TemplateName)

	// Validate imports

	bindData, ok := input.BindData.(generator.GenerateRouterData)
	assert.True(t, ok)
	assert.Equal(t, len(bindData.Routes), 2)
	assert.Equal(t, len(bindData.Imports), 7)

	r := bindData.Routes[0]
	assert.Equal(t, "function", r.Type)
	assert.Equal(t, "/orders", r.Path)
}
