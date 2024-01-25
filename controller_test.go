package raiden_test

import (
	"testing"

	"github.com/magiconair/properties/assert"
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/mock"
	"github.com/valyala/fasthttp"
)

// @type function
// @route /test-handler
func ExampleHandler(ctx raiden.Context) raiden.Presenter {
	data := map[string]any{"message": "test"}
	return ctx.SendJson(data)
}

func TestRegisterFunctionHandler(t *testing.T) {
	registry := raiden.NewControllerRegistry()
	registry.Register(ExampleHandler)

	assert.Equal(t, len(registry.Controllers), 1)

	controller := registry.Controllers[0]
	assert.Equal(t, controller.Options.Path, "/test-handler")
	assert.Equal(t, controller.Options.Method, fasthttp.MethodPost)
	assert.Equal(t, controller.Options.Type, raiden.RouteTypeFunction)
}

// @type http-handler
// @route GET /test-http-handler
func ExampleHttpHandler(ctx raiden.Context) raiden.Presenter {
	data := map[string]any{"message": "test"}
	return ctx.SendJson(data)
}

func TestRegisterHttpHandlerHandler(t *testing.T) {
	registry := raiden.NewControllerRegistry()
	registry.Register(ExampleHttpHandler)

	assert.Equal(t, len(registry.Controllers), 1)

	controller := registry.Controllers[0]
	assert.Equal(t, controller.Options.Path, "/test-http-handler")
	assert.Equal(t, controller.Options.Method, fasthttp.MethodGet)
	assert.Equal(t, controller.Options.Type, raiden.RouteTypeHttpHandler)
}

// Health Handler Test
func TestHealthController(t *testing.T) {
	type mockPresenterCustom struct {
		mock.MockPresenter
		Data any
	}

	mockContext := mock.MockContext{
		SendJsonFn: func(data any) raiden.Presenter {
			return &mockPresenterCustom{
				Data: data,
			}
		},
	}

	presenter := raiden.HealthController(&mockContext)
	presenterCustom, isMockPresenter := presenter.(*mockPresenterCustom)
	assert.Equal(t, isMockPresenter, true)

	resultData, isMap := presenterCustom.Data.(map[string]any)
	assert.Equal(t, isMap, true)
	assert.Equal(t, resultData["message"], "server up")
}
