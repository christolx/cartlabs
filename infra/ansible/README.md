# Debian home-server configuration

Ansible configures Debian 12+ laptop with checksum-verified k3s, pinned Helm 4,
operator kubectl access, embedded-etcd snapshots, host hardening, Traefik
forwarded-header trust, optional cloudflared, and optional protected deployment
runner.

```bash
cp inventory.example.yml inventory.yml
cp group_vars/all.example.yml group_vars/all.yml
export CLOUDFLARE_TUNNEL_TOKEN='...'
ansible-playbook -i inventory.yml playbook.yml
```

Set `cloudflared_enabled: true` only after remotely managed tunnel maps
`cartlabs.christofle.dev` to `http://localhost:80`. Token is written root-only to
`/etc/cloudflared/token`; never store it in vars files or Git.

Runner defaults off. To install first time, export one-hour registration token,
set `github_runner_enabled: true`, rerun playbook. Route only protected deploy
workflows to label `cartlabs-demo`; never untrusted pull requests.

Inventory and real group vars remain ignored. Verify SSH host key out of band.
Keep Kubernetes API and Traefik ports off WAN firewall.
