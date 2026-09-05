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
| `TRUSTED_PROXY_TOKEN` | Authenticates web BFF client-IP assertions to API |
| `CLOUDINARY_CLOUD_NAME` | Restricts seller image URLs |
| `NEXT_PUBLIC_CLOUDINARY_CLOUD_NAME` | Selects browser upload account |
| `NEXT_PUBLIC_CLOUDINARY_UPLOAD_PRESET` | Selects restricted unsigned browser preset |
| `BACKUP_S3_ENDPOINT` | S3-compatible backup endpoint |
| `BACKUP_S3_ACCESS_KEY` | Backup access key |
| `BACKUP_S3_SECRET_KEY` | Backup secret key |
| `BACKUP_S3_BUCKET` | Existing backup bucket |
| `GRAFANA_ADMIN_USER` | Internal Grafana administrator |
| `GRAFANA_ADMIN_PASSWORD` | Internal Grafana administrator password |

Create separate secret named by `demoReset.cloudinarySecret`. Only reset CronJob receives it:

| Key | Purpose |
| --- | --- |
| `CLOUDINARY_CLOUD_NAME` | Selects cleanup account |
| `CLOUDINARY_API_KEY` | Authenticates reset cleanup |
| `CLOUDINARY_API_SECRET` | Authenticates reset cleanup |

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

Persistent Debian server uses `values-k3s.yaml` and `.env.k3s.local` through
`make k3s-up`. It enables observability, keeps synthetic traffic opt-in, uses
Cloudflare public HTTPS with HTTP origin, and reconciles pull/application/reset
Secrets idempotently. Deployment mode defaults to complete stack:

```dotenv
K3S_DEPLOYMENT_MODE=full
```

Use backend mode when web runs on Vercel or another external platform:

```bash
K3S_DEPLOYMENT_MODE=backend make k3s-up
```

Backend mode applies `web.enabled=false` and `ingress.target=api`. Chart omits
web Deployment, Service, and smoke request; Ingress sends `/` to API port 8080.
Frontend Cloudinary secret keys become optional for lifecycle config. Direct
Helm users can set same values explicitly. See `docs/platform.md`.

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
- Reset CronJob is suspended; serialized workflow gates public/synthetic traffic
  during maintenance before creating manual job.
- Backup CronJob creates custom-format dumps and uploads them to external
  S3-compatible storage with seven-day retention.
- Demo values deploy small Prometheus and Tempo PVCs, internal Grafana, and a
  bounded synthetic buyer journey. Access Grafana with
  `kubectl -n cartlabs port-forward service/cartlabs-grafana 3001:3000`.
- Dependency ingress is limited to pods belonging to the same Helm release.
- Workloads run non-root with dropped capabilities and read-only root filesystems
  where image behavior permits.

Key topology values:

| Value | Default | Purpose |
| --- | --- | --- |
| `web.enabled` | `true` | Create in-cluster Next.js Deployment and Service |
| `ingress.target` | `web` | Route Ingress to `web:3000` or `api:8080` |

Run `make helm-check` before deployment. See `docs/platform.md` for provisioning,
rollback, backup, and restore procedures.
