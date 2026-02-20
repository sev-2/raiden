# Pub/Sub

Raiden provides a provider-agnostic Pub/Sub layer for subscribing to and publishing messages. The framework ships with a Google Cloud Pub/Sub provider and is designed to support additional providers (e.g., AWS SNS/SQS, Kafka) through the `PubSubProvider` interface.

## Architecture

```
┌────────────────────────────────────────────────────────────────────┐
│                          PubSubManager                             │
│                                                                    │
│  providers: map[PubSubProviderType]PubSubProvider                  │
│   ├── "google" ──► GooglePubSubProvider (adapter)                  │
│   │                   └── pkg/pubsub/google.Provider               │
│   └── "custom" ──► YourCustomProvider                              │
│                                                                    │
│  handlers: []SubscriberHandler                                     │
│   ├── Pull subscribers  ──► Listen() ──► provider.StartListen()    │
│   └── Push subscribers  ──► Serve()  ──► provider.Serve() ──► HTTP │
├────────────────────────────────────────────────────────────────────┤
│  Publish(ctx, provider, topic, message)                            │
│   └── providers[provider].Publish(ctx, topic, message)             │
└────────────────────────────────────────────────────────────────────┘
```

### Request Flow

**Pull Subscription:**
```
Cloud Provider ──push──► Provider.StartListen()
                              │
                              ▼
                     Consume(ctx, SubscriberMessage)
                              │
                              ▼
                      Your business logic
```

**Push Subscription:**
```
Cloud Provider ──HTTP POST──► /pubsub-endpoint/<subscription>
                                     │
                                     ▼
                            Provider.Serve() handler
                                     │
                              parse & validate
                                     │
                                     ▼
                           Consume(ctx, SubscriberMessage)
                                     │
                                     ▼
                              Your business logic
```

**Publishing:**
```
Controller / Job ──► ctx.Publish(provider, topic, data)
                          │
                          ▼
                  PubSubManager.Publish()
                          │
                          ▼
                  provider.Publish(ctx, topic, data)
                          │
                          ▼
                     Cloud Provider
```

## Quick Start

### 1. Create a Pull Subscriber

```go
package subscribers

import (
    "encoding/json"

    "github.com/sev-2/raiden"
)

type OrderCreatedSubscriber struct {
    raiden.SubscriberBase
}

func (s *OrderCreatedSubscriber) Name() string {
    return "OrderCreated"
}

func (s *OrderCreatedSubscriber) Provider() raiden.PubSubProviderType {
    return raiden.PubSubProviderGoogle
}

func (s *OrderCreatedSubscriber) Subscription() string {
    return "order.created-sub"
}

func (s *OrderCreatedSubscriber) Consume(ctx raiden.SubscriberContext, message raiden.SubscriberMessage) error {
    var order Order
    if err := json.Unmarshal(message.Data, &order); err != nil {
        return err
    }

    raiden.Info("order received", "id", order.ID)
    return nil
}
```

The subscriber embeds `raiden.SubscriberBase` which provides sensible defaults:
- `AutoAck()` returns `true`
- `SubscriptionType()` returns `SubscriptionTypePull`
- `PushEndpoint()` returns `""`

Override only the methods you need.

### 2. Create a Push Subscriber

Push subscribers receive messages via HTTP endpoints that Raiden registers automatically.

```go
type PaymentWebhookSubscriber struct {
    raiden.SubscriberBase
}

func (s *PaymentWebhookSubscriber) Name() string {
    return "PaymentWebhook"
}

func (s *PaymentWebhookSubscriber) Provider() raiden.PubSubProviderType {
    return raiden.PubSubProviderGoogle
}

func (s *PaymentWebhookSubscriber) Topic() string {
    return "payment.completed"
}

func (s *PaymentWebhookSubscriber) Subscription() string {
    return "payment.completed-push"
}

func (s *PaymentWebhookSubscriber) SubscriptionType() raiden.SubscriptionType {
    return raiden.SubscriptionTypePush
}

func (s *PaymentWebhookSubscriber) PushEndpoint() string {
    return "/payment-webhook"
}

func (s *PaymentWebhookSubscriber) Consume(ctx raiden.SubscriberContext, message raiden.SubscriberMessage) error {
    raiden.Info("payment received", "data", string(message.Data))
    return nil
}
```

