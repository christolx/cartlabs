# Local k3s setup

Short answer: choose one host path, prepare release config, ensure matching SHA
images exist in GHCR, then run `make k3s-up`.

Use the existing-host path when k3s and deployment tools already work. Use the
Ansible path for a new Debian machine or when the repository should own host
configuration, hardening, snapshots, and Cloudflare Tunnel setup.

## Path A: existing k3s host

Ansible is not required when the machine already has:

- a working single-node k3s cluster with Traefik enabled;
- `kubectl`, Helm, and `sha256sum` available to the operator;
- a kubeconfig that lets `kubectl` manage the cluster; and
- enough CPU, memory, and storage for the complete Cartlabs stack.

Verify access:

```bash
kubectl get --raw=/readyz
helm version
```

If `kubectl` is not installed separately but k3s is local, install the kubectl
CLI or create a `kubectl` wrapper backed by `k3s kubectl`. A shell alias is not
enough because the lifecycle scripts run in a separate shell.

Skip to [Prepare release config](#prepare-release-config).

## Path B: new Debian host with Ansible

Use Debian 12 or newer. From an operator machine with Ansible, `kubectl`, and
Helm installed, copy the ignored host config:

```bash
cp infra/ansible/inventory.example.yml infra/ansible/inventory.yml
cp infra/ansible/group_vars/all.example.yml infra/ansible/group_vars/all.yml
```

Set real SSH host and user in `infra/ansible/inventory.yml`:

```yaml
all:
  children:
    k3s:
      hosts:
        cartlabs-home:
          ansible_host: 192.168.1.20
          ansible_user: operator
```

Keep `cloudflared_enabled: false` for LAN-only use. For public access, create a
remotely managed Cloudflare Tunnel, route the public hostname to
`http://localhost:80`, then set:

```yaml
cloudflared_enabled: true
```

Export the tunnel token only in the provisioning shell:

```bash
export CLOUDFLARE_TUNNEL_TOKEN='replace-with-real-token'
```

Provision the host:

```bash
ansible-playbook \
  -i infra/ansible/inventory.yml \
  infra/ansible/playbook.yml
```

Ansible installs pinned checksum-verified k3s, host prerequisites, hardening,
embedded-etcd snapshots, and Traefik forwarded-header trust. When enabled, it
also installs `cloudflared` as a systemd service.

Before deploying from another machine, securely copy
`/etc/rancher/k3s/k3s.yaml`, change its server address from loopback to the
host's trusted LAN or VPN address, and export that copy as `KUBECONFIG`. Never
expose Kubernetes port `6443` to the WAN.

Verify access:

```bash
kubectl get --raw=/readyz
helm version
```

## Prepare release config

Create private config:

```bash
cp .env.k3s.local.example .env.k3s.local
chmod 600 .env.k3s.local
```

Find the current full commit:

```bash
git rev-parse HEAD
```

Set the matching immutable image tag and target hostname in
`.env.k3s.local`:

```dotenv
K3S_IMAGE_REGISTRY=ghcr.io/christolx
K3S_IMAGE_TAG=sha-FULL_40_CHARACTER_LOWERCASE_COMMIT
K3S_DOMAIN=cartlabs.christofle.dev
K3S_DEPLOYMENT_MODE=full
```

`full` remains default and hosts Next.js plus backend in k3s. Use `backend` to
omit in-cluster web workload and route Traefik directly to API:

```dotenv
K3S_DOMAIN=api-cartlabs.christofle.dev
K3S_DEPLOYMENT_MODE=backend
```

That exact SHA must already be published to GHCR for every Cartlabs runtime
image. Usually: push the commit, then wait for the delivery workflow to publish
the images. `make k3s-up` pulls images; it does not build them.

Fill read-only GHCR credentials:

```dotenv
GHCR_USERNAME=your-github-username
GHCR_PULL_TOKEN=token-with-read-packages
```

Generate independent secrets:

```bash
openssl rand -hex 32
```

Use a separate output for each value:

- `ACCESS_TOKEN_SECRET`
- `PAYMENT_WEBHOOK_SECRET`
- `SEARCH_SERVICE_TOKEN`
- `TRUSTED_PROXY_TOKEN`
- `GRAFANA_ADMIN_PASSWORD`

Generate remaining passwords:

```bash
openssl rand -hex 24
```

Use a separate output for each value:

- `MOCK_PAYMENT_API_KEY`
- `POSTGRES_PASSWORD`
- `RABBITMQ_DEFAULT_PASS`

Database and RabbitMQ passwords must contain only URL-safe characters. Hex
output satisfies that requirement.

Full mode needs real frontend plus backend Cloudinary values. Both cloud names
must match:

```dotenv
CLOUDINARY_CLOUD_NAME=your-cloud
NEXT_PUBLIC_CLOUDINARY_CLOUD_NAME=your-cloud
NEXT_PUBLIC_CLOUDINARY_UPLOAD_PRESET=your-preset
CLOUDINARY_API_KEY=your-key
CLOUDINARY_API_SECRET=your-secret
```

Backend mode still needs `CLOUDINARY_CLOUD_NAME`, API key, and API secret for
API validation/reset cleanup. `NEXT_PUBLIC_*` values may stay empty in k3s;
configure them on Vercel instead.

Keep Grafana user explicit:

```dotenv
GRAFANA_ADMIN_USER=admin
```

## Deploy

Reconcile namespace, Secrets, PVC-backed services, and Helm release:

```bash
make k3s-up
```

One-off mode override takes precedence over `.env.k3s.local`:

```bash
K3S_DEPLOYMENT_MODE=backend make k3s-up
```

Accepted modes: `full`, `backend`. Command rejects other values, waits for
enabled workloads, then runs matching Helm smoke tests before succeeding.

Inspect status and logs:

```bash
make k3s-status
K3S_LOG_COMPONENT=api make k3s-logs
```

Before enabling Cloudflare Tunnel, test Traefik locally using the configured
hostname:

```bash
curl -H 'Host: cartlabs.christofle.dev' \
  http://127.0.0.1/api/backend/health/ready
```

Backend mode health path reaches API directly:

```bash
curl -H 'Host: api-cartlabs.christofle.dev' \
  http://127.0.0.1/api/v1/health/ready
```

When running that command from a different LAN machine, replace `127.0.0.1`
with the k3s host address.

For public backend mode, configure Cloudflare Tunnel public hostname:

```text
Hostname: api-cartlabs.christofle.dev
Service URL: http://localhost:80
HTTP Host Header: api-cartlabs.christofle.dev
```

DNS hostname, HTTP Host Header, and `K3S_DOMAIN` must match.

After changing config or image SHA, reconcile and restart application
Deployments:

```bash
make k3s-rebuild
```

Status, logs, rebuild, down, and purge commands work unchanged in both modes.

## Stop or remove

Stop while retaining namespace and persistent data:

```bash
make k3s-down
```

Delete namespace and PVCs only when intentional:

```bash
K3S_PURGE=1 make k3s-purge
```

Purge permanently deletes Cartlabs data stored in that namespace.

## Local disposable deployment test

`make k3s-e2e` is a different path. It builds and imports local images into a
disposable namespace, tests install and upgrade, then removes the namespace.
Use it for deployment validation, not a persistent local installation:

```bash
make k3s-e2e
```

See [platform runbook](./platform.md) for architecture, proxy trust, rollback,
reset maintenance, and recovery limits.
