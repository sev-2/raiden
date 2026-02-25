## 1. Core Types and Interface
- [x] 1.1 Add `PubSubProviderSupabase` constant to `pubsub.go`
- [x] 1.2 Add `RealtimeChannelType` type and constants (`RealtimeChannelBroadcast`, `RealtimeChannelPresence`, `RealtimeChannelPostgresChanges`)
- [x] 1.3 Add `ChannelType()`, `Table()`, `Schema()`, `EventFilter()` methods to `SubscriberHandler` interface
- [x] 1.4 Add defaults to `SubscriberBase`: `ChannelType()=""`, `Table()=""`, `Schema()="public"`, `EventFilter()="*"`
- [x] 1.5 Update `LegacySubscriberConsumer` interface and `legacySubscriberAdapter` with new methods

## 2. Supabase Realtime Provider Package
- [x] 2.1 Create `pkg/pubsub/supabase/provider.go` with `Provider` struct and `ProviderConfig`
- [x] 2.2 Implement WebSocket client connection (connect, authenticate, heartbeat)
- [x] 2.3 Implement `StartListen()` — subscribe to channels, dispatch messages to `ConsumeFn`
- [x] 2.4 Implement `StopListen()` — graceful WebSocket close
- [x] 2.5 Implement `Publish()` — send Broadcast messages over WebSocket
- [x] 2.6 Implement `CreateSubscription()` — no-op or channel join validation
- [x] 2.7 Implement channel type handling: Broadcast, Presence, Postgres Changes
- [x] 2.8 WebSocket interfaces defined inline in provider.go (`WebSocketConn`, `Dialer`)

## 3. Root Adapter
- [x] 3.1 Create `SupabaseRealtimeProvider` adapter in `pubsub.go` (same pattern as `GooglePubSubProvider`)
- [x] 3.2 Implement `Serve()` — return error (Supabase Realtime uses WebSocket, not HTTP push)
- [x] 3.3 Map Realtime messages to `SubscriberMessage` in adapter
- [x] 3.4 Register Supabase provider in `NewPubsub()` alongside Google

## 4. Mocks and Testing
- [x] 4.1 Update `pkg/mock/pubsub.go` — add `ChannelType`, `EventFilter`, `Schema`, `Table` to mock handler
- [x] 4.2 Write unit tests for `pkg/pubsub/supabase/provider.go` (connection, topic building, publish, dispatch)
- [x] 4.3 Verify existing Google provider tests and root tests still pass

## 5. Documentation and Spec
- [x] 5.1 Update `docs/pubsub.md` — add Supabase Realtime section with examples
- [x] 5.2 Update `docs/pubsub.md` — document `SubscriberBase` dual purpose (runtime defaults + code generation marker for CLI auto-registration)
- [x] 5.3 Update `openspec/specs/pubsub/spec.md` — add SubscriberBase code generation detection requirement
- [x] 5.4 Update `openspec/specs/pubsub/spec.md` after implementation
- [x] 5.5 Add example subscriber in `examples/` (tested in external project)

## 6. Quality Assurance
- [x] 6.1 Run `golangci-lint run` and fix all linting issues on new/changed files
- [x] 6.2 Add unit tests for `SupabaseRealtimeProvider` adapter in `pubsub_test.go` (Serve error, StartListen mapping, StopListen delegation)
- [x] 6.3 Add unit tests for `DefaultWebSocketDialer` type assertion
- [x] 6.4 Add unit tests for `newSupabaseProviderAdapter` config mapping
- [x] 6.5 Ensure `pkg/pubsub/supabase/` test coverage meets project target (86.3% — target 71%)
- [x] 6.6 Run `make test` — full suite passes
- [x] 6.7 Run `go vet ./...` — no vet issues
