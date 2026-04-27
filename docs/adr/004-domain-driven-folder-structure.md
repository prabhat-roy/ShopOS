# ADR-004: Domain-Driven Folder Structure Under `src/`

Status: Accepted  
Date: 2024-01-20  
Deciders: Platform Architecture Team

---

## Context

With 154 services, we needed a folder structure that makes clear which team owns which service, groups services by business capability, and scales without becoming unwieldy.

Alternatives considered:
- Flat (`src/order-service/`, `src/cart-service/`) â€” simple but loses domain context as the count grows
- By language (`src/go/`, `src/java/`) â€” groups implementation details, not business concepts
- By domain (`src/commerce/order-service/`) â€” aligns with DDD bounded contexts

---

## Decision

Services are organised under `src/{domain}/{service-name}/` where domain matches the bounded context from Domain-Driven Design.

```
src/
â”œâ”€â”€ platform/           â† Cross-cutting infrastructure (22 services)
â”œâ”€â”€ identity/           â† Auth, users, sessions, compliance (8 services)
â”œâ”€â”€ catalog/            â† Products, pricing, search (12 services)
â”œâ”€â”€ commerce/           â† Cart, checkout, orders, payments (23 services)
â”œâ”€â”€ supply-chain/       â† Vendors, warehouses, fulfilment (13 services)
â”œâ”€â”€ financial/          â† Invoicing, accounting, payouts (11 services)
â”œâ”€â”€ customer-experience/â† Reviews, support, wishlists (14 services)
â”œâ”€â”€ communications/     â† Notifications, email, SMS, chat (9 services)
â”œâ”€â”€ content/            â† Media, CMS, i18n (8 services)
â”œâ”€â”€ analytics-ai/       â† Events, ML, recommendations (13 services)
â”œâ”€â”€ b2b/                â† Organisations, contracts, procurement (7 services)
â”œâ”€â”€ integrations/       â† ERP, marketplace, CRM adapters (10 services)
â””â”€â”€ affiliate/          â† Affiliate, referral, influencer (4 services)
```

---

## Rationale

1. Team ownership alignment â€” Each domain maps to a team. Developers navigate to their domain first, then their service.
2. Bounded context isolation â€” DDD bounded contexts map directly to folders. Cross-domain calls must go through explicit interfaces (gRPC or Kafka), preventing accidental coupling.
3. CI/CD routing â€” CI pipelines detect the changed domain from the file path and route tests and deployments to the appropriate pipeline.
4. Helm chart mirroring â€” `helm/charts/{service-name}` and ArgoCD ApplicationSets use the same service name, making the deployment topology predictable.

---

## Consequences

Positive: Domain ownership is immediately obvious from the path; bounded context boundaries are structurally enforced; CI can filter by changed domain.

Negative: Services shared across domains (e.g., audit-service) must be placed in `platform/` by convention even when consumed by other domains.

Mitigations: The `platform/` domain explicitly contains all cross-cutting concerns. Inter-domain dependencies are always via gRPC or Kafka â€” never by importing code across domain folders.
