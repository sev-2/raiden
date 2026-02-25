package raiden

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"time"

	"github.com/fasthttp/websocket"
	"github.com/oklog/run"
	"github.com/sev-2/raiden/pkg/client/net"
	"github.com/sev-2/raiden/pkg/logger"
	googlepubsub "github.com/sev-2/raiden/pkg/pubsub/google"
	supabasepubsub "github.com/sev-2/raiden/pkg/pubsub/supabase"
	"github.com/sev-2/raiden/pkg/supabase/constants"
	"github.com/valyala/fasthttp"
	"go.opentelemetry.io/otel/trace"
)

var PubSubLogger = logger.HcLog().Named("raiden.pubsub")

// ----- Type Definition -----
type PubSubProviderType string

const (
	PubSubProviderGoogle   PubSubProviderType = "google"
	PubSubProviderSupabase PubSubProviderType = "supabase"
	PubSubProviderUnknown  PubSubProviderType = "unknown"
)

type SubscriptionType string

const (
	SubscriptionTypePull SubscriptionType = "pull"
	SubscriptionTypePush SubscriptionType = "push"
)

const SubscriptionPrefixEndpoint = "pubsub-endpoint"

type RealtimeChannelType string

const (
	RealtimeChannelBroadcast       RealtimeChannelType = "broadcast"
	RealtimeChannelPresence        RealtimeChannelType = "presence"
	RealtimeChannelPostgresChanges RealtimeChannelType = "postgres_changes"
)

// SubscriberMessage is a provider-agnostic message wrapper.
type SubscriberMessage struct {
	Data       []byte
	Attributes map[string]string
	// Raw holds the original provider-specific message (e.g., *pubsub.Message).
	Raw any
}

// ----- Subscription Context -----
type SubscriberContext interface {
	Config() *Config
	Span() trace.Span
	SetSpan(span trace.Span)
	HttpRequest(method string, url string, body []byte, headers map[string]string, timeout time.Duration, response any) error
}

type subscriberContext struct {
	context.Context
	cfg  *Config
	span trace.Span
}

func (ctx *subscriberContext) Config() *Config {
	return ctx.cfg
}

func (ctx *subscriberContext) Span() trace.Span {
	return ctx.span
}

func (ctx *subscriberContext) SetSpan(span trace.Span) {
	ctx.span = span
}

func (c *subscriberContext) HttpRequest(method string, url string, body []byte, headers map[string]string, timeout time.Duration, response any) error {
	if reflect.TypeOf(response).Kind() != reflect.Ptr {
		return errors.New("response payload must be pointer")
	}

	byteData, err := net.SendRequest(method, url, body, timeout, func(req *http.Request) error {
		currentHeaders := req.Header.Clone()
		if len(headers) > 0 {
			for k, v := range headers {
				currentHeaders.Set(k, v)
			}
		}
		req.Header = currentHeaders

		return nil
	}, nil)

	if err != nil {
		return err
	}

	return json.Unmarshal(byteData, response)
}

// ----- Subscription Handler -----
type SubscriberHandler interface {
	AutoAck() bool
	Name() string
	Consume(ctx SubscriberContext, message SubscriberMessage) error
	ChannelType() RealtimeChannelType
	EventFilter() string
	Provider() PubSubProviderType
	PushEndpoint() string
	Schema() string
	Subscription() string
	SubscriptionType() SubscriptionType
	Table() string
	Topic() string
}

type SubscriberBase struct{}

func (s *SubscriberBase) AutoAck() bool {
	return true
}

func (s *SubscriberBase) ChannelType() RealtimeChannelType {
	return ""
}

func (s *SubscriberBase) EventFilter() string {
	return constants.RealtimeEventAll
}

func (s *SubscriberBase) Name() string {
	return "unknown"
}

func (s *SubscriberBase) Provider() PubSubProviderType {
	return PubSubProviderUnknown
}

func (s *SubscriberBase) Schema() string {
	return "public"
}

func (s *SubscriberBase) Subscription() string {
	return ""
}

