# Shared database module — provisions managed Postgres via Helm (for k8s-local)

variable "namespace" {
  type    = string
  default = "shopos-infra"
}

variable "postgres_version" {
  type    = string
  default = "16.2.0"
}

variable "storage_size" {
  type    = string
  default = "100Gi"
}

terraform {
  required_providers {
    helm = {
      source  = "hashicorp/helm"
      version = "~> 2.12"
    }
  }
}

resource "helm_release" "postgres" {
  name             = "postgres"
  repository       = "https://charts.bitnami.com/bitnami"
  chart            = "postgresql"
  version          = var.postgres_version
  namespace        = var.namespace
  create_namespace = false

  values = [
    yamlencode({
      primary = {
        persistence = {
          size = var.storage_size
        }
      }
      metrics = {
        enabled = true
      }
    })
  ]
}

resource "helm_release" "redis" {
  name       = "redis"
  repository = "https://charts.bitnami.com/bitnami"
  chart      = "redis"
  namespace  = var.namespace

  values = [
    yamlencode({
      architecture = "replication"
      metrics = {
        enabled = true
      }
    })
  ]
}

resource "helm_release" "mongodb" {
  name       = "mongodb"
  repository = "https://charts.bitnami.com/bitnami"
  chart      = "mongodb"
  namespace  = var.namespace

  values = [
    yamlencode({
      persistence = {
        size = "50Gi"
      }
    })
  ]
}
