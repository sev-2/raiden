# Pub/Sub

## Purpose

The Pub/Sub capability provides a provider-agnostic messaging layer for publishing and subscribing to messages. It supports multiple subscription delivery modes (pull and push) and multiple provider backends through a pluggable provider registry.

**Key files:** `pubsub.go`, `pkg/pubsub/google/provider.go`, `pkg/pubsub/google/wrapper.go`, `pkg/pubsub/supabase/provider.go`, `pkg/mock/provider.go`, `pkg/mock/pubsub.go`

## Requirements

### Requirement: Provider-Agnostic Message Envelope

The system SHALL wrap all incoming messages in a `SubscriberMessage` struct containing `Data` ([]byte), `Attributes` (map[string]string), and `Raw` (the original provider-specific message). Subscribers MUST NOT need to import provider-specific packages to access message data.

#### Scenario: Pull subscription delivers SubscriberMessage
- **WHEN** a pull subscriber receives a message from any provider
- **THEN** the `Consume` method receives a `SubscriberMessage` with `Data` populated from the provider message payload, `Attributes` from the provider metadata, and `Raw` holding the original provider object

#### Scenario: Push subscription delivers SubscriberMessage
- **WHEN** a push subscriber receives an HTTP POST from the cloud provider
- **THEN** the `Consume` method receives a `SubscriberMessage` with `Data` decoded from the push envelope body and `Attributes` containing envelope metadata (e.g., `message_id`, `publish_time`)

#### Scenario: Access to raw provider message
- **WHEN** a subscriber needs provider-specific features (e.g., manual Ack/Nack)
- **THEN** the subscriber MAY type-assert `message.Raw` to the provider's native message type

### Requirement: Multi-Provider Registry

The system SHALL maintain a map-based registry of providers (`map[PubSubProviderType]PubSubProvider`). Multiple providers MAY be active simultaneously. Each subscriber handler declares which provider it uses via `Provider()`.

#### Scenario: Default provider registration
- **WHEN** `NewPubsub(config, tracer)` is called
- **THEN** the Google Cloud Pub/Sub provider SHALL be registered under `PubSubProviderGoogle` ("google") and the Supabase Realtime provider SHALL be registered under `PubSubProviderSupabase` ("supabase")

#### Scenario: Custom provider registration
- **WHEN** `SetProvider(providerType, provider)` is called with a new provider type
- **THEN** the provider SHALL be stored in the registry and available for subscribers that declare that provider type

#### Scenario: Unknown provider lookup
- **WHEN** a subscriber or publish call references a provider type not in the registry
- **THEN** the system SHALL return an error "unsupported pubsub provider: {type}"

### Requirement: Subscriber Handler Interface

The system SHALL define a `SubscriberHandler` interface that all subscribers MUST implement. The interface includes: `AutoAck()`, `Name()`, `Consume(ctx SubscriberContext, message SubscriberMessage) error`, `ChannelType()`, `EventFilter()`, `Provider()`, `PushEndpoint()`, `Schema()`, `Subscription()`, `SubscriptionType()`, `Table()`, and `Topic()`.

#### Scenario: Embed SubscriberBase for defaults
- **WHEN** a subscriber struct embeds `SubscriberBase`
- **THEN** it inherits default values: `AutoAck() = true`, `SubscriptionType() = SubscriptionTypePull`, `Provider() = PubSubProviderUnknown`, `ChannelType() = ""`, `Schema() = "public"`, `Table() = ""`, `EventFilter() = "*"`, and empty strings for `Name`, `Subscription`, `PushEndpoint`, and `Topic`

#### Scenario: Unimplemented Consume returns error
- **WHEN** `SubscriberBase.Consume` is called without being overridden
- **THEN** it SHALL return an error "subscriber {name} is not implemented"

### Requirement: Pull Subscription Listening

The system SHALL support pull-based subscriptions where the provider actively polls or receives streamed messages from the cloud service. Pull listeners are started via `PubSubManager.Listen()`.

#### Scenario: Start pull listeners grouped by provider
- **WHEN** `Listen()` is called with registered pull-type handlers
- **THEN** the system SHALL group handlers by their `Provider()` type and call each provider's `StartListen(handlers)` concurrently using `oklog/run.Group`

#### Scenario: No pull handlers registered
- **WHEN** `Listen()` is called and no handlers have `SubscriptionType() == SubscriptionTypePull`
- **THEN** `Listen()` SHALL return immediately without starting any listeners

