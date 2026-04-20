resource "azurerm_resource_group" "this" {
  name     = "${var.cluster_name}-rg"
  location = var.location

  tags = {
    Name        = "${var.cluster_name}-rg"
    Environment = var.environment
  }
}
