resource "azurerm_resource_group" "this" {
  name     = "${var.name}-rg"
  location = var.region

  tags = {
    Name        = "${var.name}-rg"
    Environment = var.environment
  }
}
