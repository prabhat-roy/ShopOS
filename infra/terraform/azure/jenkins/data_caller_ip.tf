data "http" "caller_ip" {
  url = "https://ipv4.icanhazip.com"
}

locals {
  caller_cidr = "${chomp(data.http.caller_ip.response_body)}/32"

  default_ssh_pub_key_path = coalesce(
    var.ssh_pub_key_path,
    fileexists(pathexpand("~/.ssh/id_ed25519.pub")) ? pathexpand("~/.ssh/id_ed25519.pub") : null,
    fileexists(pathexpand("~/.ssh/id_rsa.pub")) ? pathexpand("~/.ssh/id_rsa.pub") : null,
  )
  ssh_pub_key = chomp(file(local.default_ssh_pub_key_path))
}
