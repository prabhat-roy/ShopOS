module "eks" {
  source = "./modules/eks"

  name               = var.name
  environment        = var.environment
  kubernetes_version = var.kubernetes_version
  subnet_ids         = concat(module.network.public_subnet_ids, module.network.private_subnet_ids)
  security_group_ids = [module.firewall.cluster_security_group_id]
  cluster_role_arn   = module.iam.cluster_role_arn
  node_role_arn      = module.iam.node_role_arn
}
