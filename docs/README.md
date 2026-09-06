# Cartlabs Documentation

Cartlabs is a multi-vendor marketplace inspired by Tokopedia and Shopee. Buyers
shop across independent stores in one cart; sellers manage catalog and
fulfillment; admins verify stores and enforce marketplace rules.

Last audited against runtime code, contracts, fixtures, delivery config, and CI:
2026-09-06.

## Current Reference

- [Product](./product.md) — users, scope, and marketplace behavior
- [End-to-end flows](./e2e-flow.md) — buyer, seller, admin, payment, and notification journeys
- [Architecture](./architecture.md) — system shape and technical boundaries
- [Delivery](./delivery.md) — environments, CI/CD, testing, and operations
- [Local k3s setup](./k3s-setup.md) — existing-host and new Debian installation steps
- [Demo platform](./platform.md) — provisioning, deployment, rollback, and recovery
- [Operability](./operability.md) — telemetry, recovery, limits, and runbooks
- [Search service](./search-service.md) — ownership, failure behavior, migration, and rollback
- [Roadmap](./roadmap.md) — completed phases and current platform limits
- [Milestones](./milestone.md) — completed work and exit evidence
- [Decisions](./decisions.md) — accepted and pending architecture decisions

## Principles

1. Build one complete marketplace flow before expanding breadth.
2. Keep backend modular and extraction-ready without premature microservices.
3. Treat local deployment, observability, security, and testing as product work.
4. Prefer reproducible demos over long-lived mutable demo data.
5. Document important trade-offs through short architecture decisions.
