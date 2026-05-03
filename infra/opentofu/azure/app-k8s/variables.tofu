variable "subscription_id" {
  type = string
}

variable "cluster_name" {
  type    = string
  default = "shopos-aks-dev"
}

variable "location" {
  type    = string
  default = "East US"
}

variable "environment" {
  type    = string
  default = "dev"
}

variable "kubernetes_version" {
  type    = string
  default = null
}

variable "node_count" {
  type    = number
  default = 1
}

variable "vm_size" {
  type    = string
  default = "Standard_D4s_v3"
}

variable "vnet_cidr" {
  type    = string
  default = "10.30.0.0/16"
}

variable "subnet_cidr" {
  type    = string
  default = "10.30.0.0/20"
}

variable "service_cidr" {
  type    = string
  default = "10.200.0.0/16"
}

variable "dns_service_ip" {
  type    = string
  default = "10.200.0.10"
}
