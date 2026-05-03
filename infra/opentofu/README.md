# OpenTofu — Non-Production K8s Clusters (dev / staging)

OpenTofu is the OSI-licensed Terraform fork (Linux Foundation). In this project it
provisions **dev and staging app-k8s clusters across all three clouds**. Production
goes through Terraform.

## Layout

```
opentofu/
├── aws/app-k8s/      ← Dev/staging EKS (VPC + subnets + EKS Auto Mode)
│   ├── main.tofu
│   ├── variables.tofu
│   └── terraform.tfvars.example
├── gcp/app-k8s/      ← Dev/staging GKE Autopilot (VPC + subnet + secondary ranges + GKE)
│   ├── main.tofu
│   ├── variables.tofu
│   └── terraform.tfvars.example
└── azure/app-k8s/    ← Dev/staging AKS (RG + VNet + subnet + AKS w/ Cilium overlay)
    ├── main.tofu
    ├── variables.tofu
    └── terraform.tfvars.example
```

Each workload is **self-contained**: VPC + cluster + outputs in one apply, no external
VPC dependency. Mirrors the Terraform per-workload pattern but keeps `.tofu` extension
and OpenTofu-specific backend conventions.

## Who runs what

OpenTofu is **always** run from a Jenkins pipeline, never from a laptop.
[k8s-infra.Jenkinsfile](../../ci/jenkins/k8s-infra.Jenkinsfile) chooses Terraform vs
OpenTofu via the `IaC_TOOL` parameter:

| `IaC_TOOL` | `ENVIRONMENT` | Workload dir | Why |
|---|---|---|---|
| `terraform` | `prod` | [`infra/terraform/<cloud>/app-k8s/`](../terraform/) | Single source of truth for production. |
| `opentofu` | `dev`, `staging` | [`infra/opentofu/<cloud>/app-k8s/`](.) | OSI-licensed, no HashiCorp BSL exposure on non-prod. |

The same groovy script ([k8s-detect-cloud.groovy](../../scripts/groovy/k8s-detect-cloud.groovy))
maps cloud + IaC_TOOL → workload dir, and ([k8s-tf-init.groovy](../../scripts/groovy/k8s-tf-init.groovy))
runs `tofu init` instead of `terraform init` when `IaC_TOOL=opentofu`.

## Local invocation (manual override only)

```bash
# AWS dev EKS
cd infra/opentofu/aws/app-k8s
cp terraform.tfvars.example terraform.tfvars
tofu init
tofu apply
$(tofu output -raw kubeconfig_command)

# GCP dev GKE
cd infra/opentofu/gcp/app-k8s
# Edit terraform.tfvars: set project_id
tofu init && tofu apply

# Azure dev AKS
cd infra/opentofu/azure/app-k8s
# Edit terraform.tfvars: set subscription_id
tofu init && tofu apply
```

## State backends

| Cloud | Backend | Bucket / RG |
|---|---|---|
| AWS | `s3` | `shopos-tfstate` (key prefix `opentofu/aws/app-k8s/`) |
| GCP | `gcs` | `shopos-tfstate` (prefix `opentofu/gcp/app-k8s`) |
| Azure | `azurerm` | RG `shopos-tfstate-rg`, account `shoposterraformstate`, key `opentofu/azure/app-k8s/...` |

State buckets are auto-created on first `tofu init` by [k8s-tf-init.groovy](../../scripts/groovy/k8s-tf-init.groovy).

## What OpenTofu does NOT manage

| Resource | Managed by |
|---|---|
| Production app-k8s | Terraform (`infra/terraform/<cloud>/app-k8s/`) |
| Jenkins servers | Terraform from laptop (`infra/terraform/<cloud>/jenkins/`) |
| K8s workloads | ArgoCD / Flux (GitOps) |
| K8s-native infra (DBs/queues claimed by apps) | Crossplane (`infra/crossplane/`) |
| OS config / node bootstrap | Ansible (`infra/ansible/`) |
| Immutable VM images | Packer (`infra/packer/`) |
