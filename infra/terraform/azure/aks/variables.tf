variable "subscription_id" {
  type        = string
  description = "Azure subscription ID"
}

variable "location" {
  type    = string
  default = "East US"
}

variable "environment" {
  type    = string
  default = "dev"
}

variable "cluster_name" {
  type    = string
  default = "shopos-aks"
}

variable "kubernetes_version" {
  type        = string
  default     = null
  description = "AKS Kubernetes version. null = latest stable."
}

variable "address_space" {
  type    = string
  default = "10.1.0.0/16"
}

variable "node_subnet_prefix" {
  type    = string
  default = "10.1.0.0/20"
}

variable "service_cidr" {
  type    = string
  default = "10.100.0.0/16"
}

variable "dns_service_ip" {
  type    = string
  default = "10.100.0.10"
}

variable "system_node_count" {
  type    = number
  default = 1
}

variable "system_vm_size" {
  type    = string
  default = "Standard_D4s_v3"
}

variable "availability_zones" {
  type    = list(string)
  default = ["1", "2", "3"]
}
