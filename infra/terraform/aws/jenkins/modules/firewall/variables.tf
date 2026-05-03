variable "name" {
  type = string
}

variable "environment" {
  type    = string
  default = "dev"
}

variable "vpc_id" {
  type = string
}

variable "ssh_source_cidr" {
  type = string
}

variable "ui_source_cidr" {
  type    = string
  default = "0.0.0.0/0"
}
