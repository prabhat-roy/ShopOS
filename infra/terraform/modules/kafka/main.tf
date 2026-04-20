variable "namespace" {
  type    = string
  default = "shopos-infra"
}

variable "kafka_version" {
  type    = string
  default = "26.8.5"
}

variable "replica_count" {
  type    = number
  default = 3
}

terraform {
  required_providers {
    helm = {
      source  = "hashicorp/helm"
      version = "~> 2.12"
    }
  }
}

resource "helm_release" "kafka" {
  name             = "kafka"
  repository       = "https://charts.bitnami.com/bitnami"
  chart            = "kafka"
  version          = var.kafka_version
  namespace        = var.namespace
  create_namespace = false

  values = [
    yamlencode({
      replicaCount = var.replica_count
      metrics = {
        kafka = {
          enabled = true
        }
      }
      zookeeper = {
        enabled = false
      }
      kraft = {
        enabled = true
      }
    })
  ]
}
