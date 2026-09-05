# Delivery

## Environments

| Environment | Runtime | Purpose |
| --- | --- | --- |
| Local | Docker Compose | Fast development |
| Deployment test | Disposable local k3s | `make k3s-e2e` |
| Demo backend | Persistent Debian laptop k3s | Full or backend-only release |
| Demo web (optional) | Vercel | Next.js web and same-origin BFF |

`make k3s-e2e` builds/imports `:local` images with pull policy `Never`, refuses
namespace reuse, and removes namespace afterward. Never use it as persistent
server deployment.

Default-branch and tag delivery builds runtime images on GitHub-hosted runners,
publishes immutable full-SHA GHCR tags with SBOM/provenance, then optionally
deploys from protected self-hosted runner. Persistent local operator path uses
same GHCR artifacts through `make k3s-up`; it never builds/imports images.

Both delivery workflow and local lifecycle reconcile complete application,
Cloudinary cleanup, and GHCR pull Secrets before Helm. Secrets never enter Git
or command output. Rotate input then reconcile; `global.secretRevision` rolls
affected app Pods.

Cloudflare owns public DNS/TLS/tunnel. Traefik origin remains HTTP. Helm's
external HTTPS value controls Secure cookies independently from origin TLS.
Forwarded client IP crosses BFF/API only through authenticated internal headers.

## Vercel web with k3s backend

Deploy k3s with `K3S_DEPLOYMENT_MODE=backend` and backend hostname
`api-cartlabs.christofle.dev`. Configure Vercel production environment:

```dotenv
API_INTERNAL_URL=https://api-cartlabs.christofle.dev/api/v1
TRUSTED_PROXY_TOKEN=<same-value-as-k3s>
NEXT_PUBLIC_CLOUDINARY_CLOUD_NAME=<cloud-name>
NEXT_PUBLIC_CLOUDINARY_UPLOAD_PRESET=<restricted-unsigned-preset>
```

Keep browser calls on existing `/api/backend/*` BFF path. Browser remains
same-origin -> no CORS change. `TRUSTED_PROXY_TOKEN` must match k3s secret so API
accepts BFF client-IP assertion. Store all values in Vercel environment config,
never Git.

Vercel preview deployments need same variables assigned to Preview environment
to reach shared demo backend. Previews then share demo data, accounts/session
store, rate limits, reset windows, and backend availability. Omit preview
variables when untrusted preview code must not access demo backend. Refresh
cookies remain isolated to each Vercel deployment origin because browser talks
only to BFF.

Reset is maintenance-window work. Workflow serializes with deploys, removes
public/synthetic traffic during reset, restores replicas through failure trap.

Required verification before merge:

```bash
make lint
make test
make helm-check
make infra-check
make build
```

See [local k3s setup](./k3s-setup.md) for installation and deployment commands.
See [platform runbook](./platform.md) for proxy trust, rollback, and recovery
limits.
