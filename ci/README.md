# CI Pipelines — ShopOS

ShopOS ships 18+ pipelines implemented across 15 CI/CD platforms. Jenkins is the primary
CI server; the other 14 platforms mirror the same pipeline set for portability and demo purposes.

> **Note**: This is "15 platforms × ~15 pipelines" templated, not 225 unique pipelines.
> Logic is identical across platforms; only syntax differs. The Tool Responsibility Matrix
> in [CLAUDE.md](../CLAUDE.md#cicd-platforms--each-platform-has-a-unique-responsibility)
> describes which platform has which non-overlapping primary responsibility.

---

## Platforms

| Platform | Directory | Files | Notes |
|---|---|---|---|
| Jenkins | `jenkins/` | 18 Jenkinsfiles | Primary CI server; declarative pipeline syntax. Includes [`proto-breaking-check.Jenkinsfile`](jenkins/proto-breaking-check.Jenkinsfile) blocking proto regressions on every PR. |
| Drone CI | `drone/` | 12 YAML | Drone v2; mirrors core Jenkins pipelines |
| Woodpecker CI | `woodpecker/` | 12 YAML | Drone-compatible fork; drop-in replacement |
| Dagger | `dagger/` | 12 Go modules | Portable Go SDK — run on any CI or locally |
| Tekton | `tekton/` | 12 YAML | Kubernetes CRD-native (Task + Pipeline + PipelineRun) |
| Concourse CI | `concourse/` | 12 YAML | Resource/job DAG pipelines |
| GitLab CI | `gitlab-ci/` | 12 YAML | `.gitlab-ci.yml` — native GitLab SCM integration |
| GitHub Actions | `github-actions/` | 13 YAML | Stored in `ci/github-actions/` (auto-trigger disabled). Includes [`proto-breaking.yml`](github-actions/proto-breaking.yml) for buf format/lint/breaking checks. |
| CircleCI | `circleci/` | 12 YAML | `version: 2.1` orb-based pipelines |
| GoCD | `gocd/` | 12 YAML | Stage/job pipelines with manual approval gates |
| Travis CI | `travis/` | 12 YAML | Stage-based pipelines with branch filters |
| Harness CI | `harness/` | 12 YAML | Enterprise CI/CD with built-in CD stages |
| Azure DevOps | `azure-devops/` | 12 YAML | `azure-pipelines.yml` — native Azure integration |
| AWS CodePipeline | `aws-codepipeline/` | 12 YAML | `buildspec.yml` + CodePipeline JSON definitions |
| GCP Cloud Build | `gcp-cloudbuild/` | 12 YAML | `cloudbuild.yaml` — native GCP integration |

---

## Directory Structure

```
ci/
├── jenkins/                          ← 17 Jenkinsfiles (primary)
│   ├── install-tools.Jenkinsfile     ← Bootstrap agent runtimes and CLIs
│   ├── cluster-bootstrap.Jenkinsfile ← Full cluster bring-up (6 phases)
│   ├── k8s-infra.Jenkinsfile         ← Terraform EKS/GKE/AKS provisioning
│   ├── gitops.Jenkinsfile            ← ArgoCD, Flux, Argo Rollouts, KEDA, Velero
│   ├── security.Jenkinsfile          ← Vault, Keycloak, Falco, Kyverno, cert-manager
│   ├── observability.Jenkinsfile     ← Prometheus, Grafana, Loki, Jaeger, OTel
│   ├── messaging.Jenkinsfile         ← Kafka, RabbitMQ, NATS, schema registry
│   ├── networking.Jenkinsfile        ← Istio, Traefik, Cilium, Consul
│   ├── registry.Jenkinsfile          ← Harbor, Nexus + cloud registry provisioning
│   ├── databases.Jenkinsfile         ← Postgres, MongoDB, Redis, Cassandra, ClickHouse
│   ├── streaming.Jenkinsfile         ← Debezium CDC, Apache Flink jobs
│   ├── tooling.Jenkinsfile           ← Developer tools (pgAdmin, Superset, MLflow, etc.)
│   ├── pre-deploy.Jenkinsfile        ← Git fetch → scan → compile → docker build → push
│   ├── deploy.Jenkinsfile            ← GitOps trigger → ArgoCD sync → rollout verify
│   ├── post-deploy.Jenkinsfile       ← Smoke tests → DAST → load tests → SLO validate
│   ├── api-quality.Jenkinsfile       ← Spectral lint → Hurl → Pact → Terrascan
│   └── reports.Jenkinsfile           ← Build/deploy Reports Portal web app
│
├── drone/                            ← Drone CI (same 12 pipelines, *.drone.yml)
├── woodpecker/                       ← Woodpecker CI (same 12 pipelines, *.woodpecker.yml)
├── gitlab-ci/                        ← GitLab CI (same 12 pipelines, *.gitlab-ci.yml)
│
├── github-actions/                   ← GitHub Actions (same 12 pipelines, *.yml)
│   │                                   Stored here (NOT in .github/) — auto-triggering disabled.
│   │                                   To enable: copy files to .github/workflows/ and add secrets.
│   ├── deploy.yml
│   ├── post-deploy.yml
│   ├── k8s-infra.yml
│   ├── gitops.yml
│   ├── security.yml
│   ├── observability.yml
│   ├── messaging.yml
│   ├── networking.yml
│   ├── registry.yml
│   ├── install-tools.yml
│   └── cluster-bootstrap.yml
│
├── dagger/                           ← Dagger Go SDK — one subdirectory per pipeline
│   ├── go.mod / main.go              ← root module (shared utilities)
│   ├── deploy/main.go
│   ├── security/main.go
│   ├── networking/main.go
│   ├── observability/main.go
│   ├── messaging/main.go
│   ├── k8s-infra/main.go
│   ├── gitops/main.go
│   ├── registry/main.go
│   ├── install-tools/main.go
│   ├── cluster-bootstrap/main.go
│   └── post-deploy/main.go
│
├── tekton/                           ← Tekton Pipelines (Kubernetes CRDs)
│   ├── deploy-pipeline.yml
│   ├── security-pipeline.yml
│   ├── networking-pipeline.yml
│   ├── observability-pipeline.yml
│   ├── messaging-pipeline.yml
│   ├── k8s-infra-pipeline.yml
│   ├── gitops-pipeline.yml
│   ├── registry-pipeline.yml
│   ├── install-tools-pipeline.yml
│   ├── cluster-bootstrap-pipeline.yml
│   └── post-deploy-pipeline.yml
│
├── concourse/                        ← Concourse CI (*-pipeline.yml)
├── circleci/                         ← CircleCI version: 2.1 (*.circleci.yml)
├── gocd/                             ← GoCD format_version: 10 (*.gocd.yml)
├── travis/                           ← Travis CI (*.travis.yml)
├── harness/                          ← Harness CI/CD (*-pipeline.yml)
├── azure-devops/                     ← Azure Pipelines (*.yml)
├── aws-codepipeline/                 ← AWS CodeBuild buildspecs (buildspec-*.yml)
└── gcp-cloudbuild/                   ← GCP Cloud Build (cloudbuild-*.yaml)
```

---

## Pipeline Overview

| Pipeline | Trigger | Est. Duration | Purpose |
|---|---|---|---|
| install-tools | manual | ~30 min | Bootstrap agent with runtimes, CLIs, scanners |
| cluster-bootstrap | manual | ~4 hrs | 6-phase full cluster bring-up (phases 1–6) |
| k8s-infra | manual | ~90 min | Provision / destroy EKS / GKE / AKS via Terraform |
| gitops | manual | ~20 min | Install ArgoCD, Flux, Argo Rollouts, KEDA, Velero |
| security | manual | ~30 min | Install Vault, Keycloak, Falco, Kyverno, cert-manager, Teleport |
| observability | manual | ~30 min | Install Prometheus, Grafana, Loki, Jaeger, OTel |
| messaging | manual | ~20 min | Install Kafka, RabbitMQ, NATS + create 20 topics |
| networking | manual | ~25 min | Install Istio, Traefik, Cilium, Consul |
| registry | manual | ~20 min | Install Harbor, Nexus + provision cloud registry |
| databases | manual | ~25 min | Install Postgres, MongoDB, Redis, Cassandra, ClickHouse, et al. |
| streaming | manual | ~15 min | Deploy Debezium CDC connectors and Apache Flink jobs |
| tooling | manual | ~35 min | Developer tools: pgAdmin, Superset, MLflow, Botkube, OpenCost, etc. |
| pre-deploy | manual/webhook | ~20 min | Git fetch → secret scan → SAST → SCA → compile → docker build → image scan → sign → push → GitOps update |
| deploy | manual/ArgoCD | ~10 min | Verify image in Harbor → ArgoCD sync → rollout status → healthz check |
| post-deploy | manual | ~45 min | Smoke tests → integration → Hurl → Pact → ZAP DAST → Nuclei → k6 → Locust → Gatling → SLO |
| api-quality | manual | ~30 min | Spectral OpenAPI lint → Hurl HTTP flows → Pact publish → Terrascan IaC |
| reports | manual | ~10 min | Build and deploy Reports Portal web app (central report aggregator) |

---

## CI Pipeline (every push / PR)

Runs automatically on every push to `main`, `develop`, `feature/*`, and `release/*` branches,
and on every pull request.

### Stages

```
git push / PR
  ─”─ tests (Go · Java · Kotlin · Python · Node.js · Rust · C# · Scala)
       ─”─ secret-scan (Gitleaks)
            ─”─ sast (Semgrep)
                 ─”─ sca (Trivy filesystem)
                      ─”─ iac-scan (Checkov)
                           ─”─ notify-slack
```

### Language Test Mapping

| Language | Container | Command |
|---|---|---|
| Go | `golang:1.23-alpine` | `go test ./... -race -count=1` |
| Java (Maven) | `maven:3.9-eclipse-temurin-21` | `mvn test -q` |
| Kotlin (Gradle) | `gradle:8.10-jdk21` | `gradle test -q` |
| Python | `python:3.12-slim` | `pytest -q` |
| Node.js | `node:22-alpine` | `npm ci && npm test` |
| Rust | `rust:1.81-slim` | `cargo test` |
| C# / .NET | `mcr.microsoft.com/dotnet/sdk:8.0` | `dotnet test -q` |
| Scala (sbt) | `sbtscala/scala-sbt:latest` | `sbt test` |

---

## Deploy Pipeline

Triggered manually per service. Performs the full build → scan → sign → push → deploy cycle.

| Stage | Tool | Blocking? |
|---|---|---|
| secret-scan | Gitleaks | No (warn) |
| sast | Semgrep | No (warn) |
| sonarqube | SonarQube scanner | No (warn) |
| docker-build | Docker multi-stage | Yes |
| image-scan | Trivy (CRITICAL exit-1) | No (warn) |
| docker-push | Harbor registry | Yes |
| cosign-sign | Cosign keyless → Rekor | No (warn) |
| helm-deploy | `helm upgrade --install` | Yes |
| notify-slack | curl webhook | No |

### Required Environment Variables

| Variable | Description |
|---|---|
| `SERVICE_NAME` | Service directory name (e.g., `order-service`) |
| `IMAGE_TAG` | Semver or SHA (e.g., `v1.5.0`) |
| `ENVIRONMENT` | `staging` or `production` |
| `HARBOR_REGISTRY` | Harbor hostname (e.g., `harbor.shopos.internal`) |
| `HARBOR_USERNAME` | Harbor robot account |
| `HARBOR_PASSWORD` | Harbor robot password (secret) |
| `SONAR_TOKEN` | SonarQube token (secret) |
| `SONAR_HOST_URL` | SonarQube URL |
| `KUBECONFIG_CONTENT` | Base64-encoded kubeconfig (secret) |
| `SLACK_WEBHOOK` | Slack incoming webhook URL (secret) |

---

## Cluster Bootstrap Pipeline

The `cluster-bootstrap` pipeline runs 6 sequential phases to bring a bare cluster to
production-ready state. Each phase waits for the previous to complete.

| Phase | Tools Installed |
|---|---|
| 1 — Networking | Cilium CNI, Istio service mesh, Traefik edge router |
| 2 — Security | cert-manager, HashiCorp Vault, Keycloak, Kyverno, Falco |
| 3 — Observability | Prometheus stack, Grafana, Loki, Jaeger, OTel Collector |
| 4 — Messaging | ZooKeeper, Kafka, RabbitMQ, NATS JetStream + 20 topics |
| 5 — Registry | MinIO, Harbor, Nexus + 8 MinIO buckets |
| 6 — GitOps | ArgoCD, Argo Rollouts, Argo Workflows, KEDA, Velero |

Run with `START_PHASE` and `END_PHASE` env vars to resume from a specific phase.

---

## Supply Chain Security

All images are signed with Cosign (keyless via OIDC) and the signature is recorded in
Rekor. Kyverno enforces signature verification at admission time.

```bash
# Verify an image manually
cosign verify \
  --certificate-identity=ci@shopos.internal \
  --certificate-oidc-issuer=https://token.actions.githubusercontent.com \
  harbor.shopos.internal/shopos/order-service:v1.5.0
```

---

## Dagger — Running Locally

Dagger pipelines are plain Go programs and run without any CI server.

```bash
# Run deploy pipeline locally
cd ci/dagger/deploy
HARBOR_REGISTRY=localhost:5000 \
HARBOR_USERNAME=admin \
HARBOR_PASSWORD=secret \
SERVICE_NAME=order-service \
IMAGE_TAG=local-test \
KUBECONFIG_CONTENT=$(cat ~/.kube/config | base64) \
SONAR_TOKEN=xxx \
SONAR_HOST_URL=http://sonar.local \
SLACK_WEBHOOK=https://hooks.slack.com/... \
dagger run go run main.go

# Run cluster bootstrap (phases 1-3 only)
cd ci/dagger/cluster-bootstrap
START_PHASE=1 END_PHASE=3 \
KUBECONFIG_CONTENT=$(cat ~/.kube/config | base64) \
... \
dagger run go run main.go
```

---

## Image Tagging Strategy

| Event | Tag |
|---|---|
| Pull request | `pr-{number}-{sha8}` |
| Merge to `main` | `main-{sha8}` |
| Release tag (`v*`) | `{tag}` + `latest` |

Images are immutable — each environment pins the exact SHA-tagged image built for that commit.

---

## References

- [Dagger Go SDK](https://docs.dagger.io/sdk/go)
- [Tekton Pipelines](https://tekton.dev/docs/)
- [Concourse CI](https://concourse-ci.org/docs.html)
- [Jenkins Declarative Pipeline](https://www.jenkins.io/doc/book/pipeline/syntax/)
- [Drone CI](https://docs.drone.io/) / [Woodpecker CI](https://woodpecker-ci.org/docs/)
- [Cosign / Sigstore](https://docs.sigstore.dev/)
- [ShopOS GitOps](../gitops/README.md)
- [ShopOS Security](../security/README.md)
- [Getting Started](../GETTING_STARTED.md)
