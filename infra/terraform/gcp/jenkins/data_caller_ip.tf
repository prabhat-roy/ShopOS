data "http" "caller_ip" {
  url = "https://ipv4.icanhazip.com"
}

locals {
  caller_cidr = "${chomp(data.http.caller_ip.response_body)}/32"

  # Auto-detect default OpenSSH key (same convention as ml-security):
  # ~/.ssh/id_ed25519.pub (preferred) → ~/.ssh/id_rsa.pub.
  # Result: `ssh ubuntu@<ip>` from any cmd / shell works with no -i flag.
  default_ssh_pub_key_path = coalesce(
    var.ssh_pub_key_path,
    fileexists(pathexpand("~/.ssh/id_ed25519.pub")) ? pathexpand("~/.ssh/id_ed25519.pub") : null,
    fileexists(pathexpand("~/.ssh/id_rsa.pub")) ? pathexpand("~/.ssh/id_rsa.pub") : null,
  )
  ssh_pub_key = chomp(file(local.default_ssh_pub_key_path))
}
