# ShopOS — Enterprise Commerce Platform

An enterprise-grade, cloud-native commerce platform — 230 services, 19 domains, 13 languages, full open source stack.

---

## Domains

| # | Domain | Services |
|---|---|---|
| 1 | Platform | 27 |
| 2 | Identity | 11 |
| 3 | Catalog | 15 |
| 4 | Commerce | 28 |
| 5 | Supply Chain | 17 |
| 6 | Financial | 15 |
| 7 | Customer Experience | 17 |
| 8 | Communications | 12 |
| 9 | Content | 9 |
| 10 | Analytics & AI | 13 |
| 11 | B2B | 10 |
| 12 | Integrations | 14 |
| 13 | Affiliate | 6 |
| 14 | Marketplace | 8 |
| 15 | Gamification | 6 |
| 16 | Developer Platform | 6 |
| 17 | Compliance | 5 |
| 18 | Sustainability | 5 |
| 19 | Web | 6 |
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
| TimescaleDB | 2.15 | Time-series metrics — service metrics, inventory events, page views |
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
| Conduktor Gateway | 3.3.0 | Kafka policy proxy — schema enforcement, PII masking, rate limiting |
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
| OpenAPI 3.1 | API specification — api-gateway, admin-api, developer-platform-api |
| Spectral | OpenAPI linting with custom ruleset |
| Hurl | HTTP integration testing (health, auth, catalog, checkout flows) |

### Container & Kubernetes

| Tool | Version | Role |
|---|---|---|
| Docker | latest | Container runtime, multi-stage builds |
| Kubernetes | 1.31 | Container orchestration |
| Helm | 3.x | 230 per-service charts + 30+ tool charts |
| KEDA | 2.15 | Kafka/Redis-driven autoscaling (alongside HPA) |
| Velero | 7.x | Kubernetes backup and restore |
| Skaffold | latest | Local dev hot-reload |
| Tilt | latest | Local dev hot-reload (alternative) |
| Kaniko | latest | Rootless container builds inside Kubernetes pods |

### Infrastructure as Code

| Tool | Version | Role |
|---|---|---|
| Terraform | 1.9 | EKS, GKE, AKS cluster provisioning |
| OpenTofu | 1.8 | Open source Terraform alternative (same targets) |
| Crossplane | 1.17 | Kubernetes-native IaC — database and cloud resource claims |
| Ansible | 2.17 | Kubernetes node bootstrapping |
| Terrascan | latest | IaC security scanning — Terraform + Helm |
| Docker Compose | v2 | Full local stack (230 services + infra) |

### CI/CD

| Tool | Pipelines | Role |
|---|---|---|
| Jenkins | 14 | Primary CI server — build, test, security, quality, tooling pipelines |
| Drone CI | 12 | Mirror of Jenkins — same stages, Drone syntax |
| Woodpecker CI | 12 | Drone-compatible fork — drop-in replacement |
| Dagger | 12 | Portable Go SDK pipelines — run anywhere |
| Tekton | 12 | Kubernetes-native CRD-based pipelines |
| Concourse CI | 12 | DAG resource/job pipelines |
| GitLab CI | 12 | `.gitlab-ci.yml` pipelines for GitLab SCM |
| GitHub Actions | 12 | `ci/github-actions/` — auto-trigger disabled |
| CircleCI | 12 | `version: 2.1` orb-based pipelines |
| GoCD | 12 | Stage/job pipelines with manual approval gates |
| Travis CI | 12 | Stage-based pipelines with branch filters |
| Harness CI | 12 | Enterprise CI/CD with built-in CD stages |
| Azure DevOps | 12 | `azure-pipelines.yml` — native Azure integration |
| AWS CodePipeline | 12 | `buildspec.yml` + CodePipeline JSON definitions |
| GCP Cloud Build | 12 | `cloudbuild.yaml` — native GCP integration |
| Argo Workflows | — | Kubernetes-native CI + ML training DAGs |
| Argo Events | — | GitHub webhook → pipeline triggers |

