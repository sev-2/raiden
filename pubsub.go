package raiden

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"cloud.google.com/go/pubsub"
	"github.com/oklog/run"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/pubsub/google"
	"github.com/valyala/fasthttp"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/api/option"
)

var PubSubLogger = logger.HcLog().Named("raiden.pubsub")

// ----- Type Definition -----
type PubSubProviderType string

const (
	PubSubProviderGoogle  PubSubProviderType = "google"
	PubSubProviderUnknown PubSubProviderType = "unknown"
)

type SubscriptionType string

const (
	SubscriptionTypePull SubscriptionType = "pull"
	SubscriptionTypePush SubscriptionType = "push"
)

const SubscriptionPrefixEndpoint = "pubsub-endpoint"

type PushSubscriptionMessage struct {
	Data         string `json:"data"`
	MessageId    string `json:"message_id"`
	Publish_time string `json:"publish_time"`
}

type PushSubscriptionData struct {
	Message      PushSubscriptionMessage `json:"message"`
	Subscription string                  `json:"subscription"`
}

// ----- Subscription Context -----
type SubscriberContext interface {
	Config() *Config
	Span() trace.Span
	SetSpan(span trace.Span)
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

// ----- Subscription Handler -----
type SubscriberHandler interface {
	AutoAck() bool
	Name() string
	Consume(ctx SubscriberContext, message any) error
	Provider() PubSubProviderType
	PushEndpoint() string
	Subscription() string
	SubscriptionType() SubscriptionType
	Topic() string
}

type SubscriberBase struct{}

func (s *SubscriberBase) AutoAck() bool {
	return true
}

func (s *SubscriberBase) Name() string {
	return "unknown"
}

func (s *SubscriberBase) Provider() PubSubProviderType {
	return PubSubProviderUnknown
}

func (s *SubscriberBase) Subscription() string {
	return ""
}

func (s *SubscriberBase) PushEndpoint() string {
	return ""
}

func (s *SubscriberBase) Topic() string {
	return ""
}

func (s *SubscriberBase) SubscriptionType() SubscriptionType {
	return SubscriptionTypePull
}

func (s *SubscriberBase) Consume(ctx SubscriberContext, message any) error {
	return fmt.Errorf("subscriber %s in not implemented", s.Name())
}

func (s *SubscriberBase) ParsePushSubscriptionMessage(message any) (*PushSubscriptionData, error) {
	byteData, valid := message.([]byte)
	if !valid {
		return nil, errors.New("push subscription message is not valid")
	}

	var data PushSubscriptionData
	if err := json.Unmarshal(byteData, &data); err != nil {
		return nil, err
	}

	return &data, nil
}

// ----- Subscription Server -----
type PubSub interface {
	Register(handler SubscriberHandler)
	Publish(ctx context.Context, provider PubSubProviderType, topic string, message []byte) error
	Listen()
	Serve(handle SubscriberHandler) (fasthttp.RequestHandler, error)
	Handlers() []SubscriberHandler
}

func NewPubsub(config *Config, tracer trace.Tracer) PubSub {
	return &PubSubManager{
		config: config,
		provider: pubSubProvider{
			google: &GooglePubSubProvider{
				Config: config,
				Tracer: tracer,
			},
		},
	}

}

type PubSubManager struct {
	config   *Config
	handlers []SubscriberHandler
	provider pubSubProvider
}

type pubSubProvider struct {
	// all provider
	google PubSubProvider
}

// Register implements Subscriber.
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
	switch providerType {
	case PubSubProviderGoogle:
		s.provider.google = provider
	}
}

