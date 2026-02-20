package raiden

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSubscriberContext_Config(t *testing.T) {
	cfg := &Config{ProjectId: "test-project"}
	ctx := &subscriberContext{
		cfg:     cfg,
		Context: context.Background(),
	}

	assert.Equal(t, cfg, ctx.Config())
}

func TestSubscriberContext_Span(t *testing.T) {
	ctx := &subscriberContext{
		Context: context.Background(),
	}

	assert.Nil(t, ctx.Span())
}

func TestSubscriberContext_HttpRequest_NonPointerResponse(t *testing.T) {
	ctx := &subscriberContext{
		cfg:     &Config{},
		Context: context.Background(),
	}

	// Non-pointer response should return error
	var resp string
	err := ctx.HttpRequest("GET", "http://localhost", nil, nil, 0, resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "response payload must be pointer")
}
