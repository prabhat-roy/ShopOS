resource "azurerm_public_ip" "this" {
  name                = "${var.name}-public-ip"
  location            = var.location
  resource_group_name = var.resource_group_name
  allocation_method   = "Static"
  sku                 = "Standard"

  tags = {
    Name        = "${var.name}-public-ip"
    Environment = var.environment
  }
}
