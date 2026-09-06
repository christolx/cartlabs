# Cartlabs

Cartlabs is a multi-vendor marketplace where buyers shop across independent
stores, pay once, then track each seller order separately. Sellers manage
stores, products, inventory, and fulfillment. Admins verify stores, enforce
listings, manage users, and inspect marketplace activity.

System uses Next.js, Go, PostgreSQL, Redis, RabbitMQ, and gRPC. Marketplace core
stays transactional in Go/PostgreSQL. Text search runs as an extracted service
with independent persistence, event-driven projection, and PostgreSQL fallback.

Cartlabs runs locally with Docker Compose or on single-node k3s through Helm,
Traefik, and Cloudflare Tunnel. Payments use signed mock-provider webhooks; no
real money moves.

## Prerequisites

- Go 1.26+
- Node.js 24+
- pnpm 10+
- Docker with Compose v2

## Local development

```bash
cp .env.example .env
make setup
make compose-up
make migrate
make dev
```

`make dev` runs in foreground. After search starts, run this from another
terminal in repository root:

```bash
make seed
```

Useful commands:

```bash
make help
make check
make reset
make web-e2e
```

`make test` runs Go tests, frontend type checking, and Vitest. `make web-e2e`
runs Playwright browser coverage with mocked API boundaries. `make check`
regenerates contracts, lints, tests, and builds.

## Home-server k3s

Persistent demo path targets existing or Ansible-provisioned Debian host. Helm
deploys full stack behind Traefik; Cloudflare Tunnel publishes it without
opening inbound WAN ports.

```bash
cp .env.k3s.local.example .env.k3s.local
# Set GHCR image SHA, hostname, credentials, and application secrets.
make k3s-up
make k3s-status
```

`full` mode hosts web and backend in k3s. `backend` mode hosts backend only for
external Next.js deployment:

```bash
K3S_DEPLOYMENT_MODE=backend make k3s-up
```

Matching immutable `sha-<commit>` images must already exist in GHCR. See
[`docs/k3s-setup.md`](./docs/k3s-setup.md) for setup commands and
[`docs/platform.md`](./docs/platform.md) for topology, reset, rollback, and
recovery limits.

Documentation lives in [`docs/`](./docs/README.md). REST contract lives in
[`api/openapi/openapi.yaml`](./api/openapi/openapi.yaml).
Role journeys live in [`docs/e2e-flow.md`](./docs/e2e-flow.md).

## Demo accounts

Set `DEMO_MODE=true` to enable role quick-login. Deterministic local/demo
credentials use password `demo-pass-123`:

| Role | Email |
| --- | --- |
| Buyer | `buyer@demo.cartlabs.local` |
| Seller | `seller@demo.cartlabs.local` |
| Admin | `admin@demo.cartlabs.local` |

Quick-login remains disabled by default and cannot be enabled with production's
local fallback token secret.
