# System Overview — ShopOS

ShopOS is an enterprise-grade, cloud-native commerce platform comprising **154 microservices** across **13 business domains**, written in **8 programming languages**. Services communicate via **gRPC** (synchronous reads/commands) and **Apache Kafka** (asynchronous domain events). Every service owns its own dedicated database — no shared data stores.

---

## Full System Architecture

```
╔══════════════════════════════════════════════════════════════════════════════════════╗
║                              EXTERNAL CLIENTS                                        ║
║                                                                                      ║
║   ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌────────────────────────┐ ║
║   │  Web Browser  │  │  Mobile App  │  │  Partner API  │  │  Admin / Back-Office   │ ║
║   └──────┬───────┘  └──────┬───────┘  └──────┬───────┘  └───────────┬────────────┘ ║
╚══════════╪══════════════════╪══════════════════╪════════════════════════╪════════════╝
           │                  │                  │                        │
           └──────────────────┴──────────────────┴────────────────────────┘
                                       │ HTTPS / WSS
                          ┌────────────▼─────────────┐
                          │     TRAEFIK EDGE ROUTER   │  TLS termination (cert-manager)
                          │     + Coraza WAF           │  OWASP Core Rule Set — SQLi, XSS,
                          │     + rate-limiter-service │  path traversal blocked at edge
                          └────────────┬─────────────┘
                                       │
                 ┌─────────────────────▼──────────────────────┐
                 │              API GATEWAY  [Go :8080]         │
                 │  JWT validation · request routing · tracing  │
                 │  GraphQL Gateway [Go :8086] — unified query  │
                 └──────┬──────────────┬──────────────┬────────┘
                        │              │              │
              ┌─────────▼──┐  ┌────────▼───┐  ┌──────▼──────┐
              │  Web BFF    │  │ Mobile BFF  │  │ Partner BFF  │
              │  [Go :8081] │  │[Node :8082] │  │ [Go :8083]  │
              └─────────┬──┘  └────────┬───┘  └──────┬──────┘
                        └──────────────┼──────────────┘
                                       │
              ─────────────────────────┼──────────────────────────────────────
                    gRPC / Protobuf     │    (Istio mTLS, SPIFFE SVID certs)
              ─────────────────────────┼──────────────────────────────────────
                                       │
       ┌───────────────────────────────┼──────────────────────────────────────┐
       │                               │                                      │
       ▼                               ▼                                      ▼
┌─────────────┐               ┌──────────────┐                       ┌──────────────┐
│  IDENTITY   │               │   CATALOG    │                       │   COMMERCE   │
│  8 services │               │  12 services │                       │  23 services │
│             │               │              │                       │              │
│ auth        │               │ product      │◄──CDC (Debezium)──────│ cart         │
│ user        │               │ category     │   MongoDB→Kafka       │ checkout     │
│ session     │               │ brand        │      ↓                │ order        │
│ permission  │               │ pricing      │  search-service       │ payment      │
│ mfa         │               │ inventory    │  (Elasticsearch)      │ fraud-detect │
│ gdpr        │               │ search       │                       │ promotions   │
│ api-key     │               │ seo          │                       │ flash-sale   │
│ device-fp   │               │ ...          │                       │ bnpl, ...    │
└──────┬──────┘               └──────┬───────┘                       └──────┬───────┘
       │                             │                                       │
       └────────────────────────────┬┘                                      │
                                    │                                        │
              ══════════════════════╪════════════════════════════════════════╪═══════
                                    │   KAFKA EVENT BUS (Confluent 7.7.1)   │
              ══════════════════════╪════════════════════════════════════════╪═══════
                                    │   + Schema Registry (Avro / backward   │
                                    │     compatible schemas in events/)     │
              ┌─────────────────────┼────────────────────────────────────────┘
              │                     │
    ┌─────────┴──────────┐   ┌──────┴─────────────────────────────────────────────┐
    │                    │   │                                                     │
    ▼                    ▼   ▼                                                     ▼
┌──────────┐    ┌──────────────┐    ┌───────────────┐    ┌────────────────────────┐
│  SUPPLY  │    │  FINANCIAL   │    │ COMMUNICATIONS │    │    ANALYTICS & AI      │
│  CHAIN   │    │  11 services │    │   9 services   │    │      13 services       │
│ 13 svc   │    │              │    │                │    │                        │
│          │    │ invoice      │    │ notification-  │    │ analytics    reporting  │
│ vendor   │    │ accounting   │    │ orchestrator   │    │ recommendation  ml-fs  │
│ warehouse│    │ payout       │    │ email  sms     │    │ personalization clv    │
│ fulfill  │    │ reconcile    │    │ push  in-app   │    │ price-opt ad attribution│
│ tracking │    │ kyc-aml      │    │ whatsapp       │    │                        │
│ cold-ch  │    │ chargeback   │    │ chatbot        │    │ ┌──────────────────┐   │
│ ...      │    │ ...          │    │                │    │ │  Apache Flink    │   │
└──────────┘    └──────────────┘    └────────────────┘    │ │  order-analytics │   │
                                                           │ │  fraud-detection │   │
                                                           │ └──────────────────┘   │
                                                           │   ClickHouse · Weaviate │
                                                           │   Neo4j · ScyllaDB      │
                                                           └────────────────────────┘

       ┌────────────────────────────────────────────────────────────────┐
       │                 REMAINING DOMAINS                               │
       │                                                                │
       │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐        │
       │  │   CUSTOMER   │  │   CONTENT    │  │     B2B      │        │
       │  │  EXPERIENCE  │  │  8 services  │  │  7 services  │        │
       │  │  14 services │  │              │  │              │        │
       │  │ reviews  qa  │  │ media  cms   │  │ org  contract│        │
       │  │ wishlist     │  │ video  i18n  │  │ quote  edi   │        │
       │  │ support chat │  │ documents    │  │ approval     │        │
       │  └──────────────┘  └──────────────┘  └──────────────┘        │
       │                                                                │
       │  ┌──────────────┐  ┌──────────────────────────────────┐      │
       │  │ INTEGRATIONS │  │           AFFILIATE               │      │
       │  │  10 services │  │           4 services              │      │
       │  │              │  │                                   │      │
       │  │ erp  crm     │  │ affiliate  referral               │      │
       │  │ marketplace  │  │ influencer  commission-payout     │      │
       │  │ cdp  pim     │  │                                   │      │
       │  └──────────────┘  └──────────────────────────────────┘      │
       └────────────────────────────────────────────────────────────────┘
```

