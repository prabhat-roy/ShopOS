resource "azurerm_public_ip" "jenkins" {
  name                = "${var.name}-pip"
  location            = azurerm_resource_group.jenkins.location
  resource_group_name = azurerm_resource_group.jenkins.name
  allocation_method   = "Static"
  sku                 = "Standard"

  tags = {
    Name        = "${var.name}-pip"
    Environment = var.environment
  }
}
