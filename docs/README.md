# Documentation â€” ShopOS

Architecture decision records, system design documents, and operational runbooks.

---

## Directory Structure

```
docs/
â”œâ”€â”€ architecture/
â”‚   â”œâ”€â”€ system-overview.md          â† High-level architecture and design philosophy
â”‚   â”œâ”€â”€ domain-map.md               â† 13 business domains and their service boundaries
â”‚   â”œâ”€â”€ communication-patterns.md   â† gRPC sync, Kafka async, WebSocket real-time
â”‚   â”œâ”€â”€ database-strategy.md        â† Polyglot persistence rationale per domain
â”‚   â””â”€â”€ security-model.md           â† Defence-in-depth layers and threat model
â”œâ”€â”€ runbooks/
â”‚   â”œâ”€â”€ incident-response.md        â† On-call escalation and triage procedures
â”‚   â”œâ”€â”€ database-failover.md        â† Postgres/Cassandra failover steps
â”‚   â”œâ”€â”€ kafka-recovery.md           â† Broker recovery and consumer group reset
â”‚   â””â”€â”€ rollback.md                 â† Service and cluster rollback procedures
â””â”€â”€ adr/
    â”œâ”€â”€ 001-microservice-boundaries.md
    â”œâ”€â”€ 002-polyglot-persistence.md
    â”œâ”€â”€ 003-event-driven-architecture.md
    â”œâ”€â”€ 004-api-gateway-pattern.md
    â”œâ”€â”€ 005-saga-orchestration.md
    â””â”€â”€ 006-gitops-delivery.md
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
- gRPC: synchronous request/response â€” reads, commands that need a response
- Kafka: asynchronous domain events â€” cross-domain side effects
- WebSocket: real-time push â€” live chat, in-app notifications
- REST: external-facing â€” BFF â†’ client, webhooks â†’ partners

### [database-strategy.md](architecture/database-strategy.md)
Explains why each database is assigned to each domain:

| Database | Domains | Rationale |
|---|---|---|
| PostgreSQL | Identity, Commerce, Financial, B2B, Affiliate | ACID, complex joins |
| MongoDB | Catalog, Content, Customer-Experience | Flexible nested documents |
| Redis | Platform, Identity, Commerce (cart/session) | Sub-millisecond, ephemeral |
| Cassandra | Analytics & AI | Write-heavy time-series |
| TimescaleDB | Analytics & AI, Platform | Time-series metrics, inventory events |
| Memcached | Commerce, Platform | High-throughput hot read cache |
| Elasticsearch | Catalog (search) | Full-text + faceted filtering |
| ClickHouse | Analytics & AI | OLAP aggregations |
| Weaviate | Analytics & AI | Semantic vector search + RAG |
| Neo4j | Analytics & AI | Graph-based recommendations |

### [security-model.md](architecture/security-model.md)
The layered security posture:
1. Cluster perimeter â€” Cilium CNI, network policies, Coraza WAF
2. Service mesh â€” Istio mTLS between all pods, SPIFFE/SPIRE workload identity
3. Identity â€” Keycloak OIDC, Dex federation, SPIRE X.509 SVIDs
4. Secrets â€” Vault dynamic credentials, External Secrets Operator, Sealed Secrets
5. Policy â€” OPA/Gatekeeper + Kyverno admission, OpenFGA authorisation
6. Runtime â€” Falco, Tetragon (eBPF), Tracee
7. Supply chain â€” Cosign signing, Rekor transparency, Kyverno image verification

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
Severity classification (P1â€“P4), on-call rotation, escalation paths, communication templates,
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
