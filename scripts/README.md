# Scripts — ShopOS

Helper scripts for service scaffolding, security scanning, CDN purging, and Vault bootstrap.

## Service scaffolding

| Script | Purpose |
|---|---|
| [`scaffold-service.sh`](scaffold-service.sh) | Scaffold a new Go service: `bash scripts/bash/scaffold-service.sh <domain> <name> <port>`. Creates `src/<domain>/<name>/` with `Dockerfile`, `Makefile`, `main.go`, `go.mod`, `.env.example`. |

After scaffolding a new service, manually create:
- `src/<domain>/<name>/README.md` — copy a sibling service README and adapt
- `helm/services/<name>/` — copy a sibling chart and adapt
- Add an entry to `gitops/argocd/applicationsets/all-services.yaml`
- Add an entry to `gitops/flux/base/helm-releases.yaml`
- Add an entry to `backstage/catalog-info.yaml`

## Other scripts

- [`bash/`](bash/) — Jenkins agent installer + cluster bootstrap helpers
- [`groovy/`](groovy/) — Jenkins shared-pipeline Groovy modules
- [`build/build-all-go.sh`](../build/build-all-go.sh) — build + push all Go service images via Ko
- [`networking/cdn/cdn-purge.sh`](../networking/cdn/cdn-purge.sh) — purge Varnish + Cloudflare + Fastly cache
- [`security/wiz-cli/scan.sh`](../security/wiz-cli/scan.sh) — consolidated Trivy + Grype + Syft + Checkov + Gitleaks scan
- [`security/gitguardian/install-hook.sh`](../security/gitguardian/install-hook.sh) — install ggshield pre-commit hook
- [`security/vault/bootstrap/`](../security/vault/bootstrap/) — three-step Vault HA bootstrap (auth methods → secret engines → policies)