Jenkins pipelines: `build.Jenkinsfile`, `test.Jenkinsfile`, `security.Jenkinsfile`, `gitops.Jenkinsfile`, `databases.Jenkinsfile`, `messaging.Jenkinsfile`, `observability.Jenkinsfile`, `registry.Jenkinsfile`, `infra.Jenkinsfile`, `install-tools.Jenkinsfile`, `deploy.Jenkinsfile`, `teardown.Jenkinsfile`, `tooling.Jenkinsfile`, `api-quality.Jenkinsfile`

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
| Botkube | 1.13 | Kubernetes event alerts to Slack |
| k8sGPT | latest | AI-powered Kubernetes diagnostics |
| Plausible | latest | Privacy-friendly web analytics (GDPR, no cookies) |
| OpenReplay | latest | Self-hosted session replay |
| Goldilocks | 9.x | VPA-based resource right-sizing recommendations |
| kube-state-metrics | latest | Kubernetes object metrics |
| node-exporter | latest | Node-level hardware metrics |

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
| Terrascan | IaC security scanning with SARIF output |
| OWASP ZAP | DAST — dynamic application security testing |
| Nuclei | CVE template-based scanning |
| kube-bench | CIS Kubernetes benchmark |
| kube-hunter | Kubernetes penetration testing |
| Cosign (Sigstore) | Container image signing |
| Rekor (Sigstore) | Transparency log for signed artifacts |
| Fulcio (Sigstore) | Certificate authority for Sigstore |
| Syft | SBOM (Software Bill of Materials) generation |
| CycloneDX | SBOM format standard |
| Dependency-Track | SBOM analysis and CVE tracking |
| DefectDojo | Vulnerability management and finding aggregation |
| Teleport | Zero-trust SSH and Kubernetes access |

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

### Analytics Data Stack

| Tool | Role |
|---|---|
| Apache Airflow | DAG-based workflow orchestration — daily ETL, fraud retrain |
| Apache Spark | Batch processing — order aggregation, user RFM segmentation |
| dbt | SQL transformations — staging, commerce, catalog models |
| Apache Superset | Data exploration and BI dashboards |
| Apache Atlas | Data catalog and governance |
| Marquez | OpenLineage data lineage tracking |
| Great Expectations | Data quality assertion suites |
| MLflow | 2.16 — Experiment tracking, model registry |
| Apache Flink | 1.20 — Real-time feature computation and stream ML |
| Weaviate | 1.26 — Vector store for RAG and semantic search |
| Neo4j | 5.23 — Graph-based recommendation engine |

### Database Management

| Tool | Role |
|---|---|
| pgAdmin | PostgreSQL web UI |
| Mongo Express | MongoDB web UI |
| Redis Commander | Redis web UI |
| Bytebase | Database schema change management and version control |

### Contract & Integration Testing

| Tool | Role |
|---|---|
| Pact | Consumer-driven contract testing — order↔cart, checkout↔payment |
| Testcontainers | Ephemeral database/broker containers for integration tests |
| Toxiproxy | Network fault injection for resilience testing |

### Build Tooling

| Tool | Role |
|---|---|
| Earthly | Reproducible polyglot builds — all 13 languages |
| Ko | Direct Go → OCI image builder for 100+ Go services |
| Kaniko | Rootless Docker builds inside Kubernetes pods |
| Score | Cloud-agnostic workload specification |
| DevPod | Cloud development environments (alternative to Codespaces) |

### Developer Experience

| Tool | Role |
|---|---|
| Devcontainer | VS Code / GitHub Codespaces — all languages pre-installed |
| Skaffold | Local Kubernetes dev hot-reload |
| Tilt | Local Kubernetes dev with live_update (alternative) |
| Backstage | Internal developer portal — service catalog, API docs |
| Buf CLI | Protobuf workflow — lint, format, codegen, breaking detection |
| Teleport | Zero-trust SSH + Kubernetes access for developers and ops |
| Botkube | Kubernetes alerts to Slack — pod failures, deployments |
| k8sGPT | AI-powered Kubernetes diagnostics operator |
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

