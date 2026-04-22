# ShopOS — Enterprise Commerce Platform

An enterprise-grade, cloud-native commerce platform — 230 services (224 microservices + 6 frontend apps), 19 domains, 13 languages, full open source stack.

---

## Domains

| # | Domain | Services |
|---|---|---|
| 1 | Platform | 22 |
| 2 | Identity | 8 |
| 3 | Catalog | 12 |
| 4 | Commerce | 23 |
| 5 | Supply Chain | 13 |
| 6 | Financial | 11 |
| 7 | Customer Experience | 14 |
| 8 | Communications | 9 |
| 9 | Content | 8 |
| 10 | Analytics & AI | 13 |
| 11 | B2B | 7 |
| 12 | Integrations | 10 |
| 13 | Affiliate | 4 |
| | **Total** | **230** |

---

## Technology Stack

### Languages

| Language | Version | Used In |
|---|---|---|
| Go | 1.24 | Platform, Catalog, Commerce, Supply Chain, Financial, CX, Content, B2B, Integrations, Affiliate |
| Java | 21 (Spring Boot) | Identity, Catalog, Commerce, Supply Chain, Financial, B2B, Integrations |
| Kotlin | 2.x (Spring Boot) | Catalog, Commerce, Supply Chain, Financial, B2B |
| Python | 3.12 | Analytics & AI, Supply Chain, Communications, Content |
| Node.js | 22 | Platform, Catalog, Customer Experience, Communications, Content, Integrations |
| C# | .NET 9 | Commerce (cart, return-refund) |
| Rust | 1.80 | Identity (auth), Commerce (shipping) |
| Scala | 3.x | Analytics & AI (reporting) |

### Databases

| Database | Version | Role |
|---|---|---|
| PostgreSQL | 16 | Primary transactional store — 100+ services |
| MongoDB | 8.0 | Document store — catalog, CMS, reviews, tracking |
| Redis | 7 | Cache, sessions, pub/sub, ephemeral data |
| Cassandra | 5.0 | Time-series analytics events |
| ScyllaDB | 6.1 | High-throughput Cassandra-compatible analytics |
| Elasticsearch | 8.15.3 | Full-text search, faceted filtering |
| OpenSearch | 2.17 | Log analytics, audit trail, security events |
| ClickHouse | 24.8 | OLAP — orders, events, revenue aggregation |
| Weaviate | 1.26 | Vector database — semantic search, AI recommendations |
| Neo4j | 5.23 | Graph database — product recommendations |
| MinIO | latest | Object storage — images, videos, PDFs, exports |
| etcd | 3.5 | Distributed configuration |
| Memcached | 1.6 | High-throughput simple caching |

### Messaging & Streaming

| Tool | Version | Role |
|---|---|---|
| Apache Kafka (Confluent) | 7.7.1 | Primary event streaming — domain events |
| Apache ZooKeeper | 7.7.1 | Kafka coordination |
| Schema Registry (Confluent) | 7.7.1 | Avro schema enforcement |
| RabbitMQ | 3.13 | Task queues, delayed messages, RPC |
| NATS JetStream | 2.10 | Low-latency pub/sub — chat, real-time notifications |
| Debezium | 2.7 | Change Data Capture from Postgres + MongoDB → Kafka |
| Apache Flink | 1.20 | Real-time stream processing — order analytics, fraud detection |
| ksqlDB | latest | Streaming SQL on Kafka |
| Strimzi | latest | Kafka on Kubernetes operator |
| AKHQ | latest | Kafka UI |
| kafka-ui | latest | Kafka management |

### API & Communication

| Tool | Role |
|---|---|
| gRPC | Synchronous inter-service communication (primary) |
| Protocol Buffers (protobuf) | Service contracts — 58 `.proto` files across 14 domains |
| Buf CLI | Protobuf linting, breaking-change detection, multi-language codegen |
| REST/HTTP | External-facing APIs (BFFs, webhooks, health endpoints) |
| GraphQL | Unified query API via graphql-gateway |
| WebSocket | Real-time — live chat, in-app notifications |
| Avro | Kafka event schema format — 20 event schemas |