func (s *SubscriberBase) PushEndpoint() string {
	return ""
}

func (s *SubscriberBase) Table() string {
	return ""
}

func (s *SubscriberBase) Topic() string {
	return ""
}

func (s *SubscriberBase) SubscriptionType() SubscriptionType {
	return SubscriptionTypePull
}

func (s *SubscriberBase) Consume(ctx SubscriberContext, message SubscriberMessage) error {
	return fmt.Errorf("subscriber %s is not implemented", s.Name())
}

// ----- Subscription Server -----
type PubSub interface {
	Register(handler SubscriberHandler)
	Publish(ctx context.Context, provider PubSubProviderType, topic string, message []byte) error
	Listen()
	Serve(handler SubscriberHandler) (fasthttp.RequestHandler, error)
	Handlers() []SubscriberHandler
}

func NewPubsub(config *Config, tracer trace.Tracer) PubSub {
	mgr := &PubSubManager{
		config:    config,
		providers: make(map[PubSubProviderType]PubSubProvider),
	}
	mgr.providers[PubSubProviderGoogle] = newGoogleProviderAdapter(config, tracer)
	mgr.providers[PubSubProviderSupabase] = newSupabaseProviderAdapter(config)
	return mgr
}

type PubSubManager struct {
	config    *Config
	handlers  []SubscriberHandler
	providers map[PubSubProviderType]PubSubProvider
}

// Register implements PubSub.
func (s *PubSubManager) Register(handler SubscriberHandler) {
	s.handlers = append(s.handlers, handler)
}

func (s *PubSubManager) Handlers() []SubscriberHandler {
	return s.handlers
}

func (s *PubSubManager) GetHandlerCount() int {
	return len(s.handlers)
}

func (s *PubSubManager) SetConfig(cfg *Config) {
	s.config = cfg
}

func (s *PubSubManager) SetProvider(providerType PubSubProviderType, provider PubSubProvider) {
	if s.providers == nil {
		s.providers = make(map[PubSubProviderType]PubSubProvider)
	}
	s.providers[providerType] = provider
}

func (s *PubSubManager) getProvider(providerType PubSubProviderType) (PubSubProvider, error) {
	p, ok := s.providers[providerType]
	if !ok {
		return nil, fmt.Errorf("unsupported pubsub provider: %s", providerType)
	}
	return p, nil
}

func (s *PubSubManager) Serve(handler SubscriberHandler) (fasthttp.RequestHandler, error) {
	if handler.SubscriptionType() != SubscriptionTypePush {
		return nil, fmt.Errorf("subscription %s is not push subscription", handler.Name())
	}

	p, err := s.getProvider(handler.Provider())
	if err != nil {
		return nil, err
	}

	return p.Serve(s.config, handler)
}

// Listen starts pull-subscription listeners for all registered providers.
func (s *PubSubManager) Listen() {
	// Group handlers by provider
	providerHandlers := make(map[PubSubProviderType][]SubscriberHandler)
	for _, h := range s.handlers {
		if h.SubscriptionType() == SubscriptionTypePull {
			providerHandlers[h.Provider()] = append(providerHandlers[h.Provider()], h)
		}
	}

	if len(providerHandlers) == 0 {
		return
	}

	var g run.Group
	for pt, handlers := range providerHandlers {
		p, err := s.getProvider(pt)
		if err != nil {
			PubSubLogger.Error("provider not found for listen", "provider", pt, "message", err)
			continue
		}

		provider := p
		h := handlers
		g.Add(func() error {
			return provider.StartListen(h)
		}, func(err error) {
			if err := provider.StopListen(); err != nil {
				PubSubLogger.Error("failed stop listener", "provider", pt, "message", err)
			}
		})
	}

	if err := g.Run(); err != nil {
		PubSubLogger.Error("stop subscribe", "message", err)
	}
}

func (s *PubSubManager) Publish(ctx context.Context, provider PubSubProviderType, topic string, message []byte) error {
	p, err := s.getProvider(provider)
	if err != nil {
		return err
	}
	return p.Publish(ctx, topic, message)
}

