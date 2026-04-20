# Shared networking module for ShopOS VPC creation

variable "environment" {
  type = string
}

variable "cidr_block" {
  type    = string
  default = "10.0.0.0/16"
}

variable "region" {
  type = string
}

output "vpc_cidr" {
  value = var.cidr_block
}

output "environment" {
  value = var.environment
}

# Provider-specific networking resources are defined in the cloud-specific modules.
# This module provides shared variable and output contracts.
