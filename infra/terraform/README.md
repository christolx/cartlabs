# Hetzner Cloud infrastructure

Terraform provisions one Ubuntu 24.04 k3s host, Ed25519 operator key, provider
firewall, and optional Hetzner whole-server backups. `prevent_destroy` protects
the demo node; intentionally remove it in a reviewed change before teardown.

Use encrypted, versioned S3-compatible remote state with locking:

```bash
cp terraform.tfvars.example terraform.tfvars
cp backend.example.hcl backend.hcl
export TF_VAR_hcloud_token='...'
export AWS_ACCESS_KEY_ID='...'
export AWS_SECRET_ACCESS_KEY='...'
terraform init -backend-config=backend.hcl
terraform plan
terraform apply
```

Never commit `terraform.tfvars`, `backend.hcl`, state, API tokens, backend credentials, or
generated Ansible inventory. Create DNS A/AAAA records from outputs before
running application deployment.
