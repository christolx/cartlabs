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
| Demo host | Single Hetzner Cloud `cx23` | Reproducible platform within EUR 15 monthly ceiling |
| Demo delivery | Direct Helm from protected GitHub environment | Small cluster does not justify GitOps controller overhead |
| Deployment runner | Unprivileged self-hosted runner on demo node | Outbound GitHub link keeps Kubernetes API off public CI networks |
| Demo edge | Traefik + cert-manager + external DNS records | Native k3s ingress with automated public TLS |
| Demo secrets | Protected GitHub environment to Kubernetes Secret | Audited injection without secrets in Git |
| Registry/CI | GHCR + GitHub Actions | Integrated image delivery workflow |
| Product media | One shared local demo asset | User-approved placeholder avoids premature media pipeline |
| Demo backups | External S3 dumps + etcd/provider snapshots | Independent, layered recovery with seven-day dump retention |
| Observability | OpenTelemetry + Prometheus/Grafana | Portable instrumentation, metrics, and dashboards |
| Browser testing | Agent-browser + Axe | Fast cross-role QA, responsive review, and accessibility evidence |
| Security scanning | Trivy | Scans images, dependencies, manifests, and IaC in CI |
| Event reliability | Transactional outbox + retries + DLQ | Prevents lost events and isolates persistent failures |
| Mutation safety | Idempotency keys and event deduplication | Makes retries safe across checkout and payment flows |
| Privileged actions | Append-only audit log | Makes seller and admin changes traceable |
| Demo recovery | Daily migration-and-seed reset | Restores predictable state after public use |
| Email | No email flow or local email tooling | Keeps project focused on marketplace core |

## Pending

Decide these when their roadmap phase begins:

- search evolution from PostgreSQL to dedicated engine
- first microservice extraction candidate

## Decision Record Template

For decisions needing more detail, add `docs/adr/NNNN-short-title.md`:

```markdown
# NNNN — Title

Status: proposed | accepted | superseded

## Context

## Decision

## Consequences
```
