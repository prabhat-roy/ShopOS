resource "azurerm_user_assigned_identity" "jenkins" {
  name                = "${var.name}-identity"
  location            = azurerm_resource_group.jenkins.location
  resource_group_name = azurerm_resource_group.jenkins.name

  tags = {
    Name        = "${var.name}-identity"
    Environment = var.environment
  }
}

resource "azurerm_role_assignment" "jenkins_owner" {
  scope                = "/subscriptions/${var.subscription_id}"
  role_definition_name = "Owner"
  principal_id         = azurerm_user_assigned_identity.jenkins.principal_id
}
