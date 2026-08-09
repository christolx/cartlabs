# k3s host configuration

Ansible installs a checksum-verified k3s `v1.36.3+k3s1` binary, embedded-etcd
snapshots, host hardening, Traefik (k3s default), cert-manager `v1.21.1`, and
staging/production Let's Encrypt issuers.

```bash
cp inventory.example.yml inventory.yml
cp group_vars/all.example.yml group_vars/all.yml
ansible-playbook -i inventory.yml playbook.yml
```

Inventory and real group variables stay untracked. Verify the host SSH key out
of band before first connection.

Deployment uses a dedicated unprivileged runner on this host so Kubernetes API
port 6443 stays restricted. Create a one-hour repository runner registration
token, then enable runner installation without placing the token in files:

```bash
export GITHUB_RUNNER_REGISTRATION_TOKEN='...'
ansible-playbook -i inventory.yml playbook.yml -e github_runner_enabled=true
```

Runner auto-update is disabled; update its pinned version and official checksum
through review. Never route pull-request jobs to label `cartlabs-demo`.