// ----- Pub Sub Provider -----
type PubSubProvider interface {
	Publish(ctx context.Context, topic string, message []byte) error
	CreateSubscription(SubscriberHandler) error
	Serve(config *Config, handler SubscriberHandler) (fasthttp.RequestHandler, error)
	StartListen(handler []SubscriberHandler) error
	StopListen() error
}

// ----- Google Provider Adapter -----
// GooglePubSubProvider adapts google.Provider to the PubSubProvider interface.
type GooglePubSubProvider struct {
	Config   *Config
	Provider *googlepubsub.Provider
}

func newGoogleProviderAdapter(config *Config, tracer trace.Tracer) *GooglePubSubProvider {
	var providerConfig *googlepubsub.ProviderConfig
	if config != nil {
		providerConfig = &googlepubsub.ProviderConfig{
			GoogleProjectId: config.GoogleProjectId,
			GoogleSaPath:    config.GoogleSaPath,
		}
	}

	return &GooglePubSubProvider{
		Config: config,
		Provider: &googlepubsub.Provider{
			Config: providerConfig,
			Tracer: tracer,
		},
	}
}

func (a *GooglePubSubProvider) Publish(ctx context.Context, topic string, message []byte) error {
	return a.Provider.Publish(ctx, topic, message)
}

func (a *GooglePubSubProvider) CreateSubscription(handler SubscriberHandler) error {
	if handler.SubscriptionType() != SubscriptionTypePush {
		return fmt.Errorf("%s is not push subscription", handler.Name())
	}

	return a.Provider.CreateSubscription(
		handler.Name(),
		handler.Topic(),
		handler.PushEndpoint(),
		a.Config.ServerDns,
		handler.Subscription(),
		SubscriptionPrefixEndpoint,
	)
}

// Google-specific push envelope types (used only by Serve).
type googlePushMessage struct {
	Data        string `json:"data"`
	MessageId   string `json:"message_id"`
	PublishTime string `json:"publish_time"`
}

type googlePushData struct {
	Message      googlePushMessage `json:"message"`
	Subscription string            `json:"subscription"`
}

func (a *GooglePubSubProvider) Serve(config *Config, handler SubscriberHandler) (fasthttp.RequestHandler, error) {
	if err := a.CreateSubscription(handler); err != nil {
		return nil, err
	}

	subscriptionSignature := fmt.Sprintf("projects/%s/subscriptions/%s", config.GoogleProjectId, handler.Subscription())

	return func(ctx *fasthttp.RequestCtx) {
		var subCtx = subscriberContext{
			cfg: config, Context: context.Background(),
		}

		ctx.SetContentType("application/json")

		var data googlePushData
		if err := json.Unmarshal(ctx.Request.Body(), &data); err != nil {
			_, _ = ctx.WriteString("{\"message\":\"invalid json data\"}")
			return
		}

		if data.Subscription != subscriptionSignature {
			ctx.SetStatusCode(fasthttp.StatusUnprocessableEntity)
			_, _ = ctx.WriteString("{\"message\":\"subscription validation failed: received unexpected subscription name\"}")
			return
		}

		msg := SubscriberMessage{
			Data: []byte(data.Message.Data),
			Attributes: map[string]string{
				"message_id":   data.Message.MessageId,
				"publish_time": data.Message.PublishTime,
			},
		}

		response := map[string]any{"message": "success handle"}

		if err := handler.Consume(&subCtx, msg); err != nil {
			ctx.SetStatusCode(fasthttp.StatusInternalServerError)
			response["message"] = err.Error()
		}

		resByte, err := json.Marshal(response)
		if err != nil {
			_, _ = ctx.WriteString("{\"message\":\"internal server error\"}")
			return
		}
		_, _ = ctx.Write(resByte)
	}, nil
}

