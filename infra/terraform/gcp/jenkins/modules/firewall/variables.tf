variable "name" {
  type = string
}

variable "network_name" {
  type = string
}

variable "ssh_source_cidr" {
  type = string
}

variable "ui_source_cidr" {
  type    = string
  default = "0.0.0.0/0"
}
