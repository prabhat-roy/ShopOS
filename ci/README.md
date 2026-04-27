# CI Pipelines â€” ShopOS

ShopOS ships 17 pipelines implemented across 15 CI/CD platforms. Jenkins is the primary
CI server with 17 Jenkinsfiles; all other platforms mirror the same pipeline set.

---

## Platforms

| Platform | Directory | Files | Notes |
|---|---|---|---|
| Jenkins | `jenkins/` | 17 Jenkinsfiles | Primary CI server; declarative pipeline syntax |
| Drone CI | `drone/` | 12 YAML | Drone v2; mirrors core Jenkins pipelines |
| Woodpecker CI | `woodpecker/` | 12 YAML | Drone-compatible fork; drop-in replacement |
| Dagger | `dagger/` | 12 Go modules | Portable Go SDK â€” run on any CI or locally |
| Tekton | `tekton/` | 12 YAML | Kubernetes CRD-native (Task + Pipeline + PipelineRun) |
| Concourse CI | `concourse/` | 12 YAML | Resource/job DAG pipelines |
| GitLab CI | `gitlab-ci/` | 12 YAML | `.gitlab-ci.yml` â€” native GitLab SCM integration |
| GitHub Actions | `github-actions/` | 12 YAML | Stored in `ci/github-actions/` â€” not in `.github/` so auto-triggering is disabled |
| CircleCI | `circleci/` | 12 YAML | `version: 2.1` orb-based pipelines |
| GoCD | `gocd/` | 12 YAML | Stage/job pipelines with manual approval gates |
| Travis CI | `travis/` | 12 YAML | Stage-based pipelines with branch filters |
| Harness CI | `harness/` | 12 YAML | Enterprise CI/CD with built-in CD stages |
| Azure DevOps | `azure-devops/` | 12 YAML | `azure-pipelines.yml` â€” native Azure integration |
| AWS CodePipeline | `aws-codepipeline/` | 12 YAML | `buildspec.yml` + CodePipeline JSON definitions |
| GCP Cloud Build | `gcp-cloudbuild/` | 12 YAML | `cloudbuild.yaml` â€” native GCP integration |

---

## Directory Structure

