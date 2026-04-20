resource "azurerm_virtual_network" "this" {
  name                = "${var.cluster_name}-vnet"
  location            = azurerm_resource_group.this.location
  resource_group_name = azurerm_resource_group.this.name
  address_space       = [var.address_space]

  tags = {
    Name        = "${var.cluster_name}-vnet"
    Environment = var.environment
  }
}

resource "azurerm_subnet" "nodes" {
  name                 = "${var.cluster_name}-nodes-subnet"
  resource_group_name  = azurerm_resource_group.this.name
  virtual_network_name = azurerm_virtual_network.this.name
  address_prefixes     = [var.node_subnet_prefix]
}
