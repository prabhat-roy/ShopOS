resource "azurerm_resource_group" "jenkins" {
  name     = "${var.name}-rg"
  location = var.location

  tags = {
    Name        = "${var.name}-rg"
    Environment = var.environment
  }
}
