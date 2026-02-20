## ADDED Requirements

### Requirement: Supabase Realtime Provider

The system SHALL provide a Supabase Realtime provider (`PubSubProviderSupabase`) that connects to the Supabase Realtime server via WebSocket. The provider MUST support three channel types: Broadcast, Presence, and Postgres Changes. The provider package (`pkg/pubsub/supabase/`) MUST NOT import the root `raiden` package.

#### Scenario: Register Supabase provider on startup
- **WHEN** `NewPubsub(config, tracer)` is called
- **THEN** the Supabase Realtime provider SHALL be registered under `PubSubProviderSupabase` alongside the Google provider

#### Scenario: Subscribe to Broadcast channel
- **WHEN** a subscriber declares `Provider() = PubSubProviderSupabase` and `ChannelType() = RealtimeChannelBroadcast`
- **THEN** the provider SHALL join the Broadcast channel via WebSocket and deliver messages to `Consume` as `SubscriberMessage`

#### Scenario: Subscribe to Presence channel
- **WHEN** a subscriber declares `ChannelType() = RealtimeChannelPresence`
- **THEN** the provider SHALL join the Presence channel and deliver join, leave, and sync events as `SubscriberMessage`

#### Scenario: Subscribe to Postgres Changes channel
- **WHEN** a subscriber declares `ChannelType() = RealtimeChannelPostgresChanges`
- **THEN** the provider SHALL subscribe to database change events and deliver INSERT, UPDATE, DELETE events as `SubscriberMessage`

#### Scenario: Postgres Changes with declarative filters
- **WHEN** a subscriber declares `Table() = "orders"`, `Schema() = "public"`, and `EventFilter() = constants.RealtimeEventInsert`
- **THEN** the provider SHALL subscribe only to INSERT events on `public.orders`
- **NOTE** `EventFilter()` returns `string`; framework provides constants in `pkg/supabase/constants` (`RealtimeEventAll`, `RealtimeEventInsert`, `RealtimeEventUpdate`, `RealtimeEventDelete`). Each subscription takes a single filter; use multiple subscribers for multiple events.

#### Scenario: Publish to Broadcast channel
- **WHEN** `Publish(ctx, topic, message)` is called on the Supabase provider
- **THEN** it SHALL send the message to the specified Broadcast channel via WebSocket

#### Scenario: Supabase provider Serve returns error
- **WHEN** `Serve()` is called on the Supabase Realtime provider
- **THEN** it SHALL return an error because Supabase Realtime uses WebSocket, not HTTP push endpoints

### Requirement: Realtime Channel Type

The system SHALL define a `RealtimeChannelType` with constants for `RealtimeChannelBroadcast`, `RealtimeChannelPresence`, and `RealtimeChannelPostgresChanges`. Subscribers MAY declare their channel type via `ChannelType()`.

#### Scenario: Default channel type is empty
- **WHEN** a subscriber embeds `SubscriberBase` without overriding `ChannelType()`
- **THEN** `ChannelType()` SHALL return `""` (empty string), which is ignored by providers that do not use channel types (e.g., Google)

### Requirement: Supabase Realtime Message Mapping

The system SHALL map Supabase Realtime events to `SubscriberMessage` with `Data` containing the JSON-encoded event payload, `Attributes` containing `channel_type`, `event`, and `topic` metadata, and `Raw` holding the original WebSocket message.

#### Scenario: Broadcast message mapping
- **WHEN** a Broadcast event is received
- **THEN** `SubscriberMessage.Data` SHALL contain the event payload as JSON bytes
- **THEN** `SubscriberMessage.Attributes` SHALL include `channel_type=broadcast` and the event name

#### Scenario: Postgres Changes message mapping
- **WHEN** a Postgres Changes event is received for a table INSERT
- **THEN** `SubscriberMessage.Data` SHALL contain the new row data as JSON bytes
- **THEN** `SubscriberMessage.Attributes` SHALL include `channel_type=postgres_changes`, `event=INSERT`, and the table name

