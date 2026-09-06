# Roadmap

Roadmap records completed delivery phases. Future work follows evidence from
current operation rather than adding service count.

## 0 — Foundations

**Status:** Complete

- Repository layout and developer commands
- Compose stack with PostgreSQL, Redis, and RabbitMQ
- CI lint, test, and build path
- OpenAPI conventions and first generated client
- Migration and deterministic seed framework

**Exit:** fresh clone boots, migrates, seeds, and passes health checks.

## 1 — Identity and Catalog

**Status:** Complete

- Authentication, refresh rotation, RBAC, and demo login
- Seller store management
- Products, variants, categories, images, and inventory
- Public catalog, search, filters, and product pages

**Exit:** seller publishes stock; buyer discovers it; admin can moderate it.

## 2 — Purchase Vertical Slice

**Status:** Complete

- Multi-seller cart
- Checkout totals and inventory reservation
- Parent purchase and seller orders
- Mock payment intent, checkout UI, and signed webhook
- Transactional outbox, RabbitMQ worker, and notifications

**Exit:** buyer completes a purchase spanning two sellers without overselling.

## 3 — Fulfillment and Trust

**Status:** Complete

- Per-seller fulfillment lifecycle
- Cancellation and inventory release
- Reviews gated by completed purchases
- Admin overview and audit trail

**Exit:** complete buyer-seller-admin lifecycle works end to end.

## 4 — Operability

**Status:** Complete

- Metrics, dashboards, traces, and structured logs
- Retry, dead-letter, replay, and outage tests
- Security hardening and image scanning
- Performance baseline and documented limits
- Daily demo reset

**Exit:** common failures are observable, recoverable, and documented.

## 5 — k3s Demo Platform

**Status:** Complete (2026-08-08)

- Helm chart with full-stack and backend-only modes
- Debian Ansible provisioning, embedded-etcd snapshots, and Cloudflare Tunnel
- GHCR publishing, optional protected-runner deployment, and host lifecycle
- External TLS, ingress, secret reconciliation, guarded reset, probes, and smoke tests

Database backup chart remains available but disabled on current home server
until durable destination and restore test exist.

**Exit:** reproducible k3s demo deployment from documented workflow.

## 6 — Microservice Learning Track

**Status:** Complete (2026-08-08)

- Measure and select one meaningful extraction boundary
- Define protobuf contract and ownership
- Extract service with gRPC, independent persistence, and telemetry
- Add compatibility, failure, and migration strategy

Search was selected. Notifications remained inside worker because recipient and
message ownership stay coupled to marketplace state. Orders and payments remain
inside transactional core.

**Exit:** extracted service demonstrates justified boundary, not service count.