### Container & Kubernetes

| Tool | Version | Role |
|---|---|---|
| Docker | latest | Container runtime, multi-stage builds |
| Kubernetes | 1.31 | Container orchestration |
| Helm | 3.x | 230 per-service charts + 30 tool charts |
| KEDA | 2.15 | Kafka/Redis-driven autoscaling (alongside HPA) |
| Velero | 7.x | Kubernetes backup and restore |
| Skaffold | latest | Local dev hot-reload |
| Tilt | latest | Local dev hot-reload (alternative) |

### Infrastructure as Code

| Tool | Version | Role |
|---|---|---|
| Terraform | 1.9 | EKS, GKE, AKS cluster provisioning + Jenkins VM |
| OpenTofu | 1.8 | Open source Terraform alternative (same targets) |
| Crossplane | 1.17 | Kubernetes-native IaC — database and cloud resource claims |
| Ansible | 2.17 | Kubernetes node bootstrapping |
| Docker Compose | v2 | Full local stack (230 services + infra) |

### CI/CD

| Tool | Pipelines | Role |
|---|---|---|
| Jenkins | 12 | Primary CI server — full declarative pipeline suite |
| Drone CI | 12 | Mirror of Jenkins — same stages, Drone syntax |
| Woodpecker CI | 12 | Drone-compatible fork — drop-in replacement |
| Dagger | 12 | Portable Go SDK pipelines — run anywhere |
| Tekton | 12 | Kubernetes-native CRD-based pipelines |
| Concourse CI | 12 | DAG resource/job pipelines |
| GitLab CI | 12 | `.gitlab-ci.yml` pipelines for GitLab SCM |
| GitHub Actions | 12 | `.github/workflows/` — native GitHub integration |
| CircleCI | 12 | `version: 2.1` orb-based pipelines |
| GoCD | 12 | Stage/job pipelines with manual approval gates |
| Travis CI | 12 | Stage-based pipelines with branch filters |
| Harness CI | 12 | Enterprise CI/CD with built-in CD stages |
| Azure DevOps | 12 | `azure-pipelines.yml` — native Azure integration |
| AWS CodePipeline | 12 | `buildspec.yml` + CodePipeline JSON definitions |
| GCP Cloud Build | 12 | `cloudbuild.yaml` — native GCP integration |
| Argo Workflows | — | Kubernetes-native CI + ML training DAGs |
| Argo Events | — | GitHub webhook → pipeline triggers |

### GitOps

| Tool | Version | Role |
|---|---|---|
| ArgoCD | 2.12 | GitOps continuous delivery — App-of-Apps pattern |
| Flux CD | 2.x | Alternative GitOps controller |
| Argo Rollouts | latest | Canary and blue/green deployments |
| Flagger | latest | Progressive delivery with metrics-based promotion |

### Observability

| Tool | Version | Role |
|---|---|---|
| OpenTelemetry | latest | Instrumentation SDK + Collector |
| Prometheus | 2.54 | Metrics scraping and alerting |
| Alertmanager | 0.27 | Alert routing and deduplication |
| Grafana | 11.x | Dashboards |
| Grafana Loki | 3.x | Log aggregation |
| Grafana Tempo | 2.x | Distributed tracing |
| Grafana Pyroscope | 1.7 | Continuous profiling |
| Jaeger | 2.x | Distributed tracing (alternative) |
| Zipkin | 3.4 | Lightweight tracing |
| Thanos | 0.36 | Long-term Prometheus storage |
| VictoriaMetrics | 1.x | Prometheus-compatible metrics alternative |
| Fluent Bit | 3.x | Log shipping from pods |
| Fluentd | 1.17 | Log aggregation and transformation |
| Elasticsearch | 8.15.3 | Log search (ELK stack) |
| Kibana | 8.15.3 | Log dashboards (ELK stack) |
| Logstash | 8.15.3 | Log ingestion pipeline |
| OpenSearch | 2.17 | Log analytics (alternative to ELK) |
| OpenSearch Dashboards | 2.17 | Log dashboards |
| Sentry OSS | latest | Error tracking and source maps |
| GlitchTip | latest | Sentry-compatible error tracking alternative |
| Uptime Kuma | 1.23 | Uptime monitoring and status pages |
| Pyrra | latest | SLO management |
| Sloth | latest | SLO/SLI generation from Prometheus rules |
| OpenCost | 1.43 | Kubernetes cost visibility per namespace/service |
| Goldilocks | 9.x | VPA-based resource right-sizing recommendations |
| Robusta | 0.13 | Kubernetes alerting enricher (Slack integration) |
| kube-state-metrics | latest | Kubernetes object metrics |
| node-exporter | latest | Node-level hardware metrics |
| Pushgateway | latest | Batch job metrics |
| Blackbox Exporter | latest | Endpoint probing |

