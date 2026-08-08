# Milestones

## Milestone 1 — Identity and Catalog

**Status:** Complete (2026-08-08)

### Work

- [x] Authentication, refresh rotation, RBAC, auth rate limiting, and demo login
- [x] Seller store management and admin moderation
- [x] Products, variants, categories, images, and atomic inventory ledger
- [x] Public catalog, search, filters, product pages, and role demo workflows

### Exit evidence

Seller creates a store and receives admin approval. Seller creates a product,
SKU, image, and stock, then publishes it. Product remains hidden until admin
approval. Buyer discovers approved product through search and opens its detail
page. Buyer cannot access admin APIs.

Verified through:

- Unit tests for Argon2id, JWT validation, refresh replay revocation, RBAC,
  validation, strict JSON decoding, cookies, and auth rate limiting
- Compose-backed Hurl workflow with 18 API requests
- Browser workflow across seller, admin, and buyer roles
- One shared, attributed stock product image used by all demo products
- Desktop/mobile and light/dark visual checks
- Axe checks with zero violations on catalog, product, and demo pages
- `make check`

## Milestone 2 — Purchase Vertical Slice

**Status:** Complete (2026-08-08)

### Work

- [x] Multi-seller cart
- [x] Checkout totals and inventory reservation
- [x] Parent purchase and seller orders
- [x] Mock payment intent, checkout UI, and signed webhook
- [x] Transactional outbox, RabbitMQ worker, and notifications

### Exit

Buyer completes a purchase spanning two sellers without overselling.

### Exit evidence

Buyer adds approved products from two sellers. Checkout locks variants in stable
order, snapshots prices and names, decrements inventory, and creates one parent
purchase with two seller orders. Mock payment sends a signed webhook. API makes
payment transition exactly once, worker publishes transactional outbox events,
and buyer plus both sellers receive durable notifications.

Verified through:

- Unit tests for service authorization and idempotency, payment intent failure
  cleanup, HMAC timestamp/signature validation, mock-provider delivery and retry,
  notification event mapping, and production secret validation
- PostgreSQL integration test proving concurrent checkout yields one success and
  one conflict, webhook replay creates one payment event/outbox fact, and expired
  reservations restore stock
- Compose-backed Hurl workflow with 25 purchase requests and 43 total requests,
  covering successful and failed payments, inventory restoration, RBAC, seller
  orders, notifications, idempotent checkout, and forged webhook rejection
- Browser workflow from two-seller cart through paid purchase and participating
  seller order/notification; shared local image used by every product
- Desktop/mobile and light/dark/reduced-motion visual checks
- Axe checks with zero violations on buyer and seller workflows
- `make check`

## Milestone 3 — Fulfillment and Trust

**Status:** Complete (2026-08-08)

### Work

- [x] Strict per-seller paid, processing, shipped, delivered, and cancelled lifecycle
- [x] Buyer and seller cancellation with transactional inventory restoration
- [x] One verified review per delivered purchase item
- [x] Admin marketplace overview and unified immutable audit trail
- [x] Fulfillment notifications through transactional outbox and RabbitMQ worker

### Exit

Complete buyer-seller-admin lifecycle works end to end.

### Exit evidence

Buyer pays for seller-owned inventory. Seller advances only valid state
transitions through delivery, while another seller can independently cancel and
restore its order stock. Buyer can review delivered items once, cannot review
cancelled items, and public product page exposes aggregate plus verified review.
Admin sees marketplace totals and catalog, moderation, fulfillment,
cancellation, and review activity through one immutable audit ledger.

Verified through:

- Unit tests for fulfillment authorization/input validation and every new
  notification event mapping
- PostgreSQL integration test proving seller isolation, strict transitions,
  partial cancellation stock restoration, review authorization/deduplication,
  buyer cancellation restoration, overview totals, and audit persistence
- Compose-backed Hurl workflow with 28 fulfillment/trust requests and 76 total
  requests, including negative RBAC and ownership paths
- Browser workflow spanning seller/store/product approval, buyer checkout and
  payment, seller processing through delivery, buyer review, public review,
  worker notifications, and admin overview/audit
- Shared local placeholder image used across every demo product
- Desktop/mobile and light/dark visual checks
- Axe checks with zero violations and zero incomplete checks on product and admin
  workflow targets
- `make check`

## Milestone 4 — Operability

**Status:** Complete (2026-08-08)

### Work

- [x] Prometheus metrics and provisioned Grafana dashboard
- [x] OpenTelemetry traces across HTTP, PostgreSQL, payment, outbox, and RabbitMQ
- [x] Correlated structured logs with bounded routes and request/trace IDs
- [x] Bounded retry, durable dead-letter queues, and operator replay command
- [x] Dependency outage tests and reconnect/fail-fast behavior
- [x] HTTP hardening, dependency scans, and runtime image scanning in CI
- [x] Repeatable performance baseline with explicit failure thresholds
- [x] Guarded daily demo reset and incident runbook

### Exit

Common failures are observable, recoverable, and documented.

### Exit evidence

Prometheus scrapes API and worker targets. Provisioned Grafana dashboard renders
five live panels. Tempo receives correlated API, database, payment, outbox, and
AMQP spans. JSON request logs carry request and trace IDs without raw SQL or
credentials. RabbitMQ restart preserves purchase events and worker catch-up;
Redis loss degrades readiness while existing JWT reads continue and new login
fails fast; payment-provider loss returns 503 and restores reserved inventory.

Verified through:

- Unit suite for configuration, request hardening, bounded metric labels,
  mutation limiting, retry counting, and all prior domain behavior
- PostgreSQL/RabbitMQ integration tests proving outbox backoff/dead-letter/replay,
  publisher-confirmed notification retries, dead-lettering, and replay
- Compose-backed Hurl workflow with 70 API requests across all vertical slices
- Automated real-service outage workflow across RabbitMQ, Redis, and payment
  provider, including recovery assertions
- `govulncheck` with zero reachable vulnerabilities, `pnpm audit` with zero
  known vulnerabilities, and local Trivy scans with zero HIGH/CRITICAL findings
  across six built runtime images; CI repository and full image-matrix gates
- 15-second/20-client catalog baseline: 201,682 requests, 0 failures, 13,445.0
  requests/second, 2.48 ms p95, and 3.28 ms p99
- Agent-browser production review with no console/page errors, zero Axe
  violations, 0.0 CLS, 48 ms FCP, and 13.5 ms TTFB; one indeterminate contrast
  result caused by the hero image gradient
- Guarded Compose reset and bounded SQL/RabbitMQ replay operator jobs
- `make check`
