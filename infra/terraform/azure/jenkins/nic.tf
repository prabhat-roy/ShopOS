resource "azurerm_network_interface" "jenkins" {
  name                = "${var.name}-nic"
  location            = azurerm_resource_group.jenkins.location
  resource_group_name = azurerm_resource_group.jenkins.name

  ip_configuration {
    name                          = "internal"
    subnet_id                     = azurerm_subnet.jenkins.id
    private_ip_address_allocation = "Dynamic"
    public_ip_address_id          = azurerm_public_ip.jenkins.id
  }

  tags = {
    Name        = "${var.name}-nic"
    Environment = var.environment
  }
}