---

## Infrastructure Layers

```
╔═════════════════════════════════════════════════════════════════════════╗
║                    KUBERNETES CLUSTER (EKS / GKE / AKS)                 ║
║                                                                          ║
║  ┌────────────────────────────────────────────────────────────────────┐ ║
║  │                     PLATFORM DOMAIN (22 svc)                       │ ║
║  │  api-gateway  ·  graphql-gateway  ·  web/mobile/partner-bff        │ ║
║  │  saga-orchestrator  ·  event-store  ·  audit-service               │ ║
║  │  scheduler  ·  worker-job-queue  ·  dead-letter-service            │ ║
║  │  tenant-service  ·  webhook-service  ·  config-service             │ ║
║  └────────────────────────────────────────────────────────────────────┘ ║
║                                                                          ║
║  ┌────────────────────────────────────────────────────────────────────┐ ║
║  │              WORKFLOW ORCHESTRATION                                  │ ║
║  │  Temporal 1.24 — durable sagas, retry-safe multi-step flows        │ ║
║  │  Argo Workflows — CI builds, ML training pipelines                 │ ║
║  └────────────────────────────────────────────────────────────────────┘ ║
║                                                                          ║
║  ┌────────────────────────────────────────────────────────────────────┐ ║
║  │              DATA & MESSAGING INFRASTRUCTURE                         │ ║
║  │                                                                      │ ║
║  │  Kafka (Confluent 7.7.1) + ZooKeeper + Schema Registry             │ ║
║  │  RabbitMQ 3.13 (AMQP task queues + delayed messages)               │ ║
║  │  NATS JetStream 2.10 (real-time pub/sub)                           │ ║
║  │                                                                      │ ║
║  │  PostgreSQL 16  ·  MongoDB 8.0  ·  Redis 7  ·  Memcached 1.6      │ ║
║  │  Cassandra 5.0  ·  ScyllaDB 6.1  ·  Elasticsearch 8.15            │ ║
║  │  ClickHouse 24.8  ·  Weaviate 1.26  ·  Neo4j 5.23                 │ ║
║  │  OpenSearch 2.17  ·  MinIO  ·  etcd 3.5                           │ ║
║  │                                                                      │ ║
║  │  Debezium 2.7 (CDC: Postgres + MongoDB → Kafka)                    │ ║
║  │  Apache Flink 1.20 (stream processing: fraud, order analytics)     │ ║
║  └────────────────────────────────────────────────────────────────────┘ ║
║                                                                          ║
║  ┌────────────────────────────────────────────────────────────────────┐ ║
║  │              NETWORKING & SERVICE MESH                               │ ║
║  │  Istio (mTLS, traffic management, circuit breaking)                │ ║
║  │  Cilium eBPF CNI (L3/L4 policies, Hubble observability)           │ ║
║  │  Traefik 3.1 (edge router, TLS, service discovery)                │ ║
║  │  Consul 1.19 (service discovery, health check, K/V)               │ ║
║  └────────────────────────────────────────────────────────────────────┘ ║
║                                                                          ║
║  ┌────────────────────────────────────────────────────────────────────┐ ║
║  │              OBSERVABILITY STACK                                     │ ║
║  │                                                                      │ ║
║  │  OpenTelemetry (all 8 languages auto-instrumented)                 │ ║
║  │  Prometheus + Thanos + VictoriaMetrics (metrics)                   │ ║
║  │  Grafana (dashboards) + Alertmanager (routing to PagerDuty/Slack)  │ ║
║  │  Jaeger + Grafana Tempo + Zipkin (distributed tracing)             │ ║
║  │  Loki + Fluent Bit + Fluentd + ELK + OpenSearch (logs)            │ ║
║  │  Sentry OSS + GlitchTip (error tracking)                          │ ║
║  │  Grafana Pyroscope (continuous profiling)                          │ ║
║  │  Pyrra + Sloth + Uptime Kuma (SLO management)                     │ ║
║  └────────────────────────────────────────────────────────────────────┘ ║
║                                                                          ║
║  ┌────────────────────────────────────────────────────────────────────┐ ║
║  │              SECURITY STACK                                          │ ║
║  │                                                                      │ ║
║  │  Keycloak 25.0 (SSO/OIDC) · SPIFFE/SPIRE (workload identity)      │ ║
║  │  HashiCorp Vault (dynamic secrets) · cert-manager (TLS)            │ ║
║  │  OPA/Gatekeeper · Kyverno · Kubewarden (admission policy)          │ ║
║  │  OpenFGA (relationship-based authz)                                │ ║
║  │  Falco + Tetragon + Tracee + Wazuh (runtime security/SIEM)        │ ║
║  │  Cosign + Syft + Rekor + Fulcio (supply chain)                    │ ║
║  │  Trivy + Grype + Semgrep + SonarQube + ZAP (scanning)             │ ║
║  └────────────────────────────────────────────────────────────────────┘ ║
║                                                                          ║
║  ┌────────────────────────────────────────────────────────────────────┐ ║
║  │              GITOPS & CI/CD                                          │ ║
║  │                                                                      │ ║
║  │  Jenkins + Drone CI + Dagger (build / test / scan / publish)       │ ║
║  │  ArgoCD (App-of-Apps → 154 Applications) + Flux CD                │ ║
║  │  Argo Rollouts (canary 10%→25%→50%→100%)                          │ ║
║  │  Argo Events (GitHub webhook → pipeline trigger)                   │ ║
║  │  Harbor + Nexus + Gitea (registry / artifact / git)                │ ║
║  └────────────────────────────────────────────────────────────────────┘ ║
║                                                                          ║
║  ┌────────────────────────────────────────────────────────────────────┐ ║
║  │              RESILIENCE & TESTING                                    │ ║
║  │                                                                      │ ║
║  │  KEDA (Kafka consumer lag + Redis list HPA autoscaling)            │ ║
║  │  Velero (daily cluster backups → MinIO/S3)                         │ ║
║  │  Pod Disruption Budgets (all stateful / critical services)         │ ║
║  │                                                                      │ ║
║  │  Chaos Mesh (PodChaos, NetworkChaos, StressChaos, IOChaos,        │ ║
║  │              HTTPChaos, TimeChaos — game-day schedule Sat 02:00)   │ ║
║  │  LitmusChaos (Argo Workflow pipelines + SLO probe validation)      │ ║
║  │                                                                      │ ║
║  │  k6 (smoke / load / spike / soak — CI gate)                       │ ║
║  │  Locust (exploratory load, web UI, distributed)                    │ ║
║  │  Gatling (JVM high-concurrency, HTML reports)                      │ ║
║  └────────────────────────────────────────────────────────────────────┘ ║
╚═════════════════════════════════════════════════════════════════════════╝
```

