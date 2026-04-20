terraform {
  required_version = ">= 1.8.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
  backend "s3" {
    bucket = "shopos-opentofu-state"
    key    = "aws/eks/terraform.tfstate"
    region = "us-east-1"
  }
}

provider "aws" {
  region = var.aws_region
}

# Mirror of infra/terraform/aws/eks — managed by OpenTofu
# OpenTofu is API-compatible with Terraform; this file uses identical HCL.
# Swap in OpenTofu-specific features (e.g. provider-defined functions) here as needed.

module "eks" {
  source  = "terraform-aws-modules/eks/aws"
  version = "~> 20.0"

  cluster_name                   = var.cluster_name
  cluster_version                = "1.31"
  cluster_endpoint_public_access = true

  vpc_id     = var.vpc_id
  subnet_ids = var.subnet_ids

  cluster_compute_config = {
    enabled    = true
    node_pools = ["general-purpose", "system"]
  }

  bootstrap_self_managed_addons = false

  tags = {
    Environment = var.environment
    ManagedBy   = "opentofu"
    Project     = "shopos"
  }
}