#### Scenario: Provider not found during listen
- **WHEN** a pull handler references a provider type not in the registry
- **THEN** the system SHALL log an error and skip that provider group (not crash)

#### Scenario: Listener stop on error
- **WHEN** a provider's `StartListen` returns an error or the run group is interrupted
- **THEN** the system SHALL call `StopListen()` on that provider and log any stop errors

### Requirement: Push Subscription Serving

The system SHALL support push-based subscriptions where the cloud provider sends HTTP POST requests to registered endpoints. Push endpoints are served via `PubSubManager.Serve(handler)`.

#### Scenario: Serve returns HTTP handler for push subscriber
- **WHEN** `Serve(handler)` is called with a handler whose `SubscriptionType() == SubscriptionTypePush`
- **THEN** it SHALL delegate to the handler's provider `Serve(config, handler)` and return a `fasthttp.RequestHandler`

#### Scenario: Serve rejects non-push subscriber
- **WHEN** `Serve(handler)` is called with a handler whose `SubscriptionType() != SubscriptionTypePush`
- **THEN** it SHALL return an error "subscription {name} is not push subscription"

#### Scenario: Push endpoint registration path
- **WHEN** push subscription handlers are registered in the router
- **THEN** the HTTP endpoint SHALL be mounted at `/{SubscriptionPrefixEndpoint}/{pushEndpoint}` (e.g., `/pubsub-endpoint/payment-webhook`)

### Requirement: Message Publishing

The system SHALL allow publishing messages to any registered provider via `PubSubManager.Publish(ctx, provider, topic, message)`.

#### Scenario: Publish to a valid provider
- **WHEN** `Publish(ctx, provider, topic, data)` is called with a registered provider
- **THEN** the system SHALL delegate to `provider.Publish(ctx, topic, data)`

#### Scenario: Publish to unknown provider
- **WHEN** `Publish` is called with a provider type not in the registry
- **THEN** it SHALL return an error "unsupported pubsub provider: {type}"

#### Scenario: Publish from controller context
- **WHEN** a controller calls `ctx.Publish(provider, topic, data)`
- **THEN** it SHALL delegate to the `PubSub.Publish` method injected via middleware
- **WHEN** the PubSub instance is not initialized in the context
- **THEN** it SHALL return an error "unable to publish because pubsub not initialize"

### Requirement: Subscriber Context

The system SHALL provide a `SubscriberContext` interface to subscribers within `Consume`, offering access to application config, OpenTelemetry tracing spans, and outbound HTTP request capability.

#### Scenario: Access config from subscriber
- **WHEN** a subscriber calls `ctx.Config()`
- **THEN** it SHALL return the application `*Config`

#### Scenario: Tracing span propagation
- **WHEN** a pull subscriber receives a message with trace context
- **THEN** the provider adapter SHALL extract the span and set it on the `SubscriberContext` via `SetSpan()`
- **WHEN** `ctx.Span()` is called
- **THEN** it SHALL return the propagated `trace.Span` (or nil if no tracing)

#### Scenario: Outbound HTTP from subscriber
- **WHEN** a subscriber calls `ctx.HttpRequest(method, url, body, headers, timeout, &response)`
- **THEN** it SHALL send the HTTP request and unmarshal the response JSON into the provided pointer
- **WHEN** the response parameter is not a pointer
- **THEN** it SHALL return an error "response payload must be pointer"

### Requirement: PubSubProvider Interface

The system SHALL define a `PubSubProvider` interface that all provider implementations MUST satisfy: `Publish`, `CreateSubscription`, `Serve`, `StartListen`, and `StopListen`.

#### Scenario: Provider owns push handling
- **WHEN** `Serve(config, handler)` is called on a provider
- **THEN** the provider SHALL handle envelope parsing, subscription validation, and HTTP response formatting specific to its protocol

#### Scenario: Provider owns pull handling
- **WHEN** `StartListen(handlers)` is called on a provider
- **THEN** the provider SHALL manage connection lifecycle, message receipt, auto-ack behavior, and error handling

### Requirement: Google Cloud Pub/Sub Provider

The system SHALL ship with a built-in Google Cloud Pub/Sub provider implemented in `pkg/pubsub/google/` with a root-level adapter (`GooglePubSubProvider`). The Google provider package MUST NOT import the root `raiden` package.

