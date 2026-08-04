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
make seed
make dev
```

Useful commands:

```bash
make help
make check
make reset
```

Planning lives in [`docs/`](./docs/README.md). REST contract lives in
[`api/openapi/openapi.yaml`](./api/openapi/openapi.yaml).
