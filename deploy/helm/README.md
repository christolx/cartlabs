# Cartlabs Helm chart

This chart packages the complete demo stack for a single-node k3s cluster:
API, worker, web, mock payment provider, PostgreSQL, Redis, RabbitMQ, database
hooks, reset and backup CronJobs, ingress, network policies, and smoke tests.

## Secret contract

Create the secret named by `global.existingSecret` before installation. The
default name is `cartlabs-secrets`.

| Key | Purpose |
| --- | --- |
| `ACCESS_TOKEN_SECRET` | Signs application access tokens |
| `PAYMENT_WEBHOOK_SECRET` | Verifies payment webhook signatures |
| `MOCK_PAYMENT_API_KEY` | Authenticates API calls to mock payment |
| `DATABASE_URL` | PostgreSQL connection URL |
| `POSTGRES_PASSWORD` | Initializes the bundled PostgreSQL user |
| `RABBITMQ_URL` | AMQP connection URL |
| `RABBITMQ_DEFAULT_PASS` | Initializes the bundled RabbitMQ user |
| `BACKUP_S3_ENDPOINT` | S3-compatible backup endpoint |
| `BACKUP_S3_ACCESS_KEY` | Backup access key |
| `BACKUP_S3_SECRET_KEY` | Backup secret key |
| `BACKUP_S3_BUCKET` | Existing backup bucket |

With release name `cartlabs`, internal hosts are `cartlabs-postgresql`,
`cartlabs-redis`, `cartlabs-rabbitmq`, and `cartlabs-mock-payment`. Keep URLs in
the secret synchronized if another release name is used.

## Install

Production deployment uses immutable `sha-<full commit>` tags:

```bash
helm upgrade --install cartlabs deploy/helm \
  --namespace cartlabs --create-namespace \
  --values deploy/helm/values-demo.yaml \
  --set-string global.imageRegistry=ghcr.io/OWNER \
  --set-string global.imageTag=sha-COMMIT \
  --set-string global.domain=demo.example.com \
  --rollback-on-failure --wait --timeout 12m
helm test cartlabs --namespace cartlabs --logs
```

`values-local.yaml` uses unqualified local images and `imagePullPolicy: Never`.
Import every `cartlabs-*:local` image into the local cluster before installing.

## Lifecycle

- API and worker wait for PostgreSQL, run idempotent migrations, then start.
- A Helm migration hook runs after install and before each upgrade.
- Deterministic seed runs after first install only.
- Reset CronJob is suspended; serialized GitHub workflow creates manual jobs.
- Backup CronJob creates custom-format dumps and uploads them to external
  S3-compatible storage with seven-day retention.
- Dependency ingress is limited to the release namespace.
- Workloads run non-root with dropped capabilities and read-only root filesystems
  where image behavior permits.

Run `make helm-check` before deployment. See `docs/platform.md` for provisioning,
rollback, backup, and restore procedures.
