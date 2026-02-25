package mock

import (
	"context"

	"github.com/sev-2/raiden"
	"github.com/valyala/fasthttp"
)

type MockProvider struct {
	CreateSubscriptionFn func(handler raiden.SubscriberHandler) error
	PublishFn            func(ctx context.Context, topic string, message []byte) error
	ServeFn              func(config *raiden.Config, handler raiden.SubscriberHandler) (fasthttp.RequestHandler, error)
	StartListenFn        func(handler []raiden.SubscriberHandler) error
	StopListenFn         func() error
}

// Publish implements raiden.PubSubProvider.
func (m *MockProvider) Publish(ctx context.Context, topic string, message []byte) error {
	return m.PublishFn(ctx, topic, message)
}

// Publish implements raiden.PubSubProvider.
func (m *MockProvider) CreateSubscription(handler raiden.SubscriberHandler) error {
	return m.CreateSubscriptionFn(handler)
}

// StartListen implements raiden.PubSubProvider.
func (m *MockProvider) StartListen(handler []raiden.SubscriberHandler) error {
	return m.StartListenFn(handler)
}

// StopListen implements raiden.PubSubProvider.
func (m *MockProvider) StopListen() error {
	return m.StopListenFn()
}

// Serve implements raiden.PubSubProvider.
func (m *MockProvider) Serve(config *raiden.Config, handler raiden.SubscriberHandler) (fasthttp.RequestHandler, error) {
	if m.ServeFn != nil {
		return m.ServeFn(config, handler)
	}
	return nil, nil
}
