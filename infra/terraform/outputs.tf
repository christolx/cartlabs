output "server_ipv4" {
  description = "Public IPv4 address used for DNS and Ansible inventory."
  value       = hcloud_server.k3s.ipv4_address
}

output "server_ipv6" {
  description = "Public IPv6 address used for optional AAAA DNS."
  value       = hcloud_server.k3s.ipv6_address
}

output "ansible_inventory" {
  description = "Minimal inventory fragment. Store outside version control."
  value = yamlencode({
    all = {
      children = {
        k3s = {
          hosts = {
            cartlabs-demo = {
              ansible_host = hcloud_server.k3s.ipv4_address
              ansible_user = "root"
            }
          }
        }
      }
    }
  })
}