---

## Domain Summary

| # | Domain | Services | Languages | Core Responsibility |
|---|---|---|---|---|
| 1 | Platform | 22 | Go, Java, Python | API gateways, BFFs, event store, saga orchestration, Temporal workflows |
| 2 | Identity | 8 | Rust, Java, Go | Auth, users, sessions, MFA, GDPR, SPIFFE workload identity |
| 3 | Catalog | 12 | Go, Java, Kotlin, Python, Node.js | Products, pricing, inventory, search |
| 4 | Commerce | 23 | Go, Kotlin, Java, C#, Rust, Python, Node.js | Cart, checkout, orders, payments, promotions, BNPL, flash sales |
| 5 | Supply Chain | 13 | Go, Java, Kotlin, Python, Node.js | Vendors, warehouses, fulfilment, tracking, cold chain |
| 6 | Financial | 11 | Java, Kotlin, Go | Invoicing, accounting, payouts, compliance, chargebacks |
| 7 | Customer Experience | 14 | Go, Java, Node.js | Reviews, support, wishlists, price alerts, gift registry |
| 8 | Communications | 9 | Node.js, Python, Go | Email, SMS, push, in-app, WhatsApp, chatbot |
| 9 | Content | 8 | Go, Java, Python, Node.js | Media, CMS, documents, i18n, video |
| 10 | Analytics & AI | 13 | Python, Scala, Java | Events, ML, recommendations, attribution, CLV, Flink stream jobs |
| 11 | B2B | 7 | Java, Kotlin, Go | Organisations, contracts, procurement, EDI, marketplace seller |
| 12 | Integrations | 10 | Java, Go, Node.js | ERP, CRM, marketplace, payment gateway, CDP adapters |
| 13 | Affiliate | 4 | Go | Affiliate, referral, influencer, commissions |