func (a *GooglePubSubProvider) StartListen(handlers []SubscriberHandler) error {
	listenHandlers := make([]googlepubsub.ListenHandler, len(handlers))
	for i, h := range handlers {
		handler := h
		listenHandlers[i] = googlepubsub.ListenHandler{
			Name:         handler.Name(),
			Subscription: handler.Subscription(),
			AutoAck:      handler.AutoAck(),
			ConsumeFn: func(ctx context.Context, span trace.Span, msg any) error {
				subCtx := &subscriberContext{
					cfg:     a.Config,
					Context: ctx,
				}
				if span != nil {
					subCtx.SetSpan(span)
				}
				return handler.Consume(subCtx, toSubscriberMessage(msg))
			},
		}
	}
	return a.Provider.StartListen(listenHandlers)
}

func (a *GooglePubSubProvider) StopListen() error {
	return a.Provider.StopListen()
}

// ----- Supabase Realtime Provider Adapter -----
// SupabaseRealtimeProvider adapts supabase.Provider to the PubSubProvider interface.
type SupabaseRealtimeProvider struct {
	Config   *Config
	Provider *supabasepubsub.Provider
}

func newSupabaseProviderAdapter(config *Config) *SupabaseRealtimeProvider {
	var providerConfig *supabasepubsub.ProviderConfig
	if config != nil {
		providerConfig = &supabasepubsub.ProviderConfig{
			SupabasePublicUrl: config.SupabasePublicUrl,
			SupabaseApiUrl:    config.SupabaseApiUrl,
			AnonKey:           config.AnonKey,
			ServiceKey:        config.ServiceKey,
			ProjectId:         config.ProjectId,
		}
	}

	return &SupabaseRealtimeProvider{
		Config: config,
		Provider: &supabasepubsub.Provider{
			Config: providerConfig,
			Dialer: &DefaultWebSocketDialer{},
		},
	}
}

func (a *SupabaseRealtimeProvider) Publish(ctx context.Context, topic string, message []byte) error {
	return a.Provider.Publish(ctx, topic, message)
}

func (a *SupabaseRealtimeProvider) CreateSubscription(handler SubscriberHandler) error {
	return a.Provider.CreateSubscription(handler.Name(), handler.Topic())
}

// Serve returns an error for Supabase Realtime — it uses WebSocket, not HTTP push.
func (a *SupabaseRealtimeProvider) Serve(config *Config, handler SubscriberHandler) (fasthttp.RequestHandler, error) {
	return nil, fmt.Errorf("supabase realtime does not support HTTP push endpoints; use pull/websocket mode")
}

func (a *SupabaseRealtimeProvider) StartListen(handlers []SubscriberHandler) error {
	listenHandlers := make([]supabasepubsub.ListenHandler, len(handlers))
	for i, h := range handlers {
		handler := h
		listenHandlers[i] = supabasepubsub.ListenHandler{
			Name:        handler.Name(),
			Topic:       handler.Topic(),
			ChannelType: supabasepubsub.RealtimeChannelType(handler.ChannelType()),
			Table:       handler.Table(),
			Schema:      handler.Schema(),
			EventFilter: handler.EventFilter(),
			ConsumeFn: func(ctx context.Context, event string, payload json.RawMessage) error {
				subCtx := &subscriberContext{
					cfg:     a.Config,
					Context: ctx,
				}
				msg := SubscriberMessage{
					Data: []byte(payload),
					Attributes: map[string]string{
						"channel_type": string(handler.ChannelType()),
						"event":        event,
						"topic":        handler.Topic(),
					},
					Raw: payload,
				}
				return handler.Consume(subCtx, msg)
			},
		}
	}
	return a.Provider.StartListen(listenHandlers)
}

func (a *SupabaseRealtimeProvider) StopListen() error {
	return a.Provider.StopListen()
}

// DefaultWebSocketDialer wraps github.com/fasthttp/websocket for production use.
type DefaultWebSocketDialer struct{}