### Requirement: SubscriberBase Code Generation Detection

The CLI code generator SHALL use `SubscriberBase` as a marker type to auto-detect subscriber structs. The generator scans Go source files in `internal/subscribers/` via AST parsing, finds structs embedding `raiden.SubscriberBase`, and generates `internal/bootstrap/subscribers.go` to register them with `server.RegisterSubscribers()`. This is why all subscribers MUST embed `SubscriberBase`.

#### Scenario: Auto-detect subscriber by SubscriberBase embed
- **WHEN** the CLI generator scans `internal/subscribers/`
- **THEN** it SHALL find all structs embedding `raiden.SubscriberBase` and include them in the generated registration file

#### Scenario: Subscriber without SubscriberBase is not detected
- **WHEN** a struct in `internal/subscribers/` does not embed `raiden.SubscriberBase`
- **THEN** the CLI generator SHALL NOT include it in the generated registration file

### Requirement: WebSocket Connection Lifecycle

The Supabase Realtime provider SHALL manage WebSocket connection lifecycle including authentication, heartbeat, and reconnection.

#### Scenario: Authenticate on connect
- **WHEN** the provider connects to Supabase Realtime
- **THEN** it SHALL authenticate using the existing `AnonKey` or `ServiceKey` from `Config`

#### Scenario: Reconnect on disconnect
- **WHEN** the WebSocket connection drops unexpectedly
- **THEN** the provider SHALL attempt reconnection with exponential backoff

#### Scenario: Graceful shutdown
- **WHEN** `StopListen()` is called
- **THEN** the provider SHALL close the WebSocket connection gracefully and stop all channel subscriptions

## MODIFIED Requirements

### Requirement: Subscriber Handler Interface

The system SHALL define a `SubscriberHandler` interface that all subscribers MUST implement. The interface includes: `AutoAck()`, `Name()`, `Consume(ctx SubscriberContext, message SubscriberMessage) error`, `Provider()`, `PushEndpoint()`, `Subscription()`, `SubscriptionType()`, `Topic()`, `ChannelType()`, `Table()`, `Schema()`, and `EventFilter()`. All subscriber structs MUST embed `SubscriberBase`, which serves as both a runtime default provider and a code generation marker for the CLI scanner.

#### Scenario: Embed SubscriberBase for defaults
- **WHEN** a subscriber struct embeds `SubscriberBase`
- **THEN** it inherits default values: `AutoAck() = true`, `SubscriptionType() = SubscriptionTypePull`, `Provider() = PubSubProviderUnknown`, `ChannelType() = ""`, `Table() = ""`, `Schema() = "public"`, `EventFilter() = constants.RealtimeEventAll ("*")`, and empty strings for `Name`, `Subscription`, `PushEndpoint`, and `Topic`

#### Scenario: Unimplemented Consume returns error
- **WHEN** `SubscriberBase.Consume` is called without being overridden
- **THEN** it SHALL return an error "subscriber {name} is not implemented"

### Requirement: Multi-Provider Registry

The system SHALL maintain a map-based registry of providers (`map[PubSubProviderType]PubSubProvider`). Multiple providers MAY be active simultaneously. Each subscriber handler declares which provider it uses via `Provider()`.

#### Scenario: Default provider registration
- **WHEN** `NewPubsub(config, tracer)` is called
- **THEN** the Google Cloud Pub/Sub provider SHALL be registered under `PubSubProviderGoogle` and the Supabase Realtime provider SHALL be registered under `PubSubProviderSupabase`

#### Scenario: Custom provider registration
- **WHEN** `SetProvider(providerType, provider)` is called with a new provider type
- **THEN** the provider SHALL be stored in the registry and available for subscribers that declare that provider type

#### Scenario: Unknown provider lookup
- **WHEN** a subscriber or publish call references a provider type not in the registry
- **THEN** the system SHALL return an error "unsupported pubsub provider: {type}"
