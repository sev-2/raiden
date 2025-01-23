package raiden

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"cloud.google.com/go/pubsub"
	"github.com/oklog/run"
	"github.com/sev-2/raiden/pkg/logger"
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
	Name() string
	Provider() PubSubProviderType
	Topic() string
	AutoAck() bool
	Consume(ctx SubscriberContext, message any) error
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

func (s *SubscriberBase) Topic() string {
	return ""
}

func (s *SubscriberBase) Consume(ctx SubscriberContext, message any) error {
	return fmt.Errorf("subscriber %s in not implemented", s.Name())
}

// ----- Subscription Server -----
type PubSub interface {
	Register(handler SubscriberHandler)
	Publish(ctx context.Context, provider PubSubProviderType, topic string, message []byte) error
	Listen()
}

func NewPubsub(config *Config, tracer trace.Tracer) PubSub {
	return &pubSub{
		config: config,
		provider: pubSubProvider{
			google: &googlePubSubProvider{
				config: config,
				tracer: tracer,
			},
		},
	}

}

type pubSub struct {
	config   *Config
	handlers []SubscriberHandler
	provider pubSubProvider
}

type pubSubProvider struct {
	// all provider
	google PubSubProvider
}

// Register implements Subscriber.
func (s *pubSub) Register(handler SubscriberHandler) {
	s.handlers = append(s.handlers, handler)
}

// StartListen implements Subscriber.
func (s *pubSub) Listen() {
	var g run.Group

	var googleHandlers []SubscriberHandler
	for _, h := range s.handlers {
		if h.Provider() == PubSubProviderGoogle {
			googleHandlers = append(googleHandlers, h)
		}
	}

	g.Add(func() error {
		return s.provider.google.StartListen(googleHandlers)
	}, func(err error) {
		if err := s.provider.google.StopListen(); err != nil {
			PubSubLogger.Error("failed stop listener", "message", err)
		}
	})

	if err := g.Run(); err != nil {
		PubSubLogger.Error("stop subscribe", "message", err)
	}
}

func (s *pubSub) Publish(ctx context.Context, provider PubSubProviderType, topic string, message []byte) error {
	switch provider {
	case PubSubProviderGoogle:
		return s.provider.google.Publish(ctx, topic, message)
	}
	return errors.New("unsupported pubsub provider")
}

// ----- Pub Sub Provider -----
type PubSubProvider interface {
	Publish(ctx context.Context, topic string, message []byte) error
	StartListen(handler []SubscriberHandler) error
	StopListen() error
}

type googlePubSubProvider struct {
	config *Config
	client *pubsub.Client
	tracer trace.Tracer
}

func (s *googlePubSubProvider) validate() error {
	if s.config.GoogleProjectId == "" {
		return errors.New("env GOOGLE_PROJECT_ID is required")
	}

	if s.config.GoogleSaPath == "" {
		return errors.New("env GOOGLE_SA_PATH is required")
	}

	return nil
}

func (s *googlePubSubProvider) createClient() error {
	client, err := pubsub.NewClient(context.Background(), s.config.GoogleProjectId, option.WithCredentialsFile(s.config.GoogleSaPath))
	if err != nil {
		PubSubLogger.Error("failed create google pubsub client")
		return err
	}

	s.client = client

	return nil
}

func (s *googlePubSubProvider) StartListen(handler []SubscriberHandler) error {
	if err := s.validate(); err != nil {
		return err
	}

	if s.client == nil {
		if err := s.createClient(); err != nil {
			return err
		}
	}

	var group run.Group
	for _, h := range handler {
		sub := s.client.Subscription(h.Topic())
		group.Add(s.listen(sub, h), func(err error) {})
	}

	return group.Run()
}

func (s *googlePubSubProvider) listen(subscription *pubsub.Subscription, handler SubscriberHandler) func() error {
	return func() error {
		PubSubLogger.Info("google - start subscribe", "name", handler.Name(), "subscription id", subscription.ID())
		err := subscription.Receive(context.Background(), func(ctx context.Context, msg *pubsub.Message) {
			var subCtx = subscriberContext{
				cfg:     s.config,
				Context: ctx,
			}

			if tranceID, exist := msg.Attributes["trace_id"]; exist && s.tracer != nil {
				spanId := msg.Attributes["span_id"]

				spanCtx, span := extractTraceSubscriber(context.Background(), s.tracer, tranceID, spanId, handler.Name())
				subCtx.Context = spanCtx
				subCtx.SetSpan(span)
			}

			// Process the received message
			if err := handler.Consume(&subCtx, msg); err != nil {
				PubSubLogger.Error("Failed consumer message", "topic", handler.Topic(), "message", string(msg.Data))
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

func (s *googlePubSubProvider) StopListen() error {
	if s.client != nil {
		return s.client.Close()
	}

	return nil
}

func (s *googlePubSubProvider) Publish(ctx context.Context, topic string, message []byte) error {
	if err := s.validate(); err != nil {
		return err
	}

	if s.client == nil {
		if err := s.createClient(); err != nil {
			return err
		}
	}

	// create message
	msg := pubsub.Message{}
	msg.Data = message

	// set trace
	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.HasTraceID() {
		msg.Attributes["trace_id"] = spanCtx.TraceID().String()
	}

	if spanCtx.HasSpanID() {
		msg.Attributes["span_id"] = spanCtx.SpanID().String()
	}

	serverId, err := s.client.Topic(topic).Publish(ctx, &msg).Get(context.Background())
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
