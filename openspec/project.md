# Project Context

## Purpose
Raiden is a Go framework for secure integration with Supabase. It provides a unified backend layer for RPC, Edge Functions, and REST APIs, preventing direct client-to-database calls by routing all Supabase interactions through a secure Go-based server. The framework includes a CLI for project scaffolding, code generation, schema management, and deployment.

**Key Objectives:**
- Prevent direct client-side database calls to bolster security
- Provide a unified layer for managing RPCs, Edge Functions, and APIs
- Offer tools for consistent database schema management and persistence

**Documentation:** https://raiden.sev-2.com

## Tech Stack
- **Language:** Go 1.23.0+ (toolchain 1.23.7)
- **HTTP Server:** [fasthttp](https://github.com/valyala/fasthttp) + [fasthttp/router](https://github.com/fasthttp/router)
- **CLI Framework:** [cobra](https://github.com/spf13/cobra)
- **Scheduler:** [gocron/v2](https://github.com/go-co-op/gocron/v2)
- **Pub/Sub:** [cloud.google.com/go/pubsub](https://cloud.google.com/go/pubsub) (extensible provider interface)
- **WebSocket:** [fasthttp/websocket](https://github.com/fasthttp/websocket)
- **Observability:** [OpenTelemetry](https://go.opentelemetry.io/otel) (tracing via OTLP gRPC/HTTP exporters)
- **Logging:** [hashicorp/go-hclog](https://github.com/hashicorp/go-hclog) (structured)
- **Auth:** [golang-jwt/jwt/v5](https://github.com/golang-jwt/jwt) (JWT-based)
- **Validation:** [go-playground/validator/v10](https://github.com/go-playground/validator)
- **Database:** PostgreSQL via [lib/pq](https://github.com/lib/pq), Supabase PostgREST & pg-meta
- **Config:** [ory/viper](https://github.com/ory/viper) (environment variables)
- **Testing:** [stretchr/testify](https://github.com/stretchr/testify), [jarcoal/httpmock](https://github.com/jarcoal/httpmock)
- **Build:** Makefile (cross-compile Linux/Windows/macOS, amd64/arm64)

## Project Conventions

### Code Style
- Run `go fmt ./...` before every commit (tabs, trailing newlines — Go defaults)
- **Exported APIs:** PascalCase (e.g., `NewServer`, `ExecuteRpc`)
- **Internal helpers:** lowerCamelCase (e.g., `createRouteGroups`, `handleRequest`)
- **Packages:** lower_snake_case matching directory (e.g., `pkg/pubsub`)
- **Constructors:** prefix with `New` (e.g., `NewLogger`, `NewConnector`)
- Keep files focused on a single capability (e.g., `server.go`, `router.go`, `pubsub.go`)
- Use structured logging from `pkg/logger`: `logger.HcLog().Named("raiden.component")`
- Comment only when clarification is needed; do not over-comment

### Architecture Patterns
- **Request flow:** HTTP → Server (fasthttp) → Router → Middleware → Controller → Context → Response
- **Route types:** `RouteTypeCustom`, `RouteTypeFunction`, `RouteTypeRest`, `RouteTypeRpc`, `RouteTypeRealtime`, `RouteTypeStorage`
- **Controller lifecycle:** `BeforeAll` → `Before{Method}` → Handler → `After{Method}` → `AfterAll`
- **Provider pattern:** Pub/Sub uses a `PubSubProvider` interface with a map-based registry (`map[PubSubProviderType]PubSubProvider`), allowing multiple providers (currently Google Cloud Pub/Sub; extensible to AWS, Kafka, etc.)
- **Adapter pattern:** Provider implementations live in `pkg/` (e.g., `pkg/pubsub/google/`) with no root-package imports; root adapters translate between framework types and provider types to avoid circular dependencies
- **Error handling:** return errors explicitly; use `Context.SendError()` / `Context.SendErrorWithCode()` in controllers

### Testing Strategy
- **Style:** Table-driven `TestXxx` functions colocated with source (`*_test.go`)
- **White-box tests:** use `_internal_test.go` for access to unexported symbols
- **Coverage:** test both success and failure paths; run `make test` before PRs
- **Mocks:** use `pkg/mock` to isolate Supabase, Google, and network dependencies
- **Examples:** add runnable samples under `examples/` for integration flows
- **Test command:** `make test` (runs `go test ./... -covermode=count`, outputs `coverage.out`)
- **Focused testing:** `go test ./pkg/<package> -run TestName -v`

### Git Workflow
- **Commit format:** `<type> : <summary>` — imperative mood, ≤72 characters
  - Types: `fix`, `feat`, `docs`, `test`, `refactor`, `chore`
  - Examples: `fix : rpc test`, `feat : add websocket support`, `Fix: Handle Custom Endpoints (#132)`
- **PR requirements:** summarize intent, list commands run, reference issues, include screenshots/CLI output for visible changes, update docs/examples when interfaces change
- **Security review:** required for changes to `deployments/`, `policy.go`, auth, authorization, or RLS

## Domain Context
- **Supabase integration:** Models and RPCs sync with Supabase schema via CLI commands (`generate`, `apply`, `imports`). The framework manages tables, functions, policies, roles, and storage buckets.
- **Pub/Sub:** supports pull subscriptions (long-polling from cloud provider) and push subscriptions (HTTP endpoints receiving webhook POSTs). Messages are wrapped in provider-agnostic `SubscriberMessage{Data, Attributes, Raw}`.
- **Self-hosted modes:** `bff` (Backend-for-Frontend — full Supabase route proxying) and `svc` (Service — custom routes only, connects directly to PostgREST and pg-meta).
- **CLI lifecycle:** `init` → `start` → `generate` → `apply` → `run` (dev) / `build` + `serve` (prod).

## Important Constraints
- **No circular imports:** `pkg/` packages must never import the root `raiden` package. Use adapters at the root level when framework types need to flow into provider implementations.
- **No committed credentials:** `google_sa.json` is a local-testing placeholder only. All secrets load via environment variables consumed by `config.go`. Use `.env` files (gitignored) for local dev.
- **Security-sensitive areas:** changes to `deployments/`, `policy.go`, and anything touching authentication, authorization, or RLS require security team review before merging.
- **Go version:** minimum Go 1.23.0 (toolchain 1.23.7).

## External Dependencies
- **Supabase:** PostgREST API, pg-meta API, Supabase Cloud API (auth, storage, realtime)
- **Google Cloud:** Pub/Sub (message broker), Service Account authentication
- **OpenTelemetry:** OTLP trace exporters (gRPC and HTTP) for distributed tracing
- **PostgreSQL:** primary data store (accessed via PostgREST or direct connection with lib/pq)

## Project Structure

```
raiden/
├── cmd/raiden/              # CLI entry point and subcommands
│   ├── main.go
│   └── commands/            # init, start, run, serve, generate, apply, configure, etc.
├── pkg/                     # Domain logic organized by capability
│   ├── connector/           # Supabase connection management
│   ├── pubsub/google/       # Google Cloud Pub/Sub provider (self-contained)
│   ├── resource/            # Schema and resource definitions
│   ├── controllers/         # Controller implementations
│   ├── generator/           # Code generation
│   ├── mock/                # Test mocks
│   ├── logger/              # Structured logging
│   ├── jwt/                 # JWT utilities
│   ├── tracer/              # OpenTelemetry tracing
│   ├── supabase/            # Supabase API client
│   └── utils/               # Shared utilities
├── docs/                    # Long-form documentation
├── examples/                # Runnable sample applications
├── deployments/             # Operational/deployment assets
├── scripts/                 # Helper automation scripts
├── build/                   # Build artifacts (via Makefile)
├── openspec/                # Specification-driven development
├── *.go                     # Root: server, router, controller, context, pubsub, etc.
├── *_test.go                # Tests colocated with source
├── Makefile                 # Build, test, clean targets
├── go.mod / go.sum          # Go module definition
└── README.md
```
