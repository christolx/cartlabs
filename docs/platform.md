# Demo platform

Cartlabs runs as a reproducible, single-node k3s demo. It is intentionally
small, recoverable, and educational; it is not a production high-availability
topology.

## Platform shape

| Concern | Choice |
| --- | --- |
| Host | One Hetzner Cloud Ubuntu 24.04 `cx23` node |
| Provisioning | Terraform with encrypted S3-compatible remote state |
| Configuration | Ansible, pinned checksum-verified k3s binary |
| Kubernetes | `v1.36.3+k3s1`, embedded etcd, Traefik |
| TLS | cert-manager `v1.21.1` and Let's Encrypt HTTP-01 |
| Delivery | Direct Helm on dedicated self-hosted deployment runner |
| Images | GHCR, immutable full-commit SHA tags, SBOM and provenance |
| Secrets | GitHub environment secrets reconciled to one Kubernetes Secret |
| Backups | Daily PostgreSQL dump to external S3 plus k3s etcd snapshots |
| Recovery | Helm rollback, forward database fixes, daily deterministic reset |

Base demo infrastructure must stay below EUR 15 per month before taxes. Check
the provider quote before every first apply or server-class change. External
DNS and S3-compatible storage remain provider-independent.

## Provisioning workflow

### 1. Prepare state and host inputs

Create an Ed25519 operator key and an S3 bucket with versioning and encryption
for Terraform state. Copy `infra/terraform/terraform.tfvars.example` to the
ignored `terraform.tfvars`, restrict `operator_cidrs` to trusted public CIDRs,
and export the Hetzner token through `TF_VAR_hcloud_token`.

Initialize the locked backend using the command in `infra/terraform/README.md`, then:

```bash
terraform -chdir=infra/terraform plan -out=cartlabs.tfplan
terraform -chdir=infra/terraform apply cartlabs.tfplan
```

Terraform creates one host, an operator SSH key, and a firewall. Only ports 80,
443, and ICMP are public. SSH and Kubernetes API access are restricted to
`operator_cidrs`. `prevent_destroy` blocks accidental server deletion.

### 2. Configure DNS and host

Create an A record from `server_ipv4` and, when used, an AAAA record from
`server_ipv6`. Wait for public resolution before requesting a production
certificate.

Wait for `cloud-init status --wait`. Copy Ansible examples to ignored real
files, set the domain and certificate email, verify the SSH host key out of
band, then run:

```bash
cd infra/ansible
ansible-playbook -i inventory.yml playbook.yml
```

Ansible hardens the host, disables swap, installs k3s, enables secrets
encryption and twice-daily embedded-etcd snapshots, and installs cert-manager.
Use the staging issuer while validating new DNS or ingress changes to avoid
Let's Encrypt rate limits.

Create a repository runner registration token, export it as
`GITHUB_RUNNER_REGISTRATION_TOKEN`, enable `github_runner_enabled`, and rerun
Ansible. The checksum-pinned runner uses an unprivileged account, accepts only
jobs carrying label `cartlabs-demo`, and reaches Kubernetes through the local
API. It builds no images; hosted runners retain that work. Never target this
runner from pull-request workflows because a cluster-admin job can become host
control through Kubernetes.

### 3. Configure protected GitHub environment

Fetch `/etc/rancher/k3s/k3s.yaml` over SSH, retain its `127.0.0.1` API endpoint,
base64-encode the complete file, and store it as `KUBE_CONFIG_B64` in the
protected `demo` environment. Require reviewer approval for that environment.
Keep a separate operator copy using the server TLS SAN; never place it in Git.

Set environment variable `DEMO_HOST` and these environment secrets:

- `KUBE_CONFIG_B64`, `GHCR_USERNAME`, `GHCR_PULL_TOKEN`
- `ACCESS_TOKEN_SECRET`, `PAYMENT_WEBHOOK_SECRET`, `MOCK_PAYMENT_API_KEY`
- `DATABASE_URL`, `POSTGRES_PASSWORD`
- `RABBITMQ_URL`, `RABBITMQ_DEFAULT_PASS`
- `BACKUP_S3_ENDPOINT`, `BACKUP_S3_ACCESS_KEY`, `BACKUP_S3_SECRET_KEY`,
  `BACKUP_S3_BUCKET`

Use internal hosts `cartlabs-postgresql:5432` and
`cartlabs-rabbitmq:5672` in application URLs. Grant GHCR token read-only package
scope. Create backup bucket before first deployment. Rotate a secret in GitHub,
then dispatch delivery; `global.secretRevision` rolls affected Pods.

## Delivery and verification

Default-branch delivery builds eight runtime images independently, publishes
full-SHA tags with SBOM and provenance, and deploys the exact SHA through Helm.
Deployment and reset share `cartlabs-demo-mutation` concurrency, preventing
migrations and destructive demo resets from overlapping.

Helm waits for probes, runs migrations, seeds first install, runs an in-cluster
test, then external smoke checks readiness, homepage rendering, catalog data,
request IDs, demo login, and cart access.

Operator checks:

```bash
kubectl -n cartlabs get pods,jobs,ingress
kubectl -n cartlabs get certificate
helm status cartlabs -n cartlabs
DEMO_URL=https://demo.example.com make deployment-smoke
```

## Rollback

Images are immutable, so application rollback targets a known Helm revision:

```bash
helm history cartlabs -n cartlabs
helm rollback cartlabs REVISION -n cartlabs --wait --timeout 12m
helm test cartlabs -n cartlabs --logs
```

Migrations are forward-only. Roll back application images only when intervening
schema changes are backward-compatible. Otherwise deploy a forward fix. Never
delete or edit an applied migration. Inspect hook logs and events before retrying
a failed release.

## Backup and restore

Backup CronJob runs at 02:17 UTC, waits for PostgreSQL, creates a custom-format
dump, uploads it under `cartlabs/`, and deletes objects older than 168 hours.
External S3 storage must enable encryption, versioning, and restricted
credentials. Hetzner whole-server backups and five twice-daily embedded-etcd
snapshots complement database dumps; neither replaces them.

Trigger and inspect a backup:

```bash
job="cartlabs-backup-$(date -u +%Y%m%d%H%M%S)"
kubectl -n cartlabs create job "$job" --from=cronjob/cartlabs-backup
kubectl -n cartlabs wait --for=condition=Complete "job/$job" --timeout=10m
kubectl -n cartlabs logs "job/$job" --all-containers=true
```

Restore procedure for demo incidents:

1. Disable delivery and reset dispatch; take one final backup when possible.
2. Download selected dump from external storage and verify object size/date.
3. Scale API and worker Deployments to zero, leaving PostgreSQL running.
4. Terminate sessions, drop and recreate `cartlabs`, then run `pg_restore
   --no-owner --no-privileges` as the `cartlabs` user.
5. Run table-count and migration-version checks before scaling workloads up.
6. Wait for readiness, run Helm and external smoke tests, then re-enable jobs.
7. Record selected object, timestamps, commands, and verification in incident
   notes. Retain failed database or server snapshot until review completes.

Test restores into a separate temporary database before relying on a new backup
configuration. Milestone verification restored the uploaded dump separately and
confirmed both seeded products without altering the live database.

## Limits and failure model

Single node means host, disk, control plane, and workloads share one failure
domain. Maintenance causes downtime; no autoscaling or zone redundancy exists.
Bundled PostgreSQL, Redis, and RabbitMQ serve demonstration workloads only.
Provider rebuild plus Terraform, Ansible, Helm, and external backup is the
disaster-recovery path. Move stateful dependencies off-node and add multiple
nodes before treating this design as production.
