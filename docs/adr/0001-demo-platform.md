# 0001 — Single-node k3s demo platform

Status: superseded by the implemented home-server platform in
[`docs/platform.md`](../platform.md)

The original cloud-host decision remains recorded below for historical context.
Cartlabs later moved to an existing Debian laptop with Cloudflare Tunnel to
remove provider cost and public origin ingress while retaining k3s and Helm.

## Context

Cartlabs needs a public, reproducible portfolio environment that exercises
infrastructure provisioning, Kubernetes packaging, secure delivery, TLS,
backups, and recovery. Traffic and availability requirements are low. Operating
cost and cognitive load matter more than horizontal scale.

## Decision

Provision one Hetzner Cloud Ubuntu 24.04 `cx23` host with Terraform. Restrict SSH
and Kubernetes API ingress to operator CIDRs. Configure a pinned k3s release,
embedded etcd snapshots, Traefik, cert-manager, and host hardening with Ansible.

Deploy directly from a protected GitHub environment with Helm on a dedicated,
unprivileged self-hosted runner. Runner connects outbound to GitHub and reaches
Kubernetes through loopback, keeping public API ingress restricted. Hosted
runners build immutable full-SHA GHCR images with SBOM and provenance. Reconcile
runtime secrets from GitHub environment secrets; do not commit encrypted or
plaintext secrets. Use provider-managed DNS records outside this repository and
Let's Encrypt HTTP-01 certificates.

Keep current product media as one immutable local demo asset. Send daily
PostgreSQL custom-format dumps to independent S3-compatible storage with
seven-day retention. Keep provider server backups and k3s snapshots as
additional recovery layers. Cap base platform spend at EUR 15 per month.

## Consequences

- One workflow reproduces host, k3s configuration, and application release.
- Direct Helm avoids operating a GitOps controller for one demo cluster.
- Protected environment approval remains security-critical because CI holds
  cluster and runtime credentials.
- Self-hosted runner must execute only reviewed default-branch deployment jobs;
  Kubernetes cluster-admin access can imply host control on single-node k3s.
- Single node remains a deliberate availability and capacity limit.
- Database recovery is operator-driven and periodically restore-tested.
- Production use would require multi-node control plane, external stateful
  services, stronger secret management, and automated disaster recovery.
