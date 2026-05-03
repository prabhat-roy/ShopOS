variable "name" {
  type    = string
  default = "jenkins"
}

variable "environment" {
  type    = string
  default = "dev"
}

variable "region" {
  type    = string
  default = "us-east-1"
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
  default = "t3.xlarge"
}

variable "disk_size_gb" {
  type    = number
  default = 200
}

variable "key_name" {
  type        = string
  description = "Existing AWS EC2 key-pair name registered in this region"
}

variable "ui_source_cidr" {
  type    = string
  default = "0.0.0.0/0"
}
