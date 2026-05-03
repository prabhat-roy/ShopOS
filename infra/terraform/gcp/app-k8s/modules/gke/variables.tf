variable "name" {
  type = string
}

variable "region" {
  type = string
}

variable "project_id" {
  type = string
}

variable "kubernetes_version" {
  type    = string
  default = null
}

variable "network_name" {
  type = string
}

variable "subnet_name" {
  type = string
}

variable "master_cidr" {
  type = string
}
