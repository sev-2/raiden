package raiden_test

import (
	"context"
	"errors"
	"testing"

	"cloud.google.com/go/pubsub"
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/mock"
	"github.com/sev-2/raiden/pkg/pubsub/google"
	"github.com/stretchr/testify/assert"
)

type pubsubHandler struct {
	raiden.SubscriberBase
}

func (*pubsubHandler) Provider() raiden.PubSubProviderType {
	return raiden.PubSubProviderGoogle
}

func TestPubsubInstance(t *testing.T) {
	pubsub := raiden.NewPubsub(nil, nil)
	assert.NotNil(t, pubsub)
}

func TestNewGooglePubsub(t *testing.T) {
	config := loadConfig()
	config.GoogleProjectId = "x"
	config.GoogleSaPath = "/sa.json"

	mockProvider := mock.MockProvider{
		PublishFn: func(ctx context.Context, topic string, message []byte) error {
			return nil
		},
		StartListenFn: func(handler []raiden.SubscriberHandler) error {
			return nil
		},
		StopListenFn: func() error {
			return nil
		},
	}

	pubsub := raiden.PubSubManager{}

	pubsub.SetConfig(config)
	pubsub.SetProvider(raiden.PubSubProviderGoogle, &mockProvider)

	handler := &pubsubHandler{}

	assert.Equal(t, true, handler.AutoAck())
	assert.Equal(t, "unknown", handler.Name())
	assert.Equal(t, "", handler.Topic())
	assert.Error(t, handler.Consume(nil, nil))

	// assert register
	pubsub.Register(handler)
	assert.Equal(t, 1, pubsub.GetHandlerCount())

	// assert publish
	err := pubsub.Publish(context.Background(), raiden.PubSubProviderGoogle, "test", []byte("{\"message\":\"hello\"}"))
	assert.NoError(t, err)

	mockProvider.PublishFn = func(ctx context.Context, topic string, message []byte) error { return errors.New("test error") }
	err = pubsub.Publish(context.Background(), raiden.PubSubProviderGoogle, "test", []byte("{\"message\":\"hello\"}"))
	assert.Error(t, err)

	pubsub.Listen()

	mockProvider.PublishFn = func(ctx context.Context, topic string, message []byte) error { return nil }
	err = pubsub.Publish(context.Background(), raiden.PubSubProviderUnknown, "test", []byte("{\"message\":\"hello\"}"))
	assert.Error(t, err)

	mockProvider.StopListenFn = func() error { return errors.New("fail stop") }
	pubsub.Listen()
}

func TestGooglePubSubProvider_Publish(t *testing.T) {
	mockClient := &mock.MockPubSubClient{
		Topics: map[string]*mock.MockTopic{
			"test-topic": {
				PublishFn: func(ctx context.Context, msg *pubsub.Message) google.PublishResult {
					return &mock.MockPublishResult{Result: "mock-server-id", Err: nil}
				},
			},
		},
	}

	provider := &raiden.GooglePubSubProvider{
		Config: &raiden.Config{GoogleProjectId: "test-project", GoogleSaPath: "test-path"},
		Client: mockClient,
	}

	err := provider.Publish(context.Background(), "test-topic", []byte("test message"))
	if err != nil {
		t.Errorf("Publish failed: %v", err)
	}
}

func TestGooglePubSubProvider_StartListen(t *testing.T) {
	// Mock PubSub client and subscription behavior
	mockSub := &mock.MockSubscription{
		Id: "test-subscription",
		ReceiveFn: func(ctx context.Context, f func(ctx context.Context, msg *pubsub.Message)) error {
			// Simulate a message being received
			msg := &pubsub.Message{
				Data: []byte("test message"),
				Attributes: map[string]string{
					"trace_id": "mock-trace-id",
					"span_id":  "mock-span-id",
				},
			}
			f(ctx, msg)
			return nil
		},
	}

	mockClient := &mock.MockPubSubClient{
		Subscriptions: map[string]*mock.MockSubscription{
			"test-topic": mockSub,
		},
	}

	// Mock SubscriberHandler behavior
	handlerCalled := false
	mockHandler := &mock.MockSubscriberHandler{
		TopicValue:   "test-topic",
		NameValue:    "test-handler",
		AutoAckValue: true,
		ConsumeFunc: func(ctx raiden.SubscriberContext, msg any) error {
			handlerCalled = true
			return nil
		},
	}

	// Create provider
	provider := &raiden.GooglePubSubProvider{
		Config: &raiden.Config{GoogleProjectId: "test-project", GoogleSaPath: "test-path"},
		Client: mockClient,
		Tracer: nil, // You can add a mock tracer here if needed
	}

	// Start listening
	err := provider.StartListen([]raiden.SubscriberHandler{mockHandler})
	if err != nil {
		t.Fatalf("StartListen failed: %v", err)
	}

	// Verify the handler was called
	if !handlerCalled {
		t.Error("Handler was not called")
	}
}

func TestGooglePubSubProvider_StartListenErr(t *testing.T) {
	// Mock PubSub client and subscription behavior
	mockSub := &mock.MockSubscription{
		Id: "test-subscription",
		ReceiveFn: func(ctx context.Context, f func(ctx context.Context, msg *pubsub.Message)) error {
			return errors.New("expect error")
		},
	}

	mockClient := &mock.MockPubSubClient{
		Subscriptions: map[string]*mock.MockSubscription{
			"test-topic": mockSub,
		},
	}

	// Mock SubscriberHandler behavior
	mockHandler := &mock.MockSubscriberHandler{
		TopicValue:   "test-topic",
		NameValue:    "test-handler",
		AutoAckValue: true,
		ConsumeFunc: func(ctx raiden.SubscriberContext, msg any) error {
			return nil
		},
	}

	// Create provider
	provider := &raiden.GooglePubSubProvider{
		Config: &raiden.Config{GoogleProjectId: "test-project", GoogleSaPath: "test-path"},
		Client: mockClient,
		Tracer: nil, // You can add a mock tracer here if needed
	}

	// Start listening
	err := provider.StartListen([]raiden.SubscriberHandler{mockHandler})
	assert.Error(t, err)
}

func TestGooglePubSubProvider_StartListenErrCancel(t *testing.T) {
	// Mock PubSub client and subscription behavior
	mockSub := &mock.MockSubscription{
		Id: "test-subscription",
		ReceiveFn: func(ctx context.Context, f func(ctx context.Context, msg *pubsub.Message)) error {
			return errors.New("code = Canceled")
		},
	}

	mockClient := &mock.MockPubSubClient{
		Subscriptions: map[string]*mock.MockSubscription{
			"test-topic": mockSub,
		},
	}

	// Mock SubscriberHandler behavior
	mockHandler := &mock.MockSubscriberHandler{
		TopicValue:   "test-topic",
		NameValue:    "test-handler",
		AutoAckValue: true,
		ConsumeFunc: func(ctx raiden.SubscriberContext, msg any) error {
			return nil
		},
	}

	// Create provider
	provider := &raiden.GooglePubSubProvider{
		Config: &raiden.Config{GoogleProjectId: "test-project", GoogleSaPath: "test-path"},
		Client: mockClient,
		Tracer: nil, // You can add a mock tracer here if needed
	}

	// Start listening
	err := provider.StartListen([]raiden.SubscriberHandler{mockHandler})
	assert.Nil(t, err)
}
