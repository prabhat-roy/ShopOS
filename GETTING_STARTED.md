# Getting Started — ShopOS

This guide takes you from zero to a running ShopOS environment. It covers prerequisites,
local development, Kubernetes cluster setup, CI/CD pipeline configuration, and day-two
operations. Read it top-to-bottom the first time; use it as a reference after that.

---

## Table of Contents

1. [Project at a Glance](#1-project-at-a-glance)
2. [Repository Layout](#2-repository-layout)
3. [Prerequisites](#3-prerequisites)
4. [Local Development (Docker Compose)](#4-local-development-docker-compose)
5. [Local Kubernetes (kind / minikube)](#5-local-kubernetes-kind--minikube)
6. [Hot-Reload Dev (Skaffold / Tilt)](#6-hot-reload-dev-skaffold--tilt)
7. [Provisioning a Cloud Cluster](#7-provisioning-a-cloud-cluster)
8. [Cluster Bootstrap (6 phases)](#8-cluster-bootstrap-6-phases)
9. [CI/CD Pipelines](#9-cicd-pipelines)
10. [Deploying a Service](#10-deploying-a-service)
11. [GitOps with ArgoCD](#11-gitops-with-argocd)
12. [Observability](#12-observability)
13. [Security Layer](#13-security-layer)
14. [Messaging Infrastructure](#14-messaging-infrastructure)
15. [Working with Proto / gRPC](#15-working-with-proto--grpc)
16. [Working with Kafka Events](#16-working-with-kafka-events)
17. [Adding a New Service](#17-adding-a-new-service)
18. [Common Make Targets](#18-common-make-targets)
19. [Environment Variables Reference](#19-environment-variables-reference)
20. [Troubleshooting](#20-troubleshooting)

---

## 1. Project at a Glance

ShopOS is an enterprise-grade, cloud-native commerce platform built entirely with open source
technology. It is a reference implementation — not a toy — demonstrating production patterns
at scale:

| Dimension | Number |
|---|---|
| Business domains | 13 |
| Microservices | 154 |
| Programming languages | 8 |
| CI/CD platforms | 15 |
| Kubernetes namespaces | 19 |
| Kafka topics | 20 |
| gRPC proto files | 58 |
| Helm charts | 184 (154 services + 30 tools) |
| Security tools configured | 50+ |
| Observability tools | 35+ |

**Design patterns in use:** DDD, CQRS, Event Sourcing, Saga (orchestration), BFF,
API Gateway, Strangler Fig, Outbox, Sidecar.

---

## 2. Repository Layout

```
ShopOS/
├── GETTING_STARTED.md          ← You are here
├── README.md                   ← Project overview and technology catalogue
├── Makefile                    ← Top-level build and operational commands
├── docker-compose.yml          ← Full local stack (154 services + all infra)
├── docker-compose.override.yml ← Local dev overrides (bind mounts, debug ports)
├── skaffold.yaml               ← Skaffold hot-reload config
├── Tiltfile                    ← Tilt hot-reload config
├── .env.example                ← All environment variables documented
├── .devcontainer/              ← VS Code / Codespaces dev container
│
├── src/                        ← All 154 microservices (13 domains)
│   ├── platform/               ← 22 services: api-gateway, BFFs, saga-orchestrator …
│   ├── identity/               ← 8 services: auth, user, session, MFA …
│   ├── catalog/                ← 12 services: products, pricing, inventory, search …
│   ├── commerce/               ← 23 services: cart, checkout, order, payment …
│   ├── supply-chain/           ← 13 services: vendor, warehouse, fulfillment …
│   ├── financial/              ← 11 services: invoice, accounting, payout …
│   ├── customer-experience/    ← 14 services: reviews, wishlist, support …
│   ├── communications/         ← 9 services: notifications, email, SMS …
│   ├── content/                ← 8 services: media, CMS, i18n …
│   ├── analytics-ai/           ← 13 services: ML, recommendations, analytics …
│   ├── b2b/                    ← 7 services: organisations, contracts, quotes …
│   ├── integrations/           ← 10 services: ERP, CRM, marketplace connectors …
│   └── affiliate/              ← 4 services: affiliate, referral, influencer …
│
├── proto/                      ← gRPC .proto files (58 files, 14 domains)
├── events/                     ← Kafka Avro schemas (20 event types)
│
├── ci/github-actions/          ← GitHub Actions (15th platform — 12 workflows, auto-trigger disabled)
│
├── ci/                         ← 14 other CI platforms × 12 pipelines = 168 files
│   ├── jenkins/                ← 12 Jenkinsfiles (deploy, security, networking …)
│   ├── drone/                  ← 12 Drone YAML (*.drone.yml)
│   ├── woodpecker/             ← 12 Woodpecker YAML (*.woodpecker.yml)
│   ├── dagger/                 ← 12 Go SDK pipelines (one subdirectory each)
│   ├── tekton/                 ← 12 Tekton CRD YAML (*-pipeline.yml)
│   ├── concourse/              ← 12 Concourse YAML (*-pipeline.yml)
│   ├── gitlab-ci/              ← 12 GitLab CI YAML (*.gitlab-ci.yml)
│   ├── circleci/               ← 12 CircleCI YAML (*.circleci.yml)
│   ├── gocd/                   ← 12 GoCD YAML (*.gocd.yml)
│   ├── travis/                 ← 12 Travis CI YAML (*.travis.yml)
│   ├── harness/                ← 12 Harness YAML (*-pipeline.yml)
│   ├── azure-devops/           ← 12 Azure Pipelines YAML
│   ├── aws-codepipeline/       ← 12 CodeBuild buildspecs (buildspec-*.yml)
│   └── gcp-cloudbuild/         ← 12 Cloud Build YAML (cloudbuild-*.yaml)
│
├── helm/
│   └── charts/                 ← 154 per-service Helm charts
│
├── gitops/                     ← GitOps configurations
│   ├── argocd/                 ← App-of-Apps + ApplicationSets
│   ├── flux/                   ← Flux base + cluster overlays
│   ├── argo-rollouts/          ← Canary rollout templates
│   ├── argo-workflows/         ← CI build + ML training workflows
│   ├── argo-events/            ← GitHub EventSource + Sensors
│   └── charts/                 ← Helm charts for 12 GitOps tools
│
├── infra/
│   ├── terraform/              ← EKS (aws/eks), GKE (gcp/gke), AKS (azure/aks), Jenkins VM
│   ├── opentofu/               ← Same targets as Terraform (open source alternative)
│   ├── crossplane/             ← Kubernetes-native IaC (compositions, claims)
│   └── ansible/                ← Kubernetes node bootstrapping roles
│
├── kubernetes/                 ← Raw K8s manifests
│   ├── namespaces/             ← 19 Namespace definitions
│   ├── rbac/                   ← ServiceAccounts, Roles, Bindings
│   ├── network-policies/       ← Default-deny + allow rules
│   ├── resource-quotas/        ← ResourceQuota + LimitRange per namespace
│   ├── pod-disruption-budgets/ ← PDBs for critical services
│   ├── keda/                   ← KEDA ScaledObjects (Kafka + Redis)
│   └── velero/                 ← Backup schedules
│
├── observability/              ← 35+ observability tool configs
├── security/                   ← 50+ security tool configs
├── messaging/                  ← Kafka, RabbitMQ, NATS configs
├── networking/                 ← Istio, Cilium, Traefik configs
├── registry/                   ← Harbor, MinIO, Nexus charts
│
├── databases/                  ← Specialist DB schemas
│   ├── clickhouse/             ← OLAP schema
│   ├── weaviate/               ← Vector schema
│   ├── neo4j/                  ← Graph schema
│   ├── scylladb/               ← Time-series keyspace
│   └── opensearch/             ← Index templates + ILM
│
├── streaming/                  ← CDC + stream processing
│   ├── debezium/               ← Postgres + MongoDB CDC connectors
│   └── flink/                  ← FlinkDeployment CRDs
│
├── workflow/temporal/          ← Temporal server config + workflow mapping
├── ml/mlflow/                  ← MLflow config
├── backstage/                  ← Developer portal app-config + catalog-info
│
├── chaos/                      ← Chaos engineering
│   ├── chaos-mesh/             ← 13 experiments + 2 workflows + 1 game-day schedule
│   └── litmus/                 ← 5 ChaosEngine + 2 Argo Workflow runs
│
├── load-testing/               ← Load tests
│   ├── k6/                     ← 6 scripts (smoke, browse, checkout, spike, soak …)
│   ├── locust/                 ← 3 user classes + 4 task sets
│   └── gatling/                ← Commerce + Search simulations
│
└── docs/
    ├── architecture/           ← 5 design documents
    ├── runbooks/               ← Incident response, failover, rollback
    └── adr/                    ← 6 Architecture Decision Records
```

---

## 3. Prerequisites

### Required on your workstation

| Tool | Min version | Install |
|---|---|---|
| Docker Desktop (or Docker Engine) | 25+ | [docs.docker.com](https://docs.docker.com/get-docker/) |
| Docker Compose v2 | 2.24+ | bundled with Docker Desktop |
| kubectl | 1.29+ | `brew install kubectl` |
| Helm | 3.14+ | `brew install helm` |
| Go | 1.23+ | [go.dev/dl](https://go.dev/dl/) |
| Git | any | pre-installed on most systems |

### Optional (needed for specific tasks)

| Tool | Purpose | Install |
|---|---|---|
| kind | Local Kubernetes | `brew install kind` |
| minikube | Local Kubernetes (alternative) | `brew install minikube` |
| Skaffold | Hot-reload dev | `brew install skaffold` |
| Tilt | Hot-reload dev (alternative) | `brew install tilt` |
| Terraform | Cloud cluster provisioning | `brew install terraform` |
| Buf CLI | Protobuf codegen | `brew install bufbuild/buf/buf` |
| Dagger CLI | Portable CI pipelines | `curl -L https://dl.dagger.io/dagger/install.sh | sh` |
| k9s | Kubernetes terminal UI | `brew install k9s` |
| grpcurl | gRPC API testing | `brew install grpcurl` |
| cosign | Image signing | `brew install cosign` |

### VS Code / Codespaces

The `.devcontainer/` directory provides a fully configured dev container with all 8 language
runtimes, CLIs, and VS Code extensions pre-installed. Open the repository in VS Code and
click "Reopen in Container" when prompted.

---

## 4. Local Development (Docker Compose)

The fastest way to run the full platform locally.

### First-time setup

```bash
# 1. Clone the repository
git clone https://github.com/prabhat-roy/ShopOS.git
cd ShopOS

# 2. Copy environment file
cp .env.example .env
# Edit .env — set passwords, ports, registry URLs as needed

# 3. Start the full stack (all 154 services + all infra)
docker compose up -d

# 4. Check health
docker compose ps
curl http://localhost:8080/healthz   # api-gateway
```

### Starting specific domains

```bash
# Infra only (databases, messaging)
docker compose up -d postgres mongodb redis kafka zookeeper rabbitmq nats

# Platform + Identity
docker compose up -d api-gateway web-bff identity

# Commerce domain
docker compose up -d commerce catalog

# Stop everything
docker compose down

# Stop and remove volumes (clean slate)
docker compose down -v
```

### Accessing services locally

| Service | URL |
|---|---|
| API Gateway | http://localhost:8080 |
| Web BFF | http://localhost:8081 |
| Admin Portal | http://localhost:8085 |
| GraphQL Gateway | http://localhost:8086 |
| Grafana | http://localhost:3000 (admin / admin) |
| Prometheus | http://localhost:9090 |
| Jaeger UI | http://localhost:16686 |
| ArgoCD UI | http://localhost:8088 |
| Harbor | http://localhost:5000 |
| AKHQ (Kafka UI) | http://localhost:8084 |
| RabbitMQ Management | http://localhost:15672 |

---

## 5. Local Kubernetes (kind / minikube)

For testing Kubernetes-specific features like KEDA, Helm charts, and network policies.

### kind

```bash
# Create a cluster
kind create cluster --name shopos --config=.devcontainer/kind-config.yaml

# Load a local image into the cluster
kind load docker-image shopos/order-service:local --name shopos

# Point kubectl
kubectl cluster-info --context kind-shopos
```

### minikube

```bash
minikube start --cpus=8 --memory=16g --driver=docker
eval $(minikube docker-env)   # use minikube's docker daemon
```

### Deploy with Helm

```bash
# Apply base Kubernetes manifests first
kubectl apply -f kubernetes/namespaces/
kubectl apply -f kubernetes/rbac/

# Deploy a single service
helm upgrade --install order-service helm/charts/order-service \
  --namespace commerce --create-namespace \
  --set image.repository=localhost:5000/shopos/order-service \
  --set image.tag=local

# Deploy all services (slow — ~154 helm releases)
make deploy-local
```

---

## 6. Hot-Reload Dev (Skaffold / Tilt)

### Skaffold

```bash
# Hot-reload a single service
skaffold dev --module=order-service

# Hot-reload all platform services
skaffold dev --module=platform

# One-shot build + deploy
skaffold run
```

### Tilt

```bash
# Start Tilt (core services defined in Tiltfile)
tilt up

# Target specific services
tilt up -- --services=order-service,payment-service

# Tilt dashboard: http://localhost:10350
```

---

## 7. Provisioning a Cloud Cluster

ShopOS supports AWS EKS, GCP GKE, and Azure AKS via Terraform (or OpenTofu).

### AWS EKS

```bash
cd infra/terraform/aws/eks

# Export credentials
export AWS_ACCESS_KEY_ID=...
export AWS_SECRET_ACCESS_KEY=...
export AWS_REGION=us-east-1

terraform init
terraform plan
terraform apply -auto-approve

# Update kubeconfig
aws eks update-kubeconfig --name shopos-cluster --region us-east-1
kubectl get nodes
```

### GCP GKE

```bash
cd infra/terraform/gcp/gke

gcloud auth application-default login
export GOOGLE_PROJECT=your-project-id

terraform init && terraform apply -auto-approve

gcloud container clusters get-credentials shopos-cluster --region us-central1
```

### Azure AKS

```bash
cd infra/terraform/azure/aks

az login
export ARM_SUBSCRIPTION_ID=...
export ARM_TENANT_ID=...
export ARM_CLIENT_ID=...
export ARM_CLIENT_SECRET=...

terraform init && terraform apply -auto-approve

az aks get-credentials --resource-group shopos-rg --name shopos-cluster
```

### Via CI Pipeline

Use the `k8s-infra` pipeline on any CI platform. Set `ACTION=apply` to create,
`ACTION=destroy` to tear down. Required secrets: cloud credentials + `KUBECONFIG_CONTENT`.

---

## 8. Cluster Bootstrap (6 phases)

After provisioning a bare cluster, run the bootstrap pipeline to install all platform
components in order. Each phase depends on the previous.

### Manual bootstrap (one phase at a time)

```bash
# Export base64-encoded kubeconfig
export KUBECONFIG_CONTENT=$(cat ~/.kube/config | base64 -w 0)

# Phase 1: Networking
cd ci/dagger/networking
dagger run go run main.go

# Phase 2: Security
cd ci/dagger/security
KEYCLOAK_ADMIN_PASSWORD=... KUBECONFIG_CONTENT=$KUBECONFIG_CONTENT \
dagger run go run main.go

# Phase 3: Observability
cd ci/dagger/observability
GRAFANA_ADMIN_PASSWORD=... MINIO_SECRET_KEY=... \
dagger run go run main.go

# Phase 4: Messaging
cd ci/dagger/messaging
RABBITMQ_PASSWORD=... CREATE_TOPICS=true \
dagger run go run main.go

# Phase 5: Registry
cd ci/dagger/registry
HARBOR_ADMIN_PASSWORD=... MINIO_ROOT_USER=... MINIO_ROOT_PASSWORD=... \
dagger run go run main.go

# Phase 6: GitOps
cd ci/dagger/gitops
ARGOCD_ADMIN_PASSWORD=... MINIO_ACCESS_KEY=... MINIO_SECRET_KEY=... \
dagger run go run main.go
```

### Full bootstrap via cluster-bootstrap pipeline

```bash
cd ci/dagger/cluster-bootstrap
export KUBECONFIG_CONTENT=$(cat ~/.kube/config | base64 -w 0)
export KEYCLOAK_ADMIN_PASSWORD=...
export GRAFANA_ADMIN_PASSWORD=...
export ARGOCD_ADMIN_PASSWORD=...
export HARBOR_ADMIN_PASSWORD=...
export MINIO_ROOT_USER=minio
export MINIO_ROOT_PASSWORD=...
export RABBITMQ_PASSWORD=...
export MINIO_SECRET_KEY=...
# Optionally resume from a phase:
export START_PHASE=3
export END_PHASE=6

dagger run go run main.go
```

### What gets installed

| Phase | Duration | Components |
|---|---|---|
| 1 — Networking | ~20 min | Cilium CNI, Istio (base + istiod + gateway), Traefik |
| 2 — Security | ~25 min | cert-manager, Vault HA, Keycloak, Kyverno, Falco |
| 3 — Observability | ~30 min | Prometheus stack, Grafana, Loki, Jaeger, OTel Collector |
| 4 — Messaging | ~20 min | ZooKeeper, Kafka (3 brokers), RabbitMQ, NATS, 20 Kafka topics |
| 5 — Registry | ~25 min | MinIO (4 nodes), Harbor (with Trivy), Nexus, 8 MinIO buckets |
| 6 — GitOps | ~20 min | ArgoCD, Argo Rollouts, Argo Workflows, KEDA, Velero |

---

## 9. CI/CD Pipelines

ShopOS ships 12 pipeline definitions across 15 CI platforms. Pick the one matching your
CI server and configure the required secrets.

### Required secrets (all platforms)

| Secret | Description |
|---|---|
| `KUBECONFIG_CONTENT` | Base64-encoded kubeconfig for the target cluster |
| `HARBOR_REGISTRY` | Harbor hostname (e.g., `harbor.shopos.internal`) |
| `HARBOR_USERNAME` | Harbor robot account username |
| `HARBOR_PASSWORD` | Harbor robot account password |
| `SONAR_TOKEN` | SonarQube authentication token |
| `SONAR_HOST_URL` | SonarQube server URL |
| `SLACK_WEBHOOK` | Incoming webhook URL for Slack notifications |

### Jenkins

1. Install Jenkins with the Pipeline + Docker + Kubernetes plugins
2. Add credentials in Jenkins Credentials Store
3. Create a Pipeline job pointing to `ci/jenkins/Jenkinsfile`
4. Trigger manually or set up a webhook from your SCM

```bash
# Validate a Jenkinsfile locally
docker run --rm -v $(pwd):/workspace \
  jenkins/jenkins:lts \
  java -jar /var/jenkins_home/war/WEB-INF/lib/cli.jar -s http://jenkins:8080 \
  declarative-linter < ci/jenkins/Jenkinsfile
```

### Drone CI / Woodpecker CI

1. Connect your Git repository to Drone/Woodpecker
2. Add secrets via the UI or CLI
3. Pipeline triggers automatically on push/PR (`.drone.yml`) or manually for infra pipelines

```bash
# Run locally with Drone CLI
drone exec ci/drone/deploy.drone.yml \
  --secret SERVICE_NAME=order-service \
  --secret IMAGE_TAG=v1.5.0
```

### Dagger (run anywhere)

Dagger pipelines are plain Go programs — no CI server needed.

```bash
cd ci/dagger/deploy
go run main.go   # (with all env vars set)

# Or via the Dagger CLI
dagger run go run main.go
```

### GitHub Actions

Workflow files live at `ci/github-actions/` (NOT `.github/workflows/`), so they do **not**
auto-trigger on push or PR. To activate them, copy the files into `.github/workflows/` and
add the required secrets in repository Settings → Secrets and variables → Actions.

```bash
# Activate GitHub Actions
mkdir -p .github/workflows && cp ci/github-actions/*.yml .github/workflows/
```

```bash
# List workflows
gh workflow list

# Trigger a workflow manually
gh workflow run deploy.yml \
  -f SERVICE_NAME=order-service \
  -f IMAGE_TAG=v1.5.0 \
  -f ENVIRONMENT=staging

# View run status
gh run list --workflow=deploy.yml
```

### GitLab CI

Copy files from `ci/gitlab-ci/` to your GitLab repository root.
Secrets are set in Settings → CI/CD → Variables.

---

## 10. Deploying a Service

### Manual Helm deploy

```bash
helm upgrade --install order-service helm/charts/order-service \
  --namespace commerce \
  --set image.repository=harbor.shopos.internal/shopos/order-service \
  --set image.tag=v1.5.0 \
  --set environment=production \
  --wait --timeout 5m
```

### Via the deploy pipeline

Set these pipeline parameters / environment variables:

```
SERVICE_NAME=order-service
IMAGE_TAG=v1.5.0
ENVIRONMENT=production
K8S_NAMESPACE=commerce
```

Then trigger the `deploy` pipeline on your CI platform of choice.

### Verify deployment

```bash
kubectl rollout status deployment/order-service -n commerce
kubectl get pods -n commerce -l app=order-service
curl http://order-service.commerce.svc/healthz
```

### Rollback

```bash
# Helm rollback to previous revision
helm rollback order-service -n commerce

# Or via ArgoCD
argocd app rollback order-service
```

---

## 11. GitOps with ArgoCD

ArgoCD watches the `gitops/argocd/` directory and automatically reconciles the cluster
state to match Git.

### Access ArgoCD

```bash
# Port-forward
kubectl port-forward svc/argocd-server -n argocd 8088:443

# Get initial admin password
kubectl get secret argocd-initial-admin-secret -n argocd \
  -o jsonpath="{.data.password}" | base64 -d
```

Open https://localhost:8088 — login with `admin` and the password above.

### App-of-Apps pattern

The root application in `gitops/argocd/app-of-apps.yaml` creates one ArgoCD Application
per domain. Each domain application manages all services within that domain.

```bash
# Bootstrap — apply root app
kubectl apply -f gitops/argocd/app-of-apps.yaml -n argocd

# Sync all apps
argocd app sync --all

# Sync a specific domain
argocd app sync commerce
```

### Canary deployments (Argo Rollouts)

Services with a `Rollout` manifest (in `gitops/argo-rollouts/`) use progressive delivery:
20% canary → automated metric check → 100% promote or rollback.

```bash
# Check rollout status
kubectl argo rollouts get rollout order-service -n commerce --watch

# Manually promote to 100%
kubectl argo rollouts promote order-service -n commerce

# Abort rollout
kubectl argo rollouts abort order-service -n commerce
```

---

## 12. Observability

### Grafana dashboards

```bash
kubectl port-forward svc/grafana -n observability 3000:80
# open http://localhost:3000 (admin / <GRAFANA_ADMIN_PASSWORD>)
```

Pre-built dashboards:
- **ShopOS Overview** — cross-domain request rates and error budgets
- **Service Health** — per-service latency p50/p95/p99, error rate, saturation
- **Kafka Lag** — consumer group lag per topic
- **SLO Dashboard** — error budget burn rates (Pyrra)
- **Cost** — per-namespace cost (OpenCost)
- **Infrastructure** — node CPU, memory, disk (kube-state-metrics + node-exporter)

### Traces (Jaeger)

```bash
kubectl port-forward svc/jaeger-query -n observability 16686:16686
# open http://localhost:16686
```

### Logs (Loki / Grafana)

In Grafana → Explore → Select datasource "Loki":

```logql
# All errors in the commerce namespace
{namespace="commerce"} |= "error"

# Specific service
{namespace="commerce", app="order-service"} | json | level="error"

# Kafka consumer lag warnings
{namespace="messaging"} |= "consumer lag"
```

### Alerts

Alert rules are in `observability/prometheus/`. Alerts route to Slack via Alertmanager.
Critical alerts page on-call via PagerDuty integration (configure in
`observability/alertmanager/alertmanager.yaml`).

---

## 13. Security Layer

### Vault — Secrets Management

```bash
kubectl port-forward svc/vault -n security 8200:8200
export VAULT_ADDR=http://localhost:8200

# Unseal (first time after install)
vault operator init
vault operator unseal <unseal-key>

# Read a secret
vault kv get secret/shopos/order-service
```

Services fetch secrets at startup via the Vault Agent Injector sidecar (configured in
each Helm chart's `values.yaml`).

### Keycloak — Identity / SSO

```bash
kubectl port-forward svc/keycloak -n security 8443:443
# open https://localhost:8443 — admin / <KEYCLOAK_ADMIN_PASSWORD>
```

Realm `shopos` is pre-configured with:
- OIDC clients for each BFF
- Roles: `admin`, `merchant`, `customer`, `support`
- LDAP federation (configure your directory in realm settings)

### Checking policy compliance

```bash
# Kyverno policy reports
kubectl get policyreport -A

# OPA decisions (if Gatekeeper installed)
kubectl get constrainttemplate

# Falco alerts
kubectl logs -l app=falco -n falco --tail=50
```

### Image signing verification

```bash
cosign verify \
  --certificate-identity=ci@shopos.internal \
  --certificate-oidc-issuer=https://token.actions.githubusercontent.com \
  harbor.shopos.internal/shopos/order-service:v1.5.0
```

---

## 14. Messaging Infrastructure

### Kafka

```bash
# Port-forward Kafka UI (AKHQ)
kubectl port-forward svc/akhq -n messaging 8084:80
# open http://localhost:8084

# Produce a test message via kubectl exec
KAFKA_POD=$(kubectl get pods -n messaging -l app.kubernetes.io/name=kafka -o jsonpath='{.items[0].metadata.name}')
kubectl exec -n messaging "$KAFKA_POD" -- \
  kafka-console-producer.sh --bootstrap-server kafka.messaging.svc:9092 \
  --topic commerce.order.placed
```

### List all topics

```bash
kubectl exec -n messaging "$KAFKA_POD" -- \
  kafka-topics.sh --list --bootstrap-server kafka.messaging.svc:9092
```

### RabbitMQ

```bash
kubectl port-forward svc/rabbitmq -n messaging 15672:15672
# open http://localhost:15672 — admin / <RABBITMQ_PASSWORD>
```

### NATS JetStream

```bash
# Install NATS CLI
brew install nats-io/nats-tools/nats

# Port-forward
kubectl port-forward svc/nats -n messaging 4222:4222

# Subscribe to a subject
nats sub "notifications.>"
```

---

## 15. Working with Proto / gRPC

All service contracts live in `proto/`. Use [Buf CLI](https://buf.build/docs/) for
linting, breaking-change detection, and code generation.

```bash
cd proto

# Lint all protos
buf lint

# Check for breaking changes against main
buf breaking --against '.git#branch=main'

# Generate code for all languages
buf generate

# Generate for a specific service
buf generate --path commerce/order.proto
```

Generated code output (per `buf.gen.yaml`):
- Go: `src/{domain}/{service}/internal/gen/`
- Java/Kotlin: `src/{domain}/{service}/src/main/java/`
- Python: `src/{domain}/{service}/gen/`
- Node.js: `src/{domain}/{service}/src/gen/`

### Testing a gRPC endpoint

```bash
# List services
grpcurl -plaintext order-service.commerce.svc:50082 list

# Call a method
grpcurl -plaintext \
  -d '{"order_id": "ord-123"}' \
  order-service.commerce.svc:50082 \
  commerce.OrderService/GetOrder
```

---

## 16. Working with Kafka Events

All Kafka events use Avro schemas defined in `events/`. The Schema Registry enforces
these schemas at produce time.

### Schema format

```json
// events/commerce.order.placed.avsc
{
  "type": "record",
  "name": "OrderPlaced",
  "namespace": "com.shopos.commerce",
  "fields": [
    {"name": "order_id", "type": "string"},
    {"name": "customer_id", "type": "string"},
    {"name": "total_amount", "type": "double"},
    {"name": "currency", "type": "string"},
    {"name": "placed_at", "type": "long", "logicalType": "timestamp-millis"}
  ]
}
```

### Registering schemas

```bash
# Register a schema
curl -X POST http://schema-registry.messaging.svc:8081/subjects/commerce.order.placed-value/versions \
  -H 'Content-Type: application/vnd.schemaregistry.v1+json' \
  -d "{\"schema\": $(cat events/commerce.order.placed.avsc | jq -Rs .)}"

# List registered subjects
curl http://schema-registry.messaging.svc:8081/subjects
```

### Topic naming convention

`{domain}.{entity}.{event}` — e.g., `commerce.order.placed`, `identity.user.registered`

All 20 topics: see [events/README.md](events/README.md).

---

## 17. Adding a New Service

Follow these steps to add a new microservice.

### 1. Choose the domain and assign a port

Refer to the port ranges in [src/README.md](src/README.md). Pick the next available port
in your domain's range and add it to the service registry in [README.md](README.md).

### 2. Create the service directory

```bash
mkdir -p src/{domain}/{service-name}
```

### 3. Scaffold the service

Use the language matching the domain's convention (see [src/README.md](src/README.md)).

```bash
# Go
cd src/platform/my-new-service
go mod init github.com/shopos/my-new-service
# Add main.go with /healthz endpoint and gRPC health check
```

Required files for every service:
- `main.go` / `index.js` / `main.py` / `Application.java` etc.
- `Dockerfile` — multi-stage build, non-root user
- `Makefile` — build, test, lint, run targets
- `.env.example` — all environment variables
- `README.md` — service description

### 4. Add the proto file

```bash
# Create proto definition
cat > proto/{domain}/my-service.proto << 'EOF'
syntax = "proto3";
package {domain};
option go_package = "github.com/shopos/{domain}/my-service";

service MyService {
  rpc GetItem (GetItemRequest) returns (GetItemResponse);
}
...
EOF

# Generate code
cd proto && buf generate
```

### 5. Add the Helm chart

```bash
cp -r helm/charts/api-gateway helm/charts/my-new-service
# Edit Chart.yaml, values.yaml, templates/deployment.yaml
```

### 6. Add to Docker Compose

Add a service block to `docker-compose.yml` following the pattern of existing services
in the same domain.

### 7. Register in ArgoCD / GitOps

Add the service to the domain's ApplicationSet in `gitops/argocd/`.

---

## 18. Common Make Targets

```bash
make help                     # list all targets

# Build
make build SERVICE=order-service    # build single service image
make build-all                      # build all 154 images

# Test
make test SERVICE=order-service     # test single service
make test-all                       # test all services

# Push
make push SERVICE=order-service IMAGE_TAG=v1.5.0
make push-all IMAGE_TAG=v1.5.0

# Proto
make proto-generate                 # buf generate all
make proto-lint                     # buf lint

# Local dev
make compose-up                     # docker compose up -d
make compose-down                   # docker compose down
make dev SERVICE=order-service      # skaffold dev --module=

# Kubernetes
make k8s-bootstrap                  # apply all kubernetes/ manifests
make deploy SERVICE=order-service   # helm upgrade --install

# Cleanup
make clean                          # remove build artefacts
make prune                          # docker system prune
```

---

## 19. Environment Variables Reference

Copy `.env.example` to `.env` and populate. Key variables:

| Variable | Description | Example |
|---|---|---|
| `HARBOR_REGISTRY` | Harbor hostname | `harbor.shopos.internal` |
| `HARBOR_USERNAME` | Harbor robot account | `robot$ci` |
| `HARBOR_PASSWORD` | Harbor robot password | (secret) |
| `POSTGRES_PASSWORD` | PostgreSQL superuser password | (secret) |
| `MONGODB_PASSWORD` | MongoDB root password | (secret) |
| `REDIS_PASSWORD` | Redis auth password | (secret) |
| `RABBITMQ_PASSWORD` | RabbitMQ admin password | (secret) |
| `GRAFANA_ADMIN_PASSWORD` | Grafana admin password | (secret) |
| `ARGOCD_ADMIN_PASSWORD` | ArgoCD admin password | (secret) |
| `KEYCLOAK_ADMIN_PASSWORD` | Keycloak admin password | (secret) |
| `VAULT_ADDR` | Vault server address | `http://vault.security.svc:8200` |
| `SONAR_TOKEN` | SonarQube auth token | (secret) |
| `SONAR_HOST_URL` | SonarQube URL | `http://sonar.shopos.internal` |
| `SLACK_WEBHOOK` | Slack incoming webhook | `https://hooks.slack.com/...` |
| `KUBECONFIG_CONTENT` | Base64-encoded kubeconfig | (secret) |
| `MINIO_ROOT_USER` | MinIO root user | `minio` |
| `MINIO_ROOT_PASSWORD` | MinIO root password | (secret) |
| `MINIO_SECRET_KEY` | MinIO secret access key | (secret) |

---

## 20. Troubleshooting

### Docker Compose — service won't start

```bash
docker compose logs order-service --tail=50
docker compose ps       # check health status
docker compose restart order-service
```

### Pod stuck in CrashLoopBackOff

```bash
kubectl describe pod <pod-name> -n commerce
kubectl logs <pod-name> -n commerce --previous
# Check environment variables and secrets
kubectl get secret <secret-name> -n commerce -o yaml
```

### Helm release failed

```bash
helm status order-service -n commerce
helm history order-service -n commerce
helm rollback order-service -n commerce   # roll back to previous
```

### Kafka consumer lag is growing

```bash
# Check lag
kubectl exec -n messaging $KAFKA_POD -- \
  kafka-consumer-groups.sh --bootstrap-server kafka.messaging.svc:9092 \
  --describe --group order-service-group

# Check KEDA ScaledObject
kubectl get scaledobject order-service -n commerce
kubectl describe hpa keda-hpa-order-service -n commerce
```

### ArgoCD app is OutOfSync

```bash
argocd app get commerce
argocd app diff commerce
argocd app sync commerce --force
```

### Vault sealed after restart

```bash
kubectl exec -it vault-0 -n security -- vault operator unseal <key-1>
kubectl exec -it vault-0 -n security -- vault operator unseal <key-2>
kubectl exec -it vault-0 -n security -- vault operator unseal <key-3>
```

### cert-manager certificate not issuing

```bash
kubectl describe certificate <cert-name> -n <namespace>
kubectl describe certificaterequest -n <namespace>
kubectl logs -l app=cert-manager -n cert-manager --tail=50
```

---

## Further Reading

| Topic | Document |
|---|---|
| Architecture overview | [docs/architecture/system-overview.md](docs/architecture/system-overview.md) |
| Domain boundaries | [docs/architecture/domain-map.md](docs/architecture/domain-map.md) |
| Communication patterns | [docs/architecture/communication-patterns.md](docs/architecture/communication-patterns.md) |
| Database strategy | [docs/architecture/database-strategy.md](docs/architecture/database-strategy.md) |
| Security model | [docs/architecture/security-model.md](docs/architecture/security-model.md) |
| ADRs | [docs/adr/](docs/adr/) |
| CI/CD pipelines | [ci/README.md](ci/README.md) |
| Helm charts | [helm/README.md](helm/README.md) |
| Infrastructure as Code | [infra/README.md](infra/README.md) |
| GitOps | [gitops/README.md](gitops/README.md) |
| Observability | [observability/README.md](observability/README.md) |
| Security configs | [security/README.md](security/README.md) |
| Kubernetes manifests | [kubernetes/README.md](kubernetes/README.md) |
| Messaging | [messaging/README.md](messaging/README.md) |
| Streaming / CDC | [streaming/README.md](streaming/README.md) |
| Databases | [databases/README.md](databases/README.md) |
| Chaos Engineering | [chaos/README.md](chaos/README.md) |
| Load Testing | [load-testing/README.md](load-testing/README.md) |
| Temporal Workflows | [workflow/README.md](workflow/README.md) |
| Proto / gRPC | [proto/README.md](proto/README.md) |
| Kafka event schemas | [events/README.md](events/README.md) |
| Backstage portal | [backstage/README.md](backstage/README.md) |
| Service catalogue | [src/README.md](src/README.md) |
