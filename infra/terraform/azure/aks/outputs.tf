output "cluster_name" {
  value = azurerm_kubernetes_cluster.this.name
}

output "cluster_endpoint" {
  value = azurerm_kubernetes_cluster.this.kube_config[0].host
}

output "cluster_version" {
  value = azurerm_kubernetes_cluster.this.kubernetes_version
}

output "cluster_certificate_authority" {
  value     = azurerm_kubernetes_cluster.this.kube_config[0].cluster_ca_certificate
  sensitive = true
}

output "resource_group_name" {
  value = azurerm_resource_group.this.name
}

output "vnet_id" {
  value = azurerm_virtual_network.this.id
}

output "node_subnet_id" {
  value = azurerm_subnet.nodes.id
}

output "kubeconfig_command" {
  value = "az aks get-credentials --resource-group ${azurerm_resource_group.this.name} --name ${var.cluster_name} --overwrite-existing"
}