Push endpoints are registered at `/<SubscriptionPrefixEndpoint>/<push-endpoint>` (e.g., `/pubsub-endpoint/payment-webhook`).

### 3. Register Subscribers

Register subscribers in your module:

```go
func (m *AppModule) Subscribers() []raiden.SubscriberHandler {
    return []raiden.SubscriberHandler{
        &subscribers.OrderCreatedSubscriber{},
        &subscribers.PaymentWebhookSubscriber{},
    }
}
```

### 4. Publish Messages

Publish from any controller or handler via the context:

```go
func (c *OrderController) Post(ctx raiden.Context) error {
    order := Order{ID: "123", Item: "Widget"}
    data, _ := json.Marshal(order)

    err := ctx.Publish(raiden.PubSubProviderGoogle, "order.created", data)
    if err != nil {
        return ctx.SendErrorWithCode(500, err)
    }

    return ctx.SendJson(order)
}
```

## Core Types

### SubscriberMessage

`SubscriberMessage` is a provider-agnostic message envelope:

```go
type SubscriberMessage struct {
    Data       []byte            // Message payload
    Attributes map[string]string // Message metadata / headers
    Raw        any               // Original provider message (e.g., *pubsub.Message)
}
```

- **`Data`**: The message body as bytes. Use `json.Unmarshal` to decode.
- **`Attributes`**: Key-value metadata. For pull subscriptions, these come from the provider. For Google push subscriptions, includes `message_id` and `publish_time`.
- **`Raw`**: The original provider-specific message. Use this when you need provider-specific features (e.g., `msg.Ack()`, `msg.Nack()` for Google Pub/Sub). Access via type assertion: `raw := message.Raw.(*pubsub.Message)`.

### SubscriberHandler

The interface every subscriber must implement:

```go
type SubscriberHandler interface {
    AutoAck() bool
    Name() string
    Consume(ctx SubscriberContext, message SubscriberMessage) error
    Provider() PubSubProviderType
    PushEndpoint() string
    Subscription() string
    SubscriptionType() SubscriptionType
    Topic() string
}
```

Embed `raiden.SubscriberBase` and override only what you need.

### SubscriberContext

Available inside `Consume`:

```go
type SubscriberContext interface {
    Config() *Config
    Span() trace.Span
    SetSpan(span trace.Span)
    HttpRequest(method, url string, body []byte, headers map[string]string, timeout time.Duration, response any) error
}
```

- `Config()` — access application configuration.
- `Span()` / `SetSpan()` — OpenTelemetry tracing support.
- `HttpRequest()` — make outbound HTTP calls within the subscriber.

### PubSubProvider

The interface for implementing new providers:

```go
type PubSubProvider interface {
    Publish(ctx context.Context, topic string, message []byte) error
    CreateSubscription(SubscriberHandler) error
    Serve(config *Config, handler SubscriberHandler) (fasthttp.RequestHandler, error)
    StartListen(handler []SubscriberHandler) error
    StopListen() error
}
```

## Adding a Custom Provider

To add a new Pub/Sub provider (e.g., AWS SQS):

### 1. Define a Provider Type

```go
const PubSubProviderAWS raiden.PubSubProviderType = "aws"
```

### 2. Implement PubSubProvider

```go
type AWSSQSProvider struct {
    // your AWS client, config, etc.
}

func (p *AWSSQSProvider) Publish(ctx context.Context, topic string, message []byte) error {
    // Publish to SNS/SQS
}

func (p *AWSSQSProvider) CreateSubscription(handler raiden.SubscriberHandler) error {
    // Create/verify SQS queue subscription
}

func (p *AWSSQSProvider) Serve(config *raiden.Config, handler raiden.SubscriberHandler) (fasthttp.RequestHandler, error) {
    // Return HTTP handler for push-style webhooks
}

func (p *AWSSQSProvider) StartListen(handlers []raiden.SubscriberHandler) error {
    // Start polling SQS queues
}

func (p *AWSSQSProvider) StopListen() error {
    // Stop polling
}
```

### 3. Register the Provider

```go
pubsub := raiden.NewPubsub(config, tracer)
manager := pubsub.(*raiden.PubSubManager)
manager.SetProvider(PubSubProviderAWS, &AWSSQSProvider{})
```

### 4. Use in Subscribers

```go
func (s *MySubscriber) Provider() raiden.PubSubProviderType {
    return PubSubProviderAWS
}
```

## Configuration

