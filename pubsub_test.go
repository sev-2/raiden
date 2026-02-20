package raiden_test

import (
	"context"
	"errors"
	"testing"

	"cloud.google.com/go/pubsub"
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/mock"
	googlepubsub "github.com/sev-2/raiden/pkg/pubsub/google"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/trace"
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
	assert.Equal(t, "", handler.Subscription())
	assert.Error(t, handler.Consume(nil, raiden.SubscriberMessage{}))

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
				PublishFn: func(ctx context.Context, msg *pubsub.Message) googlepubsub.PublishResult {
					return &mock.MockPublishResult{Result: "mock-server-id", Err: nil}
				},
			},
		},
	}

	provider := &raiden.GooglePubSubProvider{
		Config: &raiden.Config{GoogleProjectId: "test-project", GoogleSaPath: "test-path"},
		Provider: &googlepubsub.Provider{
			Config: &googlepubsub.ProviderConfig{GoogleProjectId: "test-project", GoogleSaPath: "test-path"},
			Client: mockClient,
		},
	}

	err := provider.Publish(context.Background(), "test-topic", []byte("test message"))
	if err != nil {
		t.Errorf("Publish failed: %v", err)
	}
}

func TestGooglePubSubProvider_PublishWithSpan(t *testing.T) {
	mockClient := &mock.MockPubSubClient{
		Topics: map[string]*mock.MockTopic{
			"test-topic": {
				PublishFn: func(ctx context.Context, msg *pubsub.Message) googlepubsub.PublishResult {
					return &mock.MockPublishResult{Result: "mock-server-id", Err: nil}
				},
			},
		},
	}

	provider := &raiden.GooglePubSubProvider{
		Config: &raiden.Config{GoogleProjectId: "test-project", GoogleSaPath: "test-path"},
		Provider: &googlepubsub.Provider{
			Config: &googlepubsub.ProviderConfig{GoogleProjectId: "test-project", GoogleSaPath: "test-path"},
			Client: mockClient,
		},
	}

	ctx := trace.ContextWithSpanContext(context.Background(), trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		SpanID:     [8]byte{1, 2, 3, 4, 5, 6, 7, 8},
		TraceFlags: trace.FlagsSampled,
	}))

	err := provider.Publish(ctx, "test-topic", []byte("test message"))
	if err != nil {
		t.Errorf("Publish failed: %v", err)
	}
}

