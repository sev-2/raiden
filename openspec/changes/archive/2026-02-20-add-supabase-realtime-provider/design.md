## Context
Raiden's pubsub layer uses a provider-agnostic interface (`PubSubProvider`) with a map-based registry. Currently only Google Cloud Pub/Sub is implemented. Supabase Realtime offers three channel types (Broadcast, Presence, Postgres Changes) over WebSocket, making it a second provider candidate.

The existing WebSocket proxy in `websocket.go` (BFF mode only) forwards raw connections to Supabase's realtime server. This proposal adds a **managed provider** that connects as a WebSocket client, enabling subscribers in both BFF and SVC modes without exposing raw WebSocket plumbing to users.

## Goals / Non-Goals
- **Goals:**
  - Implement Supabase Realtime as a `PubSubProvider`
  - Support all three channel types: Broadcast, Presence, Postgres Changes
  - Support publishing to Broadcast channels
  - Use WebSocket client connection (works in BFF and SVC modes)
  - Follow the adapter pattern (no root imports in `pkg/pubsub/supabase/`)
  - Maintain backward compatibility with existing subscribers

- **Non-Goals:**
  - Replace or modify the existing WebSocket proxy in BFF mode
  - Support direct PostgreSQL LISTEN/NOTIFY (may be added later)
  - Presence state management APIs beyond receiving change events
  - Admin/management operations on Supabase Realtime channels

## Decisions

### WebSocket client library
- **Decision:** Use `github.com/fasthttp/websocket` (already a direct dependency) for the Realtime WebSocket client
- **Alternatives:** `gorilla/websocket` (deprecated maintainer), `nhooyr.io/websocket` (extra dependency). fasthttp/websocket is already in go.mod and aligns with the HTTP server choice

### Configuration reuse
- **Decision:** Reuse existing config fields — no new environment variables needed
  - **Realtime URL:** Derive from `SupabasePublicUrl` (cloud: `wss://{host}/realtime/v1/websocket`) or `SupabaseApiUrl` (self-hosted)
  - **Auth:** Use `AnonKey` for standard access, `ServiceKey` for elevated access (both already in `Config`)
  - **Project:** Use `ProjectId` for channel namespace
- **Alternatives:** Add dedicated `SUPABASE_REALTIME_URL` env var. Rejected because the URL is deterministic from existing config and adding more env vars increases configuration burden

### Channel type modeling
- **Decision:** Add a `ChannelType()` method returning a `RealtimeChannelType` to `SubscriberHandler` interface, with a default of `""` in `SubscriberBase` (non-breaking since all subscribers embed `SubscriberBase`)
- **Alternatives:** Encode channel type in the `Topic()` string with a prefix convention (e.g., `broadcast:my-topic`). Rejected because it's implicit and error-prone

### Postgres Changes filters
- **Decision:** Add `Table()`, `Schema()`, `EventFilter()` methods to `SubscriberHandler` interface with defaults in `SubscriberBase`: `Table()=""`, `Schema()="public"`, `EventFilter()="*"`. This matches the declarative pattern of `Topic()`/`Subscription()` and keeps filter logic out of `Consume`
- **Alternatives:** Filter inside `Consume` (simpler interface but pushes work to every subscriber). Rejected for consistency

### Heartbeat
- **Decision:** Use Supabase defaults (30s heartbeat interval). No config field needed
- **Alternatives:** Configurable via `Config` field. Rejected — unnecessary complexity

### Message mapping
- **Decision:** Map Realtime payloads to `SubscriberMessage` as follows:
  - `Data`: JSON-encoded event payload
  - `Attributes`: `{"channel_type": "broadcast|presence|postgres_changes", "event": "<event_name>", "topic": "<channel_topic>"}`
  - `Raw`: Original WebSocket message frame (for advanced use)

### Provider package isolation
- **Decision:** `pkg/pubsub/supabase/` uses its own config struct (`ProviderConfig`) and types, no `raiden` imports. Root adapter translates, matching the Google provider pattern.

### Authentication
- **Decision:** Use existing `AnonKey` or `ServiceKey` from config. The WebSocket connection sends the key as a query parameter (`apikey=`) per Supabase Realtime protocol. No new config fields needed.

## Risks / Trade-offs
- **WebSocket reconnection** → Implement exponential backoff with jitter; log disconnections. If connection drops during `StartListen`, return error to trigger `run.Group` interrupt/restart
- **Message ordering** → Supabase Realtime does not guarantee strict ordering. Document this limitation
- **Presence state** → Presence events (join/leave/sync) are delivered as messages but the provider does not maintain a local presence map. Users needing state tracking must build it themselves
- **Breaking interface change** → Adding `ChannelType()` to `SubscriberHandler` is safe because all subscribers MUST embed `SubscriberBase` (the CLI generator uses it as a marker type to detect and auto-register subscribers via AST scanning). The default `ChannelType()` in `SubscriberBase` returns `""`, so existing subscribers compile and register without changes

## Open Questions
- (Resolved) Postgres Changes filters: declarative methods on handler (`Table()`, `Schema()`, `EventFilter()`)
- (Resolved) Heartbeat: use Supabase defaults (30s)
