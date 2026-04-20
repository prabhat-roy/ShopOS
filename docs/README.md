# Documentation — ShopOS

Architecture decision records, system design documents, and operational runbooks.

---

## Directory Structure

```
docs/
├── architecture/
│   ├── system-overview.md          ← High-level architecture and design philosophy
│   ├── domain-map.md               ← 13 business domains and their service boundaries
│   ├── communication-patterns.md   ← gRPC sync, Kafka async, WebSocket real-time
│   ├── database-strategy.md        ← Polyglot persistence rationale per domain
│   └── security-model.md           ← Defence-in-depth layers and threat model
├── runbooks/
│   ├── incident-response.md        ← On-call escalation and triage procedures
│   ├── database-failover.md        ← Postgres/Cassandra failover steps
│   ├── kafka-recovery.md           ← Broker recovery and consumer group reset
│   └── rollback.md                 ← Service and cluster rollback procedures
└── adr/
    ├── 001-microservice-boundaries.md
    ├── 002-polyglot-persistence.md
    ├── 003-event-driven-architecture.md
    ├── 004-api-gateway-pattern.md
    ├── 005-saga-orchestration.md
    └── 006-gitops-delivery.md
```

---

## Architecture Documents

### [system-overview.md](architecture/system-overview.md)
The overall system topology: how the 154 microservices are grouped into 13 domains, how they
communicate (sync gRPC vs async Kafka), and the shared platform services they rely on
(api-gateway, saga-orchestrator, event-store-service, config-service).

### [domain-map.md](architecture/domain-map.md)
A detailed map of each domain's responsibilities, its bounded context boundaries, and the
events it publishes and consumes. Includes a domain dependency graph showing which domains
call which synchronously vs. event-driven.

### [communication-patterns.md](architecture/communication-patterns.md)
When to use each communication mechanism:
- **gRPC**: synchronous request/response — reads, commands that need a response
- **Kafka**: asynchronous domain events — cross-domain side effects
- **WebSocket**: real-time push — live chat, in-app notifications
- **REST**: external-facing — BFF → client, webhooks → partners

### [database-strategy.md](architecture/database-strategy.md)
Explains why each database is assigned to each domain:

| Database | Domains | Rationale |
|---|---|---|
| PostgreSQL | Identity, Commerce, Financial, B2B, Affiliate | ACID, complex joins |
| MongoDB | Catalog, Content, Customer-Experience | Flexible nested documents |
| Redis | Platform, Identity, Commerce (cart/session) | Sub-millisecond, ephemeral |
| Cassandra / ScyllaDB | Analytics & AI | Write-heavy time-series |
| Elasticsearch | Catalog (search) | Full-text + faceted filtering |
| ClickHouse | Analytics & AI | OLAP aggregations |
| Weaviate | Analytics & AI | Semantic vector search + RAG |
| Neo4j | Analytics & AI | Graph-based recommendations |

### [security-model.md](architecture/security-model.md)
The layered security posture:
1. **Cluster perimeter** — Cilium CNI, network policies, Coraza WAF
2. **Service mesh** — Istio mTLS between all pods, SPIFFE/SPIRE workload identity
3. **Identity** — Keycloak OIDC, Dex federation, SPIRE X.509 SVIDs
4. **Secrets** — Vault dynamic credentials, External Secrets Operator, Sealed Secrets
5. **Policy** — OPA/Gatekeeper + Kyverno admission, OpenFGA authorisation
6. **Runtime** — Falco, Tetragon (eBPF), Tracee
7. **Supply chain** — Cosign signing, Rekor transparency, Kyverno image verification

---

## Architecture Decision Records (ADRs)

| ADR | Title | Status | Decision |
|---|---|---|---|
| [001](adr/001-microservice-boundaries.md) | Microservice Boundaries | Accepted | Domain-driven design; one DB per service |
| [002](adr/002-polyglot-persistence.md) | Polyglot Persistence | Accepted | Right database for each access pattern |
| [003](adr/003-event-driven-architecture.md) | Event-Driven Architecture | Accepted | Kafka for cross-domain async, gRPC for sync |
| [004](adr/004-api-gateway-pattern.md) | API Gateway Pattern | Accepted | Single Go gateway + domain BFFs |
| [005](adr/005-saga-orchestration.md) | Saga Orchestration | Accepted | Centralised orchestrator over choreography |
| [006](adr/006-gitops-delivery.md) | GitOps Delivery | Accepted | ArgoCD App-of-Apps as primary delivery model |

---

## Runbooks

### [incident-response.md](runbooks/incident-response.md)
Severity classification (P1–P4), on-call rotation, escalation paths, communication templates,
and postmortem process.

### [database-failover.md](runbooks/database-failover.md)
Step-by-step procedures for PostgreSQL primary failover (Patroni), Cassandra node replacement,
and MongoDB replica set re-election.

### [kafka-recovery.md](runbooks/kafka-recovery.md)
Broker crash recovery, consumer group offset reset, under-replicated partition remediation,
and ZooKeeper quorum recovery.

### [rollback.md](runbooks/rollback.md)
Service rollback via `helm rollback`, cluster rollback via Velero restore, and ArgoCD
`app rollback` procedures.