func (s *PubSubManager) Serve(handler SubscriberHandler) (fasthttp.RequestHandler, error) {

	if handler.SubscriptionType() != SubscriptionTypePush {
		return nil, fmt.Errorf("subscription %s is not push subscription", handler.Name())
	}

	if err := s.provider.google.CreateSubscription(handler); err != nil {
		return nil, err
	}

	return func(ctx *fasthttp.RequestCtx) {
		var subCtx = subscriberContext{
			cfg: s.config, Context: context.Background(),
		}

		ctx.SetContentType("application/json")
		response := map[string]any{
			"message": "success handle",
		}

		if err := handler.Consume(&subCtx, ctx.Request.Body()); err != nil {
			ctx.SetStatusCode(http.StatusInternalServerError)
			response["message"] = err.Error()
			resByte, err := json.Marshal(response)
			if err != nil {
				errMsg := fmt.Sprintf("{\"message\":\"%s\"}", err.Error())
				if _, writeErr := ctx.WriteString(errMsg); writeErr != nil {
					// Log the error if necessary
					PubSubLogger.Error(fmt.Sprintf("%s endpoint handler ctx.WriteString() error", handler.Name()), "message", writeErr)
				}
				return
			}

			if _, err := ctx.Write(resByte); err != nil {
				PubSubLogger.Error(fmt.Sprintf("%s endpoint handler ctx.Write() error", handler.Name()), "message", err)
			}
			return
		}

		resByte, err := json.Marshal(response)
		if err != nil {
			errMsg := fmt.Sprintf("{\"message\":\"%s\"}", err.Error())
			if _, writeErr := ctx.WriteString(errMsg); writeErr != nil {
				// Log the error if necessary
				PubSubLogger.Error(fmt.Sprintf("%s endpoint handler ctx.WriteString() error in last step", handler.Name()), "message", writeErr)
			}
			return
		}
		if _, err := ctx.Write(resByte); err != nil {
			PubSubLogger.Error(fmt.Sprintf("%s endpoint handler ctx.Write() error ins last step", handler.Name()), "message", err)
		}
	}, nil

}

// StartListen implements Subscriber.
func (s *PubSubManager) Listen() {
	var g run.Group

	var googlePullHandlers []SubscriberHandler
	for _, h := range s.handlers {
		if h.Provider() == PubSubProviderGoogle && h.SubscriptionType() == SubscriptionTypePull {
			googlePullHandlers = append(googlePullHandlers, h)
		}
	}

	if len(googlePullHandlers) > 0 {
		g.Add(func() error {
			return s.provider.google.StartListen(googlePullHandlers)
		}, func(err error) {
			if err := s.provider.google.StopListen(); err != nil {
				PubSubLogger.Error("failed stop listener", "message", err)
			}
		})

		if err := g.Run(); err != nil {
			PubSubLogger.Error("stop subscribe", "message", err)
		}
	}

}

func (s *PubSubManager) Publish(ctx context.Context, provider PubSubProviderType, topic string, message []byte) error {
	switch provider {
	case PubSubProviderGoogle:
		return s.provider.google.Publish(ctx, topic, message)
	}
	return errors.New("unsupported pubsub provider")
}

// ----- Pub Sub Provider -----
type PubSubProvider interface {
	Publish(ctx context.Context, topic string, message []byte) error
	CreateSubscription(SubscriberHandler) error
	StartListen(handler []SubscriberHandler) error
	StopListen() error
}

type GooglePubSubProvider struct {
	Config *Config
	Client google.PubSubClient
	Tracer trace.Tracer
}

func (s *GooglePubSubProvider) validate() error {
	if s.Config.GoogleProjectId == "" {
		return errors.New("env GOOGLE_PROJECT_ID is required")
	}

	if s.Config.GoogleSaPath == "" {
		return errors.New("env GOOGLE_SA_PATH is required")
	}

	return nil
}

func (s *GooglePubSubProvider) createClient() error {
	client, err := pubsub.NewClient(context.Background(), s.Config.GoogleProjectId, option.WithCredentialsFile(s.Config.GoogleSaPath))
	if err != nil {
		PubSubLogger.Error("failed create google pubsub client")
		return err
	}

	s.Client = &google.GooglePubSubClient{Client: client}
	return nil
}

func (s *GooglePubSubProvider) CreateSubscription(handler SubscriberHandler) error {
	if handler.SubscriptionType() != SubscriptionTypePush {
		return fmt.Errorf("%s is not push subscription", handler.Name())
	}

	if handler.Topic() == "" {
		return fmt.Errorf("topic in subscription %s ir required", handler.Name())
	}

	if handler.PushEndpoint() == "" {
		return fmt.Errorf("push endpoint in subscription %s ir required", handler.Name())
	}

	if s.Config.ServerDns == "" {
		return errors.New("the SERVER_DSN configuration is required when using the 'push' subscription type.")
	}

	if s.Client == nil {
		if err := s.createClient(); err != nil {
			return err
		}
	}

	t := s.Client.Topic(handler.Topic())

	var endpoint string
	if strings.HasPrefix("/", handler.PushEndpoint()) {
		endpoint = fmt.Sprintf("%s/%s%s", s.Config.ServerDns, SubscriptionPrefixEndpoint, handler.PushEndpoint())
	} else {
		endpoint = fmt.Sprintf("%s/%s/%s", s.Config.ServerDns, SubscriptionPrefixEndpoint, handler.PushEndpoint())
	}

	_, err := s.Client.CreateSubscription(context.Background(), handler.Subscription(), pubsub.SubscriptionConfig{
		Topic: t.GetInstance(),
		PushConfig: pubsub.PushConfig{
			Endpoint: endpoint,
		},
	})

	if strings.Contains(err.Error(), "Resource already exists in the project") {
		PubSubLogger.Info("google - subscription already exists in the project", "topic", handler.Topic(), "subscription-id", handler.Subscription())
		return nil
	}

	return err
}