---

## Data Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                     DATA FLOW ARCHITECTURE                               │
│                                                                          │
│  OPERATIONAL DATABASES (source of truth per service)                    │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐   │
│  │ Postgres  │ │ MongoDB  │ │  Redis   │ │Cassandra │ │  MinIO   │   │
│  │  (ACID)   │ │  (docs)  │ │(cache)   │ │(timeseries│ │(objects) │   │
│  └─────┬─────┘ └────┬─────┘ └──────────┘ └──────────┘ └──────────┘   │
│        │             │                                                   │
│        └──────── Debezium CDC ──────────────────────────────────┐      │
│                       │                                          │      │
│                       ▼                                          │      │
│              ┌─────────────────┐                                 │      │
│              │   KAFKA TOPICS   │                                 │      │
│              │  *.cdc (changes) │                                 │      │
│              └────────┬────────┘                                 │      │
│                       │                                          │      │
│          ┌────────────┼────────────────────────┐                 │      │
│          ▼            ▼                        ▼                 │      │
│  ┌──────────────┐ ┌──────────────┐   ┌──────────────────┐       │      │
│  │ Elasticsearch│ │  ClickHouse  │   │   OpenSearch     │       │      │
│  │ (product     │ │ (OLAP reports│   │  (log analytics  │       │      │
│  │  search)     │ │  revenue MV) │   │   audit search)  │       │      │
│  └──────────────┘ └──────────────┘   └──────────────────┘       │      │
│                       ▲                                          │      │
│          ┌────────────┘                                          │      │
│          │                                                       │      │
│  ┌──────────────────────────────────────────────────────────┐   │      │
│  │              APACHE FLINK (stream processing)             │   │      │
│  │  order-analytics job:  Kafka → windowed revenue → CH     │◄──┘      │
│  │  fraud-detection job:  Kafka → velocity check → Kafka    │          │
│  └──────────────────────────────────────────────────────────┘          │
│                                                                          │
│  SPECIALIST STORES                                                       │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐                  │
│  │ Weaviate │ │  Neo4j   │ │ScyllaDB  │ │Memcached │                  │
│  │(vectors  │ │(product  │ │(high-tput│ │(hot read │                  │
│  │ ANN)     │ │ graph)   │ │ events)  │ │ cache)   │                  │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘                  │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## Request Lifecycle — Full Purchase Journey

```
Browser                 Traefik          API Gateway         Services
   │                       │                  │                  │
   │── POST /checkout ─────►│                  │                  │
   │                       │── TLS ──────────►│                  │
   │                       │  WAF check        │── JWT validate ─►│ auth-service
   │                       │                  │◄─ token valid ───│
   │                       │                  │                  │
   │                       │                  │── gRPC ─────────►│ cart-service
   │                       │                  │◄─ cart contents ─│
   │                       │                  │                  │
   │                       │                  │── gRPC ─────────►│ inventory-service
   │                       │                  │── gRPC ─────────►│ tax-service
   │                       │                  │── gRPC ─────────►│ shipping-service
   │                       │                  │── gRPC ─────────►│ promotions-service
   │                       │                  │                  │
   │                       │                  │── Temporal ─────►│ saga-orchestrator
   │                       │                  │  (order saga)    │── gRPC ──► payment-service
   │                       │                  │                  │── gRPC ──► order-service
   │                       │                  │                  │
   │                       │                  │  order created   │── Kafka ──► fulfillment
   │                       │                  │                  │── Kafka ──► loyalty
   │                       │                  │                  │── Kafka ──► notifications
   │                       │                  │                  │── Kafka ──► analytics
   │                       │                  │                  │── Kafka ──► fraud-detection
   │                       │                  │                  │
   │◄── 201 Created ───────│◄─────────────────│                  │
   │   { orderId: "..." }  │                  │                  │
```

---

## Key Architectural Patterns

| Pattern | Implementation | Why |
|---|---|---|
| **API Gateway** | `api-gateway` (Go) | Single ingress; JWT validation, rate limiting, and routing in one place |
| **BFF** | `web-bff`, `mobile-bff`, `partner-bff` | Different clients need different response shapes; BFF tailors the API per client |
| **CQRS** | `order-service`, `accounting-service` | High-read reporting queries must not compete with write transactions |
| **Event Sourcing** | `event-store-service` (Postgres, append-only) | Full audit trail; enables replay and temporal queries |
| **Saga** | `saga-orchestrator` (choreography) + Temporal (orchestration) | Distributed transactions across services without 2PC; each step is compensatable |
| **Database-per-service** | 154 separate databases | No coupling between services; each chooses the optimal store for its access pattern |
| **Outbox pattern** | Services write to outbox table → Debezium CDC → Kafka | Guarantees at-least-once event delivery even if Kafka is briefly unavailable |
| **CQRS read models** | Debezium → ClickHouse / Elasticsearch / OpenSearch | Reporting queries run against optimised read stores, not transactional databases |
| **Stream processing** | Apache Flink (Kafka → stateful jobs → Kafka/ClickHouse) | Fraud detection and analytics require windowed aggregations across millions of events |
| **Circuit breaker** | All gRPC clients (exponential backoff + jitter, 3 attempts) | Prevents cascade failure when a downstream service is slow or unavailable |
| **Progressive delivery** | Argo Rollouts canary (10%→25%→50%→100%) | Reduce blast radius of a broken release; auto-rollback on metric degradation |

---

## Observability Stack

```
┌─────────────────────────────────────────────────────────────────────────┐
│                     OBSERVABILITY PIPELINE                               │
│                                                                          │
│  Service A  →  OpenTelemetry SDK  →  OTel Collector  →  ┬── Jaeger      │
│                 (traces, metrics,                         ├── Tempo       │
│                  logs, profiles)                          ├── Prometheus  │
│                                                           ├── Loki        │
│                                                           └── Pyroscope   │
│                                                                          │
│  Prometheus → Thanos (long-term storage, dedup, global query)           │
│  Loki ← Fluent Bit (log shipping from all pods) ← Fluentd               │
│  OpenSearch ← Logstash ← application structured logs                    │
│                                                                          │
│  Grafana ← all of the above (unified dashboards)                        │
│  Alertmanager ← Prometheus alerts → Slack / PagerDuty / email          │
│                                                                          │
│  Sentry OSS + GlitchTip ← application exception tracking               │
│  Uptime Kuma ← external HTTP health checks (public status page)         │
│  Pyrra + Sloth ← SLO burn rate recording rules                          │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## Resilience Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    RESILIENCE LAYERS                                      │
│                                                                          │
│  APPLICATION LEVEL                                                       │
│  ├── Circuit breakers on all gRPC clients (grpc-go retry interceptor)  │
│  ├── Timeouts configured per RPC method (per proto service definition)  │
│  ├── Idempotency keys on all payment and order mutations                │
│  └── Saga compensation handlers for every distributed transaction step  │
│                                                                          │
│  INFRASTRUCTURE LEVEL                                                    │
│  ├── Pod Disruption Budgets — minAvailable=1 on all stateful services   │
│  ├── KEDA autoscaling — Kafka consumer lag + Redis queue depth triggers │
│  ├── Argo Rollouts — canary with Prometheus metric gates                │
│  └── Velero daily backup — namespace-level restore to clean cluster     │
│                                                                          │
│  CHAOS ENGINEERING                                                       │
│  ├── Chaos Mesh experiments (automated, namespace-scoped)               │
│  │   ├── PodChaos    — api-gateway, order, payment, kafka (33%)        │
│  │   ├── NetworkChaos — delay (100ms), loss (10%), partition, bw limit │
│  │   ├── StressChaos  — CPU 80% on checkout, memory 256MB on recommend │
│  │   ├── IOChaos      — disk delay 100ms on product-catalog            │
│  │   ├── HTTPChaos    — 30% abort on payment, 3s delay on checkout     │
│  │   └── TimeChaos    — -10m clock skew on order-service               │
│  ├── Chaos Mesh Workflows — multi-phase resilience scenarios            │
│  ├── Game-day schedule    — Saturdays 02:00 UTC (automated)             │
│  └── LitmusChaos          — Argo Workflow pipelines with SLO probes    │
│                                                                          │
│  LOAD TESTING (CI-gated SLO validation)                                 │
│  ├── k6 smoke test  — every deploy (1 VU, 2m, p95 < 2s)               │
│  ├── k6 load test   — nightly (50 VUs, checkout p95 < 5s)             │
│  ├── k6 spike test  — weekly (500 VUs burst, recovery rate > 90%)      │
│  ├── k6 soak test   — weekly overnight (30 VUs, 2h, no p95 drift)      │
│  └── Gatling        — pre-release (JVM concurrency, HTML reports)       │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## CI/CD & GitOps Pipeline

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    CI/CD PIPELINE                                         │
│                                                                          │
│  1. CODE PUSH → GitHub / Gitea                                          │
│     └── Argo Events (GitHub webhook EventSource → Sensor → trigger)    │
│                                                                          │
│  2. CI BUILD (Jenkins / Drone CI / Dagger)                              │
│     ├── Checkout + compile (language-specific)                          │
│     ├── Unit + integration tests                                        │
│     ├── SonarQube SAST (quality gate)                                   │
│     ├── Semgrep custom security rules                                   │
│     ├── Checkov / KICS IaC scan (if infra changed)                     │
│     ├── Docker multi-stage build (non-root, minimal image)             │
│     ├── Trivy scan → block on CRITICAL CVE                             │
│     ├── Grype scan → block on CRITICAL CVE (second opinion)            │
│     ├── Syft SBOM generation (CycloneDX + SPDX)                        │
│     ├── Cosign keyless sign (Fulcio cert + Rekor log entry)            │
│     └── Push to Harbor registry                                         │
│                                                                          │
│  3. GITOPS SYNC (ArgoCD — pull model)                                   │
│     ├── ArgoCD detects Helm chart change in Git                         │
│     ├── Pre-flight checks (cluster health, secret existence, PDB)      │
│     ├── Kyverno admission: verify Cosign signature before pod starts   │
│     └── Argo Rollouts canary: 10% → metrics check → 25% → ... → 100%  │
│                                                                          │
│  4. POST-DEPLOY VALIDATION                                               │
│     ├── k6 smoke test against new deployment                           │
│     ├── OWASP ZAP DAST scan (nightly on staging)                       │
│     └── Nuclei CVE template scan (nightly on staging)                  │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## Infrastructure as Code

| Tool | Provider | Location |
|---|---|---|
| Terraform | AWS EKS, GCP GKE, Azure AKS, Jenkins EC2 | `infra/terraform/` |
| OpenTofu | AWS, GCP, Azure (open-source Terraform fork) | `infra/opentofu/` |
| Crossplane | K8s-native IaC (compositions + claims) | `infra/crossplane/` |
| Ansible | K8s node bootstrapping, OS configuration | `infra/ansible/` |

---

## References

- [Communication Patterns](communication-patterns.md)
- [Database Strategy](database-strategy.md)
- [Domain Map](domain-map.md)
- [Security Model](security-model.md)
- [ADR-001: gRPC for Internal Communication](../adr/001-grpc-for-internal-communication.md)
- [ADR-002: Kafka for Async Events](../adr/002-kafka-for-async-events.md)
- [ADR-005: Database-per-Service](../adr/005-database-per-service.md)
- [ADR-006: GitOps with ArgoCD](../adr/006-gitops-with-argocd.md)
- [Chaos Engineering Runbook](../../chaos/README.md)
- [Load Testing Guide](../../load-testing/README.md)
