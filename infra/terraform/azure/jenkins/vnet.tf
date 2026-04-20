resource "azurerm_virtual_network" "jenkins" {
  name                = "${var.name}-vnet"
  location            = azurerm_resource_group.jenkins.location
  resource_group_name = azurerm_resource_group.jenkins.name
  address_space       = [var.address_space]

  tags = {
    Name        = "${var.name}-vnet"
    Environment = var.environment
  }
}

resource "azurerm_subnet" "jenkins" {
  name                 = "${var.name}-subnet"
  resource_group_name  = azurerm_resource_group.jenkins.name
  virtual_network_name = azurerm_virtual_network.jenkins.name
  address_prefixes     = [var.subnet_prefix]
}
