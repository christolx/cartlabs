variable "hcloud_token" {
  description = "Hetzner Cloud API token. Supply through TF_VAR_hcloud_token."
  type        = string
  sensitive   = true
}

variable "project_name" {
  description = "Prefix for provisioned resources."
  type        = string
  default     = "cartlabs-demo"

  validation {
    condition     = can(regex("^[a-z0-9-]{3,40}$", var.project_name))
    error_message = "project_name must contain 3-40 lowercase letters, digits, or hyphens."
  }
}

variable "location" {
  description = "Hetzner location."
  type        = string
  default     = "nbg1"
}

variable "server_type" {
  description = "Single-node demo server class."
  type        = string
  default     = "cx23"
}

variable "server_image" {
  description = "Pinned operating-system family."
  type        = string
  default     = "ubuntu-24.04"
}

variable "ssh_public_key" {
  description = "Operator SSH public key."
  type        = string
  sensitive   = true

  validation {
    condition     = can(regex("^(ssh-ed25519|sk-ssh-ed25519@openssh.com) ", var.ssh_public_key))
    error_message = "Use an Ed25519 SSH public key."
  }
}

variable "operator_cidrs" {
  description = "CIDRs allowed to reach SSH and Kubernetes API. Never use a global CIDR."
  type        = list(string)

  validation {
    condition = length(var.operator_cidrs) > 0 && alltrue([
      for cidr in var.operator_cidrs : can(cidrnetmask(cidr)) && cidr != "0.0.0.0/0" && cidr != "::/0"
    ])
    error_message = "operator_cidrs must contain valid, non-global CIDRs."
  }
}

variable "enable_provider_backups" {
  description = "Enable Hetzner weekly backups in addition to database backups."
  type        = bool
  default     = true
}
