# Terraform — All Cloud Infrastructure (AWS + GCP + Azure in Parallel)

**Scope: All three clouds — AWS, GCP, and Azure — run in parallel, not sequentially.**

Terraform manages all cloud infrastructure. AWS, GCP, and Azure workspaces are planned and applied
concurrently using Atlantis. Jenkins server (master + agents) is provisioned by Terraform — it is
never created manually or from a laptop. All infrastructure changes go through Git → Atlantis.

## Why Terraform (not OpenTofu) for AWS Production?

- HashiCorp BSL licence is acceptable for internal infrastructure automation
- Mature AWS provider (>1000 resources), battle-tested in production
- Atlantis GitOps integration — plan on PR, apply on merge with audit trail
- Infracost cost estimation on every PR
- Driftctl weekly drift detection

## Directory Structure

```
terraform/
├── aws/
│   ├── eks/            ← EKS Auto Mode cluster (production)
│   │   ├── main.tf     ← Cluster, node groups, IRSA, OIDC provider
│   │   ├── vpc.tf      ← VPC, subnets, NAT gateways, route tables
│   │   ├── rds.tf      ← PostgreSQL 16 RDS Multi-AZ (all domains)
│   │   ├── elasticache.tf ← Redis 7 ElastiCache (sessions, cache)
│   │   ├── msk.tf      ← Amazon MSK (Kafka 3.7 — production only)
│   │   ├── s3.tf       ← S3 buckets (MinIO alternative for prod)
│   │   ├── acm.tf      ← TLS certificates via ACM
│   │   ├── route53.tf  ← DNS zones and records
│   │   ├── waf.tf      ← AWS WAF v2 (backup to Cloudflare WAF)
│   │   └── variables.tf
│   └── jenkins/        ← Jenkins master EC2 + EBS + EIP
│       ├── main.tf
│       ├── sg.tf       ← Security groups
│       └── iam.tf      ← IAM roles for Jenkins agents
└── modules/
    ├── k8s-cluster/    ← Reusable EKS module
    ├── databases/      ← RDS + ElastiCache module
    ├── kafka/          ← MSK module
    └── networking/     ← VPC + subnets module
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
Developer opens PR → Atlantis runs terraform plan → Posts plan as PR comment
Tech lead approves PR → Atlantis runs terraform apply → Updates state in S3
Driftctl runs weekly → Reports drift between state and actual AWS resources
Infracost runs on PR → Reports cost delta (must be < $500/month per PR)
```

## Usage

```bash
cd infra/terraform/aws/eks
terraform init -backend-config=../../backend.hcl
terraform plan -var-file=../../prod.tfvars
# Apply only via Atlantis (not manually in production)
```
