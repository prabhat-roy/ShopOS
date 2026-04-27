# Domain Map â€” ShopOS

ShopOS organises 230 services (224 microservices + 6 frontend apps) into 19 bounded contexts. Each domain owns its data, publishes events via Kafka (Avro), and exposes capabilities via gRPC. All domains run on Kubernetes, connected by Istio mTLS, and observed by OpenTelemetry.

---

## Bounded Context Overview

```
â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”
â”‚                        ShopOS â€” 19 Domains Â· 230 Services                                     â”‚
â”‚                                                                                               â”‚
â”‚  â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”   â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”   â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”   â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”                  â”‚
â”‚  â”‚   PLATFORM   â”‚   â”‚   IDENTITY   â”‚   â”‚   CATALOG    â”‚   â”‚  COMMERCE    â”‚                  â”‚
â”‚  â”‚  27 services â”‚   â”‚  11 services â”‚   â”‚ 15 services  â”‚   â”‚ 28 services  â”‚                  â”‚
â”‚  â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜   â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜   â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜   â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜                  â”‚
â”‚                                                                                               â”‚
â”‚  â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”   â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”   â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”   â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”                  â”‚
â”‚  â”‚ SUPPLY CHAIN â”‚   â”‚  FINANCIAL   â”‚   â”‚  CUSTOMER    â”‚   â”‚    COMMS     â”‚                  â”‚
â”‚  â”‚ 17 services  â”‚   â”‚ 15 services  â”‚   â”‚  EXPERIENCE  â”‚   â”‚ 12 services  â”‚                  â”‚
â”‚  â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜   â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜   â”‚ 17 services  â”‚   â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜                  â”‚
â”‚                                        â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜                                      â”‚
â”‚  â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”   â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”   â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”   â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”                  â”‚
â”‚  â”‚   CONTENT    â”‚   â”‚ ANALYTICS/AI â”‚   â”‚     B2B      â”‚   â”‚ INTEGRATIONS â”‚                  â”‚
â”‚  â”‚  9 services  â”‚   â”‚ 13 services  â”‚   â”‚ 10 services  â”‚   â”‚ 14 services  â”‚                  â”‚
â”‚  â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜   â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜   â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜   â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜                  â”‚
â”‚                                                                                               â”‚
â”‚  â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”   â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”   â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”   â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”   â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â” â”‚
â”‚  â”‚  AFFILIATE   â”‚   â”‚ MARKETPLACE  â”‚   â”‚ GAMIFICATION â”‚   â”‚  DEV PLATFORMâ”‚   â”‚ COMPLIANCE â”‚ â”‚
â”‚  â”‚  6 services  â”‚   â”‚  8 services  â”‚   â”‚  6 services  â”‚   â”‚  6 services  â”‚   â”‚ 5 services â”‚ â”‚
â”‚  â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜   â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜   â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜   â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜   â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜ â”‚
â”‚                                                                                               â”‚
â”‚  â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”   â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â” â”‚
â”‚  â”‚SUSTAINABILITYâ”‚   â”‚                   WEB FRONTENDS (6 apps)                              â”‚ â”‚
â”‚  â”‚  5 services  â”‚   â”‚  Next.js Â· React+Vite Â· Vue.js 3 Â· Angular 18 Â· React Native        â”‚ â”‚
â”‚  â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜   â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜ â”‚
â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜
```

---

## Domain Ownership Map

### 1. Platform Domain
Owns: Cross-cutting infrastructure capabilities  
Services: api-gateway, web-bff, mobile-bff, partner-bff, graphql-gateway, config-service, feature-flag-service, rate-limiter-service, health-check-service, saga-orchestrator, event-store-service, cache-warming-service, webhook-service, scheduler-service, worker-job-queue, audit-service, load-generator, admin-portal, dead-letter-service, geolocation-service, event-replay-service, tenant-service, notification-preferences-service, circuit-breaker-service, idempotency-service, correlation-id-service, data-masking-service  
Databases: Postgres (audit, saga, event-store, scheduler, webhook), Redis (rate-limiter, worker queue), etcd (config)  
Workflow engine: Temporal â€” orchestrates complex sagas (order, refund, subscription) and coordinates long-running flows that span multiple domains  
Publishes: `platform.audit.recorded`, `platform.config.changed`

