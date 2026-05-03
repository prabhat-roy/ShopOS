variable "name" {
  type = string
}

variable "environment" {
  type    = string
  default = "dev"
}

variable "region" {
  type = string
}

variable "vpc_cidr" {
  type = string
}

variable "subnet_cidr" {
  type = string
}
