# Change: Add Supabase Realtime as a PubSub Provider

## Why
The pubsub system currently only supports Google Cloud Pub/Sub. Supabase Realtime (Broadcast, Presence, Postgres Changes) is a natural fit as a provider since Raiden is a Supabase-first framework. Adding it as a `PubSubProvider` lets subscribers listen to Realtime channels using the same `SubscriberHandler` contract, unifying the messaging model.

## What Changes
- Add `PubSubProviderSupabase` provider type constant
- Create `pkg/pubsub/supabase/` with a self-contained Realtime provider (WebSocket client)
- Create root-level `SupabaseRealtimeProvider` adapter (same pattern as `GooglePubSubProvider`)
- Support three Realtime channel types: **Broadcast**, **Presence**, **Postgres Changes**
- Support **publishing** to Broadcast channels via `ctx.Publish()`
- Register the Supabase provider in `NewPubsub()` alongside Google
- `SubscriberHandler` interface gains `ChannelType()` method (default in `SubscriberBase` — non-breaking since all subscribers must embed `SubscriberBase` for code generation detection)

## Impact
- Affected specs: `pubsub`
- Affected code: `pubsub.go`, `pkg/pubsub/supabase/` (new), `pkg/mock/provider.go`, `pkg/mock/pubsub.go`, `server.go`
- No new config fields — reuses existing `SupabasePublicUrl`/`SupabaseApiUrl` (derive Realtime WebSocket URL), `AnonKey`/`ServiceKey` (auth)
- No new external dependencies — uses `fasthttp/websocket` (already in go.mod)
