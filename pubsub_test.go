package raiden_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"cloud.google.com/go/pubsub"
	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/mock"
	google "github.com/sev-2/raiden/pkg/pubsub/google"
	supabasepubsub "github.com/sev-2/raiden/pkg/pubsub/supabase"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
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
				PublishFn: func(ctx context.Context, msg *pubsub.Message) google.PublishResult {
					return &mock.MockPublishResult{Result: "mock-server-id", Err: nil}
				},
			},
		},
	}

	provider := &raiden.GooglePubSubProvider{
		Config: &raiden.Config{GoogleProjectId: "test-project", GoogleSaPath: "test-path"},
		Provider: &google.Provider{
			Config: &google.ProviderConfig{GoogleProjectId: "test-project", GoogleSaPath: "test-path"},
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
				PublishFn: func(ctx context.Context, msg *pubsub.Message) google.PublishResult {
					return &mock.MockPublishResult{Result: "mock-server-id", Err: nil}
				},
			},
		},
	}

	provider := &raiden.GooglePubSubProvider{
		Config: &raiden.Config{GoogleProjectId: "test-project", GoogleSaPath: "test-path"},
		Provider: &google.Provider{
			Config: &google.ProviderConfig{GoogleProjectId: "test-project", GoogleSaPath: "test-path"},
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
		Provider: &google.Provider{
			Config: &google.ProviderConfig{GoogleProjectId: "test-project", GoogleSaPath: "test-path"},
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
		Provider: &google.Provider{
			Config: &google.ProviderConfig{GoogleProjectId: "test-project", GoogleSaPath: "test-path"},
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
		Provider: &google.Provider{
			Config: &google.ProviderConfig{GoogleProjectId: "test-project", GoogleSaPath: "test-path"},
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
		Provider: &google.Provider{
			Config: &google.ProviderConfig{GoogleProjectId: "test-project", GoogleSaPath: "test-path"},
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
		Provider: &google.Provider{
			Config: &google.ProviderConfig{GoogleProjectId: "test-project", GoogleSaPath: "test-path"},
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

// ----- SubscriberBase Default Tests -----

func TestSubscriberBase_AllDefaults(t *testing.T) {
	base := &raiden.SubscriberBase{}
	assert.Equal(t, raiden.PubSubProviderUnknown, base.Provider())
	assert.Equal(t, "", base.PushEndpoint())
	assert.Equal(t, "", base.Topic())
	assert.Equal(t, true, base.AutoAck())
	assert.Equal(t, "unknown", base.Name())
	assert.Equal(t, raiden.SubscriptionTypePull, base.SubscriptionType())
	assert.Equal(t, "", base.Subscription())
	assert.Equal(t, raiden.RealtimeChannelType(""), base.ChannelType())
	assert.Equal(t, "public", base.Schema())
	assert.Equal(t, "", base.Table())
	assert.Equal(t, "*", base.EventFilter())
	assert.Error(t, base.Consume(nil, raiden.SubscriberMessage{}))
}

// ----- PubSubManager.Serve missing branch tests -----

func TestPubSubManager_Serve_NotPush(t *testing.T) {
	mgr := &raiden.PubSubManager{}
	mgr.SetConfig(loadConfig())
	mgr.SetProvider(raiden.PubSubProviderGoogle, &mock.MockProvider{})

	handler := &mock.MockSubscriberHandler{
		NameValue:             "test",
		ProviderValue:         raiden.PubSubProviderGoogle,
		SubscriptionTypeValue: raiden.SubscriptionTypePull,
	}

	_, err := mgr.Serve(handler)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not push subscription")
}

func TestPubSubManager_Serve_UnknownProvider(t *testing.T) {
	mgr := &raiden.PubSubManager{}
	mgr.SetConfig(loadConfig())

	handler := &mock.MockSubscriberHandler{
		NameValue:             "test",
		ProviderValue:         raiden.PubSubProviderType("nonexistent"),
		SubscriptionTypeValue: raiden.SubscriptionTypePush,
	}

	_, err := mgr.Serve(handler)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported pubsub provider")
}

// ----- Legacy Adapter Full Coverage -----

func TestWrapLegacySubscriber_AllMethods(t *testing.T) {
	legacy := &legacyHandler{}
	wrapped := raiden.WrapLegacySubscriber(legacy)

	assert.Equal(t, true, wrapped.AutoAck())
	assert.Equal(t, raiden.RealtimeChannelType(""), wrapped.ChannelType())
	assert.Equal(t, "*", wrapped.EventFilter())
	assert.Equal(t, "legacy-test", wrapped.Name())
	assert.Equal(t, raiden.PubSubProviderGoogle, wrapped.Provider())
	assert.Equal(t, "", wrapped.PushEndpoint())
	assert.Equal(t, "public", wrapped.Schema())
	assert.Equal(t, "legacy-sub", wrapped.Subscription())
	assert.Equal(t, raiden.SubscriptionTypePull, wrapped.SubscriptionType())
	assert.Equal(t, "", wrapped.Table())
	assert.Equal(t, "", wrapped.Topic())
}

// ----- SupabaseRealtimeProvider Adapter Direct Tests -----

func TestSupabaseRealtimeProvider_CreateSubscription(t *testing.T) {
	adapter := &raiden.SupabaseRealtimeProvider{
		Config:   loadConfig(),
		Provider: &supabasepubsub.Provider{},
	}

	handler := &mock.MockSubscriberHandler{
		NameValue:  "test-sub",
		TopicValue: "test-topic",
	}

	err := adapter.CreateSubscription(handler)
	assert.NoError(t, err)
}

func TestSupabaseRealtimeProvider_StopListen(t *testing.T) {
	adapter := &raiden.SupabaseRealtimeProvider{
		Config:   loadConfig(),
		Provider: &supabasepubsub.Provider{},
	}

	err := adapter.StopListen()
	assert.NoError(t, err)
}

func TestSupabaseRealtimeProvider_StartListen(t *testing.T) {
	adapter := &raiden.SupabaseRealtimeProvider{
		Config: loadConfig(),
		Provider: &supabasepubsub.Provider{
			Config: &supabasepubsub.ProviderConfig{
				SupabasePublicUrl: "https://test.supabase.co",
				AnonKey:           "test-key",
				ProjectId:         "test-project",
			},
			Dialer: &mockDialer{err: fmt.Errorf("dial failed")},
		},
	}

	handler := &mock.MockSubscriberHandler{
		NameValue:             "test-handler",
		TopicValue:            "test-room",
		ChannelTypeValue:      raiden.RealtimeChannelBroadcast,
		SchemaValue:           "public",
		TableValue:            "",
		EventFilterValue:      "*",
		SubscriptionTypeValue: raiden.SubscriptionTypePull,
		ConsumeFunc: func(ctx raiden.SubscriberContext, msg raiden.SubscriberMessage) error {
			return nil
		},
	}

	// StartListen will fail at connect because dialer returns error
	err := adapter.StartListen([]raiden.SubscriberHandler{handler})
	assert.Error(t, err)
}

// Mock Dialer for Supabase adapter tests
type mockDialer struct {
	conn supabasepubsub.WebSocketConn
	err  error
}

func (m *mockDialer) Dial(url string) (supabasepubsub.WebSocketConn, error) {
	return m.conn, m.err
}

// ----- Google Provider Adapter Tests -----

func TestGooglePubSubProvider_CreateSubscription_NotPush(t *testing.T) {
	provider := &raiden.GooglePubSubProvider{
		Config: &raiden.Config{GoogleProjectId: "test-project"},
		Provider: &google.Provider{
			Config: &google.ProviderConfig{GoogleProjectId: "test-project"},
		},
	}

	handler := &mock.MockSubscriberHandler{
		NameValue:             "test",
		SubscriptionTypeValue: raiden.SubscriptionTypePull,
	}

	err := provider.CreateSubscription(handler)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not push subscription")
}

func TestGooglePubSubProvider_Serve(t *testing.T) {
	mockClient := &mock.MockPubSubClient{
		CreateSubscriptionFn: func(ctx context.Context, id string, cfg pubsub.SubscriptionConfig) (google.Subscription, error) {
			return &mock.MockSubscription{Id: id}, nil
		},
		Topics: map[string]*mock.MockTopic{
			"test-topic": {
				PublishFn: func(ctx context.Context, msg *pubsub.Message) google.PublishResult {
					return &mock.MockPublishResult{Result: "ok", Err: nil}
				},
			},
		},
	}

	config := &raiden.Config{GoogleProjectId: "test-project", ServerDns: "http://localhost"}
	provider := &raiden.GooglePubSubProvider{
		Config: config,
		Provider: &google.Provider{
			Config: &google.ProviderConfig{GoogleProjectId: "test-project"},
			Client: mockClient,
		},
	}

	handler := &mock.MockSubscriberHandler{
		NameValue:             "test-handler",
		TopicValue:            "test-topic",
		SubscriptionValue:     "test-sub",
		PushEndpointValue:     "/webhook",
		SubscriptionTypeValue: raiden.SubscriptionTypePush,
		ConsumeFunc: func(ctx raiden.SubscriberContext, msg raiden.SubscriberMessage) error {
			return nil
		},
	}

	httpHandler, err := provider.Serve(config, handler)
	assert.NoError(t, err)
	assert.NotNil(t, httpHandler)

	// Test with valid JSON body
	reqCtx := &fasthttp.RequestCtx{}
	reqCtx.Request.SetBodyString(`{"message":{"data":"dGVzdA==","message_id":"123","publish_time":"2026-01-01"},"subscription":"projects/test-project/subscriptions/test-sub"}`)
	httpHandler(reqCtx)
	assert.Equal(t, fasthttp.StatusOK, reqCtx.Response.StatusCode())

	// Test with invalid JSON body
	reqCtx2 := &fasthttp.RequestCtx{}
	reqCtx2.Request.SetBodyString(`not json`)
	httpHandler(reqCtx2)
	assert.Contains(t, string(reqCtx2.Response.Body()), "invalid json data")

	// Test with wrong subscription
	reqCtx3 := &fasthttp.RequestCtx{}
	reqCtx3.Request.SetBodyString(`{"message":{"data":"test"},"subscription":"wrong-sub"}`)
	httpHandler(reqCtx3)
	assert.Equal(t, fasthttp.StatusUnprocessableEntity, reqCtx3.Response.StatusCode())
}

func TestGooglePubSubProvider_Serve_ConsumeError(t *testing.T) {
	mockClient := &mock.MockPubSubClient{
		CreateSubscriptionFn: func(ctx context.Context, id string, cfg pubsub.SubscriptionConfig) (google.Subscription, error) {
			return &mock.MockSubscription{Id: id}, nil
		},
		Topics: map[string]*mock.MockTopic{
			"test-topic": {
				PublishFn: func(ctx context.Context, msg *pubsub.Message) google.PublishResult {
					return &mock.MockPublishResult{Result: "ok", Err: nil}
				},
			},
		},
	}

	config := &raiden.Config{GoogleProjectId: "test-project", ServerDns: "http://localhost"}
	provider := &raiden.GooglePubSubProvider{
		Config: config,
		Provider: &google.Provider{
			Config: &google.ProviderConfig{GoogleProjectId: "test-project"},
			Client: mockClient,
		},
	}

	handler := &mock.MockSubscriberHandler{
		NameValue:             "test-handler",
		TopicValue:            "test-topic",
		SubscriptionValue:     "test-sub",
		PushEndpointValue:     "/webhook",
		SubscriptionTypeValue: raiden.SubscriptionTypePush,
		ConsumeFunc: func(ctx raiden.SubscriberContext, msg raiden.SubscriberMessage) error {
			return errors.New("consume failed")
		},
	}

	httpHandler, err := provider.Serve(config, handler)
	assert.NoError(t, err)

	reqCtx := &fasthttp.RequestCtx{}
	reqCtx.Request.SetBodyString(`{"message":{"data":"test","message_id":"1","publish_time":"t"},"subscription":"projects/test-project/subscriptions/test-sub"}`)
	httpHandler(reqCtx)
	assert.Equal(t, fasthttp.StatusInternalServerError, reqCtx.Response.StatusCode())
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
