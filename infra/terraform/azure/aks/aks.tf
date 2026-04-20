resource "azurerm_kubernetes_cluster" "this" {
  name                = var.cluster_name
  location            = azurerm_resource_group.this.location
  resource_group_name = azurerm_resource_group.this.name
  kubernetes_version  = var.kubernetes_version
  dns_prefix          = var.cluster_name

  default_node_pool {
    name                = "system"
    vm_size             = var.system_vm_size
    node_count          = var.system_node_count
    vnet_subnet_id      = azurerm_subnet.nodes.id
    zones               = var.availability_zones
    auto_scaling_enabled = true
    min_count           = 1
    max_count           = 10

    upgrade_settings {
      max_surge = "33%"
    }
  }

  identity {
    type         = "UserAssigned"
    identity_ids = [azurerm_user_assigned_identity.this.id]
  }

  network_profile {
    network_plugin = "azure"
    service_cidr   = var.service_cidr
    dns_service_ip = var.dns_service_ip
    load_balancer_sku = "standard"
  }

  # Node Auto Provisioning — AKS equivalent of EKS Auto Mode compute
  node_provisioning_profile {
    mode = "Auto"
  }

  automatic_upgrade_channel = "stable"

  oidc_issuer_enabled       = true
  workload_identity_enabled = true

  tags = {
    Name        = var.cluster_name
    Environment = var.environment
  }

  depends_on = [azurerm_role_assignment.network_contributor]
}
