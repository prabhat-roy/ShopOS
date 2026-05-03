# Infrastructure — ShopOS

All cloud + on-prem infrastructure: IaC, control-plane bootstrap, HA databases, connection
pooling, alternative caches, search engines, GitOps for Terraform, image baking, and
non-K8s workloads (Nomad).

---

## Layout

```
infra/
├── terraform/         Primary IaC: EKS / GKE / AKS modules + Jenkins on each cloud
├── opentofu/          OpenTofu mirror (OSI-licensed alternative; same module interfaces)
├── crossplane/        K8s-native IaC (developer self-service: claim a database/queue/bucket)
├── ansible/           K8s node bootstrap roles (common, docker, k8s-node, jenkins)
├── packer/            VM image automation (Jenkins agent AMI, K8s node base AMI, GCE images)
├── nomad/             HashiCorp Nomad — non-containerized + batch workloads (legacy + ML training)
├── atlantis/          Terraform GitOps — plan on PR comment, apply on merge (production AWS only)
├── patroni/           HA PostgreSQL (3-node primary/replica with etcd failover)
├── pgbouncer/         PostgreSQL connection pooler (transaction mode, 2000 max client conns)
├── dragonfly/         Redis-compatible multi-threaded store (4× throughput vs Redis)
└── meilisearch/       Fast typo-tolerant product search engine
```

---

## Who runs what — laptop vs Jenkins

The chicken-and-egg rule: **the Jenkins server itself is provisioned from a laptop**.
Once Jenkins is up, **everything else** is provisioned by Jenkins pipelines.

| Layer | Run from | Tool | Pipeline / path |
|---|---|---|---|
| Jenkins server (3 clouds) | **Laptop** | Terraform | [`terraform/aws/jenkins/`](terraform/aws/jenkins/), [`terraform/gcp/jenkins/`](terraform/gcp/jenkins/), [`terraform/azure/jenkins/`](terraform/azure/jenkins/) |
| K8s clusters (prod) | Jenkins | Terraform | [`ci/jenkins/k8s-infra.Jenkinsfile`](../ci/jenkins/k8s-infra.Jenkinsfile) → [`terraform/<cloud>/app-k8s/`](terraform/) |
| K8s clusters (dev/staging) | Jenkins | OpenTofu | same pipeline, IaC_TOOL=opentofu → [`opentofu/<cloud>/app-k8s/`](opentofu/) |
| Node bootstrap | Jenkins | Ansible | [`scripts/groovy/run-ansible.groovy`](../scripts/groovy/run-ansible.groovy) (RUN_ANSIBLE=true on k8s-infra) |
| Immutable AMI/GCE images | Jenkins | Packer | [`ci/jenkins/infra-images.Jenkinsfile`](../ci/jenkins/infra-images.Jenkinsfile) → [`packer/jenkins-agent/`](packer/jenkins-agent/), [`packer/k8s-node-base/`](packer/k8s-node-base/) |
| Cost / drift / lint | Jenkins | Infracost + Driftctl + tflint + Atlantis-validate + inframap | [`ci/jenkins/infra-quality.Jenkinsfile`](../ci/jenkins/infra-quality.Jenkinsfile) (weekly cron + manual) |
| Runtime self-service | Jenkins | Crossplane | [`ci/jenkins/cluster-bootstrap.Jenkinsfile`](../ci/jenkins/cluster-bootstrap.Jenkinsfile) → claims/compositions in [`crossplane/`](crossplane/) |
| Atlantis (Terraform GitOps) | Jenkins | Atlantis | [`ci/jenkins/tooling.Jenkinsfile`](../ci/jenkins/tooling.Jenkinsfile) installs Atlantis from [`atlantis/`](atlantis/) |
| Non-K8s workloads | Jenkins | Nomad + Waypoint + Boundary | [`ci/jenkins/tooling.Jenkinsfile`](../ci/jenkins/tooling.Jenkinsfile) |
| Data-tier HA | Jenkins | Patroni + PgBouncer + Dragonfly + Meilisearch | [`ci/jenkins/databases.Jenkinsfile`](../ci/jenkins/databases.Jenkinsfile) (also installs CockroachDB / YugabyteDB / SurrealDB / EventStore / Valkey / Typesense / Manticore / SeaweedFS) |

## Tool responsibility (non-overlapping scope)

| Tool | Scope |
|---|---|
| Terraform | All three clouds — production. Single source of truth. Used from laptop for Jenkins, from Jenkins for app-k8s. |
| OpenTofu | Non-prod (dev / staging) ephemeral environments. Same module shape, OSI-licensed. |
| Crossplane | Runtime developer self-service in K8s (claim DB / queue / bucket). |
| Ansible | Post-provisioning OS configuration — hardening, K8s node bootstrap, STIG, also runs as a Packer provisioner. |
| Packer | Bake immutable VM/AMI base images (Jenkins agent + K8s node). Speed up cold starts. |
| Nomad | Non-containerised workloads + batch (ML training, long-running pipelines, periodic tasks). |
| Atlantis | GitOps for Terraform — plan on PR comment, apply on merge. AWS prod only. |
| Infracost | Cost delta on every Terraform PR (blocks > $500/month increase). |
| Driftctl | Weekly scheduled scan: Terraform state vs actual cloud resources. |
| Waypoint | App deployment abstraction — one config for K8s/Nomad/ECS. Used for non-core services. |
| Boundary | Zero-trust SSH/RDP without VPN, full session recording. |
| Patroni | PostgreSQL HA — 3-node Raft-managed cluster with etcd failover. |
| PgBouncer | Connection pooler — transaction mode, in front of Patroni primary. |
| Dragonfly | Redis-compatible cache when single-threaded Redis is the bottleneck. |
| Meilisearch | Fast product search alternative to Elasticsearch. |