### 1. Platform (27 services)

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
| notification-preferences-service | Go | User notification channel and frequency preferences |
| circuit-breaker-service | Go | Circuit breaker state management via Redis |
| idempotency-service | Go | Idempotency key storage and deduplication |
| correlation-id-service | Go | Request correlation ID propagation across services |
| data-masking-service | Go | PII masking and tokenisation for logs and events |

---

### 2. Identity (11 services)

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
| sso-service | Go | Single sign-on federation and session bridging |
| password-policy-service | Go | Password complexity rules and breach detection |
| bot-detection-service | Go | Bot and credential-stuffing attack detection |

---

### 3. Catalog (15 services)

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
| product-label-service | Go | Product labels, badges, and promotional tags |
| variant-service | Go | Product variant matrix management |
| stock-reservation-service | Go | Atomic stock reservations via Redis |

---

### 4. Commerce (28 services)

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
| split-payment-service | Go | Multi-method payment splitting at checkout |
| installment-service | Go | Fixed-term instalment payment plans |
| dynamic-pricing-service | Go | Real-time demand-based price adjustment |
| coupon-service | Go | Coupon lifecycle management and validation |
| order-amendment-service | Go | Post-placement order modification and repricing |

---

### 5. Supply Chain (17 services)

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
| route-optimization-service | Go | Delivery route optimisation across carriers |
| packaging-service | Go | Packaging material selection and cost optimisation |
| cross-dock-service | Go | Cross-docking flow management — inbound to outbound |
| duty-drawback-service | Go | Import duty drawback claim management |

---

### 6. Financial (15 services)

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
| escrow-service | Go | Marketplace escrow — fund holding and release |
| forex-service | Go | Foreign exchange rate management and hedging |
| audit-trail-service | Java | Immutable financial audit trail |
| dunning-service | Go | Failed payment retry and dunning communication |

---

### 7. Customer Experience (17 services)

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
| loyalty-tier-service | Go | Loyalty tier evaluation and benefit management |
| accessibility-service | Node.js | WCAG accessibility audit and remediation guidance |
| return-portal-service | Go | Self-service customer return initiation |

---

### 8. Communications (12 services)

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
| telegram-service | Node.js | Telegram Bot API messaging |
| voice-service | Go | Voice call notifications via Twilio-compatible API |
| webhook-delivery-service | Go | Reliable outbound webhook delivery with retries |

---

### 9. Content (9 services)

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
| ab-content-service | Go | A/B content variant management and assignment |

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

### 11. B2B (10 services)

| Service | Language | Responsibility |
|---|---|---|
| organization-service | Java | B2B company accounts and team management |
| contract-service | Kotlin | B2B pricing contracts and SLA terms |
| quote-rfq-service | Go | Request for quote (RFQ) creation and negotiation |
| approval-workflow-service | Go | Multi-level purchase approval workflows |
| b2b-credit-limit-service | Go | Credit limit management for business accounts |
| edi-service | Java | Electronic Data Interchange (EDI 850/856/810) |
| marketplace-seller-service | Java | Marketplace seller onboarding and commission management |
| rfp-service | Go | Request for proposal (RFP) management |
| vendor-onboarding-service | Java | Structured vendor onboarding workflow |
| purchase-requisition-service | Kotlin | Internal purchase requisition and approval |

---

### 12. Integrations (14 services)

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
| webhook-ingestion-service | Go | Receives and validates inbound partner webhooks |
| etl-service | Python | Extract-transform-load for third-party data sources |
| data-sync-service | Go | Near-real-time bidirectional data synchronisation |
| ipaas-connector-service | Go | iPaaS integration connector (Zapier/Make-compatible) |

---

### 13. Affiliate (6 services)

