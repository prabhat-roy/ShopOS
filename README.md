# ShopOS — Enterprise Commerce Platform

An enterprise-grade, cloud-native commerce platform — 303 services, 22 domains, 19 languages, full open source stack.

---

## Quick links

- [Getting started](GETTING_STARTED.md) — from clone to local stack in ~10 min
- [Architecture overview](docs/architecture/system-overview.md)
- [Service catalog (Backstage)](backstage/catalog-info.yaml) — 348 entries
- [Runbooks](docs/runbooks/) — deployment, incident, rollback, postgres-failover, kafka-consumer-lag
- [CLAUDE.md](CLAUDE.md) — authoritative source-of-truth (read this if you're an AI agent)

## Repository layout

```
src/                 # 296 backend services + src/web/ 7 frontends
proto/               # gRPC contracts (58 files, 14 domains)
events/              # 20 Avro Kafka event schemas
helm/services/       # 303 per-service Helm charts
gitops/              # ArgoCD ApplicationSet + Flux HelmReleases (303 each)
ci/                  # 15 CI platforms x 15 pipelines
infra/               # Terraform, OpenTofu, Crossplane, Ansible, Patroni, PgBouncer, Atlantis, Nomad
kubernetes/          # raw K8s manifests, RBAC, NetworkPolicies, PDBs, Velero, KEDA, Karpenter, VPA
networking/          # Traefik, Istio, Cilium, Consul, Caddy, Anubis, MetalLB, Kube-VIP, edge/spin
security/            # Vault, Keycloak, Falco, OPA, Kyverno, Cosign, Teleport, Trivy, Kubescape, ESO, ...
observability/       # Prometheus, Grafana, Loki, Tempo, OTel, Mimir, Quickwit, Parca, Komodor, ...
messaging/           # Kafka, RabbitMQ, NATS, Redpanda, Zilla, Conduktor Gateway
databases/           # Postgres Flyway migrations + ClickHouse/Weaviate/Neo4j/.../LakeFS, Dgraph, YugabyteDB
storage/             # Longhorn, Rook-Ceph (PV providers)
data/                # Airflow, dbt, Spark, Airbyte, Cube, Metabase, OpenLineage, Great Expectations
ml/                  # MLflow, Feast (Phase 5)
streaming/           # Debezium CDC + Apache Flink
workflow/            # Temporal
api-management/      # Apache APISIX, Hasura, Tyk
api-testing/         # Hurl, Spectral
testing/             # Pact (9 contracts), Playwright, Karate, Testcontainers, Artillery
load-testing/        # k6, Locust, Gatling
chaos/               # Chaos Mesh, LitmusChaos
backstage/           # Developer portal (348 entries)
dev/                 # Coder, DevSpace, n8n, Windmill, Score, Backstage Scaffolder
feature-flags/       # Unleash
finops/              # Kubecost
incident/            # Cachet, Grafana Incident, Grafana OnCall
build/               # Earthly, Ko, Kaniko
registry/            # Harbor, Nexus, Gitea, Zot, ChartMuseum
openapi/             # OpenAPI 3.1 specs
docs/                # ADRs, runbooks, architecture
scripts/             # Service scaffolder (bash) + Jenkins helpers (groovy)
```

---

## Domains

| # | Domain | Services |
|---|---|---|
| 1 | Platform | 40 |
| 2 | Identity | 14 |
| 3 | Catalog | 19 |
| 4 | Commerce | 32 |
| 5 | Supply Chain | 20 |
| 6 | Financial | 20 |
| 7 | Customer Experience | 20 |
| 8 | Communications | 14 |
| 9 | Content | 13 |
| 10 | Analytics & AI | 13 |
| 11 | B2B | 11 |
| 12 | Integrations | 18 |
| 13 | Affiliate | 7 |
| 14 | Marketplace | 10 |
| 15 | Gamification | 7 |
| 16 | Developer Platform | 8 |
| 17 | Compliance | 7 |
| 18 | Sustainability | 6 |
| 19 | Web | 7 |
| 20 | Events & Ticketing | 7 |
| 21 | Auction | 5 |
| 22 | Rental | 5 |
| | Total | 303 |

---

## Technology Stack

### Languages

| Language | Version | Used In |
|---|---|---|
| Go | 1.24 | Platform, Catalog, Commerce, Supply Chain, Financial, CX, Content, B2B, Integrations, Affiliate, Events, Auction, Rental |
| Java | 21 (Spring Boot) | Identity, Catalog, Commerce, Supply Chain, Financial, B2B, Integrations, Auction |
| Kotlin | 2.x (Spring Boot) | Catalog, Commerce, Supply Chain, Financial, B2B, Rental |
| Python | 3.12 | Analytics & AI, Supply Chain, Communications, Content |
| Node.js | 22 | Platform, Catalog, Customer Experience, Communications, Content, Integrations |
| C# | .NET 9 | Commerce (cart, return-refund) |
| Rust | 1.80 | Identity (auth), Commerce (shipping) |
| Scala | 3.x | Analytics & AI (reporting) |
| Elixir | 1.17 (OTP 27) | Platform (presence, realtime, pubsub), Events & Ticketing, Auction — real-time concurrent services |
| Haskell | GHC 9.6 | Financial (rules engine) — type-safe pure functional calculations |
| PHP | 8.3 (Laravel 11) | Integrations (Magento/WooCommerce adapters) |
| Ruby | 3.3 (Sinatra 4) | Content (CMS adapter) |
| Dart | 3.4 (Flutter) | Web (mobile-flutter-service) — native iOS + Android |
| Swift | 5.10 (Vapor 4) | Platform (iOS push gateway) |
| Clojure | 1.12 | Platform (event transform) — immutable stream transformation |
| Crystal | 1.13 | Content (webhook service) — Ruby-like, C performance |
| Zig | 0.13 | Platform (rate-limiter-core) — zero-overhead systems-level |
| Gleam | 1.4 | Platform (event pipeline) — type-safe on BEAM/OTP |

### Databases

| Database | Version | Role |
|---|---|---|
| PostgreSQL | 16 | Primary transactional store — 100+ services |
| MongoDB | 8.0 | Document store — catalog, CMS, reviews, tracking |
| Redis | 7 | Cache, sessions, pub/sub, ephemeral data |
| Cassandra | 5.0 | Time-series analytics events |
| TimescaleDB | 2.15 | Time-series metrics — service metrics, inventory events, page views |
| CockroachDB | 24.2 | Distributed PostgreSQL — geo-distributed ACID transactions |
| SurrealDB | 2.1 | Multi-model — SQL + document + graph + time-series in one |
| EventStoreDB | 24.10 | Purpose-built event store for event sourcing |
| Valkey | 8.0 | Redis-compatible cache (Linux Foundation open fork) |
| Typesense | 27.0 | Typo-tolerant instant search engine |
| Manticore Search | 6.3 | Fast full-text search, Sphinx-compatible |
| SeaweedFS | 3.78 | Distributed object + file storage, S3-compatible |
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
| Apache Kafka (Confluent) | 7.7.1 | Primary event streaming — domain events; 20 Avro topics + 3 DLQs as Strimzi `KafkaTopic` CRDs |
| Apache ZooKeeper | 7.7.1 | Kafka coordination |
| Schema Registry (Confluent) | 7.7.1 | Avro schema enforcement |
| RabbitMQ | 3.13 | Task queues, delayed messages, RPC |
| NATS JetStream | 2.10 | Low-latency pub/sub — chat, real-time notifications |
| Redpanda | 5.9 | Kafka-API-compatible low-latency alternative for analytics |
| Zilla | 0.9 | Kafka → REST/SSE/MQTT proxy for browsers and IoT |
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
| Helm | 3.x | 303 per-service charts + tool charts (all vendored, no internet at deploy time) |
| KEDA | 2.15 | Kafka/Redis-driven autoscaling (alongside HPA) |
| Velero | 7.x | Kubernetes backup and restore |
| Skaffold | latest | Local dev hot-reload |
| Tilt | latest | Local dev hot-reload (alternative) |
| Kaniko | latest | Rootless container builds inside Kubernetes pods |
| Bazel | 7.x | Google's hermetic monorepo build system — 1000+ targets |
| Nx | 20.x | Monorepo orchestration for TypeScript/JavaScript frontend services |
| Turborepo | 2.x | Fast JS monorepo builds with remote cache |
| Packer | 1.11 | Automated VM image builder for cloud AMIs and GCE images |

### Infrastructure as Code

| Tool | Version | Role |
|---|---|---|
| Terraform | 1.9 | EKS, GKE, AKS cluster provisioning |
| OpenTofu | 1.8 | Open source Terraform alternative (same targets) |
| Crossplane | 1.17 | Kubernetes-native IaC — database and cloud resource claims |
| Ansible | 2.17 | Kubernetes node bootstrapping |
| Terrascan | latest | IaC security scanning — Terraform + Helm |
| Docker Compose | v2 | Full local stack (303 services + infra) |
| Karpenter | 1.0 | Node autoscaler — provisions right-sized EC2 on demand |
| VPA + Goldilocks | 4/9 | Vertical resource recommendations + UI |
| Longhorn | 1.7 | Distributed RWO block storage on local disks |
| Rook-Ceph | v1.15 | Block, file, and S3 storage operator |

### CI/CD

| Tool | Pipelines | Role |
|---|---|---|
| Jenkins | 15 | Primary CI server — build, test, security, quality, tooling pipelines |
| Drone CI | 15 | Mirror of Jenkins — same stages, Drone syntax |
| Woodpecker CI | 15 | Drone-compatible fork — drop-in replacement |
| Dagger | 15 | Portable Go SDK pipelines — run anywhere |
| Tekton | 15 | Kubernetes-native CRD-based pipelines |
| Concourse CI | 15 | DAG resource/job pipelines |
| GitLab CI | 15 | `.gitlab-ci.yml` pipelines for GitLab SCM |
| GitHub Actions | 15 | `ci/github-actions/` — auto-trigger disabled |
| CircleCI | 15 | `version: 2.1` orb-based pipelines |
| GoCD | 15 | Stage/job pipelines with manual approval gates |
| Travis CI | 15 | Stage-based pipelines with branch filters |
| Harness CI | 15 | Enterprise CI/CD with built-in CD stages |
| Azure DevOps | 15 | `azure-pipelines.yml` — native Azure integration |
| AWS CodePipeline | 15 | `buildspec.yml` + CodePipeline JSON definitions |
| GCP Cloud Build | 15 | `cloudbuild.yaml` — native GCP integration |
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
| OpenTelemetry | 0.108 | Instrumentation SDK + Collector (agent DaemonSet + gateway Deployment with tail-sampling) |
| Fluent Bit | 3.x | Log collection DaemonSet → Loki + VictoriaLogs split routing |
| Prometheus | 2.54 | Metrics scraping and alerting (per-domain rules: identity, catalog, financial, supply-chain, cx, comms, marketplace+b2b, infra) |
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
| Grafana Mimir | 2.14 | Horizontally scalable long-term Prometheus storage |
| VictoriaLogs | 1.4 | Fast, cheap log storage at high volume |
| Pixie | latest | eBPF auto-instrumentation — zero-code traces per pod |
| SigNoz | 0.53 | Full-stack observability — traces + metrics + logs, OTel-native |
| Netdata | 1.47 | Real-time per-second metrics per container |
| Kiali | 2.0 | Istio service mesh observability UI — topology + traffic |
| Perses | 0.49 | GitOps-native dashboard-as-code |
| Goldilocks | 9.x | VPA-based resource right-sizing recommendations |
| kube-state-metrics | latest | Kubernetes object metrics |
| node-exporter | latest | Node-level hardware metrics |
| Quickwit | 0.7 | S3-backed log search engine (cold logs alternative to Loki/ES) |
| Parca | 0.22 | Continuous profiling (eBPF, complement to Pyroscope) |
| Komodor | 4.4 | K8s troubleshooting timelines per resource |
| Healthchecks | 0.7 | Self-hosted cron monitoring with miss alerts |
| Kubecost | 2.4 | Multi-cluster cost allocation + idle resource recommendations |

### Security

| Tool | Role |
|---|---|
| HashiCorp Vault | HA Raft + KMS unseal, K8s/OIDC/JWT/AppRole auth, KV-per-domain, dynamic Postgres roles, AWS IAM, PKI int+root, Transit, TOTP, SSH-CA |
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
| Wazuh | SIEM + HIDS — log correlation, compliance, intrusion detection |
| Suricata | Network IDS/IPS — deep packet inspection on cluster traffic |
| Zeek | Network traffic analysis — behavioral detection, TLS fingerprinting |
| OpenVAS | External attack surface vulnerability scanning |
| Pomerium | Identity-aware access proxy — zero-trust for internal tools |
| External Secrets Operator | Sync secrets from Vault/AWS SSM into Kubernetes |
| Sealed Secrets | Encrypt secrets for safe GitOps storage |
| OpenSSF Scorecard | Automated security best-practice scoring |
| Sigstore Policy Controller | Admission-time Cosign verification of all images |
| Trivy Operator | Continuous in-cluster vulnerability + misconfig + secret scanning |
| Kubescape | NSA / CISA / MITRE ATT&CK posture scoring |
| Cedar | AWS Cedar policies for resource-scoped authz (sellers, orgs) |
| GitGuardian (ggshield) | Secrets scanning across 200+ providers |

### Networking & Service Mesh

| Tool | Version | Role |
|---|---|---|
| Traefik | 3.1 | Edge router — ingress, automatic TLS, service discovery |
| Istio | 1.23 | Service mesh — STRICT mTLS, AuthZ deny-all + named allows, DestinationRules, VirtualServices |
| Linkerd | 2.x | Lightweight service mesh (alternative) |
| Cilium | 1.16 | eBPF CNI + L7 NetworkPolicies (HTTP-aware payment + auth filters) |
| Calico | 3.28 | CNI and NetworkPolicy alternative |
| Consul | 1.19 | Service discovery, health checking, K/V config |
| Kong | 3.x | API gateway (alternative to Traefik) |
| NGINX | 1.27 | Reverse proxy and static file serving |
| Caddy | 2.x | Auto-TLS ingress for non-prod / preview environments |
| Varnish | 7.x | HTTP cache in front of storefront / admin portals |
| Anubis | 1.10 | Proof-of-work anti-bot / anti-AI-scraper for storefront |
| ngrok-operator | 0.16 | Public ingress for PR-preview environments |
| MetalLB | 0.14 | Bare-metal LoadBalancer for on-prem clusters |
| Kube-VIP | latest | HA control-plane VIP for bare-metal clusters |
| Spin / SpinKube | 0.4 | Wasm serverless edge functions for storefront personalization |

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
| Metabase | Self-serve BI alternative to Superset |
| Cube | Semantic layer + BI API on top of ClickHouse/Postgres |
| Airbyte | ELT from 300+ sources (Stripe, Salesforce, Postgres) into ClickHouse |
| LakeFS | Git-like data versioning over MinIO |
| Apache Atlas | Data catalog and governance |
| Marquez | OpenLineage data lineage tracking |
| Great Expectations | Data quality assertion suites |
| MLflow | 2.16 — Experiment tracking, model registry |
| Apache Flink | 1.20 — Real-time feature computation and stream ML |
| Weaviate | 1.26 — Vector store for RAG and semantic search |
| Dgraph | Distributed graph DB — complement to Neo4j for recommendations at scale |
| Neo4j | 5.23 — Graph-based recommendation engine |
| YugabyteDB | Distributed Postgres-compatible SQL DB — geo-distributed alternative to CockroachDB |

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
| Pact | Consumer-driven contract testing — 9 contracts (cart←”catalog, checkout←”promotions, order←”inventory, notif←”template, web-bff←”search, payout←”wallet, affiliate←”commission, order←”cart, checkout←”payment) |
| Buf breaking-change CI | Jenkins + GitHub Actions blocking merge on proto regressions |
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

### Feature Flags

| Tool | Role |
|---|---|
| Unleash | Open source feature flag platform — SDK, UI, audit trail |
| OpenFeature | Open standard SDK for feature flags — wraps any backend |

### Incident Management

| Tool | Role |
|---|---|
| Grafana OnCall | Open source on-call scheduling, escalation, Slack integration |
| Cachet | Open source public status page for customer-facing uptime |

### API Management

| Tool | Role |
|---|---|
| Apache APISIX | Cloud-native API gateway with Lua plugin ecosystem |
| Tyk | Open source API management + built-in developer portal |
| Hasura | Instant GraphQL engine from PostgreSQL |

### Infrastructure Automation

| Tool | Role |
|---|---|
| Atlantis | Terraform GitOps — plan on PR, apply on merge |
| Infracost | Cloud cost estimation on every Terraform PR |
| Driftctl | Detect drift between Terraform state and cloud resources |
| Packer | Automated VM image builder (AMI, GCE) |
| Nomad | HashiCorp workload orchestrator for mixed containerised/bare-metal |
| Boundary | HashiCorp zero-trust SSH/RDP access without VPN |
| Waypoint | HashiCorp app deployment abstraction across K8s/Nomad/ECS |

### Developer Experience

| Tool | Role |
|---|---|
| Devcontainer | VS Code / GitHub Codespaces — all 19 languages pre-installed |
| Skaffold | Local Kubernetes dev hot-reload |
| Tilt | Local Kubernetes dev with live_update (alternative) |
| Backstage | Internal developer portal — service catalog (348 entries), API docs, scaffolder |
| Backstage Software Templates | One-click Go-service scaffolding via [`dev/scaffolder/`](dev/scaffolder/) |
| Coder | Self-hosted cloud development environments (Terraform-defined templates) |
| n8n | Low-code workflow automation for ops |
| Windmill | Internal scripts/APIs as a service (Retool/Lambda alternative) |
| Buf CLI | Protobuf workflow — lint, format, codegen, breaking detection |
| Teleport | Zero-trust SSH + Kubernetes access for developers and ops |
| Telepresence | Run one local service against live cluster — instant debugging |
| Garden | Full environment automation — spin up entire stack per PR |
| Signadot | Kubernetes sandbox per PR — route traffic to feature branch |
| Devspace | Kubernetes dev tool — live sync, port-forward, log aggregation |
| Botkube | Kubernetes alerts to Slack — pod failures, deployments |
| k8sGPT | AI-powered Kubernetes diagnostics operator |
| k9s | Terminal-based Kubernetes UI |
| grpcurl | gRPC API testing |

### Workflow Orchestration (Data)

| Tool | Role |
|---|---|
| Prefect | Python-native workflow orchestration — complement to Airflow |
| Dagster | Data asset orchestration — lineage-first pipeline management |
| Apache Camel | Enterprise integration patterns — 300+ connectors |

### Chaos & Load Testing

| Tool | Role |
|---|---|
| Chaos Mesh | Kubernetes-native chaos injection |
| Litmus | Chaos engineering framework |
| k6 | JavaScript-based load testing |
| Locust | Python-based distributed load testing |
| Gatling | Scala-based load and performance testing |
| Artillery | Node.js load testing for realistic browser scenarios |
| Vegeta | Go HTTP load testing — deterministic rate, ideal for soak tests |
| k6 Operator | Run k6 load tests natively as Kubernetes Jobs |

### Testing Extensions

| Tool | Role |
|---|---|
| Playwright | E2E browser testing for all 7 frontend services |
| WireMock | API mock server — stub third-party APIs in integration tests |
| Karate | Java BDD-style API + UI testing framework |

---

## Services

### 1. Platform (40 services)

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
| presence-service | Elixir | Real-time user presence tracking using BEAM/OTP |
| realtime-gateway-service | Elixir | WebSocket gateway for real-time event streaming |
| pubsub-router-service | Elixir | Pub/sub message routing using Phoenix.PubSub |
| api-versioning-service | Go | API version negotiation and deprecation lifecycle |
| event-transform-service | Clojure | Kafka event transformation with immutable data pipelines |
| ios-push-gateway-service | Swift | Apple APNs push notification gateway |
| rate-limiter-core | Zig | Ultra-low-latency rate limiting core in systems-level Zig |
| reports-portal-service | Go | Internal reporting portal aggregating per-domain reports |
| secrets-rotation-service | Go | Rotates Vault dynamic credentials, JWT signing keys, DB passwords on schedule |
| service-registry-service | Go | Consul wrapper for service discovery and health-check aggregation |
| distributed-lock-service | Go | Redis-backed distributed lock primitive for cross-pod critical sections |
| chaos-control-service | Go | Programmatic API to trigger Chaos Mesh / Litmus experiments |
| graphql-federation-service | Go | Apollo Federation gateway composing per-domain GraphQL subgraphs |

---

### 2. Identity (14 services)

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
| passkey-service | Go | WebAuthn / FIDO2 passkey registration and authentication |
| risk-scoring-service | Go | Login risk score from device fingerprint, geo, velocity, behaviour |
| account-linking-service | Go | Merge duplicate user accounts (email, social, passkey) into canonical identity |

---

### 3. Catalog (19 services)

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
| search-suggestion-service | Go | Autocomplete and typeahead suggestions via Redis |
| product-feed-service | Go | Google Shopping / Meta / TikTok product feed generation |
| catalog-translation-service | Go | Auto-translates product titles and descriptions across locales |
| merchandising-service | Go | Hero banners, featured collections, sort overrides per category |

---

### 4. Commerce (32 services)

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
| tip-service | Go | Gratuity at checkout (delivery, in-store) with per-staff distribution |
| reorder-service | Go | One-click reorder of past orders with substitution suggestions |
| cart-recovery-service | Go | Abandoned-cart recovery: targeted email/SMS, discount escalation |
| tax-jurisdiction-service | Go | Resolves tax jurisdiction for any address worldwide |

---

### 5. Supply Chain (20 services)

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
| zone-pricing-service | Go | Geo-zone based shipping price matrices |
| returns-grading-service | Go | Grades returned items A/B/C/scrap and routes to disposition |
| dropship-service | Go | Vendor-direct fulfillment with split-shipment handling |

---

### 6. Financial (20 services)

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
| financial-rules-engine | Haskell | Type-safe pure functional financial calculation engine |
| tax-exemption-service | Go | B2B tax-exempt purchase handling and certificate management |
| multi-currency-account-service | Go | Multi-currency ledger and account management per customer |
| cash-flow-forecast-service | Go | Rolling 13-week cash-flow forecast with scenario analysis |
| revenue-share-service | Go | Splits inbound revenue across partners with idempotent ledger writes |

---

### 7. Customer Experience (20 services)

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
| review-summary-service | Go | Aggregated review analytics and summary per product |
| in-store-pickup-service | Go | BOPIS flow — store selection, pickup window, ready alerts |
| notification-frequency-service | Go | Caps notification frequency per user / channel to prevent fatigue |

---

### 8. Communications (14 services)

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
| line-service | Go | LINE Messaging API integration (Japan/Asia channel) |
| rcs-service | Go | Rich Communication Services outbound (Android-rich SMS replacement) |

---

### 9. Content (13 services)

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
| storefront-cms-adapter | Ruby | Magento/Spree-compatible CMS adapter service |
| content-webhook-service | Crystal | High-throughput inbound content webhook processor |
| moderation-service | Go | UGC moderation: rules + ML classifier + human-review queue |
| dam-service | Go | Digital Asset Management — taxonomy, tagging, search, rights |

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

### 11. B2B (11 services)

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
| punchout-service | Go | OCI / cXML PunchOut — B2B procurement system integration |

---

### 12. Integrations (18 services)

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
| magento-sync-service | PHP | Bi-directional sync with Magento 2 catalog and orders |
| woocommerce-adapter-service | PHP | WooCommerce product, order, and customer sync adapter |
| zapier-connector-service | Go | Public Zapier app exposing ShopOS triggers and actions |
| make-connector-service | Go | Make.com (Integromat) connector |

---

### 13. Affiliate (7 services)

| Service | Language | Responsibility |
|---|---|---|
| affiliate-service | Go | Affiliate partner accounts, tracking links, commission tiers |
| referral-service | Go | Customer referral codes, conversion tracking, reward triggers |
| influencer-service | Go | Influencer campaign management and UTM attribution |
| commission-payout-service | Go | Commission aggregation, tax rules, and payout batching |
| click-tracking-service | Go | High-throughput affiliate click recording via Redis |
| fraud-prevention-affiliate-service | Go | Affiliate fraud detection — click stuffing, cookie dropping |
| brand-partner-service | Go | Brand-tier partnerships — co-marketing, exclusive products |

---

### 14. Marketplace (10 services)

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
| seller-onboarding-service | Go | Walks new sellers through KYC, store setup, payout configuration |
| seller-tier-service | Go | Bronze/silver/gold tiers based on volume, ratings, disputes |

---

### 15. Gamification (7 services)

| Service | Language | Responsibility |
|---|---|---|
| points-service | Go | Points accrual and balance management via Redis |
| badge-service | Go | Badge definitions, award triggers, and display |
| leaderboard-service | Go | Real-time leaderboards via Redis sorted sets |
| challenge-service | Go | Timed challenges and progress tracking |
| reward-redemption-service | Go | Reward catalogue and redemption workflow |
| streak-service | Go | Daily/weekly streak tracking and incentives |
| quest-service | Go | Multi-step quests with state-machine progress tracking |

---

### 16. Developer Platform (8 services)

| Service | Language | Responsibility |
|---|---|---|
| api-management-service | Go | API product management — plans, versions, quotas |
| sandbox-service | Go | Isolated sandbox environments for API testing |
| developer-portal-backend | Node.js | Developer portal API — apps, credentials, docs |
| oauth-client-service | Go | OAuth2 client registration and credential management |
| api-analytics-service | Go | Per-developer API usage analytics |
| webhook-management-service | Go | Developer webhook subscription management |
| sdk-generator-service | Go | Generates TS/Python/Go/Java SDKs from OpenAPI + buf descriptors |
| api-changelog-service | Go | Tracks API breaking/additive changes; publishes changelog |

---

### 17. Compliance (7 services)

| Service | Language | Responsibility |
|---|---|---|
| data-retention-service | Go | Data retention policy enforcement and deletion |
| consent-audit-service | Go | Consent record keeping and audit trail |
| privacy-request-service | Go | GDPR/CCPA data subject request orchestration |
| compliance-reporting-service | Java | Regulatory report generation (GDPR, PCI, SOC2) |
| data-lineage-service | Go | Data flow tracking across services |
| pci-scope-service | Go | Tracks PCI-DSS scope — which services touch cardholder data |
| soc2-evidence-service | Go | Continuously collects SOC2 evidence into audit packs |

---

### 18. Sustainability (6 services)

| Service | Language | Responsibility |
|---|---|---|
| carbon-tracker-service | Go | Carbon footprint calculation per order and shipment |
| eco-score-service | Go | Product environmental impact scoring |
| green-shipping-service | Go | Low-carbon carrier selection and routing |
| sustainability-reporting-service | Go | ESG metrics aggregation and reporting |
| offset-service | Go | Carbon offset purchase and certification tracking |
| circular-economy-service | Go | Repair/resale/refurbish/recycle tracking; diversion-from-landfill metric |

---

### 19. Web (7 services)

| Service | Framework | Responsibility |
|---|---|---|
| storefront-service | Next.js 14 (React/TS) | Customer-facing shopping experience — SSR for SEO |
| admin-dashboard-service | React + Vite (TS) | Admin and merchant management portal |
| seller-portal-service | Vue.js 3 (TS) | Marketplace seller portal — listings, analytics, payouts |
| partner-portal-service | Angular 18 (TS) | B2B partner portal — contracts, orders, invoices |
| mobile-app-service | React Native / Expo (TS) | iOS + Android customer app |
| developer-portal-service | React + Vite (TS) | Developer portal — API docs, sandbox, OAuth apps |
| mobile-flutter-service | Dart / Flutter | iOS + Android customer app (Flutter native alternative) |

---

### 20. Events & Ticketing (7 services)

| Service | Language | Responsibility |
|---|---|---|
| event-service | Elixir | Event creation, scheduling, and lifecycle management |
| ticket-service | Go | Ticket issuance, validation, and transfer |
| seat-map-service | Go | Venue seat map management and availability |
| venue-service | Go | Venue profiles, capacity, and facility management |
| booking-service | Go | Event booking orchestration and confirmation |
| check-in-service | Go | QR-code check-in and attendance tracking via Redis |
| waitlist-event-service | Go | Sold-out event waitlist with auto-purchase on release |

---

### 21. Auction (5 services)

| Service | Language | Responsibility |
|---|---|---|
| auction-service | Elixir | Real-time auction lifecycle using BEAM/OTP concurrency |
| bidding-service | Go | Bid submission, validation, and real-time leaderboard |
| reserve-price-service | Go | Reserve price management and auto-extension rules |
| auction-settlement-service | Java | Auction close, winner selection, and payment initiation |
| proxy-bid-service | Go | Manages proxy bids — auto-counter on behalf of holder |

---

### 22. Rental (5 services)

| Service | Language | Responsibility |
|---|---|---|
| rental-service | Go | Rental agreement creation and lifecycle management |
| lease-service | Kotlin | Long-term lease management and renewal scheduling |
| damage-deposit-service | Go | Security deposit collection, hold, and release |
| availability-calendar-service | Go | Real-time rental item availability calendar |
| insurance-rider-service | Go | Optional insurance rider (quote → bind → claim) for rentals |

---

## Docs

| Topic | Path |
|---|---|
| Getting started | [GETTING_STARTED.md](GETTING_STARTED.md) |
| Authoritative project guide (read this if you're an AI agent) | [CLAUDE.md](CLAUDE.md) |
| Architecture | [docs/architecture/](docs/architecture/) |
| ADRs | [docs/adr/](docs/adr/) |
| Runbooks (deployment, incident, rollback, postgres-failover, kafka-consumer-lag) | [docs/runbooks/](docs/runbooks/) |
| CI/CD (15 platforms) | [ci/README.md](ci/README.md) |
| Helm Charts (303 per-service) | [helm/README.md](helm/README.md) |
| Infrastructure (Terraform/OpenTofu/Crossplane/Ansible/Patroni/Atlantis/Nomad) | [infra/README.md](infra/README.md) |
| GitOps (ArgoCD ApplicationSet + Flux HelmReleases, 303 each) | [gitops/README.md](gitops/README.md) |
| Observability (Prom + Grafana + Loki + Tempo + OTel + ...) | [observability/README.md](observability/README.md) |
| Security (Vault + Istio mTLS + OPA + Kyverno + Falco + ...) | [security/README.md](security/README.md) |
| Kubernetes (RBAC, NetworkPolicies, PDBs, Velero, KEDA, Karpenter, VPA, scaling/, manifests/) | [kubernetes/README.md](kubernetes/README.md) |
| Messaging (Kafka + Strimzi topics, RabbitMQ, NATS, Redpanda, Zilla) | [messaging/README.md](messaging/README.md) |
| Networking (Istio + Cilium + Traefik + Caddy + Anubis + Varnish + edge/spin) | [networking/README.md](networking/README.md) |
| Databases (Flyway 11 schemas, ClickHouse/Weaviate/Neo4j/TimescaleDB/LakeFS/Dgraph/Yugabyte) | [databases/README.md](databases/README.md) |
| Storage (Longhorn, Rook-Ceph PV providers) | [storage/README.md](storage/README.md) |
| Data (Airflow/dbt/Spark/Airbyte/Cube/Metabase/OpenLineage/GreatExpectations) | [data/README.md](data/README.md) |
| Streaming (Debezium CDC + Apache Flink) | [streaming/README.md](streaming/README.md) |
| Registry (Harbor, Nexus, Gitea, Zot, ChartMuseum) | [registry/README.md](registry/README.md) |
| ML Platform (MLflow + Feast) | [ml/README.md](ml/README.md) |
| Chaos Engineering (Chaos Mesh + LitmusChaos) | [chaos/README.md](chaos/README.md) |
| Load Testing (k6 + Locust + Gatling) | [load-testing/README.md](load-testing/README.md) |
| Proto / gRPC (58 protobuf files, Buf, breaking-change CI) | [proto/README.md](proto/README.md) |
| Kafka Events (20 Avro schemas + Strimzi KafkaTopic CRDs) | [events/README.md](events/README.md) |
| Backstage developer portal (348 entries) | [backstage/README.md](backstage/README.md) |
| Workflow / Temporal | [workflow/README.md](workflow/README.md) |
| Dev Experience (Coder/DevSpace/n8n/Windmill/Score/Backstage Templates) | [dev/README.md](dev/README.md) |
| Feature flags (Unleash + OpenFeature) | [feature-flags/README.md](feature-flags/README.md) |
| FinOps (Kubecost + OpenCost) | [finops/README.md](finops/README.md) |
| Incident management (Cachet, Grafana Incident, Grafana OnCall) | [incident/README.md](incident/README.md) |
| API management (APISIX, Tyk, Hasura) | [api-management/README.md](api-management/README.md) |
| API testing (Hurl + Spectral) | [api-testing/hurl/README.md](api-testing/hurl/README.md) |
| Testing (Pact 9 contracts, Playwright, Karate, Testcontainers, Artillery) | [testing/](testing/) |
| Build tooling (Earthly, Ko, Kaniko) | [build/](build/) |
| OpenAPI specs | [openapi/README.md](openapi/README.md) |
| Generators | [scripts/](scripts/) |

## License

Apache 2.0
