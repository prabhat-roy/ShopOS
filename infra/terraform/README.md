# Terraform â€” All Cloud Infrastructure (AWS + GCP + Azure in Parallel)

Scope: All three clouds â€” AWS, GCP, and Azure â€” run in parallel, not sequentially.

Terraform manages all cloud infrastructure. AWS, GCP, and Azure workspaces are planned and applied
concurrently using Atlantis. Jenkins server (master + agents) is provisioned by Terraform â€” it is
never created manually or from a laptop. All infrastructure changes go through Git â†’ Atlantis.

## Why Terraform (not OpenTofu) for AWS Production?

- HashiCorp BSL licence is acceptable for internal infrastructure automation
- Mature AWS provider (>1000 resources), battle-tested in production
- Atlantis GitOps integration â€” plan on PR, apply on merge with audit trail
- Infracost cost estimation on every PR
- Driftctl weekly drift detection

## Directory Structure

```
terraform/
â”œâ”€â”€ aws/
â”‚   â”œâ”€â”€ eks/            â† EKS Auto Mode cluster (production)
â”‚   â”‚   â”œâ”€â”€ main.tf     â† Cluster, node groups, IRSA, OIDC provider
â”‚   â”‚   â”œâ”€â”€ vpc.tf      â† VPC, subnets, NAT gateways, route tables
â”‚   â”‚   â”œâ”€â”€ rds.tf      â† PostgreSQL 16 RDS Multi-AZ (all domains)
â”‚   â”‚   â”œâ”€â”€ elasticache.tf â† Redis 7 ElastiCache (sessions, cache)
â”‚   â”‚   â”œâ”€â”€ msk.tf      â† Amazon MSK (Kafka 3.7 â€” production only)
â”‚   â”‚   â”œâ”€â”€ s3.tf       â† S3 buckets (MinIO alternative for prod)
â”‚   â”‚   â”œâ”€â”€ acm.tf      â† TLS certificates via ACM
â”‚   â”‚   â”œâ”€â”€ route53.tf  â† DNS zones and records
â”‚   â”‚   â”œâ”€â”€ waf.tf      â† AWS WAF v2 (backup to Cloudflare WAF)
â”‚   â”‚   â””â”€â”€ variables.tf
â”‚   â””â”€â”€ jenkins/        â† Jenkins master EC2 + EBS + EIP
â”‚       â”œâ”€â”€ main.tf
â”‚       â”œâ”€â”€ sg.tf       â† Security groups
â”‚       â””â”€â”€ iam.tf      â† IAM roles for Jenkins agents
â””â”€â”€ modules/
    â”œâ”€â”€ k8s-cluster/    â† Reusable EKS module
    â”œâ”€â”€ databases/      â† RDS + ElastiCache module
    â”œâ”€â”€ kafka/          â† MSK module
    â””â”€â”€ networking/     â† VPC + subnets module
```

## What Terraform Does NOT Manage

| Resource Type | Managed By |
|---|---|
| GCP GKE cluster | OpenTofu (`infra/opentofu/gcp/`) |
| Azure AKS cluster | OpenTofu (`infra/opentofu/azure/`) |
| Dev/staging EKS | OpenTofu (`infra/opentofu/aws/`) |
| K8s workloads | ArgoCD / Flux (GitOps) |
| K8s-native infra (DBs, queues) | Crossplane (`infra/crossplane/`) |
| Server OS config | Ansible (`infra/ansible/`) |
| VM images (AMIs) | Packer (`infra/packer/`) |
| Batch workloads | Nomad (`infra/nomad/`) |

## GitOps Workflow (Atlantis)

```
Developer opens PR â†’ Atlantis runs terraform plan â†’ Posts plan as PR comment
Tech lead approves PR â†’ Atlantis runs terraform apply â†’ Updates state in S3
Driftctl runs weekly â†’ Reports drift between state and actual AWS resources
Infracost runs on PR â†’ Reports cost delta (must be < $500/month per PR)
```

## Usage

```bash
cd infra/terraform/aws/eks
terraform init -backend-config=../../backend.hcl
terraform plan -var-file=../../prod.tfvars
# Apply only via Atlantis (not manually in production)
```