| Service | Language | Responsibility |
|---|---|---|
| affiliate-service | Go | Affiliate partner accounts, tracking links, commission tiers |
| referral-service | Go | Customer referral codes, conversion tracking, reward triggers |
| influencer-service | Go | Influencer campaign management and UTM attribution |
| commission-payout-service | Go | Commission aggregation, tax rules, and payout batching |
| click-tracking-service | Go | High-throughput affiliate click recording via Redis |
| fraud-prevention-affiliate-service | Go | Affiliate fraud detection — click stuffing, cookie dropping |

---

### 14. Marketplace (8 services)

| Service | Language | Responsibility |
|---|---|---|
| seller-registration-service | Go | Marketplace seller account creation and KYC |
| listing-approval-service | Go | Product listing review and approval workflow |
| marketplace-commission-service | Go | Commission rate management per seller and category |
| dispute-resolution-service | Java | Buyer-seller dispute intake and resolution |
| seller-analytics-service | Go | Seller performance metrics and dashboards |
| product-syndication-service | Go | Seller product syndication to multiple channels |
| storefront-service | Node.js | Per-seller branded storefront rendering |
| seller-payout-service | Go | Marketplace seller payout calculation and disbursement |

---

### 15. Gamification (6 services)

| Service | Language | Responsibility |
|---|---|---|
| points-service | Go | Points accrual and balance management via Redis |
| badge-service | Go | Badge definitions, award triggers, and display |
| leaderboard-service | Go | Real-time leaderboards via Redis sorted sets |
| challenge-service | Go | Timed challenges and progress tracking |
| reward-redemption-service | Go | Reward catalogue and redemption workflow |
| streak-service | Go | Daily/weekly streak tracking and incentives |

---

### 16. Developer Platform (6 services)

| Service | Language | Responsibility |
|---|---|---|
| api-management-service | Go | API product management — plans, versions, quotas |
| sandbox-service | Go | Isolated sandbox environments for API testing |
| developer-portal-backend | Node.js | Developer portal API — apps, credentials, docs |
| oauth-client-service | Go | OAuth2 client registration and credential management |
| api-analytics-service | Go | Per-developer API usage analytics |
| webhook-management-service | Go | Developer webhook subscription management |

---

### 17. Compliance (5 services)

| Service | Language | Responsibility |
|---|---|---|
| data-retention-service | Go | Data retention policy enforcement and deletion |
| consent-audit-service | Go | Consent record keeping and audit trail |
| privacy-request-service | Go | GDPR/CCPA data subject request orchestration |
| compliance-reporting-service | Java | Regulatory report generation (GDPR, PCI, SOC2) |
| data-lineage-service | Go | Data flow tracking across services |

---

### 18. Sustainability (5 services)

| Service | Language | Responsibility |
|---|---|---|
| carbon-tracker-service | Go | Carbon footprint calculation per order and shipment |
| eco-score-service | Go | Product environmental impact scoring |
| green-shipping-service | Go | Low-carbon carrier selection and routing |
| sustainability-reporting-service | Go | ESG metrics aggregation and reporting |
| offset-service | Go | Carbon offset purchase and certification tracking |

---

### 19. Web (6 services)

| Service | Framework | Responsibility |
|---|---|---|
| storefront-service | Next.js 14 (React/TS) | Customer-facing shopping experience — SSR for SEO |
| admin-dashboard-service | React + Vite (TS) | Admin and merchant management portal |
| seller-portal-service | Vue.js 3 (TS) | Marketplace seller portal — listings, analytics, payouts |
| partner-portal-service | Angular 18 (TS) | B2B partner portal — contracts, orders, invoices |
| mobile-app-service | React Native / Expo (TS) | iOS + Android customer app |
| developer-portal-service | React + Vite (TS) | Developer portal — API docs, sandbox, OAuth apps |

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
- [Databases](databases/README.md) — ClickHouse, Weaviate, Neo4j, TimescaleDB, Memcached, OpenSearch
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
