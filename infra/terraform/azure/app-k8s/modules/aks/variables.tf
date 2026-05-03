variable "name" {
  type = string
}

variable "environment" {
  type    = string
  default = "dev"
}

variable "resource_group_name" {
  type = string
}

variable "location" {
  type = string
}

variable "kubernetes_version" {
  type    = string
  default = null
}

variable "subnet_id" {
  type = string
}

variable "identity_id" {
  type = string
}

variable "service_cidr" {
  type = string
}

variable "dns_service_ip" {
  type = string
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