### 2. Identity Domain
Owns: Authentication, authorisation, user lifecycle, sessions  
Services: auth-service, user-service, session-service, permission-service, mfa-service, gdpr-service, api-key-service, device-fingerprint-service, sso-service, password-policy-service, bot-detection-service  
Databases: Postgres (users, permissions, MFA), Redis (sessions, device fingerprints)  
Publishes: `identity.user.registered`, `identity.user.deleted`, `identity.login.failed`

### 3. Catalog Domain
Owns: Product data, pricing, inventory levels, search index  
Services: product-catalog-service, category-service, brand-service, pricing-service, inventory-service, bundle-service, configurator-service, subscription-product-service, search-service, seo-service, product-import-service, price-list-service, product-label-service, variant-service, stock-reservation-service  
Databases: MongoDB (products, configurator), Postgres (categories, brands, pricing, inventory), Elasticsearch (search)  
Publishes: `catalog.product.created`, `catalog.product.updated`, `catalog.inventory.updated`, `catalog.price.changed`

### 4. Commerce Domain
Owns: Shopping cart, checkout flow, orders, payments, promotions  
Services: cart-service, checkout-service, order-service, payment-service, shipping-service, currency-service, tax-service, promotions-service, loyalty-service, return-refund-service, subscription-billing-service, fraud-detection-service, wallet-service, ab-testing-service, gift-card-service, address-validation-service, digital-goods-service, voucher-service, pre-order-service, backorder-service, waitlist-service, flash-sale-service, bnpl-service, split-payment-service, installment-service, dynamic-pricing-service, coupon-service, order-amendment-service  
Databases: Postgres (orders, payments, promotions, loyalty), Redis (cart, waitlist, flash-sale)  
Publishes: `commerce.order.placed`, `commerce.order.cancelled`, `commerce.order.fulfilled`, `commerce.payment.processed`, `commerce.payment.failed`, `commerce.cart.abandoned`

### 5. Supply Chain Domain
Owns: Vendors, warehouses, fulfilment, carrier integrations, tracking  
Services: vendor-service, purchase-order-service, warehouse-service, fulfillment-service, tracking-service, label-service, carrier-integration-service, demand-forecast-service, customs-duties-service, returns-logistics-service, supplier-portal-service, cold-chain-service, supplier-rating-service, route-optimization-service, packaging-service, cross-dock-service, duty-drawback-service  
Databases: Postgres (vendors, POs, warehouses, fulfilment), MongoDB (tracking, cold-chain)  
Publishes: `supplychain.shipment.created`, `supplychain.shipment.updated`, `supplychain.inventory.low`, `supplychain.inventory.restocked`

### 6. Financial Domain
Owns: Invoices, payouts, accounting ledger, reconciliation, compliance  
Services: invoice-service, accounting-service, payout-service, reconciliation-service, tax-reporting-service, expense-management-service, credit-service, kyc-aml-service, budget-service, chargeback-service, revenue-recognition-service, escrow-service, forex-service, audit-trail-service, dunning-service  
Databases: Postgres (all â€” ACID required for financial data)  
Publishes: `financial.invoice.created`, `financial.payout.initiated`, `financial.journal.entry.created`

### 7. Customer Experience Domain
Owns: Reviews, Q&A, wishlists, support tickets, live chat, consent  
Services: review-rating-service, qa-service, wishlist-service, compare-service, recently-viewed-service, support-ticket-service, live-chat-service, consent-management-service, age-verification-service, survey-service, feedback-service, price-alert-service, back-in-stock-service, gift-registry-service, loyalty-tier-service, accessibility-service, return-portal-service  
Databases: MongoDB (reviews, Q&A), Postgres (support, wishlist, consent), Redis (compare, recently-viewed, price-alert, back-in-stock)  
Publishes: `cx.review.submitted`, `cx.support.ticket.created`

### 8. Communications Domain
Owns: Notification delivery across all channels (email, SMS, push, in-app, WhatsApp)  
Services: notification-orchestrator, email-service, sms-service, push-notification-service, template-service, in-app-notification-service, digest-service, whatsapp-service, chatbot-service, telegram-service, voice-service, webhook-delivery-service  
Databases: MongoDB (templates), Redis (in-app), Postgres (digest)  
Consumes: `notification.email.requested`, `notification.sms.requested`, `notification.push.requested`