func TestGooglePubSubProvider_StartListen(t *testing.T) {
	mockSub := &mock.MockSubscription{
		Id: "test-subscription",
		ReceiveFn: func(ctx context.Context, f func(ctx context.Context, msg *pubsub.Message)) error {
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

	handlerCalled := false
	mockHandler := &mock.MockSubscriberHandler{
		TopicValue:   "test-topic",
		NameValue:    "test-handler",
		AutoAckValue: true,
		ConsumeFunc: func(ctx raiden.SubscriberContext, msg raiden.SubscriberMessage) error {
			handlerCalled = true
			assert.NotNil(t, msg.Data)
			assert.NotNil(t, msg.Raw)
			return nil
		},
		SubscriptionTypeValue: raiden.SubscriptionTypePush,
		SubscriptionValue:     "test-topic",
	}

	provider := &raiden.GooglePubSubProvider{
		Config: &raiden.Config{GoogleProjectId: "test-project", GoogleSaPath: "test-path"},
		Provider: &googlepubsub.Provider{
			Config: &googlepubsub.ProviderConfig{GoogleProjectId: "test-project", GoogleSaPath: "test-path"},
			Client: mockClient,
		},
	}

	err := provider.StartListen([]raiden.SubscriberHandler{mockHandler})
	if err != nil {
		t.Fatalf("StartListen failed: %v", err)
	}

	if !handlerCalled {
		t.Error("Handler was not called")
	}
}

func TestGooglePubSubProvider_StartListenErr(t *testing.T) {
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

	mockHandler := &mock.MockSubscriberHandler{
		TopicValue:   "test-topic",
		NameValue:    "test-handler",
		AutoAckValue: true,
		ConsumeFunc: func(ctx raiden.SubscriberContext, msg raiden.SubscriberMessage) error {
			return nil
		},
		SubscriptionTypeValue: raiden.SubscriptionTypePush,
		SubscriptionValue:     "test-topic",
	}

	provider := &raiden.GooglePubSubProvider{
		Config: &raiden.Config{GoogleProjectId: "test-project", GoogleSaPath: "test-path"},
		Provider: &googlepubsub.Provider{
			Config: &googlepubsub.ProviderConfig{GoogleProjectId: "test-project", GoogleSaPath: "test-path"},
			Client: mockClient,
		},
	}

	err := provider.StartListen([]raiden.SubscriberHandler{mockHandler})
	assert.Error(t, err)
}

func TestGooglePubSubProvider_StartListenErrCancel(t *testing.T) {
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

	mockHandler := &mock.MockSubscriberHandler{
		TopicValue:   "test-topic",
		NameValue:    "test-handler",
		AutoAckValue: true,
		ConsumeFunc: func(ctx raiden.SubscriberContext, msg raiden.SubscriberMessage) error {
			return nil
		},
		SubscriptionTypeValue: raiden.SubscriptionTypePush,
		SubscriptionValue:     "test-topic",
	}

	provider := &raiden.GooglePubSubProvider{
		Config: &raiden.Config{GoogleProjectId: "test-project", GoogleSaPath: "test-path"},
		Provider: &googlepubsub.Provider{
			Config: &googlepubsub.ProviderConfig{GoogleProjectId: "test-project", GoogleSaPath: "test-path"},
			Client: mockClient,
		},
	}

	err := provider.StartListen([]raiden.SubscriberHandler{mockHandler})
	assert.Nil(t, err)
}

func TestGooglePubSubProvider_StartListenWithTrace(t *testing.T) {
	mockSub := &mock.MockSubscription{
		Id: "test-subscription",
		ReceiveFn: func(ctx context.Context, f func(ctx context.Context, msg *pubsub.Message)) error {
			msg := &pubsub.Message{
				Data: []byte("test message"),
				Attributes: map[string]string{
					"trace_id": "0123456789abcdef0123456789abcdef",
					"span_id":  "0123456789abcdef",
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

	handlerCalled := false
	mockHandler := &mock.MockSubscriberHandler{
		TopicValue:   "test-topic",
		NameValue:    "test-handler",
		AutoAckValue: true,
		ConsumeFunc: func(ctx raiden.SubscriberContext, msg raiden.SubscriberMessage) error {
			handlerCalled = true
			return nil
		},
		SubscriptionTypeValue: raiden.SubscriptionTypePush,
		SubscriptionValue:     "test-topic",
	}

	provider := &raiden.GooglePubSubProvider{
		Config: &raiden.Config{GoogleProjectId: "test-project", GoogleSaPath: "test-path"},
		Provider: &googlepubsub.Provider{
			Config: &googlepubsub.ProviderConfig{GoogleProjectId: "test-project", GoogleSaPath: "test-path"},
			Client: mockClient,
			Tracer: &mock.MockTracer{},
		},
	}

	err := provider.StartListen([]raiden.SubscriberHandler{mockHandler})
	if err != nil {
		t.Fatalf("StartListen failed: %v", err)
	}

	if !handlerCalled {
		t.Error("Handler was not called")
	}
}

func TestGooglePubSubProvider_StopLister(t *testing.T) {
	mockClient := &mock.MockPubSubClient{
		CloseFn: func() error {
			return nil
		},
	}

	provider := &raiden.GooglePubSubProvider{
		Config: &raiden.Config{GoogleProjectId: "test-project", GoogleSaPath: "test-path"},
		Provider: &googlepubsub.Provider{
			Config: &googlepubsub.ProviderConfig{GoogleProjectId: "test-project", GoogleSaPath: "test-path"},
			Client: mockClient,
		},
	}

	err := provider.StopListen()
	assert.NoError(t, err)
}

// ----- Legacy Backward Compatibility Tests -----

// legacyHandler implements the old Consume(ctx, message any) pattern.
type legacyHandler struct {
	raiden.SubscriberBase
	consumed any
}

func (h *legacyHandler) Name() string                        { return "legacy-test" }
func (h *legacyHandler) Provider() raiden.PubSubProviderType { return raiden.PubSubProviderGoogle }
func (h *legacyHandler) Subscription() string                { return "legacy-sub" }
func (h *legacyHandler) Consume(ctx raiden.SubscriberContext, message any) error {
	h.consumed = message
	return nil
}

func TestWrapLegacySubscriber(t *testing.T) {
	legacy := &legacyHandler{}
	wrapped := raiden.WrapLegacySubscriber(legacy)

	assert.Equal(t, "legacy-test", wrapped.Name())
	assert.Equal(t, raiden.PubSubProviderGoogle, wrapped.Provider())
	assert.Equal(t, "legacy-sub", wrapped.Subscription())

	msg := raiden.SubscriberMessage{
		Data:       []byte("hello"),
		Attributes: map[string]string{"key": "val"},
		Raw:        "original-raw",
	}

	err := wrapped.Consume(nil, msg)
	assert.NoError(t, err)
	// Legacy consumer should receive the Raw value
	assert.Equal(t, "original-raw", legacy.consumed)
}

func TestWrapLegacySubscriber_NilRaw(t *testing.T) {
	legacy := &legacyHandler{}
	wrapped := raiden.WrapLegacySubscriber(legacy)

	msg := raiden.SubscriberMessage{
		Data: []byte("data-only"),
	}

	err := wrapped.Consume(nil, msg)
	assert.NoError(t, err)
	// When Raw is nil, legacy consumer should receive the full SubscriberMessage
	assert.Equal(t, msg, legacy.consumed)
}

func TestPushSubscriptionMessageAlias(t *testing.T) {
	// Verify the deprecated type alias still works
	var msg raiden.PushSubscriptionMessage
	msg.Data = "test-data"
	msg.MessageId = "msg-123"
	msg.PublishTime = "2026-01-01T00:00:00Z"
	assert.Equal(t, "test-data", msg.Data)

	var data raiden.PushSubscriptionData
	data.Message = msg
	data.Subscription = "projects/test/subscriptions/sub"
	assert.Equal(t, "projects/test/subscriptions/sub", data.Subscription)
}

// ----- Supabase Realtime Provider Adapter Tests -----

func TestNewPubsub_RegistersSupabaseProvider(t *testing.T) {
	config := loadConfig()
	ps := raiden.NewPubsub(config, nil)
	mgr := ps.(*raiden.PubSubManager)

	// Publish to supabase provider should not return "unsupported" error
	// (it will fail on connection, but the provider is registered)
	err := mgr.Publish(context.Background(), raiden.PubSubProviderSupabase, "test", []byte("hello"))
	// Error should be about connection, not "unsupported pubsub provider"
	assert.Error(t, err)
	assert.NotContains(t, err.Error(), "unsupported pubsub provider")
}

func TestSupabaseRealtimeProvider_ServeReturnsError(t *testing.T) {
	config := loadConfig()
	ps := raiden.NewPubsub(config, nil)

	handler := &mock.MockSubscriberHandler{
		NameValue:             "test-supabase",
		ProviderValue:         raiden.PubSubProviderSupabase,
		SubscriptionTypeValue: raiden.SubscriptionTypePush,
		ConsumeFunc: func(ctx raiden.SubscriberContext, msg raiden.SubscriberMessage) error {
			return nil
		},
	}

	_, err := ps.Serve(handler)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not support HTTP push")
}

func TestSupabaseRealtimeProvider_SubscriberBaseDefaults(t *testing.T) {
	handler := &pubsubHandler{}

	assert.Equal(t, raiden.RealtimeChannelType(""), handler.ChannelType())
	assert.Equal(t, "public", handler.Schema())
	assert.Equal(t, "", handler.Table())
	assert.Equal(t, "*", handler.EventFilter())
}

type supabaseHandler struct {
	raiden.SubscriberBase
	consumeCalled bool
	lastMsg       raiden.SubscriberMessage
}

func (s *supabaseHandler) Name() string                        { return "supabase-test" }
func (s *supabaseHandler) Provider() raiden.PubSubProviderType { return raiden.PubSubProviderSupabase }
func (s *supabaseHandler) Topic() string                       { return "test-room" }
func (s *supabaseHandler) ChannelType() raiden.RealtimeChannelType {
	return raiden.RealtimeChannelBroadcast
}
func (s *supabaseHandler) Consume(ctx raiden.SubscriberContext, msg raiden.SubscriberMessage) error {
	s.consumeCalled = true
	s.lastMsg = msg
	return nil
}

func TestSupabaseRealtimeProvider_HandlerRegistration(t *testing.T) {
	mgr := &raiden.PubSubManager{}
	mgr.SetConfig(loadConfig())
	mgr.SetProvider(raiden.PubSubProviderSupabase, &mock.MockProvider{
		PublishFn: func(ctx context.Context, topic string, message []byte) error {
			return nil
		},
	})

	handler := &supabaseHandler{}
	mgr.Register(handler)

	assert.Equal(t, 1, mgr.GetHandlerCount())
	assert.Equal(t, raiden.PubSubProviderSupabase, mgr.Handlers()[0].Provider())
	assert.Equal(t, raiden.RealtimeChannelBroadcast, mgr.Handlers()[0].ChannelType())
}

func TestSupabaseRealtimeProvider_PublishViaMock(t *testing.T) {
	publishCalled := false
	mgr := &raiden.PubSubManager{}
	mgr.SetConfig(loadConfig())
	mgr.SetProvider(raiden.PubSubProviderSupabase, &mock.MockProvider{
		PublishFn: func(ctx context.Context, topic string, message []byte) error {
			publishCalled = true
			assert.Equal(t, "test-topic", topic)
			assert.Equal(t, []byte(`{"text":"hello"}`), message)
			return nil
		},
	})

	err := mgr.Publish(context.Background(), raiden.PubSubProviderSupabase, "test-topic", []byte(`{"text":"hello"}`))
	assert.NoError(t, err)
	assert.True(t, publishCalled)
}
