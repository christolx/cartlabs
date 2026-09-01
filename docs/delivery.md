# Delivery

## Environments

| Environment | Runtime | Purpose |
| --- | --- | --- |
| Local | Docker Compose | Fast development |
| Deployment test | Disposable local k3s | `make k3s-e2e` |
| Demo | Persistent Debian laptop k3s | Public home-server release |

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

See [platform runbook](./platform.md) for provisioning, lifecycle, proxy trust,
rollback, and recovery limits.