### Security

| Tool | Role |
|---|---|
| HashiCorp Vault | Secrets management, PKI, dynamic credentials |
| Keycloak | Identity and Access Management, SSO, OIDC |
| Dex | OIDC federation |
| Authentik | Identity provider (alternative to Keycloak) |
| SPIFFE/SPIRE | Workload identity — X.509 SVIDs for mTLS |
| OpenFGA | Relationship-based authorisation (Google Zanzibar model) |
| OPA / Gatekeeper | Policy as code — Rego admission policies |
| Kyverno | Kubernetes admission controller policies |
| Kubewarden | Wasm-based policy engine |
| Falco | Runtime threat detection |
| Tetragon | eBPF security enforcement |
| Tracee | eBPF event collection and analysis |
| Istio | Service mesh — mTLS, traffic management |
| Linkerd | Lightweight service mesh (alternative) |
| Cilium | eBPF CNI — network policies, L7 visibility |
| Calico | CNI alternative with network policy |
| cert-manager | Automatic TLS certificate provisioning |
| Coraza WAF | OWASP WAF (ModSecurity rules) |
| Trivy | Container and IaC vulnerability scanning |
| Grype | CVE scanner for container images |
| Semgrep | SAST — custom security rules |
| SonarQube | Static code analysis and quality gates |
| Checkov | IaC security scanning (Terraform/Helm/K8s) |
| KICS | IaC scanning — extended rule set |
| OWASP ZAP | DAST — dynamic application security testing |
| Nuclei | CVE template-based scanning |
| kube-bench | CIS Kubernetes benchmark |
| kube-hunter | Kubernetes penetration testing |
| Cosign (Sigstore) | Container image signing |
| Rekor (Sigstore) | Transparency log for signed artifacts |
| Fulcio (Sigstore) | Certificate authority for Sigstore |
| Syft | SBOM (Software Bill of Materials) generation |
| CycloneDX | SBOM format standard |

### Networking & Service Mesh

| Tool | Version | Role |
|---|---|---|
| Traefik | 3.1 | Edge router — ingress, automatic TLS, service discovery |
| Istio | 1.23 | Service mesh — mTLS, traffic management, canary |
| Linkerd | 2.x | Lightweight service mesh (alternative) |
| Cilium | 1.16 | eBPF CNI — NetworkPolicy + Hubble observability |
| Calico | 3.28 | CNI and NetworkPolicy alternative |
| Consul | 1.19 | Service discovery, health checking, K/V config |
| Kong | 3.x | API gateway (alternative to Traefik) |
| NGINX | 1.27 | Reverse proxy and static file serving |

### Artifact & Container Registry

| Tool | Version | Role |
|---|---|---|
| Harbor | 2.11 | Container registry with vulnerability scanning |
| Nexus | 3.71 | Artifact repository (Maven, npm, PyPI, Go, Docker) |
| Gitea | 1.22 | Self-hosted Git server for GitOps |
| Zot | 2.x | OCI-native container registry |
| ChartMuseum | 0.16 | Helm chart repository |

### Workflow Orchestration