func (d *DefaultWebSocketDialer) Dial(wsUrl string) (supabasepubsub.WebSocketConn, error) {
	conn, _, err := websocket.DefaultDialer.Dial(wsUrl, nil)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// ----- Backward Compatibility -----
// Deprecated: PushSubscriptionMessage is an alias for backward compatibility.
// Use SubscriberMessage instead.
type PushSubscriptionMessage = googlePushMessage

// Deprecated: PushSubscriptionData is an alias for backward compatibility.
// Use SubscriberMessage instead.
type PushSubscriptionData = googlePushData

// LegacySubscriberConsumer allows old subscribers with Consume(ctx, message any)
// to be wrapped as a SubscriberHandler via WrapLegacySubscriber.
type LegacySubscriberConsumer interface {
	AutoAck() bool
	Name() string
	// Consume receives the raw provider message (e.g. *pubsub.Message).
	// Deprecated: implement SubscriberHandler.Consume(ctx, SubscriberMessage) instead.
	Consume(ctx SubscriberContext, message any) error
	Provider() PubSubProviderType
	PushEndpoint() string
	Subscription() string
	SubscriptionType() SubscriptionType
	Topic() string
}

// legacySubscriberAdapter wraps a LegacySubscriberConsumer as a SubscriberHandler.
type legacySubscriberAdapter struct {
	legacy LegacySubscriberConsumer
}

// WrapLegacySubscriber wraps a subscriber that uses the old Consume(ctx, message any)
// signature so it satisfies the current SubscriberHandler interface.
//
// Deprecated: migrate to SubscriberHandler.Consume(ctx, SubscriberMessage) directly.
// The Raw field of SubscriberMessage contains the original provider message.
func WrapLegacySubscriber(s LegacySubscriberConsumer) SubscriberHandler {
	return &legacySubscriberAdapter{legacy: s}
}

func (a *legacySubscriberAdapter) AutoAck() bool                    { return a.legacy.AutoAck() }
func (a *legacySubscriberAdapter) ChannelType() RealtimeChannelType { return "" }
func (a *legacySubscriberAdapter) EventFilter() string              { return constants.RealtimeEventAll }
func (a *legacySubscriberAdapter) Name() string                     { return a.legacy.Name() }
func (a *legacySubscriberAdapter) Provider() PubSubProviderType     { return a.legacy.Provider() }
func (a *legacySubscriberAdapter) PushEndpoint() string             { return a.legacy.PushEndpoint() }
func (a *legacySubscriberAdapter) Schema() string                   { return "public" }
func (a *legacySubscriberAdapter) Subscription() string             { return a.legacy.Subscription() }
func (a *legacySubscriberAdapter) SubscriptionType() SubscriptionType {
	return a.legacy.SubscriptionType()
}
func (a *legacySubscriberAdapter) Table() string { return "" }
func (a *legacySubscriberAdapter) Topic() string { return a.legacy.Topic() }
func (a *legacySubscriberAdapter) Consume(ctx SubscriberContext, message SubscriberMessage) error {
	// Forward Raw (original provider message) to legacy consumer.
	// For push subscriptions Raw may be nil; fall back to the full SubscriberMessage.
	msg := message.Raw
	if msg == nil {
		msg = message
	}
	return a.legacy.Consume(ctx, msg)
}

// toSubscriberMessage converts a provider-specific message to SubscriberMessage.
func toSubscriberMessage(msg any) SubscriberMessage {
	type pubsubMsg interface {
		GetData() []byte
		GetAttributes() map[string]string
	}

	// Try duck-typing first for testability
	if m, ok := msg.(pubsubMsg); ok {
		return SubscriberMessage{
			Data:       m.GetData(),
			Attributes: m.GetAttributes(),
			Raw:        msg,
		}
	}

	// Direct field access for cloud.google.com/go/pubsub.Message
	v := reflect.ValueOf(msg)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	sm := SubscriberMessage{Raw: msg}
	if dataField := v.FieldByName("Data"); dataField.IsValid() {
		if b, ok := dataField.Interface().([]byte); ok {
			sm.Data = b
		}
	}
	if attrField := v.FieldByName("Attributes"); attrField.IsValid() {
		if a, ok := attrField.Interface().(map[string]string); ok {
			sm.Attributes = a
		}
	}
	return sm
}
