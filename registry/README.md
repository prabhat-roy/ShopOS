# Registry â€” ShopOS

Helm charts for artifact and container registries used in the CI/CD pipeline.

## Directory Structure

```
registry/
â””â”€â”€ charts/
    â”œâ”€â”€ harbor/         â† Harbor container/Helm registry (primary)
    â”œâ”€â”€ nexus/          â† Nexus 3 â€” Maven, npm, PyPI, Go, Docker proxy
    â”œâ”€â”€ gitea/          â† Gitea â€” self-hosted Git server (GitOps source of truth)
    â”œâ”€â”€ chartmuseum/    â† Helm chart repository
    â”œâ”€â”€ zot/            â† OCI-native container registry (lightweight alternative)
    â”œâ”€â”€ quay/           â† Quay.io self-hosted (Red Hat)
    â”œâ”€â”€ gitlab/         â† GitLab self-hosted (Git + registry + CI)
    â”œâ”€â”€ forgejo/        â† Forgejo (Gitea fork)
    â”œâ”€â”€ distribution/   â† Docker Distribution (registry:2)
    â”œâ”€â”€ dragonfly/      â† Dragonfly P2P image distribution
    â”œâ”€â”€ athens/         â† Athens Go module proxy
    â”œâ”€â”€ goproxy/        â† Go module proxy
    â”œâ”€â”€ devpi/          â† Python package index
    â”œâ”€â”€ pypiserver/     â† Simple PyPI server
    â”œâ”€â”€ verdaccio/      â† npm private registry
    â”œâ”€â”€ cnpmjs/         â† npm registry mirror
    â”œâ”€â”€ conan-server/   â† Conan C/C++ package server
    â”œâ”€â”€ baget/          â† BaGet NuGet server
    â”œâ”€â”€ kellnr/         â† Rust crate registry
    â”œâ”€â”€ reposilite/     â† Lightweight Maven repository
    â”œâ”€â”€ alexandrie/     â† Rust crate registry alternative
    â”œâ”€â”€ aptly/          â† Debian package repository
    â”œâ”€â”€ pulp/           â† Pulp content management
    â”œâ”€â”€ geminabox/      â† Ruby gem server
    â”œâ”€â”€ gitbucket/      â† GitBucket (Git hosting)
    â”œâ”€â”€ gogs/           â† Gogs (lightweight Git service)
    â”œâ”€â”€ onedev/         â† OneDev all-in-one DevOps
    â”œâ”€â”€ quetz/          â† Conda package server
    â””â”€â”€ terrareg/       â† Terraform module registry
```

## Deployed Stack

| Component | Version | Role |
|---|---|---|
| Harbor | latest | Primary OCI container registry + Helm chart repository + image vulnerability scanning |
| Nexus | 3.71 | Universal artifact proxy â€” Maven (Java/Kotlin/Scala), npm (Node.js), PyPI (Python), Go modules, Docker layers |
| Gitea | 1.22 | Self-hosted Git server â€” GitOps source of truth for ArgoCD and Flux |
| ChartMuseum | latest | Helm chart repository for per-service charts |

## Image Build & Push Flow

```
CI Build (Jenkins / Drone)
  â”‚
  â”œâ”€â”€ Build Docker image (multi-stage, non-root)
  â”œâ”€â”€ Scan with Trivy + Grype (block on CRITICAL CVEs)
  â”œâ”€â”€ Generate SBOM with Syft
  â”œâ”€â”€ Sign image with Cosign (Sigstore)
  â””â”€â”€ Push to Harbor registry
         â””â”€â”€ Harbor scans with Trivy on push
```

All images are pulled from Harbor during Kubernetes deployments. Direct pulls from Docker Hub are blocked by network policy.

## Harbor Configuration

- Project per domain (e.g., `shopos/commerce/`, `shopos/catalog/`)
- Robot accounts per CI system (Jenkins token, Drone token)
- Replication rules: mirror critical base images (golang, openjdk, python, node) from Docker Hub â†’ Harbor on a schedule
- Image retention policy: keep last 10 tags per repository; delete untagged manifests older than 7 days

## Nexus Configuration

- Maven proxy: proxies Maven Central; local caching for Java/Kotlin/Scala builds
- npm proxy: proxies npmjs.com; local caching for Node.js builds  
- PyPI proxy: proxies PyPI; local caching for Python builds
- Go proxy: `GOPROXY=http://nexus:8081/repository/go-proxy/` â€” all Go modules pulled through Nexus
- Docker proxy: mirrors Docker Hub rate limits; all base image pulls go through Nexus

## Gitea Configuration

- Mirrors the primary GitHub repository on a 5-minute sync schedule
- ArgoCD and Flux point to Gitea for GitOps (avoids GitHub rate limits in prod)
- Webhook triggers Drone CI on push events

## References

- [CI/CD Pipelines](../ci/README.md)
- [GitOps Configs](../gitops/README.md)
- [ADR-006: GitOps with ArgoCD](../docs/adr/006-gitops-with-argocd.md)