func (s *GooglePubSubProvider) StartListen(handler []SubscriberHandler) error {
	if err := s.validate(); err != nil {
		return err
	}

	if s.Client == nil {
		if err := s.createClient(); err != nil {
			return err
		}
	}

	var group run.Group
	for _, h := range handler {
		sub := s.Client.Subscription(h.Subscription())
		group.Add(s.listen(sub, h), func(err error) {
			os.Exit(1)
		})
	}

	return group.Run()
}

func (s *GooglePubSubProvider) listen(subscription google.Subscription, handler SubscriberHandler) func() error {
	return func() error {
		PubSubLogger.Info("google - start subscribe", "name", handler.Name(), "subscription id", subscription.ID())
		err := subscription.Receive(context.Background(), func(ctx context.Context, msg *pubsub.Message) {
			var subCtx = subscriberContext{
				cfg:     s.Config,
				Context: ctx,
			}

			if tranceID, exist := msg.Attributes["trace_id"]; exist && s.Tracer != nil {
				spanId := msg.Attributes["span_id"]

				spanCtx, span := extractTraceSubscriber(context.Background(), s.Tracer, tranceID, spanId, handler.Name())
				subCtx.Context = spanCtx
				subCtx.SetSpan(span)
			}

			// Process the received message
			if err := handler.Consume(&subCtx, msg); err != nil {
				PubSubLogger.Error("Failed consumer message", "topic", handler.Subscription(), "message", string(msg.Data))
			}

			// Acknowledge the message
			if handler.AutoAck() {
				msg.Ack()
			}
		})
		if err == nil {
			return nil
		}

		errMessage := err.Error()
		if strings.Contains(errMessage, "code = Canceled") {
			PubSubLogger.Info("google - stop listen", "name", handler.Name(), "subscription-id", subscription.ID())
			return nil
		}

		PubSubLogger.Error("google - stop listen", "name", handler.Name(), "subscription-id", subscription.ID(), "message", err)
		return err
	}
}

func (s *GooglePubSubProvider) StopListen() error {
	if s.Client != nil {
		return s.Client.Close()
	}

	return nil
}

func (s *GooglePubSubProvider) Publish(ctx context.Context, topic string, message []byte) error {
	if err := s.validate(); err != nil {
		return err
	}

	if s.Client == nil {
		if err := s.createClient(); err != nil {
			return err
		}
	}

	// create message
	msg := pubsub.Message{
		Attributes: make(map[string]string),
	}
	msg.Data = message

	// set trace
	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.HasTraceID() {
		msg.Attributes["trace_id"] = spanCtx.TraceID().String()
	}

	if spanCtx.HasSpanID() {
		msg.Attributes["span_id"] = spanCtx.SpanID().String()
	}

	serverId, err := s.Client.Topic(topic).Publish(ctx, &msg).Get(context.Background())
	if err != nil {
		return err
	}

	PubSubLogger.Info("success publish message to server", "server_id", serverId)
	return nil
}

func extractTraceSubscriber(ctx context.Context, tracer trace.Tracer, traceId string, spanId string, subscriberName string) (rCtx context.Context, span trace.Span) {
	spanName := fmt.Sprintf("subscriber - %s", subscriberName)
	if traceId == "" {
		return tracer.Start(ctx, spanName)
	}

	var spanContextConfig trace.SpanContextConfig
	spanContextConfig.TraceID, _ = trace.TraceIDFromHex(traceId)

	if spanId != "" {
		spanContextConfig.SpanID, _ = trace.SpanIDFromHex(spanId)
	}

	spanContextConfig.TraceFlags = 01
	spanContextConfig.Remote = true

	spanContext := trace.NewSpanContext(spanContextConfig)
	traceCtx := trace.ContextWithSpanContext(ctx, spanContext)
	return tracer.Start(traceCtx, spanName)
}
