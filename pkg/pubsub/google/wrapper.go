package google

import (
	"context"

	"cloud.google.com/go/pubsub"
)

// PubSubClient defines the methods required for Pub/Sub client interaction.
type PubSubClient interface {
	CreateSubscription(ctx context.Context, id string, cfg pubsub.SubscriptionConfig) (Subscription, error)
	Subscription(id string) Subscription
	Topic(id string) Topic
	Close() error
}

// Subscription defines the methods required for subscriptions.
type Subscription interface {
	ID() string
	Receive(ctx context.Context, f func(ctx context.Context, msg *pubsub.Message)) error
}

// Topic defines the methods required for publishing messages.
type Topic interface {
	Publish(ctx context.Context, msg *pubsub.Message) PublishResult
	GetInstance() *pubsub.Topic
}

// PublishResult abstracts the result of publishing a message.
type PublishResult interface {
	Get(ctx context.Context) (string, error)
}

type GooglePubSubClient struct {
	Client *pubsub.Client
}

func (g *GooglePubSubClient) Subscription(id string) Subscription {
	return &GoogleSubscription{subscription: g.Client.Subscription(id)}
}

func (g *GooglePubSubClient) CreateSubscription(ctx context.Context, id string, cfg pubsub.SubscriptionConfig) (Subscription, error) {
	sub, err := g.Client.CreateSubscription(context.Background(), id, cfg)
	return &GoogleSubscription{subscription: sub}, err
}

func (g *GooglePubSubClient) Topic(id string) Topic {
	return &GoogleTopic{topic: g.Client.Topic(id)}
}

func (g *GooglePubSubClient) Close() error {
	return g.Client.Close()
}

type GoogleSubscription struct {
	subscription *pubsub.Subscription
}

func (g *GoogleSubscription) ID() string {
	return g.subscription.ID()
}

func (g *GoogleSubscription) Receive(ctx context.Context, f func(ctx context.Context, msg *pubsub.Message)) error {
	return g.subscription.Receive(ctx, f)
}

type GoogleTopic struct {
	topic *pubsub.Topic
}

func (g *GoogleTopic) Publish(ctx context.Context, msg *pubsub.Message) PublishResult {
	return &GooglePublishResult{result: g.topic.Publish(ctx, msg)}
}

func (g *GoogleTopic) GetInstance() *pubsub.Topic {
	return g.topic
}

type GooglePublishResult struct {
	result *pubsub.PublishResult
}

func (g *GooglePublishResult) Get(ctx context.Context) (string, error) {
	return g.result.Get(ctx)
}