### 9. Content Domain
Owns: Media assets, CMS pages, documents, internationalisation  
Services: media-asset-service, image-processing-service, document-service, cms-service, video-service, sitemap-service, i18n-l10n-service, data-export-service, ab-content-service  
Databases: MinIO (binaries), MongoDB (CMS), Postgres (i18n)  
Publishes: `content.asset.uploaded`, `content.page.published`

### 10. Analytics & AI Domain
Owns: Event tracking, ML models, recommendations, reporting, attribution  
Services: analytics-service, reporting-service, recommendation-service, sentiment-analysis-service, price-optimization-service, ml-feature-store, personalization-service, data-pipeline-service, ad-service, event-tracking-service, attribution-service, clv-service, search-analytics-service  
Databases: Cassandra (events), Postgres (features, CLV), MongoDB (personalisation), ClickHouse (OLAP reporting), Weaviate (vectors), Neo4j (product graph), TimescaleDB (metrics)  
Stream processing: Apache Flink â€” `order-analytics` job (revenue aggregations to ClickHouse) + `fraud-detection` job (velocity checks across order and login streams)  
ML platform: MLflow â€” experiment tracking and model registry for recommendation and price-optimization models  
Consumes: All domain events for analytics aggregation

### 11. B2B Domain
Owns: Organisations, contracts, RFQ/quotes, approval workflows, EDI  
Services: organization-service, contract-service, quote-rfq-service, approval-workflow-service, b2b-credit-limit-service, edi-service, marketplace-seller-service, rfp-service, vendor-onboarding-service, purchase-requisition-service  
Databases: Postgres (all)  
Publishes: `b2b.quote.approved`, `b2b.contract.signed`

### 12. Integrations Domain
Owns: External system connectors (ERP, CRM, marketplaces, logistics providers)  
Services: erp-integration-service, marketplace-connector-service, social-commerce-service, crm-integration-service, payment-gateway-integration, logistics-provider-integration, tax-provider-integration, pim-integration-service, cdp-integration-service, accounting-integration-service, webhook-ingestion-service, etl-service, data-sync-service, ipaas-connector-service  
Databases: Stateless â€” transforms and forwards; no persistent store  
Pattern: Anti-Corruption Layer â€” translates external models to internal domain events

### 13. Affiliate Domain
Owns: Affiliate programs, referrals, influencer tracking, commissions  
Services: affiliate-service, referral-service, influencer-service, commission-payout-service, click-tracking-service, fraud-prevention-affiliate-service  
Databases: Postgres (all), Redis (click-tracking)  
Publishes: `affiliate.commission.earned`, `affiliate.referral.converted`

### 14. Marketplace Domain
Owns: Seller registration, listing approvals, marketplace commissions, disputes  
Services: seller-registration-service, listing-approval-service, marketplace-commission-service, dispute-resolution-service, seller-analytics-service, product-syndication-service, storefront-service, seller-payout-service  
Databases: Postgres (all)  
Publishes: `marketplace.seller.approved`, `marketplace.dispute.opened`

### 15. Gamification Domain
Owns: Points, badges, leaderboards, challenges, streaks  
Services: points-service, badge-service, leaderboard-service, challenge-service, reward-redemption-service, streak-service  
Databases: Redis (points, leaderboard, streaks), Postgres (badges, challenges, rewards)  
Publishes: `gamification.badge.earned`, `gamification.points.awarded`

### 16. Developer Platform Domain
Owns: API management, sandboxes, OAuth clients, developer analytics  
Services: api-management-service, sandbox-service, developer-portal-backend, oauth-client-service, api-analytics-service, webhook-management-service  
Databases: Postgres (all)  
Publishes: `developer.api.key.created`

### 17. Compliance Domain
Owns: Data retention, consent auditing, privacy requests, regulatory reporting  
Services: data-retention-service, consent-audit-service, privacy-request-service, compliance-reporting-service, data-lineage-service  
Databases: Postgres (all)  
Consumes: Events from all domains for audit trail

### 18. Sustainability Domain
Owns: Carbon tracking, eco scoring, green shipping, sustainability reporting  
Services: carbon-tracker-service, eco-score-service, green-shipping-service, sustainability-reporting-service, offset-service  
Databases: Postgres (all)  
Publishes: `sustainability.carbon.tracked`, `sustainability.offset.purchased`

