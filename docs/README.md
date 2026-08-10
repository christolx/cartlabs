# Cartlabs Planning

Cartlabs is a portfolio-grade, multi-vendor e-commerce platform inspired by
Tokopedia and Shopee. It exists to demonstrate full-stack engineering and
platform operations, not production-scale commerce.

## Documents

- [Product](./product.md) — users, scope, and marketplace behavior
- [Architecture](./architecture.md) — system shape and technical boundaries
- [Delivery](./delivery.md) — environments, CI/CD, testing, and operations
- [Demo platform](./platform.md) — provisioning, deployment, rollback, and recovery
- [Operability](./operability.md) — telemetry, recovery, limits, and runbooks
- [Search service](./search-service.md) — ownership, failure behavior, migration, and rollback
- [Roadmap](./roadmap.md) — incremental implementation plan
- [Milestones](./milestone.md) — completed work and exit evidence
- [Decisions](./decisions.md) — accepted and pending architecture decisions

## Principles

1. Build one complete marketplace flow before expanding breadth.
2. Keep backend modular and extraction-ready without premature microservices.
3. Treat local deployment, observability, security, and testing as product work.
4. Prefer reproducible demos over long-lived mutable demo data.
5. Document important trade-offs through short architecture decisions.
