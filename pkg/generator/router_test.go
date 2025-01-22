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

	assert.Equal(t, 3, len(routes))

	var barRoute, fooRoute *generator.GenerateRouteItem
	for i := range routes {
		r := routes[i]
		switch r.Controller {
		case "testdata.FooController{}":
			fooRoute = &r
		case "testdata.BarController{}":
			barRoute = &r

		}
	}

	assert.NotNil(t, fooRoute)
	assert.NotNil(t, barRoute)

	assert.Equal(t, "raiden.RouteTypeCustom", fooRoute.Type)
	assert.Equal(t, "\"/foo/{name}\"", fooRoute.Path)
	assert.Equal(t, "[]string{fasthttp.MethodPost, fasthttp.MethodPatch, fasthttp.MethodPut, fasthttp.MethodDelete, fasthttp.MethodHead, fasthttp.MethodOptions, fasthttp.MethodGet}", fooRoute.Methods)
	assert.Equal(t, "testdata.FooController{}", fooRoute.Controller)

	assert.Equal(t, "raiden.RouteTypeFunction", barRoute.Type)
	assert.Equal(t, "\"/bar\"", barRoute.Path)
	assert.Equal(t, "[]string{fasthttp.MethodPost}", barRoute.Methods)
	assert.Equal(t, "testdata.BarController{}", barRoute.Controller)
}
