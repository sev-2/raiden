package google

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

var ProviderLogger = logger.HcLog().Named("raiden.pubsub.google")

// ProviderConfig holds Google-specific configuration.
type ProviderConfig struct {
	GoogleProjectId string
	GoogleSaPath    string
}

// ListenHandler describes a pull-subscription handler for StartListen.
type ListenHandler struct {
	Name         string
	Subscription string
	AutoAck      bool
	ConsumeFn    func(ctx context.Context, span trace.Span, msg any) error
}

// Provider implements Google Cloud Pub/Sub operations.
type Provider struct {
	Config *ProviderConfig
	Client PubSubClient
	Tracer trace.Tracer
}

func (s *Provider) Validate() error {
	if s.Config.GoogleProjectId == "" {
		return errors.New("env GOOGLE_PROJECT_ID is required")
	}

	if s.Config.GoogleSaPath == "" {
		return errors.New("env GOOGLE_SA_PATH is required")
	}

	return nil
}

func (s *Provider) CreateClient() error {
	client, err := pubsub.NewClient(context.Background(), s.Config.GoogleProjectId, option.WithCredentialsFile(s.Config.GoogleSaPath))
	if err != nil {
		ProviderLogger.Error("failed create google pubsub client")
		return err
	}

	s.Client = &GooglePubSubClient{Client: client}
	return nil
}

func (s *Provider) CreateSubscription(name, topic, pushEndpoint, serverDns, subscription, prefixEndpoint string) error {
	if topic == "" {
		return fmt.Errorf("topic in subscription %s is required", name)
	}

	if pushEndpoint == "" {
		return fmt.Errorf("push endpoint in subscription %s is required", name)
	}

	if serverDns == "" {
		return errors.New("the SERVER_DNS configuration is required when using the 'push' subscription type")
	}

	if s.Client == nil {
		if err := s.CreateClient(); err != nil {
			return err
		}
	}

	t := s.Client.Topic(topic)

	var endpoint string
	if strings.HasPrefix(pushEndpoint, "/") {
		endpoint = fmt.Sprintf("%s/%s%s", serverDns, prefixEndpoint, pushEndpoint)
	} else {
		endpoint = fmt.Sprintf("%s/%s/%s", serverDns, prefixEndpoint, pushEndpoint)
	}

	_, err := s.Client.CreateSubscription(context.Background(), subscription, pubsub.SubscriptionConfig{
		Topic: t.GetInstance(),
		PushConfig: pubsub.PushConfig{
			Endpoint: endpoint,
		},
	})

	if err != nil {
		if strings.Contains(err.Error(), "Resource already exists in the project") {
			ProviderLogger.Info("subscription already exists in the project", "topic", topic, "subscription-id", subscription)
			return nil
		}
		return err
	}

	return nil
}

func (s *Provider) StartListen(handlers []ListenHandler) error {
	if err := s.Validate(); err != nil {
		return err
	}

	if s.Client == nil {
		if err := s.CreateClient(); err != nil {
			return err
		}
	}

	var group run.Group
	for _, h := range handlers {
		sub := s.Client.Subscription(h.Subscription)
		group.Add(s.listen(sub, h), func(err error) {
			if err != nil {
				ProviderLogger.Error("s.listen()", "message", err.Error())
			}
		})
	}

	return group.Run()
}

func (s *Provider) listen(subscription Subscription, handler ListenHandler) func() error {
	return func() error {
		ProviderLogger.Info("start subscribe", "name", handler.Name, "subscription id", subscription.ID())
		err := subscription.Receive(context.Background(), func(ctx context.Context, msg *pubsub.Message) {
			var span trace.Span

			if traceID, exist := msg.Attributes["trace_id"]; exist && s.Tracer != nil {
				spanId := msg.Attributes["span_id"]
				ctx, span = ExtractTraceSubscriber(context.Background(), s.Tracer, traceID, spanId, handler.Name)
			}

			if err := handler.ConsumeFn(ctx, span, msg); err != nil {
				ProviderLogger.Error("failed consume message", "subscription", handler.Subscription, "message", string(msg.Data))
			}

			if handler.AutoAck {
				msg.Ack()
			}
		})
		if err == nil {
			return nil
		}

		errMessage := err.Error()
		if strings.Contains(errMessage, "code = Canceled") {
			ProviderLogger.Info("stop listen", "name", handler.Name, "subscription-id", subscription.ID())
			return nil
		}

		ProviderLogger.Error("stop listen", "name", handler.Name, "subscription-id", subscription.ID(), "message", err)
		return err
	}
}

func (s *Provider) StopListen() error {
	if s.Client != nil {
		return s.Client.Close()
	}

	return nil
}

func (s *Provider) Publish(ctx context.Context, topic string, message []byte) error {
	if err := s.Validate(); err != nil {
		return err
	}

	if s.Client == nil {
		if err := s.CreateClient(); err != nil {
			return err
		}
	}

	msg := pubsub.Message{
		Attributes: make(map[string]string),
	}
	msg.Data = message

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

	ProviderLogger.Info("success publish message to server", "server_id", serverId)
	return nil
}

// ExtractTraceSubscriber reconstructs a span context from message attributes.
func ExtractTraceSubscriber(ctx context.Context, tracer trace.Tracer, traceId string, spanId string, subscriberName string) (rCtx context.Context, span trace.Span) {
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