| Tool | Version | Role |
|---|---|---|
| Temporal | 1.24 | Durable workflow engine — checkout sagas, billing cycles, KYC |
| Argo Workflows | 3.x | DAG-based workflow execution on Kubernetes |

### ML Platform

| Tool | Version | Role |
|---|---|---|
| MLflow | 2.16 | Experiment tracking, model registry, model serving |
| Apache Flink | 1.20 | Real-time feature computation and stream ML |
| Weaviate | 1.26 | Vector store for RAG and semantic search |
| Neo4j | 5.23 | Graph-based recommendation engine |

### Developer Experience

| Tool | Role |
|---|---|
| Devcontainer | VS Code / GitHub Codespaces — all 8 languages pre-installed |
| Skaffold | Local Kubernetes dev hot-reload |
| Tilt | Local Kubernetes dev with live_update (alternative) |
| Backstage | Internal developer portal — service catalog, API docs |
| Buf CLI | Protobuf workflow — lint, format, codegen, breaking detection |
| k9s | Terminal-based Kubernetes UI |
| grpcurl | gRPC API testing |

### Chaos & Load Testing

| Tool | Role |
|---|---|
| Chaos Mesh | Kubernetes-native chaos injection |
| Litmus | Chaos engineering framework |
| k6 | JavaScript-based load testing |
| Locust | Python-based distributed load testing |
| Gatling | Scala-based load and performance testing |

---

## Services

### 1. Platform (22 services)

| Service | Language | Responsibility |
|---|---|---|
| api-gateway | Go | Single entry point — routing, rate limiting, JWT validation |
| web-bff | Go | Backend-for-frontend for web clients |
| mobile-bff | Node.js | Backend-for-frontend for mobile clients |
| partner-bff | Go | Backend-for-frontend for external partner APIs |
| config-service | Go | Centralised runtime configuration via etcd |
| feature-flag-service | Go | Feature toggles and A/B flag evaluation |
| rate-limiter-service | Go | Distributed rate limiting via Redis |
| health-check-service | Go | Deep health checks across all service dependencies |
| saga-orchestrator | Go | Distributed transaction coordination (Saga pattern) |
| event-store-service | Go | Append-only event log for event sourcing |
| cache-warming-service | Go | Pre-populates caches on deploy or cache miss |
| webhook-service | Go | Outbound webhook delivery to partner systems |
| scheduler-service | Go | Cron-based and delayed job scheduling |
| worker-job-queue | Go | Async background task processing |
| audit-service | Java | Immutable audit log of all state-changing operations |
| load-generator | Python | Simulates realistic user traffic |
| admin-portal | Java | Internal administration web UI |
| graphql-gateway | Go | Unified GraphQL API aggregating domain services |
| dead-letter-service | Go | Dead-letter queue processing and replay |
| geolocation-service | Go | IP-to-location and address geolocation |
| event-replay-service | Go | Replays historical events for projections and recovery |
| tenant-service | Go | Multi-tenancy management and tenant isolation |

---

### 2. Identity (8 services)

| Service | Language | Responsibility |
|---|---|---|
| auth-service | Rust | OAuth2/OIDC token issuance and validation |
| user-service | Java | User profiles, preferences, and account management |
| session-service | Go | JWT session lifecycle backed by Redis |
| permission-service | Go | RBAC/ABAC policy evaluation |
| mfa-service | Go | TOTP/WebAuthn multi-factor authentication |
| gdpr-service | Go | Data subject requests — access, erasure, portability |
| api-key-service | Go | API key issuance, rotation, and scoping |
| device-fingerprint-service | Go | Device recognition and trust scoring |

---

### 3. Catalog (12 services)

