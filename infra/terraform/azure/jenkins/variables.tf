variable "subscription_id" {
  type        = string
  description = "Azure subscription ID"
}

variable "name" {
  type    = string
  default = "jenkins"
}

variable "environment" {
  type    = string
  default = "dev"
}

variable "region" {
  type        = string
  default     = "East US"
  description = "Azure location"
}

variable "vpc_cidr" {
  type    = string
  default = "10.0.0.0/16"
}

variable "subnet_cidr" {
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

variable "ui_source_cidr" {
  type    = string
  default = "0.0.0.0/0"
}

variable "ssh_pub_key_path" {
  type        = string
  default     = null
  description = "Optional override; defaults to ~/.ssh/id_ed25519.pub or id_rsa.pub"
}