### 19. Web Domain (Frontend Apps)
Owns: Customer-facing and operator-facing web/mobile interfaces  
Apps: storefront (Next.js 14 Â· :3000), admin-dashboard (React+Vite Â· :3001), seller-portal (Vue.js 3 Â· :3002), partner-portal (Angular 18 Â· :3003), mobile-app (React Native / Expo), developer-portal-ui (React+Vite Â· :3004)  
Consumes: All services via BFF layer (Web BFF, Mobile BFF, Partner BFF)

---

## Inter-Domain Dependency Matrix

Arrows show event consumption (`â†’`) or gRPC calls (`â‡’`).

| Producer Domain | Consumer Domain | Channel | Event / Call |
|---|---|---|---|
| Identity | Commerce | Kafka | `identity.user.registered` â†’ loyalty provisioning |
| Identity | Communications | Kafka | `identity.user.registered` â†’ welcome email |
| Identity | Analytics/AI | Kafka | `identity.user.registered` â†’ user tracking |
| Commerce | Supply Chain | Kafka | `commerce.order.placed` â†’ fulfillment |
| Commerce | Financial | Kafka | `commerce.payment.processed` â†’ accounting |
| Commerce | Communications | Kafka | `commerce.order.placed` â†’ order confirmation |
| Commerce | Loyalty | gRPC | checkout â‡’ loyalty (apply points) |
| Commerce | Promotions | gRPC | checkout â‡’ promotions (apply discounts) |
| Commerce | Analytics/AI | Kafka | `commerce.order.placed` â†’ conversion tracking |
| Supply Chain | Communications | Kafka | `supplychain.shipment.updated` â†’ tracking notification |
| Supply Chain | Analytics/AI | Kafka | shipment events â†’ delivery analytics |
| Financial | Communications | Kafka | `financial.invoice.created` â†’ email receipt |
| Catalog | Search | CDC | MongoDB product changes â†’ Elasticsearch index |
| Catalog | Analytics/AI | CDC | catalog changes â†’ reporting sync |
| Affiliate | Commerce | Kafka | referral applied at checkout |
| Affiliate | Financial | Kafka | `affiliate.commission.earned` â†’ payout |

---

## Context Map Legend

- Partnership: Identity â†” Commerce (shared kernel: user ID)
- Customer/Supplier: Commerce â†’ Supply Chain (Commerce places orders; Supply Chain fulfils)
- Conformist: Integrations â†’ External Systems (Integration adapts to external APIs)
- Anti-Corruption Layer: Integrations translates external models to internal domain events
- Open Host Service: Catalog exposes product data to all consumers via gRPC
- Published Language: Avro schemas in `events/` are the shared language across all domains

---

---

## Platform Concerns (Cross-Domain Infrastructure)

The following capabilities are not owned by any single domain but support all 13:

| Concern | Tool | Location |
|---|---|---|
| Service mesh + mTLS | Istio + Linkerd | `networking/istio/`, `networking/linkerd/` |
| eBPF CNI + network policy | Cilium + Calico | `networking/cilium/`, `networking/calico/` |
| Observability instrumentation | OpenTelemetry | All services â€” auto-injected by Istio |
| Metrics + alerting | Prometheus + Alertmanager + Grafana | `observability/` |
| Distributed tracing | Jaeger + Tempo | `observability/jaeger/`, `observability/tempo/` |
| Log aggregation | Loki + Fluentd + ELK + OpenSearch | `observability/` |
| Profiling | Grafana Pyroscope | `observability/pyroscope/` |
| SLO management | Pyrra + Sloth | `observability/pyrra/`, `observability/sloth/` |
| GitOps delivery | ArgoCD + Flux + Argo Rollouts | `gitops/` |
| Durable workflows | Temporal | `workflow/temporal/` |
| Stream processing | Apache Flink | `streaming/flink/` |
| Change data capture | Debezium | `streaming/debezium/` |
| Chaos engineering | Chaos Mesh + LitmusChaos | `chaos/` |
| Load testing | k6 + Locust + Gatling | `load-testing/` |
| Developer portal | Backstage | `backstage/` |
| ML experiment tracking | MLflow | `ml/mlflow/` |

---

## References

- [System Overview](system-overview.md)
- [Communication Patterns](communication-patterns.md)
- [Database Strategy](database-strategy.md)
- [Security Model](security-model.md)
- [ADR-004: Domain-Driven Folder Structure](../adr/004-domain-driven-folder-structure.md)