```
ci/
â”œâ”€â”€ jenkins/                          â† 17 Jenkinsfiles (primary)
â”‚   â”œâ”€â”€ install-tools.Jenkinsfile     â† Bootstrap agent runtimes and CLIs
â”‚   â”œâ”€â”€ cluster-bootstrap.Jenkinsfile â† Full cluster bring-up (6 phases)
â”‚   â”œâ”€â”€ k8s-infra.Jenkinsfile         â† Terraform EKS/GKE/AKS provisioning
â”‚   â”œâ”€â”€ gitops.Jenkinsfile            â† ArgoCD, Flux, Argo Rollouts, KEDA, Velero
â”‚   â”œâ”€â”€ security.Jenkinsfile          â† Vault, Keycloak, Falco, Kyverno, cert-manager
â”‚   â”œâ”€â”€ observability.Jenkinsfile     â† Prometheus, Grafana, Loki, Jaeger, OTel
â”‚   â”œâ”€â”€ messaging.Jenkinsfile         â† Kafka, RabbitMQ, NATS, schema registry
â”‚   â”œâ”€â”€ networking.Jenkinsfile        â† Istio, Traefik, Cilium, Consul
â”‚   â”œâ”€â”€ registry.Jenkinsfile          â† Harbor, Nexus + cloud registry provisioning
â”‚   â”œâ”€â”€ databases.Jenkinsfile         â† Postgres, MongoDB, Redis, Cassandra, ClickHouse
â”‚   â”œâ”€â”€ streaming.Jenkinsfile         â† Debezium CDC, Apache Flink jobs
â”‚   â”œâ”€â”€ tooling.Jenkinsfile           â† Developer tools (pgAdmin, Superset, MLflow, etc.)
â”‚   â”œâ”€â”€ pre-deploy.Jenkinsfile        â† Git fetch â†’ scan â†’ compile â†’ docker build â†’ push
â”‚   â”œâ”€â”€ deploy.Jenkinsfile            â† GitOps trigger â†’ ArgoCD sync â†’ rollout verify
â”‚   â”œâ”€â”€ post-deploy.Jenkinsfile       â† Smoke tests â†’ DAST â†’ load tests â†’ SLO validate
â”‚   â”œâ”€â”€ api-quality.Jenkinsfile       â† Spectral lint â†’ Hurl â†’ Pact â†’ Terrascan
â”‚   â””â”€â”€ reports.Jenkinsfile           â† Build/deploy Reports Portal web app
â”‚
â”œâ”€â”€ drone/                            â† Drone CI (same 12 pipelines, *.drone.yml)
â”œâ”€â”€ woodpecker/                       â† Woodpecker CI (same 12 pipelines, *.woodpecker.yml)
â”œâ”€â”€ gitlab-ci/                        â† GitLab CI (same 12 pipelines, *.gitlab-ci.yml)
â”‚
â”œâ”€â”€ github-actions/                   â† GitHub Actions (same 12 pipelines, *.yml)
â”‚   â”‚                                   Stored here (NOT in .github/) â€” auto-triggering disabled.
â”‚   â”‚                                   To enable: copy files to .github/workflows/ and add secrets.
â”‚   â”œâ”€â”€ deploy.yml
â”‚   â”œâ”€â”€ post-deploy.yml
â”‚   â”œâ”€â”€ k8s-infra.yml
â”‚   â”œâ”€â”€ gitops.yml
â”‚   â”œâ”€â”€ security.yml
â”‚   â”œâ”€â”€ observability.yml
â”‚   â”œâ”€â”€ messaging.yml
â”‚   â”œâ”€â”€ networking.yml
â”‚   â”œâ”€â”€ registry.yml
â”‚   â”œâ”€â”€ install-tools.yml
â”‚   â””â”€â”€ cluster-bootstrap.yml
â”‚
â”œâ”€â”€ dagger/                           â† Dagger Go SDK â€” one subdirectory per pipeline
â”‚   â”œâ”€â”€ go.mod / main.go              â† root module (shared utilities)
â”‚   â”œâ”€â”€ deploy/main.go
â”‚   â”œâ”€â”€ security/main.go
â”‚   â”œâ”€â”€ networking/main.go
â”‚   â”œâ”€â”€ observability/main.go
â”‚   â”œâ”€â”€ messaging/main.go
â”‚   â”œâ”€â”€ k8s-infra/main.go
â”‚   â”œâ”€â”€ gitops/main.go
â”‚   â”œâ”€â”€ registry/main.go
â”‚   â”œâ”€â”€ install-tools/main.go
â”‚   â”œâ”€â”€ cluster-bootstrap/main.go
â”‚   â””â”€â”€ post-deploy/main.go
â”‚
â”œâ”€â”€ tekton/                           â† Tekton Pipelines (Kubernetes CRDs)
â”‚   â”œâ”€â”€ deploy-pipeline.yml
â”‚   â”œâ”€â”€ security-pipeline.yml
â”‚   â”œâ”€â”€ networking-pipeline.yml
â”‚   â”œâ”€â”€ observability-pipeline.yml
â”‚   â”œâ”€â”€ messaging-pipeline.yml
â”‚   â”œâ”€â”€ k8s-infra-pipeline.yml
â”‚   â”œâ”€â”€ gitops-pipeline.yml
â”‚   â”œâ”€â”€ registry-pipeline.yml
â”‚   â”œâ”€â”€ install-tools-pipeline.yml
â”‚   â”œâ”€â”€ cluster-bootstrap-pipeline.yml
â”‚   â””â”€â”€ post-deploy-pipeline.yml
â”‚
â”œâ”€â”€ concourse/                        â† Concourse CI (*-pipeline.yml)
â”œâ”€â”€ circleci/                         â† CircleCI version: 2.1 (*.circleci.yml)
â”œâ”€â”€ gocd/                             â† GoCD format_version: 10 (*.gocd.yml)
â”œâ”€â”€ travis/                           â† Travis CI (*.travis.yml)
â”œâ”€â”€ harness/                          â† Harness CI/CD (*-pipeline.yml)
â”œâ”€â”€ azure-devops/                     â† Azure Pipelines (*.yml)
â”œâ”€â”€ aws-codepipeline/                 â† AWS CodeBuild buildspecs (buildspec-*.yml)
â””â”€â”€ gcp-cloudbuild/                   â† GCP Cloud Build (cloudbuild-*.yaml)
```

