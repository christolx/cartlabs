# Home-server k3s platform

For copy-paste installation and deployment steps, see
[Local k3s setup](./k3s-setup.md). This document explains platform design and
operational constraints.

Cartlabs backend runs persistently on one Debian 12+ laptop. Full mode also runs
web there. Backend mode allows Vercel-hosted web. Cloudflare Tunnel publishes
matching Helm hostname; no inbound WAN port is required.

```text
Full:    Browser -> Cloudflare -> cloudflared -> Traefik -> web BFF -> API
Backend: Browser -> Vercel web BFF -> Cloudflare -> cloudflared -> Traefik -> API
```

Cloudflare terminates public TLS. Traefik keeps host routing. Helm sets
`ingress.externalHTTPS=true`, so API emits Secure refresh cookies even though
origin Ingress uses HTTP. This is single-node demo infrastructure, not HA.

## Provision host

Install Debian, create sudo-capable operator, enable SSH only on trusted LAN/VPN,
then copy Ansible examples:

```bash
cp infra/ansible/inventory.example.yml infra/ansible/inventory.yml
cp infra/ansible/group_vars/all.example.yml infra/ansible/group_vars/all.yml
```

Create remotely managed Cloudflare Tunnel. Dashboard public hostname must route
`cartlabs.christofle.dev` to `http://localhost:80`. Export token only for
provisioning, enable role in ignored group vars, run playbook:

```bash
export CLOUDFLARE_TUNNEL_TOKEN='...'
ansible-playbook -i infra/ansible/inventory.yml infra/ansible/playbook.yml
```

Ansible installs checksum-verified k3s, embedded-etcd snapshots, Traefik trust
for host tunnel path, official Cloudflare APT package, root-only token file, and
systemd service. GitHub runner is opt-in. Never expose ports 80, 443, or 6443 to
WAN; tunnel uses outbound connections.

## Prepare release input

Hosted CI publishes every runtime image to GHCR using `sha-<full-commit>` tags.
Copy operator example and fill all values:

```bash
cp .env.k3s.local.example .env.k3s.local
```

Use read-only GHCR package token. Generate independent random app secrets. DB
and RabbitMQ passwords must use URL-safe characters; lifecycle script derives
cluster service URLs. File stays ignored.

Choose `K3S_DEPLOYMENT_MODE=full` (default) with
`K3S_DOMAIN=cartlabs.christofle.dev`, or `backend` with
`K3S_DOMAIN=api-cartlabs.christofle.dev`. Both existing-host and Ansible host
paths use same lifecycle input.

## Lifecycle

```bash
make k3s-up       # reconcile namespace, Secrets, immutable Helm release, tests
make k3s-status
make k3s-logs     # K3S_LOG_COMPONENT=api narrows stream
make k3s-rebuild  # reconcile same/new SHA, restart app Deployments
make k3s-down     # uninstall release, retain namespace and PVCs
```

`K3S_DEPLOYMENT_MODE=backend make k3s-up` overrides file for one reconcile.
Backend mode removes web Deployment/Service, routes Ingress to API port 8080,
and omits web smoke check. Other backend and observability workloads remain.

Normal down never deletes persistent data. Explicit destructive removal:

```bash
K3S_PURGE=1 make k3s-purge
```

Observability defaults on. Synthetic traffic defaults off to avoid startup
noise; opt in with `K3S_SYNTHETIC_TRAFFIC=true`. Grafana stays cluster-internal:

```bash
kubectl -n cartlabs port-forward service/cartlabs-grafana 3001:3000
```

## Proxy trust chain

`cloudflared` is only origin client. k3s Traefik trusts `X-Forwarded-*` from
loopback and node address, not arbitrary peers. Web BFF validates first
`X-Forwarded-For` address and converts it to internal `X-Cartlabs-Client-IP`.
API accepts that header only with matching `TRUSTED_PROXY_TOKEN`; otherwise
rate limiting uses direct socket peer. Never expose web service directly outside
trusted host path.

## Reset maintenance

Reset drops/recreates demo tables. Host command `make k3s-reset` scales each
present public web, API, worker, and synthetic Deployment to zero, runs reset
Job, then restores replicas even on failure. Missing web Deployment in backend
mode is ignored. Shared lock prevents overlap with host lifecycle mutations;
durable replica state supports recovery after interruption. Expect maintenance
downtime; this is not zero-downtime reset. Never run host-side `make reset` for
cluster data.

## Recovery

Use `helm history`, immutable SHA rollback, and Helm tests for application
failure. Migrations remain forward-only. k3s stores twice-daily embedded-etcd
snapshots locally; copy snapshots off laptop before claiming disaster recovery.
Database backup remains disabled until local/external destination and restore
test exist.

Single laptop means disk, power, network, and control plane share one failure
domain. Keep Debian, cloudflared, k3s, images, and snapshots patched/monitored.
