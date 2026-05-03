# Feature Flags — ShopOS

Runtime toggles for rollouts, kill-switches, and A/B tests. Wraps an open-source backend
behind the OpenFeature standard so any service can switch providers transparently.

## Layout

| Subdir | Tool | Role |
|---|---|---|
| [unleash/](unleash/) | Unleash | Open source feature flag platform — SDK (all 19 languages), UI, audit trail, gradual rollout, segment targeting |

## Conventions

- All flags use the OpenFeature SDK (services do not import Unleash directly)
- Flag naming: `<domain>.<service>.<flag>` — e.g. `commerce.checkout.split-payment-v2`
- Every flag has an owner + expiry date in Unleash UI; flags older than 90d are reviewed weekly
- Kill-switches go in the `safety` project so they're never auto-cleaned

## Related

- Per-tenant feature-flag service (DB-backed, complements Unleash for tenant-scoped flags): [`src/platform/feature-flag-service/`](../src/platform/feature-flag-service/)
- Argo Rollouts uses flag-style traffic weights for canary: [`gitops/argo-rollouts/`](../gitops/argocd/)