#### Scenario: Google pull subscription
- **WHEN** a Google pull subscriber is started
- **THEN** the provider SHALL create a subscription client, receive messages via the Google Pub/Sub SDK, and auto-ack messages when `AutoAck()` is true

#### Scenario: Google push subscription validation
- **WHEN** a push HTTP POST is received by the Google provider's `Serve` handler
- **THEN** it SHALL parse the Google-specific JSON envelope (`message.data`, `message.message_id`, `message.publish_time`, `subscription`)
- **WHEN** the envelope `subscription` field does not match `projects/{projectId}/subscriptions/{subscriptionId}`
- **THEN** it SHALL respond with HTTP 422 and `{"message":"subscription validation failed: received unexpected subscription name"}`

#### Scenario: Google push subscription invalid JSON
- **WHEN** the push HTTP POST body is not valid JSON
- **THEN** the provider SHALL respond with HTTP 200 and `{"message":"invalid json data"}`

#### Scenario: Google push consume error
- **WHEN** `handler.Consume` returns an error during push handling
- **THEN** the provider SHALL respond with HTTP 500 and include the error message in the JSON response

#### Scenario: No circular dependency
- **WHEN** `pkg/pubsub/google/` is compiled
- **THEN** it SHALL have zero imports of the `github.com/sev-2/raiden` root package

### Requirement: Backward Compatibility

The system SHALL provide deprecated type aliases and adapter functions to allow old code using the previous API to compile and function correctly.

#### Scenario: Deprecated PushSubscriptionMessage alias
- **WHEN** existing code references `raiden.PushSubscriptionMessage`
- **THEN** it SHALL compile successfully as an alias to the internal Google push message type

#### Scenario: Deprecated PushSubscriptionData alias
- **WHEN** existing code references `raiden.PushSubscriptionData`
- **THEN** it SHALL compile successfully as an alias to the internal Google push data type

#### Scenario: WrapLegacySubscriber adapter
- **WHEN** an old subscriber implements `Consume(ctx SubscriberContext, message any)`
- **THEN** `WrapLegacySubscriber(oldSubscriber)` SHALL return a `SubscriberHandler` that forwards `message.Raw` to the legacy `Consume` method
- **WHEN** `message.Raw` is nil (e.g., push subscriptions)
- **THEN** the adapter SHALL forward the entire `SubscriberMessage` as the `any` argument

### Requirement: Server Integration

The system SHALL integrate Pub/Sub into the server lifecycle. The server creates the PubSub manager, registers subscriber handlers from modules, starts pull listeners in a goroutine, and passes the manager to the router for push endpoint registration.

#### Scenario: Server starts pull listeners
- **WHEN** the server starts and subscriber handlers are registered
- **THEN** it SHALL call `runSubscriberServer()` which creates a `PubSub` instance, registers all handlers, and calls `go pubsub.Listen()`

#### Scenario: Router registers push endpoints
- **WHEN** the router builds HTTP handlers
- **THEN** for each push-type subscriber handler, it SHALL call `pubSub.Serve(handler)` and mount the returned handler at the push endpoint path

#### Scenario: Middleware injects PubSub into context
- **WHEN** an HTTP request is processed through middleware
- **THEN** the PubSub instance SHALL be injected into `Ctx` so controllers can call `ctx.Publish()`

### Requirement: Supabase Realtime Provider

The system SHALL ship with a built-in Supabase Realtime provider implemented in `pkg/pubsub/supabase/` with a root-level adapter (`SupabaseRealtimeProvider`). The Supabase provider package MUST NOT import the root `raiden` package.

#### Scenario: Supabase Realtime channel types
- **WHEN** a Supabase Realtime subscriber is registered
- **THEN** it SHALL support three channel types: `RealtimeChannelBroadcast` ("broadcast"), `RealtimeChannelPresence` ("presence"), and `RealtimeChannelPostgresChanges` ("postgres_changes")

#### Scenario: Supabase WebSocket connection
- **WHEN** `StartListen()` is called for Supabase subscribers
- **THEN** the provider SHALL connect via WebSocket to `wss://{host}/realtime/v1/websocket?apikey={key}&vsn=1.0.0`, deriving the URL from `SupabasePublicUrl` or `SupabaseApiUrl` config, and authenticating with `ServiceKey` (preferred) or `AnonKey`