---

## Pipeline Overview

| Pipeline | Trigger | Est. Duration | Purpose |
|---|---|---|---|
| install-tools | manual | ~30 min | Bootstrap agent with runtimes, CLIs, scanners |
| cluster-bootstrap | manual | ~4 hrs | 6-phase full cluster bring-up (phases 1â€“6) |
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
| pre-deploy | manual/webhook | ~20 min | Git fetch â†’ secret scan â†’ SAST â†’ SCA â†’ compile â†’ docker build â†’ image scan â†’ sign â†’ push â†’ GitOps update |
| deploy | manual/ArgoCD | ~10 min | Verify image in Harbor â†’ ArgoCD sync â†’ rollout status â†’ healthz check |
| post-deploy | manual | ~45 min | Smoke tests â†’ integration â†’ Hurl â†’ Pact â†’ ZAP DAST â†’ Nuclei â†’ k6 â†’ Locust â†’ Gatling â†’ SLO |
| api-quality | manual | ~30 min | Spectral OpenAPI lint â†’ Hurl HTTP flows â†’ Pact publish â†’ Terrascan IaC |
| reports | manual | ~10 min | Build and deploy Reports Portal web app (central report aggregator) |

---

## CI Pipeline (every push / PR)

Runs automatically on every push to `main`, `develop`, `feature/*`, and `release/*` branches,
and on every pull request.

### Stages

```
git push / PR
  â””â”€ tests (Go Â· Java Â· Kotlin Â· Python Â· Node.js Â· Rust Â· C# Â· Scala)
       â””â”€ secret-scan (Gitleaks)
            â””â”€ sast (Semgrep)
                 â””â”€ sca (Trivy filesystem)
                      â””â”€ iac-scan (Checkov)
                           â””â”€ notify-slack
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

Triggered manually per service. Performs the full build â†’ scan â†’ sign â†’ push â†’ deploy cycle.

| Stage | Tool | Blocking? |
|---|---|---|
| secret-scan | Gitleaks | No (warn) |
| sast | Semgrep | No (warn) |
| sonarqube | SonarQube scanner | No (warn) |
| docker-build | Docker multi-stage | Yes |
| image-scan | Trivy (CRITICAL exit-1) | No (warn) |
| docker-push | Harbor registry | Yes |
| cosign-sign | Cosign keyless â†’ Rekor | No (warn) |
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
| 1 â€” Networking | Cilium CNI, Istio service mesh, Traefik edge router |
| 2 â€” Security | cert-manager, HashiCorp Vault, Keycloak, Kyverno, Falco |
| 3 â€” Observability | Prometheus stack, Grafana, Loki, Jaeger, OTel Collector |
| 4 â€” Messaging | ZooKeeper, Kafka, RabbitMQ, NATS JetStream + 20 topics |
| 5 â€” Registry | MinIO, Harbor, Nexus + 8 MinIO buckets |
| 6 â€” GitOps | ArgoCD, Argo Rollouts, Argo Workflows, KEDA, Velero |

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

## Dagger â€” Running Locally

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

Images are immutable â€” each environment pins the exact SHA-tagged image built for that commit.

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
