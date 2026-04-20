resource "azurerm_user_assigned_identity" "this" {
  name                = "${var.cluster_name}-identity"
  location            = azurerm_resource_group.this.location
  resource_group_name = azurerm_resource_group.this.name

  tags = {
    Name        = "${var.cluster_name}-identity"
    Environment = var.environment
  }
}

resource "azurerm_role_assignment" "network_contributor" {
  scope                = azurerm_virtual_network.this.id
  role_definition_name = "Network Contributor"
  principal_id         = azurerm_user_assigned_identity.this.principal_id
}
