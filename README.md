# Cartlabs

Portfolio-grade multi-vendor marketplace built with Next.js, Go, PostgreSQL,
Redis, RabbitMQ, and MinIO.

## Prerequisites

- Go 1.26+
- Node.js 24+
- pnpm 10+
- Docker with Compose v2

## Start

```bash
cp .env.example .env
make setup
make compose-up
make migrate
make dev
```

`make dev` runs in foreground. After Search starts, run this from another
terminal in repository root:

```bash
make seed
```

Useful commands:

```bash
make help
make check
make reset
```

Planning lives in [`docs/`](./docs/README.md). REST contract lives in
[`api/openapi/openapi.yaml`](./api/openapi/openapi.yaml).

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
