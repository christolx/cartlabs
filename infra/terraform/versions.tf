terraform {
  required_version = ">= 1.10.0"

  required_providers {
    hcloud = {
      source  = "hetznercloud/hcloud"
      version = "1.68.0"
    }
  }

  backend "s3" {}
}

provider "hcloud" {
  token = var.hcloud_token
}
