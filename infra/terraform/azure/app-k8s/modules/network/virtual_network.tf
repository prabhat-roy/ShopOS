resource "azurerm_virtual_network" "this" {
  name                = "${var.name}-vpc"
  location            = azurerm_resource_group.this.location
  resource_group_name = azurerm_resource_group.this.name
  address_space       = [var.vpc_cidr]

  tags = {
    Name        = "${var.name}-vpc"
    Environment = var.environment
  }
}