#### Scenario: Supabase channel join
- **WHEN** a WebSocket connection is established
- **THEN** the provider SHALL send a `phx_join` message for each subscriber's channel with appropriate config payload (broadcast settings, presence settings, or postgres_changes filter with schema/table/event)

#### Scenario: Supabase heartbeat
- **WHEN** a WebSocket connection is active
- **THEN** the provider SHALL send heartbeat messages to the `"phoenix"` topic every 30 seconds

#### Scenario: Supabase message dispatch
- **WHEN** a message is received on the WebSocket
- **THEN** the provider SHALL dispatch it to the matching handler's `ConsumeFn` based on the message topic, skipping system events (`phx_reply`, `phx_close`, `heartbeat`)

#### Scenario: Supabase message mapping
- **WHEN** a Supabase Realtime message is delivered to a subscriber
- **THEN** `SubscriberMessage.Data` SHALL contain the JSON payload, `Attributes` SHALL contain `{"channel_type", "event", "topic"}`, and `Raw` SHALL hold the original `json.RawMessage`

#### Scenario: Supabase Publish broadcast
- **WHEN** `Publish(ctx, topic, message)` is called on the Supabase provider
- **THEN** it SHALL send a broadcast event to `realtime:{topic}` over the WebSocket connection

#### Scenario: Supabase Serve returns error
- **WHEN** `Serve()` is called on the Supabase provider
- **THEN** it SHALL return an error because Supabase Realtime uses WebSocket, not HTTP push endpoints

#### Scenario: Supabase reconnection
- **WHEN** the WebSocket connection is lost
- **THEN** the provider SHALL attempt to reconnect with exponential backoff up to a maximum number of retries, and re-join all channels on successful reconnection

#### Scenario: Supabase no circular dependency
- **WHEN** `pkg/pubsub/supabase/` is compiled
- **THEN** it SHALL have zero imports of the `github.com/sev-2/raiden` root package

#### Scenario: Supabase config reuse
- **WHEN** the Supabase Realtime provider is initialized
- **THEN** it SHALL reuse existing Raiden config fields (`SupabasePublicUrl`, `SupabaseApiUrl`, `AnonKey`, `ServiceKey`, `ProjectId`) without requiring new environment variables

### Requirement: Postgres Changes Filtering

The system SHALL support declarative filtering for Postgres Changes subscribers via `Table()`, `Schema()`, and `EventFilter()` methods on `SubscriberHandler`. `EventFilter()` returns `string`; the framework SHALL provide string constants in `pkg/supabase/constants` (`RealtimeEventAll`, `RealtimeEventInsert`, `RealtimeEventUpdate`, `RealtimeEventDelete`) for convenience. Each subscription accepts a single event filter value; to listen for multiple specific events, create multiple subscribers.

#### Scenario: Default Postgres Changes filter
- **WHEN** a Postgres Changes subscriber does not override filter methods
- **THEN** `Schema()` SHALL default to `"public"`, `Table()` SHALL default to `""` (all tables), and `EventFilter()` SHALL default to `"*"` (all events, using `constants.RealtimeEventAll`)

#### Scenario: Filtered Postgres Changes subscription
- **WHEN** a Postgres Changes subscriber overrides `Table()` and `EventFilter()`
- **THEN** the channel join payload SHALL include the specified table and event filter in the `postgres_changes` config

#### Scenario: Event filter constants
- **WHEN** a subscriber sets `EventFilter()` return value
- **THEN** it MAY use string constants from `pkg/supabase/constants`: `RealtimeEventAll` (`"*"`), `RealtimeEventInsert` (`"INSERT"`), `RealtimeEventUpdate` (`"UPDATE"`), `RealtimeEventDelete` (`"DELETE"`)

### Requirement: SubscriberBase Code Generation Detection

The system SHALL use `SubscriberBase` as a code generation marker. The CLI code generator (`pkg/generator/subscriber_register.go`) SHALL scan subscriber files via AST to detect structs embedding `SubscriberBase` and auto-register them in the server.

#### Scenario: Subscriber detection via AST
- **WHEN** the code generator processes files in `internal/subscribers/`
- **THEN** it SHALL find all structs that embed `SubscriberBase` using `getStructByBaseName(path, "SubscriberBase")` and generate registration code for each

#### Scenario: Adding new interface methods is non-breaking
- **WHEN** new methods are added to `SubscriberHandler` interface with defaults in `SubscriberBase`
- **THEN** existing subscribers SHALL continue to compile because they embed `SubscriberBase` which provides the defaults
