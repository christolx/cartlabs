# Delivery

## Environments

| Environment | Runtime | Purpose |
| --- | --- | --- |
| Local | Docker Compose | Fast development and integration testing |
| CI | Compose or service containers | Automated verification |
| Demo | k3s + Helm | Portfolio deployment and operations practice |

Infrastructure platform remains intentionally undecided until application MVP is
stable. Terraform provisions infrastructure; Ansible configures hosts and k3s.
Helm packages application workloads.

## Repository Shape

```text
apps/
  web/
  api/
  worker/
  mock-payment/
  migrate/
  seed/
  reset/
api/
  openapi/
deploy/
  compose/
  helm/
infra/
  terraform/
  ansible/
docs/
```

Go API and worker may share one Go module and container build stages while
remaining separate runtime commands.

## CI

Pull requests should run:

1. Formatting, linting, and static analysis
2. Unit and module integration tests
3. OpenAPI linting and generated-client drift check
4. Database migration tests
5. Compose-backed end-to-end tests
6. Container build and vulnerability scan
7. Helm lint/render and infrastructure validation

Default-branch builds publish immutable images to GHCR using commit SHA tags.
Release tags add semantic-version tags. Avoid mutable `latest` in deployments.

## CD

- GitHub Actions authenticates with least privilege.
- Helm values reference immutable image tags or digests.
- Demo deployment runs migrations as a controlled pre-deploy job.
- Smoke tests verify deployment before success is reported.
- Daily scheduled workflow resets database through migration-and-seed job.
- Reset job is mutually exclusive with active migration/deployment jobs.

## Testing Strategy

- **Unit:** domain state transitions and pure business rules
- **Integration:** PostgreSQL repositories, Redis, RabbitMQ, and payment webhooks
- **Contract:** OpenAPI request/response behavior and event schemas
- **End-to-end:** Hurl API vertical slices; Playwright buyer, seller, and admin journeys
- **Security:** authorization matrix, webhook replay, rate limits, dependency scan
- **Resilience:** duplicate events, worker restart, dependency outage, retry behavior

Use Testcontainers for backend integration tests where practical. Keep a small,
deterministic end-to-end suite for CI speed.

## Local Developer Experience

- One command starts required services.
- Compose profiles enable optional observability and tooling.
- Health checks gate dependent containers.
- Example environment file contains no secrets.
- Makefile or task runner exposes consistent commands for lint, test, seed, reset,
  generate, and run.
- `make e2e-api` runs native Hurl workflows against a running local API. Override
  `E2E_API_URL` when targeting another environment.

## Security Baseline

- Secrets come from environment or deployment secret store, never Git.
- Containers run non-root with read-only filesystem where practical.
- Images use pinned minimal bases and multi-stage builds.
- CORS, CSRF, cookie, upload, and proxy trust policies are explicit.
- Uploaded media validates size and type; object storage is recommended over DB.
- Admin and seller mutations produce audit records.
- Demo quick login is impossible outside demo mode.

## Recommended Supporting Tools

- S3-compatible object storage such as MinIO for local product media
- OpenTelemetry Collector, Prometheus, Grafana, and optionally Loki/Tempo
- Trivy for images and IaC; Dependabot or Renovate for dependencies
- Hurl for black-box API slice workflows; Playwright for browser flows
- `golangci-lint`, TypeScript ESLint, and OpenAPI linting
