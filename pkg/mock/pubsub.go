package mock

import (
	"context"

	"cloud.google.com/go/pubsub"
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/pubsub/google"
)

type MockPubSubClient struct {
	Subscriptions map[string]*MockSubscription
	Topics        map[string]*MockTopic
}

func (m *MockPubSubClient) Subscription(id string) google.Subscription {
	return m.Subscriptions[id]
}

func (m *MockPubSubClient) Topic(id string) google.Topic {
	return m.Topics[id]
}

func (m *MockPubSubClient) Close() error {
	return nil
}

type MockSubscription struct {
	Id        string
	ReceiveFn func(ctx context.Context, f func(ctx context.Context, msg *pubsub.Message)) error
}

func (m *MockSubscription) ID() string {
	return m.Id
}

func (m *MockSubscription) Receive(ctx context.Context, f func(ctx context.Context, msg *pubsub.Message)) error {
	return m.ReceiveFn(ctx, f)
}

type MockTopic struct {
	PublishFn func(ctx context.Context, msg *pubsub.Message) google.PublishResult
}

func (m *MockTopic) Publish(ctx context.Context, msg *pubsub.Message) google.PublishResult {
	return m.PublishFn(ctx, msg)
}

type MockPublishResult struct {
	Result string
	Err    error
}

func (m *MockPublishResult) Get(ctx context.Context) (string, error) {
	return m.Result, m.Err
}

type MockSubscriberHandler struct {
	TopicValue    string
	NameValue     string
	AutoAckValue  bool
	ProviderValue raiden.PubSubProviderType
	ConsumeFunc   func(ctx raiden.SubscriberContext, msg any) error
}

func (m *MockSubscriberHandler) Provider() raiden.PubSubProviderType {
	return m.ProviderValue
}

func (m *MockSubscriberHandler) Topic() string {
	return m.TopicValue
}

func (m *MockSubscriberHandler) Name() string {
	return m.NameValue
}

func (m *MockSubscriberHandler) AutoAck() bool {
	return m.AutoAckValue
}

func (m *MockSubscriberHandler) Consume(ctx raiden.SubscriberContext, msg any) error {
	return m.ConsumeFunc(ctx, msg)
}
