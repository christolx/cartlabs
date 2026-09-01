# Decisions

## Accepted

| Decision | Choice | Reason |
| --- | --- | --- |
| Product | Multi-vendor marketplace | Matches portfolio goal and inspiration |
| Actors | Buyer, seller, admin | Covers core marketplace responsibilities |
| Backend | Modular Go monolith + worker | Strong boundaries without premature distribution |
| API | Spec-first REST/OpenAPI | Clear contract and generated TypeScript client |
| Future RPC | Internal gRPC | Suitable after justified service extraction |
| Database | PostgreSQL | Transactional system of record |
| Messaging | RabbitMQ | Durable async workflows and event practice |
| Redis | Cache and ephemeral coordination | Fast state without replacing source of truth |
| Orders | Parent purchase + seller child orders | Correct multi-vendor payment and fulfillment model |
| Payments | Separate mock provider + signed webhook | Realistic workflow without real money |
| Auth | Application-owned | Demonstrates identity fundamentals |
| Product model | Products with SKU variants | Useful complexity without catalog sprawl |
| Local runtime | Docker Compose | Low-friction development and testing |
| Demo runtime | k3s + Helm | Platform-learning target |
| Demo host | Existing Debian laptop | Persistent 24/7 home server without provider infrastructure |
| Demo delivery | Direct Helm from protected GitHub environment | Small cluster does not justify GitOps controller overhead |
| Deployment runner | Optional unprivileged self-hosted runner | Protected deploy only; local Make lifecycle remains primary |
| Demo edge | Cloudflare Tunnel + Traefik | Outbound-only publication; Cloudflare owns public TLS |
| Demo secrets | Protected GitHub environment to Kubernetes Secret | Audited injection without secrets in Git |
| Registry/CI | GHCR + GitHub Actions | Integrated image delivery workflow |
| Product media | One shared local demo asset | User-approved placeholder avoids premature media pipeline |
| Demo backups | Local etcd snapshots initially | DB backup disabled until destination and restore test exist |
| Observability | OpenTelemetry + Prometheus/Grafana | Portable instrumentation, metrics, and dashboards |
| Browser testing | Agent-browser + Axe | Fast cross-role QA, responsive review, and accessibility evidence |
| Security scanning | Trivy | Scans images, dependencies, manifests, and IaC in CI |
| Event reliability | Transactional outbox + retries + DLQ | Prevents lost events and isolates persistent failures |
| Mutation safety | Idempotency keys and event deduplication | Makes retries safe across checkout and payment flows |
| Privileged actions | Append-only audit log | Makes seller and admin changes traceable |
| Demo recovery | Daily migration-and-seed reset | Restores predictable state after public use |
| Email | No email flow or local email tooling | Keeps project focused on marketplace core |
| First extraction | Search candidate retrieval over gRPC | Eventual read projection teaches service boundaries without splitting commerce transactions |

## Pending

Decide these when their roadmap phase begins:

- search engine evolution after ranking or scale evidence

## Decision Record Template

For decisions needing more detail, add `docs/adr/NNNN-short-title.md`:

```markdown
# NNNN — Title

Status: proposed | accepted | superseded

## Context

## Decision

## Consequences
```