| Service | Language | Responsibility |
|---|---|---|
| product-catalog-service | Go | Core product data — listing, detail, variants |
| category-service | Go | Product taxonomy and category hierarchy |
| brand-service | Go | Brand profiles and brand-level filtering |
| pricing-service | Java | Dynamic pricing rules, tiers, and overrides |
| inventory-service | Go | Stock levels, reservations, and availability |
| bundle-service | Go | Product bundles and kit definitions |
| configurator-service | Node.js | Configurable products (colour, size, material) |
| subscription-product-service | Kotlin | Subscription and recurring product management |
| search-service | Python | Full-text and faceted search via Elasticsearch |
| seo-service | Node.js | Meta tags, canonical URLs, and structured data |
| product-import-service | Go | Bulk product import via CSV/XLSX/API |
| price-list-service | Java | Customer-group and channel-specific price lists |

---

### 4. Commerce (23 services)

| Service | Language | Responsibility |
|---|---|---|
| cart-service | C# | Shopping cart backed by Redis |
| checkout-service | Go | Orchestrates the full checkout flow |
| order-service | Kotlin | Order lifecycle management and state machine |
| payment-service | Java | Payment processing, capture, and refund |
| shipping-service | Rust | Shipping cost calculation and carrier selection |
| currency-service | Node.js | Real-time currency conversion |
| tax-service | Go | Tax rate lookup and calculation |
| promotions-service | Java | Coupon codes, discount rules, flash sales |
| loyalty-service | Go | Points accrual, tier management, and redemption |
| return-refund-service | C# | Returns lifecycle and refund processing |
| subscription-billing-service | Go | Recurring billing and subscription renewals |
| fraud-detection-service | Python | ML-based fraud scoring on transactions |
| wallet-service | Go | Customer credit wallet and prepaid balance |
| ab-testing-service | Go | Experiment assignment and flag evaluation |
| gift-card-service | Go | Gift card issuance, balance, and redemption |
| address-validation-service | Go | Postal address validation and normalisation |
| digital-goods-service | Go | Digital product delivery and licence management |
| voucher-service | Go | Single-use and multi-use voucher management |
| pre-order-service | Go | Pre-order capture, deposits, and release scheduling |
| backorder-service | Go | Backorder management and fulfilment queue |
| waitlist-service | Go | Customer waitlists for out-of-stock items |
| flash-sale-service | Go | Time-limited flash sale events and inventory locks |
| bnpl-service | Go | Buy-now-pay-later instalment plan management |

---

### 5. Supply Chain (13 services)

| Service | Language | Responsibility |
|---|---|---|
| vendor-service | Java | Supplier profiles, contracts, and onboarding |
| purchase-order-service | Kotlin | PO creation, approval, and tracking |
| warehouse-service | Go | Warehouse management — bins, zones, movements |
| fulfillment-service | Go | Multi-warehouse order routing and pick/pack/ship |
| tracking-service | Node.js | Real-time shipment tracking from carriers |
| label-service | Python | Shipping label generation (PDF/ZPL) |
| carrier-integration-service | Go | Adapter layer for UPS, FedEx, DHL, etc. |
| demand-forecast-service | Python | ML-based inventory demand forecasting |
| customs-duties-service | Go | International trade — HS codes, duties calculation |
| returns-logistics-service | Go | Reverse logistics and returned goods routing |
| supplier-portal-service | Java | Self-service portal for vendor onboarding and PO management |
| cold-chain-service | Go | Cold-chain temperature monitoring and compliance tracking |
| supplier-rating-service | Go | Supplier performance scoring and rating aggregation |

---

### 6. Financial (11 services)

| Service | Language | Responsibility |
|---|---|---|
| invoice-service | Java | PDF invoice generation and delivery |
| accounting-service | Kotlin | General ledger — journal entries and chart of accounts |
| payout-service | Java | Vendor and seller payout scheduling |
| reconciliation-service | Kotlin | Payment-to-order reconciliation |
| tax-reporting-service | Go | VAT/GST report generation |
| expense-management-service | Go | Internal expense tracking and approval |
| credit-service | Go | Customer credit lines and scoring |
| kyc-aml-service | Java | Know Your Customer and Anti-Money Laundering |
| budget-service | Go | Departmental budget tracking and alerts |
| chargeback-service | Java | Payment chargeback intake and dispute resolution |
| revenue-recognition-service | Kotlin | ASC 606/IFRS 15 compliant revenue recognition |

