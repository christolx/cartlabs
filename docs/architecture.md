# Architecture

## System Shape

Start with an extraction-ready modular monolith. Future phases may split selected
modules into gRPC services on k3s.

```text
Browser
  |
  v
Next.js web ---> Go REST API ---> PostgreSQL
                    |   |
                    |   +----------> Redis
                    v
                 RabbitMQ ---> Go worker
                    ^
                    |
              mock-payment
                 webhook
```

## Deployables

| Component | Responsibility |
| --- | --- |
| `web` | Next.js and TypeScript user interface |
| `api` | Go REST API and marketplace domain logic |
| `worker` | Go consumers for asynchronous jobs and events |
| `mock-payment` | External-gateway simulation and signed webhooks |
| PostgreSQL | Source of truth |
| RabbitMQ | Durable async jobs and domain-event delivery |
| Redis | Cache, rate limits, and short-lived coordination state |

## Backend Modules

- Identity and access
- Users and stores
- Catalog and media
- Inventory
- Cart
- Checkout
- Orders
- Payments
- Fulfillment
- Reviews
- Administration and audit

Each module owns its domain model and persistence access. Cross-module calls use
explicit interfaces. No module reads another module's tables directly.

## Contracts

### HTTP

- OpenAPI is written first and stored in repository.
- Public endpoints use `/api/v1`.
- TypeScript client is generated from validated specification.
- API errors use one documented problem format.
- Pagination, filtering, sorting, authentication, and idempotency stay consistent.

### Events

- Events use versioned names and schemas.
- Consumers are idempotent and tolerate redelivery.
- Transactional outbox publishes events after database commits.
- Failed messages use bounded retries and dead-letter queues.
- Events describe completed facts; commands request work.

### Future Services

- Browser continues using REST at platform edge.
- Extracted internal services communicate through protobuf/gRPC.
- Module interfaces and event ownership guide extraction boundaries.
- Extraction happens only after profiling or learning goals justify it.

## Data Rules

- PostgreSQL remains authoritative; Redis never stores irreplaceable state.
- Money uses integer minor units plus currency code; never floating point.
- IDs use one consistent sortable format.
- Timestamps use UTC; presentation handles local time zones.
- Schema migrations are forward-only and tested on representative data.
- Inventory changes use atomic operations and an auditable ledger.
- Checkout snapshots product names, prices, and seller details.

## Authentication and Authorization

- Application-owned email/password authentication
- Argon2id password hashing
- Short-lived access tokens and rotating refresh tokens
- Refresh token delivered through `HttpOnly`, `Secure`, `SameSite` cookie
- Role checks at route and domain layers
- Resource ownership checks for every seller operation
- Rate limits for authentication and sensitive mutations
- Optional OIDC integration deferred

## Payment Simulation

`mock-payment` exposes intent creation and a small checkout UI. It can simulate
success, failure, cancellation, duplicate callbacks, and delayed callbacks. Go API
verifies webhook signature, deduplicates provider event ID, and applies explicit
payment state transitions.

Provider integration sits behind an interface so a sandbox provider can replace
the mock without changing checkout domain logic.

## Frontend Defaults

- Next.js App Router with strict TypeScript
- Server rendering for public catalog pages
- Generated OpenAPI client as network boundary
- Accessible semantic UI and responsive layouts
- URL-owned search/filter state
- Role-specific buyer, seller, and admin areas

## Observability

- Structured JSON logs with request, user, order, and trace correlation IDs
- OpenTelemetry traces across HTTP, database, RabbitMQ, and worker jobs
- Prometheus metrics and Grafana dashboards
- Health, readiness, and startup probes
- Local optional observability profile to control resource use