---

## Kubernetes clusters

| Cloud | Service | Mode | Path |
|---|---|---|---|
| AWS | EKS | Auto Mode (managed nodes) | [`terraform/aws/app-k8s/`](terraform/aws/app-k8s/) |
| GCP | GKE | Autopilot | [`terraform/gcp/app-k8s/`](terraform/gcp/app-k8s/) |
| Azure | AKS | Node Auto Provisioning | [`terraform/azure/app-k8s/`](terraform/azure/app-k8s/) |

All three are private clusters across 3 AZs with Workload Identity / IRSA / Managed Identity
enabled. Terraform.tfvars.example is provided in each module.

```bash
# AWS EKS Auto Mode
cd infra/terraform/aws/app-k8s && cp terraform.tfvars.example terraform.tfvars && terraform init && terraform apply
aws eks update-kubeconfig --region us-east-1 --name shopos-eks

# GCP GKE Autopilot
cd infra/terraform/gcp/app-k8s && terraform apply
gcloud container clusters get-credentials shopos-gke --region us-central1

# Azure AKS NAP
cd infra/terraform/azure/app-k8s && terraform apply
az aks get-credentials --resource-group shopos-aks-rg --name shopos-aks
```

---

## Jenkins CI server (laptop-provisioned)

Jenkins is **always** provisioned by Terraform from a laptop — never from another Jenkins
job (chicken-and-egg). Cloud-init runs `scripts/bash/jenkins-install.sh` on first boot to
install Jenkins, Docker, kubectl, helm, terraform, and 30+ other tools. The default
credentials are `admin/admin` — change after first login.

| Cloud | Path | Instance | OS |
|---|---|---|---|
| AWS | [`terraform/aws/jenkins/`](terraform/aws/jenkins/) | t3.xlarge | Ubuntu 24.04 |
| GCP | [`terraform/gcp/jenkins/`](terraform/gcp/jenkins/) | n2-standard-4 | Ubuntu 24.04 |
| Azure | [`terraform/azure/jenkins/`](terraform/azure/jenkins/) | Standard_D4s_v3 | Ubuntu 24.04 |

```bash
cd infra/terraform/aws/jenkins
terraform apply       # ~8–12 min including Jenkins install
terraform output jenkins_url
```

---

## High-availability data tier

```bash
# Patroni Postgres HA (3 nodes)
kubectl apply -f infra/patroni/

# PgBouncer in front of Patroni primary
kubectl apply -f infra/pgbouncer/

# Dragonfly (Redis-compatible) for high-throughput cache hot paths
kubectl apply -f infra/dragonfly/

# Meilisearch for product search
kubectl apply -f infra/meilisearch/
```

The application-tier services in [`src/`](../src/) are wired to point at:
- `postgres-primary.databases.svc:5432` → PgBouncer → Patroni primary
- `dragonfly.databases.svc:6379` (Redis-compatible)
- `meilisearch.databases.svc:7700`

---

## Atlantis GitOps for Terraform

Atlantis runs in [`atlantis/`](atlantis/). On every PR touching `infra/terraform/aws/`,
Atlantis runs `terraform plan`, posts the diff as a PR comment, and `terraform apply` on
merge. Required for all AWS production changes.

---

## Nomad

[`nomad/`](nomad/) contains:
- `nomad-values.yaml` — Nomad cluster config
- `nomad.hcl` — server config
- `batch-jobs/demand-forecast.nomad` — example batch ML job
- `legacy-services/` — non-containerized legacy workloads

Used when Kubernetes is overkill (one-shot batch jobs, legacy binaries that won't containerize).

---

## Environment parity

| Config | dev | staging | prod |
|---|---|---|---|
| Node pool size | 3 | 6 | 12–30 (Karpenter autoscaled) |
| Postgres | `db-f1-micro` | `db-n1-standard-2` | `db-n1-standard-8` HA via Patroni |
| Redis | Single node | Single node | Dragonfly cluster, 3 replicas |
| Multi-AZ | No | No | Yes (3 AZ) |
| Remote state bucket | `shopos-tf-dev` | `shopos-tf-staging` | `shopos-tf-prod` |
| Atlantis-gated | No | No | Yes |
| Cosign verify on admission | No | Yes (warn) | Yes (block) |

---

## References

- [GitOps (ArgoCD/Flux on top of these clusters)](../gitops/README.md)
- [Helm charts](../helm/README.md)
- [Crossplane database claim example](crossplane/)