---

### 7. Customer Experience (14 services)

| Service | Language | Responsibility |
|---|---|---|
| review-rating-service | Node.js | Product reviews, ratings, and moderation |
| qa-service | Node.js | Product question and answer threads |
| wishlist-service | Go | Named wishlists and sharing |
| compare-service | Go | Side-by-side product comparison |
| recently-viewed-service | Go | Browsing history and recently viewed products |
| support-ticket-service | Java | Customer support tickets and case management |
| live-chat-service | Go | Real-time customer support chat via WebSocket |
| consent-management-service | Node.js | Cookie consent and marketing preferences |
| age-verification-service | Go | Age gate for age-restricted products |
| survey-service | Node.js | Post-purchase and NPS surveys |
| feedback-service | Node.js | General feedback collection and routing |
| price-alert-service | Go | Customer price drop alerts |
| back-in-stock-service | Go | Back-in-stock notifications |
| gift-registry-service | Go | Gift registry creation, sharing, and purchase tracking |

---

### 8. Communications (9 services)

| Service | Language | Responsibility |
|---|---|---|
| notification-orchestrator | Node.js | Routes notifications to the correct channel |
| email-service | Python | Transactional email delivery |
| sms-service | Node.js | SMS delivery abstraction |
| push-notification-service | Go | Mobile push via FCM and APNs |
| template-service | Node.js | Message template management |
| in-app-notification-service | Go | Real-time in-app notifications via WebSocket |
| digest-service | Go | Batches and schedules notification digests |
| whatsapp-service | Node.js | WhatsApp Business API messaging |
| chatbot-service | Python | Conversational chatbot engine |

---

### 9. Content (8 services)

| Service | Language | Responsibility |
|---|---|---|
| media-asset-service | Go | Image/video upload, storage (MinIO), CDN URLs |
| image-processing-service | Python | Resizing, compression, and format conversion |
| document-service | Java | PDF generation — invoices, packing slips, reports |
| cms-service | Node.js | Headless CMS — pages, blog posts, banners |
| video-service | Go | Video upload and HLS streaming |
| sitemap-service | Go | XML sitemap generation for SEO |
| i18n-l10n-service | Go | Translation strings and locale management |
| data-export-service | Python | CSV/Excel data exports for merchants |

---

### 10. Analytics & AI (13 services)

| Service | Language | Responsibility |
|---|---|---|
| analytics-service | Python | Real-time event ingestion and metric aggregation |
| reporting-service | Scala | Batch reporting and dashboards |
| recommendation-service | Python | Collaborative filtering product recommendations |
| sentiment-analysis-service | Python | NLP sentiment scoring on reviews |
| price-optimization-service | Python | ML-driven dynamic pricing suggestions |
| ml-feature-store | Python | Centralised ML feature engineering and serving |
| personalization-service | Python | Personalised content and product ranking |
| data-pipeline-service | Python | ETL/ELT pipelines — raw events → data warehouse |
| ad-service | Java | Context-aware advertisement serving |
| event-tracking-service | Python | Behavioural event capture and stream processing |
| attribution-service | Python | Multi-touch marketing attribution modelling |
| clv-service | Python | Customer lifetime value scoring and cohort analysis |
| search-analytics-service | Python | Search query analytics and relevance metrics |

---

### 11. B2B (7 services)

| Service | Language | Responsibility |
|---|---|---|
| organization-service | Java | B2B company accounts and team management |
| contract-service | Kotlin | B2B pricing contracts and SLA terms |
| quote-rfq-service | Go | Request for quote (RFQ) creation and negotiation |
| approval-workflow-service | Go | Multi-level purchase approval workflows |
| b2b-credit-limit-service | Go | Credit limit management for business accounts |
| edi-service | Java | Electronic Data Interchange (EDI 850/856/810) |
| marketplace-seller-service | Java | Marketplace seller onboarding and commission management |

---

