# Cartlabs Helm chart

This chart packages the complete demo stack for a single-node k3s cluster:
application services, PostgreSQL, Redis, RabbitMQ, database jobs, recovery jobs,
Prometheus, Tempo, internal Grafana, synthetic traffic, ingress, network policies,
and smoke tests.

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
| `SEARCH_DATABASE_URL` | Independent search-owned PostgreSQL database URL |
| `SEARCH_SERVICE_TOKEN` | Authenticates internal search gRPC calls |
| `BACKUP_S3_ENDPOINT` | S3-compatible backup endpoint |
| `BACKUP_S3_ACCESS_KEY` | Backup access key |
| `BACKUP_S3_SECRET_KEY` | Backup secret key |
| `BACKUP_S3_BUCKET` | Existing backup bucket |
| `GRAFANA_ADMIN_USER` | Internal Grafana administrator |
| `GRAFANA_ADMIN_PASSWORD` | Internal Grafana administrator password |

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
Build, import, install, upgrade, and verify an isolated local release with:

```bash
make k3s-e2e
```

The target refuses to reuse an existing namespace, removes only the namespace it
created, and imports every `cartlabs-*:local` image immediately before install.
When direct k3s containerd access is unavailable, `pkexec` opens the desktop
password prompt. Override `K3S_NAMESPACE` to choose another disposable namespace.
Set `K3S_BUILD=0` to reuse existing Docker images while still re-importing them.
Set `K3S_KEEP_NAMESPACE=1` only when failed-release inspection is needed.

## Lifecycle

- API and worker wait for PostgreSQL, run idempotent migrations, then start.
- Search owns a separate database, runs its own migration, and receives catalog events through RabbitMQ.
- Search reindex hook repairs initial state and drift after install or upgrade.
- A Helm migration hook runs after install and before each upgrade.
- Deterministic seed runs after first install only.
- Reset CronJob is suspended; serialized GitHub workflow creates manual jobs.
- Backup CronJob creates custom-format dumps and uploads them to external
  S3-compatible storage with seven-day retention.
- Demo values deploy small Prometheus and Tempo PVCs, internal Grafana, and a
  bounded synthetic buyer journey. Access Grafana with
  `kubectl -n cartlabs port-forward service/cartlabs-grafana 3001:3000`.
- Dependency ingress is limited to pods belonging to the same Helm release.
- Workloads run non-root with dropped capabilities and read-only root filesystems
  where image behavior permits.

Run `make helm-check` before deployment. See `docs/platform.md` for provisioning,
rollback, backup, and restore procedures.
