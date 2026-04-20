# Registry — ShopOS

Helm charts for artifact and container registries used in the CI/CD pipeline.

## Directory Structure

```
registry/
└── charts/
    ├── harbor/         ← Harbor container/Helm registry (primary)
    ├── nexus/          ← Nexus 3 — Maven, npm, PyPI, Go, Docker proxy
    ├── gitea/          ← Gitea — self-hosted Git server (GitOps source of truth)
    ├── chartmuseum/    ← Helm chart repository
    ├── zot/            ← OCI-native container registry (lightweight alternative)
    ├── quay/           ← Quay.io self-hosted (Red Hat)
    ├── gitlab/         ← GitLab self-hosted (Git + registry + CI)
    ├── forgejo/        ← Forgejo (Gitea fork)
    ├── distribution/   ← Docker Distribution (registry:2)
    ├── dragonfly/      ← Dragonfly P2P image distribution
    ├── athens/         ← Athens Go module proxy
    ├── goproxy/        ← Go module proxy
    ├── devpi/          ← Python package index
    ├── pypiserver/     ← Simple PyPI server
    ├── verdaccio/      ← npm private registry
    ├── cnpmjs/         ← npm registry mirror
    ├── conan-server/   ← Conan C/C++ package server
    ├── baget/          ← BaGet NuGet server
    ├── kellnr/         ← Rust crate registry
    ├── reposilite/     ← Lightweight Maven repository
    ├── alexandrie/     ← Rust crate registry alternative
    ├── aptly/          ← Debian package repository
    ├── pulp/           ← Pulp content management
    ├── geminabox/      ← Ruby gem server
    ├── gitbucket/      ← GitBucket (Git hosting)
    ├── gogs/           ← Gogs (lightweight Git service)
    ├── onedev/         ← OneDev all-in-one DevOps
    ├── quetz/          ← Conda package server
    └── terrareg/       ← Terraform module registry
```

## Deployed Stack

| Component | Version | Role |
|---|---|---|
| **Harbor** | latest | Primary OCI container registry + Helm chart repository + image vulnerability scanning |
| **Nexus** | 3.71 | Universal artifact proxy — Maven (Java/Kotlin/Scala), npm (Node.js), PyPI (Python), Go modules, Docker layers |
| **Gitea** | 1.22 | Self-hosted Git server — GitOps source of truth for ArgoCD and Flux |
| **ChartMuseum** | latest | Helm chart repository for per-service charts |

## Image Build & Push Flow

```
CI Build (Jenkins / Drone)
  │
  ├── Build Docker image (multi-stage, non-root)
  ├── Scan with Trivy + Grype (block on CRITICAL CVEs)
  ├── Generate SBOM with Syft
  ├── Sign image with Cosign (Sigstore)
  └── Push to Harbor registry
         └── Harbor scans with Trivy on push
```

All images are pulled from Harbor during Kubernetes deployments. Direct pulls from Docker Hub are blocked by network policy.

## Harbor Configuration

- Project per domain (e.g., `shopos/commerce/`, `shopos/catalog/`)
- Robot accounts per CI system (Jenkins token, Drone token)
- Replication rules: mirror critical base images (golang, openjdk, python, node) from Docker Hub → Harbor on a schedule
- Image retention policy: keep last 10 tags per repository; delete untagged manifests older than 7 days

## Nexus Configuration

- **Maven proxy**: proxies Maven Central; local caching for Java/Kotlin/Scala builds
- **npm proxy**: proxies npmjs.com; local caching for Node.js builds  
- **PyPI proxy**: proxies PyPI; local caching for Python builds
- **Go proxy**: `GOPROXY=http://nexus:8081/repository/go-proxy/` — all Go modules pulled through Nexus
- **Docker proxy**: mirrors Docker Hub rate limits; all base image pulls go through Nexus

## Gitea Configuration

- Mirrors the primary GitHub repository on a 5-minute sync schedule
- ArgoCD and Flux point to Gitea for GitOps (avoids GitHub rate limits in prod)
- Webhook triggers Drone CI on push events

## References

- [CI/CD Pipelines](../ci/README.md)
- [GitOps Configs](../gitops/README.md)
- [ADR-006: GitOps with ArgoCD](../docs/adr/006-gitops-with-argocd.md)
