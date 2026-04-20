terraform {
  required_version = ">= 1.8.0"
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 4.0"
    }
  }
  backend "azurerm" {
    resource_group_name  = "shopos-tfstate"
    storage_account_name = "shoposopentofu"
    container_name       = "tfstate"
    key                  = "azure/aks/terraform.tfstate"
  }
}

provider "azurerm" {
  features {}
}

resource "azurerm_kubernetes_cluster" "shopos" {
  name                = var.cluster_name
  location            = var.location
  resource_group_name = var.resource_group
  dns_prefix          = "shopos"
  kubernetes_version  = "1.31"

  default_node_pool {
    name       = "default"
    node_count = var.node_count
    vm_size    = var.vm_size
  }

  identity {
    type = "SystemAssigned"
  }

  tags = {
    Environment = "prod"
    ManagedBy   = "opentofu"
  }
}
