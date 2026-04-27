# Harbor â€” Private OCI Container Registry

## Role in ShopOS

Harbor is the private container image registry for ShopOS. It stores all Docker images built
from the 130 microservices, enforces vulnerability scanning with Trivy before images can be deployed,
signs images with Cosign for supply chain integrity, and provides project-level RBAC to restrict
who can push and pull images per domain team.

Key capabilities used in ShopOS:
- OCI image storage for all 130 service images
- Integrated Trivy scanning â€” images with critical CVEs are automatically quarantined
- Cosign content trust â€” Kubernetes admission controller rejects unsigned images (Phase 3)
- Project-level RBAC: each domain team (platform, identity, catalog, â€¦) has its own Harbor project
- Proxy cache â€” caches Docker Hub, GCR, and GHCR images locally to avoid rate limits
- Webhook notifications to ArgoCD on successful image push

---

## CI Pipeline â†’ Harbor â†’ Kubernetes

```mermaid
flowchart LR
    subgraph CI["CI Pipeline (Tekton / Jenkins)"]
        SRC[Source Code Push<br/>Gitea / GitHub]
        BLD[docker build<br/>multi-stage Dockerfile]
        SCAN[Trivy Scan<br/>in CI step]
        SIGN[Cosign Sign<br/>keyless OIDC]
    end

    subgraph Harbor["Harbor Registry (:5000)"]
        PROJ[Harbor Projects<br/>per domain]
        TRIVY[Harbor Trivy<br/>re-scan on push]
        TRUST[Content Trust<br/>Cosign signatures]
        CACHE[Proxy Cache<br/>docker.io, gcr.io]
    end

    subgraph GitOps["GitOps (ArgoCD)"]
        WH[Webhook Receiver<br/>image push event]
        SYNC[ArgoCD Sync<br/>updates image tag in Git]
    end

    subgraph K8s["Kubernetes"]
        ADM[Admission Controller<br/>Kyverno â€” verify signature]
        DEP[Deployment<br/>pulls from Harbor]
    end

    SRC --> BLD
    BLD --> SCAN
    SCAN -->|passed| SIGN
    SIGN -->|docker push| PROJ
    PROJ --> TRIVY
    TRIVY -->|webhook on clean| WH
    WH --> SYNC
    SYNC --> ADM
    ADM -->|pull| DEP
    CACHE -.->|upstream pull-through| DEP
```

---

## Project Structure in Harbor

| Harbor Project | Domain | Repositories |
|---|---|---|
| `platform` | platform | api-gateway, web-bff, mobile-bff, saga-orchestrator, â€¦ |
| `identity` | identity | auth-service, user-service, session-service, â€¦ |
| `catalog` | catalog | product-catalog-service, pricing-service, search-service, â€¦ |
| `commerce` | commerce | cart-service, order-service, payment-service, â€¦ |
| `supply-chain` | supply-chain | warehouse-service, fulfillment-service, tracking-service, â€¦ |
| `financial` | financial | invoice-service, payout-service, reconciliation-service, â€¦ |
| `cx` | customer-experience | review-rating-service, support-ticket-service, â€¦ |
| `analytics-ai` | analytics-ai | recommendation-service, fraud-detection-service, â€¦ |
| `proxy-cache` | â€” | docker.io, gcr.io, ghcr.io pull-through |

---

## Setup Instructions

### 1. Install Harbor (Docker Compose)

```bash
# Download Harbor installer
wget https://github.com/goharbor/harbor/releases/latest/download/harbor-online-installer.tgz
tar xzf harbor-online-installer.tgz
cd harbor

# Copy and edit config
cp harbor.yml.tmpl harbor.yml
# Edit hostname, admin password, database password

# Run installer
./install.sh --with-trivy
```

### 2. Configure Docker Daemon (dev)

```json
// /etc/docker/daemon.json
{
  "insecure-registries": ["harbor.shopos.internal:5000"]
}
```

### 3. Login and Push

```bash
docker login harbor.shopos.internal:5000 -u admin -p Harbor12345
docker tag order-service:latest harbor.shopos.internal:5000/commerce/order-service:v1.0.0
docker push harbor.shopos.internal:5000/commerce/order-service:v1.0.0
```

### 4. Create Robot Accounts (for CI)

Create a robot account per CI pipeline with push access scoped to the relevant Harbor project.
Robot account credentials are stored in Vault (`secret/ci/harbor/robot-account`).

---

## Trivy Scanning Policy

| Severity | Policy |
|---|---|
| `CRITICAL` | Block push â€” image quarantined |
| `HIGH` | Allow push, alert via webhook to Slack |
| `MEDIUM` | Allow push, log to audit |
| `LOW` / `NEGLIGIBLE` | Allow, visible in Harbor UI |

`ignore_unfixed: false` means vulnerabilities with no available fix are still reported and counted
against policy.

---

## Connection Details

| Parameter | Value |
|---|---|
| HTTP Port | 5000 |
| Admin User | `admin` |
| Admin Password | Set in `harbor.yml` â€” rotate immediately in production |
| Data Volume | `/data` |
| Database | PostgreSQL (bundled or external) |
