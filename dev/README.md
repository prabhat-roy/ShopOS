# Developer Experience — ShopOS

Tools that improve the inner loop: cloud IDEs, live cluster sync, low-code automation,
and Backstage scaffolder templates.

## Layout

| Subdir | Tool | Role |
|---|---|---|
| [coder/](coder/) | Coder | Self-hosted cloud development environments (Terraform-defined templates), Keycloak SSO |
| [devspace/](devspace/) | DevSpace | `devspace dev <service>` — live sync to remote cluster |
| [n8n/](n8n/) | n8n | Low-code workflow automation for ops (PagerDuty → Slack handoff, daily PR summaries) |
| [windmill/](windmill/) | Windmill | Internal scripts/APIs as a service (Retool/AWS Lambda alternative) |
| [score/](score/) | Score | Cloud-agnostic workload spec — one `score.yaml` replaces Helm + Compose + K8s manifests |
| [scaffolder/](scaffolder/) | Backstage Software Templates | One-click Go-service scaffolder, opens PR with full chart + CI |

## Local development cycle

```bash
# 1. Spin up your environment in Coder (browser-based VS Code)
coder create dev-$(whoami) --template shopos-go

# 2. Inside the workspace, start a service with hot-reload
cd src/commerce/cart-service
devspace dev

# 3. Telepresence intercept (run cart-service locally, hit live cluster deps)
telepresence intercept cart-service --port 50080:50080

# 4. Use Score for portable workload spec
score-compose generate score.yaml > docker-compose.gen.yaml

# 5. Scaffold a brand-new service via Backstage UI or CLI
backstage-cli new --template shopos-go-service --name foo-service --domain platform
```

## Related

- Backstage portal config: [`backstage/`](../backstage/)
- Service generator script: [`scripts/bash/scaffold-service.sh`](../scripts/bash/scaffold-service.sh)
