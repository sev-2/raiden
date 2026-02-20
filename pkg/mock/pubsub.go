package mock

import (
	"context"

	"cloud.google.com/go/pubsub"
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/pubsub/google"
)

type MockPubSubClient struct {
	Subscriptions        map[string]*MockSubscription
	Topics               map[string]*MockTopic
	CloseFn              func() error
	CreateSubscriptionFn func(ctx context.Context, id string, cfg pubsub.SubscriptionConfig) (google.Subscription, error)
}

func (m *MockPubSubClient) Subscription(id string) google.Subscription {
	return m.Subscriptions[id]
}

func (m *MockPubSubClient) CreateSubscription(ctx context.Context, id string, cfg pubsub.SubscriptionConfig) (google.Subscription, error) {
	return m.CreateSubscriptionFn(ctx, id, cfg)
}

func (m *MockPubSubClient) Topic(id string) google.Topic {
	return m.Topics[id]
}

func (m *MockPubSubClient) Close() error {
	return m.CloseFn()
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

func (m *MockTopic) GetInstance() *pubsub.Topic {
	return nil
}

type MockPublishResult struct {
	Result string
	Err    error
}

func (m *MockPublishResult) Get(ctx context.Context) (string, error) {
	return m.Result, m.Err
}

type MockSubscriberHandler struct {
	TopicValue            string
	SubscriptionValue     string
	NameValue             string
	PushEndpointValue     string
	AutoAckValue          bool
	ChannelTypeValue      raiden.RealtimeChannelType
	EventFilterValue      string
	SchemaValue           string
	TableValue            string
	ProviderValue         raiden.PubSubProviderType
	SubscriptionTypeValue raiden.SubscriptionType
	ConsumeFunc           func(ctx raiden.SubscriberContext, msg raiden.SubscriberMessage) error
}

func (m *MockSubscriberHandler) Provider() raiden.PubSubProviderType {
	return m.ProviderValue
}

func (m *MockSubscriberHandler) Subscription() string {
	return m.SubscriptionValue
}

func (m *MockSubscriberHandler) Topic() string {
	return m.TopicValue
}

func (m *MockSubscriberHandler) PushEndpoint() string {
	return m.PushEndpointValue
}

func (m *MockSubscriberHandler) SubscriptionType() raiden.SubscriptionType {
	return m.SubscriptionTypeValue
}

func (m *MockSubscriberHandler) Name() string {
	return m.NameValue
}

func (m *MockSubscriberHandler) AutoAck() bool {
	return m.AutoAckValue
}

func (m *MockSubscriberHandler) ChannelType() raiden.RealtimeChannelType {
	return m.ChannelTypeValue
}

func (m *MockSubscriberHandler) EventFilter() string {
	if m.EventFilterValue == "" {
		return "*"
	}
	return m.EventFilterValue
}

func (m *MockSubscriberHandler) Schema() string {
	if m.SchemaValue == "" {
		return "public"
	}
	return m.SchemaValue
}

func (m *MockSubscriberHandler) Table() string {
	return m.TableValue
}

func (m *MockSubscriberHandler) Consume(ctx raiden.SubscriberContext, msg raiden.SubscriberMessage) error {
	return m.ConsumeFunc(ctx, msg)
}
