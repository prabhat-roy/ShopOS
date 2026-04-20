terraform {
  backend "azurerm" {
    # resource_group_name, storage_account_name, container_name, key, access_key
    # are passed via -backend-config in k8s-tf-init.groovy
    # resource_group_name  = "shopos-tfstate-rg"
    # storage_account_name = "shoposterraformstate"
    # container_name       = "tfstate"
    # key                  = "aks/terraform.tfstate"
  }
}
