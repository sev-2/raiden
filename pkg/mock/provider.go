package mock

import (
	"context"

	"github.com/sev-2/raiden"
)

type MockProvider struct {
	PublishFn     func(ctx context.Context, topic string, message []byte) error
	StartListenFn func(handler []raiden.SubscriberHandler) error
	StopListenFn  func() error
}

// Publish implements raiden.PubSubProvider.
func (m *MockProvider) Publish(ctx context.Context, topic string, message []byte) error {
	return m.PublishFn(ctx, topic, message)
}

// StartListen implements raiden.PubSubProvider.
func (m *MockProvider) StartListen(handler []raiden.SubscriberHandler) error {
	return m.StartListenFn(handler)
}

// StopListen implements raiden.PubSubProvider.
func (m *MockProvider) StopListen() error {
	return m.StopListenFn()
}
