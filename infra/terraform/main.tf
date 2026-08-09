locals {
  common_labels = {
    project     = "cartlabs"
    environment = "demo"
    managed_by  = "terraform"
  }
}

resource "hcloud_ssh_key" "operator" {
  name       = "${var.project_name}-operator"
  public_key = var.ssh_public_key
  labels     = local.common_labels
}

resource "hcloud_firewall" "k3s" {
  name   = "${var.project_name}-firewall"
  labels = local.common_labels

  rule {
    direction   = "in"
    protocol    = "tcp"
    port        = "22"
    source_ips  = var.operator_cidrs
    description = "Operator SSH"
  }

  rule {
    direction   = "in"
    protocol    = "tcp"
    port        = "6443"
    source_ips  = var.operator_cidrs
    description = "Kubernetes API"
  }

  rule {
    direction   = "in"
    protocol    = "tcp"
    port        = "80"
    source_ips  = ["0.0.0.0/0", "::/0"]
    description = "HTTP certificate challenge and redirect"
  }

  rule {
    direction   = "in"
    protocol    = "tcp"
    port        = "443"
    source_ips  = ["0.0.0.0/0", "::/0"]
    description = "HTTPS"
  }

  rule {
    direction   = "in"
    protocol    = "icmp"
    source_ips  = ["0.0.0.0/0", "::/0"]
    description = "Path MTU and reachability"
  }
}

resource "hcloud_server" "k3s" {
  name         = var.project_name
  image        = var.server_image
  server_type  = var.server_type
  location     = var.location
  ssh_keys     = [hcloud_ssh_key.operator.id]
  firewall_ids = [hcloud_firewall.k3s.id]
  backups      = var.enable_provider_backups
  labels       = local.common_labels
  user_data    = templatefile("${path.module}/templates/cloud-init.yaml.tftpl", {})

  public_net {
    ipv4_enabled = true
    ipv6_enabled = true
  }

  lifecycle {
    prevent_destroy = true
  }
}
