output "cluster_name" {
  value = module.aks.cluster_name
}

output "cluster_fqdn" {
  value = module.aks.cluster_fqdn
}

output "resource_group_name" {
  value = module.network.resource_group_name
}

output "kubeconfig_command" {
  value = "az aks get-credentials --resource-group ${module.network.resource_group_name} --name ${module.aks.cluster_name}"
}