### 12. Integrations (10 services)

| Service | Language | Responsibility |
|---|---|---|
| erp-integration-service | Java | Bi-directional sync with ERP systems |
| marketplace-connector-service | Go | Publish products to Amazon, eBay, etc. |
| social-commerce-service | Node.js | Instagram Shopping and TikTok Shop integration |
| crm-integration-service | Go | Sync customers and orders to CRM systems |
| payment-gateway-integration | Go | Adapter layer for Stripe, Adyen, PayPal |
| logistics-provider-integration | Go | Adapter layer for UPS/FedEx/DHL rate APIs |
| tax-provider-integration | Go | Adapter layer for tax calculation providers |
| pim-integration-service | Go | Bi-directional sync with external PIM systems |
| cdp-integration-service | Go | Streams customer events to Customer Data Platform |
| accounting-integration-service | Java | Pushes financial transactions to external accounting systems |

---

### 13. Affiliate (4 services)

| Service | Language | Responsibility |
|---|---|---|
| affiliate-service | Go | Affiliate partner accounts, tracking links, commission tiers |
| referral-service | Go | Customer referral codes, conversion tracking, reward triggers |
| influencer-service | Go | Influencer campaign management and UTM attribution |
| commission-payout-service | Go | Commission aggregation, tax rules, and payout batching |

### 14–18. New Domains (marketplace, gamification, developer-platform, compliance, sustainability)

See [CLAUDE.md](CLAUDE.md) for full service registry of all 224 microservices.

### 19. Web Frontend (6 apps)

| App | Framework | Responsibility |
|---|---|---|
| storefront | Next.js 14 (React/TS) | Customer-facing shopping experience — SSR for SEO |
| admin-dashboard | React + Vite (TS) | Admin and merchant management portal |
| seller-portal | Vue.js 3 (TS) | Marketplace seller portal — listings, analytics, payouts |
| partner-portal | Angular 18 (TS) | B2B partner portal — contracts, orders, invoices |
| mobile-app | React Native / Expo (TS) | iOS + Android customer app |
| developer-portal-ui | React + Vite (TS) | Developer portal — API docs, sandbox, OAuth apps |

---

## Docs

- [Getting Started](GETTING_STARTED.md) — complete from-scratch setup guide
- [Architecture](docs/architecture/) — system overview, domain map, communication patterns, database strategy, security model
- [ADRs](docs/adr/) — architecture decision records (001–006)
- [CI/CD](ci/README.md) — all 15 CI platform pipelines
- [Helm Charts](helm/README.md) — per-service Kubernetes deployment
- [Infrastructure](infra/README.md) — Terraform, OpenTofu, Crossplane, Ansible
- [GitOps](gitops/README.md) — ArgoCD, Flux, Argo Rollouts
- [Observability](observability/README.md) — metrics, logs, traces, SLOs
- [Security](security/README.md) — 50+ security tools and policies
- [Kubernetes](kubernetes/README.md) — namespaces, RBAC, network policies, KEDA, Velero
- [Messaging](messaging/README.md) — Kafka, RabbitMQ, NATS, Debezium, Flink
- [Networking](networking/README.md) — Istio, Cilium, Traefik, Consul
- [Databases](databases/README.md) — ClickHouse, Weaviate, Neo4j, ScyllaDB, OpenSearch
- [Streaming](streaming/README.md) — Debezium CDC + Apache Flink
- [Registry](registry/README.md) — Harbor, Nexus, MinIO, ChartMuseum
- [ML Platform](ml/README.md) — MLflow
- [Chaos Engineering](chaos/README.md) — Chaos Mesh, LitmusChaos
- [Load Testing](load-testing/README.md) — k6, Locust, Gatling
- [Proto / gRPC](proto/README.md) — 58 protobuf files, Buf CLI
- [Kafka Events](events/README.md) — 20 Avro event schemas
- [Backstage](backstage/README.md) — developer portal
- [Workflow / Temporal](workflow/README.md) — durable workflow engine

## License

Apache 2.0
