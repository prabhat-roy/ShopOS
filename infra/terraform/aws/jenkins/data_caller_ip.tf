data "http" "caller_ip" {
  url = "https://ipv4.icanhazip.com"
}

locals {
  caller_cidr = "${chomp(data.http.caller_ip.response_body)}/32"
}