The following environment variables are used for Google Pub/Sub:

| Variable | Description |
|---|---|
| `GOOGLE_PROJECT_ID` | Google Cloud project ID |
| `GOOGLE_SA_PATH` | Path to Google service account JSON key file |
| `SERVER_DNS` | Server DNS used for push subscription endpoint registration |

These are loaded via `raiden.Config` and passed to providers automatically.

## Project Structure

```
raiden/
├── pubsub.go                      # Core interfaces, PubSubManager, Google adapter
├── pubsub_test.go                 # Tests
├── pkg/
│   ├── pubsub/
│   │   └── google/
│   │       ├── provider.go        # Self-contained Google Cloud Pub/Sub implementation
│   │       └── wrapper.go         # Google SDK wrapper interfaces (for testing)
│   └── mock/
│       ├── provider.go            # MockProvider for testing
│       └── pubsub.go              # Mock Pub/Sub client & handler
```

The Google provider in `pkg/pubsub/google/` is self-contained with no imports back to the root `raiden` package, avoiding circular dependencies. The root `GooglePubSubProvider` acts as a thin adapter that translates between raiden types and the Google provider's internal types.

## Migration Guide

### From `Consume(ctx, message any)` to `Consume(ctx, message SubscriberMessage)`

The `Consume` method signature changed from accepting `any` to `SubscriberMessage`.

**Before (old):**
```go
func (s *MySub) Consume(ctx raiden.SubscriberContext, message any) error {
    msg := message.(*pubsub.Message)
    data := msg.Data
    // ...
}
```

**After (new):**
```go
func (s *MySub) Consume(ctx raiden.SubscriberContext, message raiden.SubscriberMessage) error {
    data := message.Data           // []byte — no type assertion needed
    attrs := message.Attributes    // map[string]string

    // If you still need the original *pubsub.Message:
    raw := message.Raw.(*pubsub.Message)
    // ...
}
```

### Quick Fix with WrapLegacySubscriber

If you have many subscribers and need a quick migration path, use `WrapLegacySubscriber` to wrap old-style subscribers without changing their code:

```go
// Old subscriber — no changes needed
type MyOldSub struct { raiden.SubscriberBase }

func (s *MyOldSub) Consume(ctx raiden.SubscriberContext, message any) error {
    msg := message.(*pubsub.Message)
    // ... existing logic
}

// Register with wrapper
func (m *AppModule) Subscribers() []raiden.SubscriberHandler {
    return []raiden.SubscriberHandler{
        raiden.WrapLegacySubscriber(&MyOldSub{}),
    }
}
```

The wrapper forwards `message.Raw` (the original provider message) to the legacy `Consume` method, so existing type assertions continue to work.

> **Note:** `WrapLegacySubscriber` is deprecated. Migrate to `SubscriberMessage` when possible for type safety and provider independence.

### From `PushSubscriptionMessage` / `PushSubscriptionData`

These types are deprecated but still available as aliases for backward compatibility. Migrate to `SubscriberMessage` instead:

**Before:**
```go
func (s *MySub) Consume(ctx raiden.SubscriberContext, message any) error {
    data := message.(raiden.PushSubscriptionMessage)
    // ...
}
```

**After:**
```go
func (s *MySub) Consume(ctx raiden.SubscriberContext, message raiden.SubscriberMessage) error {
    // Data is already decoded into message.Data
    // Attributes contain message_id and publish_time
    // ...
}
```

## Testing

### Mock Provider

Use `mock.MockProvider` in tests:

```go
import "github.com/sev-2/raiden/pkg/mock"

provider := &mock.MockProvider{
    PublishFn: func(ctx context.Context, topic string, message []byte) error {
        assert.Equal(t, "my-topic", topic)
        return nil
    },
}

mgr := raiden.NewPubsub(config, nil).(*raiden.PubSubManager)
mgr.SetProvider(raiden.PubSubProviderGoogle, provider)
```

### Mock Subscriber Handler

Use `mock.MockSubscriberHandler` to test subscriber registration and consumption:

```go
handler := &mock.MockSubscriberHandler{
    NameVal:     "test-handler",
    ProviderVal: raiden.PubSubProviderGoogle,
    ConsumeFunc: func(ctx raiden.SubscriberContext, msg raiden.SubscriberMessage) error {
        assert.Equal(t, []byte("hello"), msg.Data)
        return nil
    },
}
```
