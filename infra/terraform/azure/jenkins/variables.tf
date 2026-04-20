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

variable "name" {
  type    = string
  default = "jenkins"
}

variable "address_space" {
  type    = string
  default = "10.0.0.0/16"
}

variable "subnet_prefix" {
  type    = string
  default = "10.0.1.0/24"
}

variable "vm_size" {
  type    = string
  default = "Standard_D4s_v3"
}

variable "disk_size_gb" {
  type    = number
  default = 200
}

variable "admin_username" {
  type    = string
  default = "ubuntu"
}

variable "ssh_pub_key_path" {
  type        = string
  description = "Path to the SSH public key file for VM access"
}

variable "private_key_path" {
  type        = string
  description = "Path to the SSH private key file used by the provisioner"
}
