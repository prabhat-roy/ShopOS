# Services — ShopOS

303 services (296 backend microservices + 7 frontend) across 22 business domains. Each
service is a self-contained unit with its own codebase, database, Dockerfile, Helm chart,
and CI pipeline.

---

## Domain Overview

| # | Domain | Directory | Services | Languages |
|---|---|---|---|---|
| 1 | Platform | `platform/` | 40 | Go, Java, Python, Node.js, Elixir, Clojure, Swift, Zig |
| 2 | Identity | `identity/` | 14 | Go, Java, Rust |
| 3 | Catalog | `catalog/` | 19 | Go, Java, Kotlin, Python, Node.js |
| 4 | Commerce | `commerce/` | 32 | Go, Java, Kotlin, Python, C#, Rust, Node.js |
| 5 | Supply Chain | `supply-chain/` | 20 | Go, Java, Kotlin, Python, Node.js |
| 6 | Financial | `financial/` | 20 | Go, Java, Kotlin, Haskell |
| 7 | Customer Experience | `customer-experience/` | 20 | Go, Java, Node.js |
| 8 | Communications | `communications/` | 14 | Go, Python, Node.js |
| 9 | Content | `content/` | 13 | Go, Java, Python, Node.js, Ruby, Crystal |
| 10 | Analytics & AI | `analytics-ai/` | 13 | Python, Java, Scala |
| 11 | B2B | `b2b/` | 11 | Go, Java, Kotlin |
| 12 | Integrations | `integrations/` | 18 | Go, Java, Node.js, PHP |
| 13 | Affiliate | `affiliate/` | 7 | Go |
| 14 | Marketplace | `marketplace/` | 10 | Go, Java, Node.js |
| 15 | Gamification | `gamification/` | 7 | Go |
| 16 | Developer Platform | `developer-platform/` | 8 | Go, Node.js |
| 17 | Compliance | `compliance/` | 7 | Go, Java |
| 18 | Sustainability | `sustainability/` | 6 | Go |
| 19 | Events & Ticketing | `events-ticketing/` | 7 | Go, Elixir |
| 20 | Auction | `auction/` | 5 | Go, Java, Elixir |
| 21 | Rental | `rental/` | 5 | Go, Kotlin |
| 22 | Web (Frontend) | `web/` | 7 | Next.js, React, Vue.js, Angular, React Native, Flutter, Dart |
| | Total | | 303 | 19 languages |

See the per-service registry table in [`../README.md`](../README.md) and the
authoritative service catalog at [`../backstage/catalog-info.yaml`](../backstage/catalog-info.yaml).

---

## Adding a new service

```bash
bash scripts/bash/scaffold-service.sh <domain> <name> <port>          # Go skeleton
```

See [`scripts/README.md`](../scripts/README.md) for details.

---

## Service Structure

Every service follows the same layout regardless of language:

```
src/{domain}/{service-name}/
├── Dockerfile           # Multi-stage build, non-root user, minimal base
├── Makefile             # build, test, lint, run targets
├── .env.example         # All environment variables documented
├── README.md            # Service-specific docs
├── catalog-info.yaml    # Backstage entity (optional; root catalog covers all)
└── (language-specific)
    ├── Go         → main.go, go.mod, internal/
    ├── Java       → pom.xml, src/main/java/com/enterprise/{pkg}/Application.java
    ├── Kotlin     → build.gradle.kts, src/main/kotlin/com/enterprise/{pkg}/Application.kt
    ├── Python     → main.py, requirements.txt
    ├── Node.js    → index.js, package.json
    ├── C#         → Program.cs, {Service}.csproj
    ├── Rust       → src/main.rs, Cargo.toml
    ├── Scala      → src/main/scala/com/enterprise/{pkg}/Main.scala, build.sbt
    ├── Elixir     → lib/<app>.ex, mix.exs
    ├── Haskell    → src/Main.hs, <service>.cabal, stack.yaml
    ├── PHP        → public/index.php, composer.json
    ├── Ruby       → app.rb, Gemfile
    ├── Dart       → lib/main.dart, pubspec.yaml          (Flutter frontend)
    ├── Swift      → Sources/App/main.swift, Package.swift
    ├── Clojure    → src/<ns>/core.clj, project.clj
    ├── Crystal    → src/main.cr, shard.yml
    ├── Zig        → src/main.zig, build.zig
    └── Gleam      → src/app.gleam, gleam.toml
```

---

## Service Contracts

Every service exposes:

| Endpoint | Purpose |
|---|---|
| `GET /healthz` | Returns `{"status":"ok"}` — used by Kubernetes liveness/readiness probes |
| `GET /metrics` | Prometheus metrics (scraped via OpenTelemetry agent or directly) |
| `grpc.health.v1.Health/Check` | gRPC health check protocol |

---

## Communication Rules

1. Synchronous: gRPC for reads and commands that need a response
2. Asynchronous: Kafka events for cross-domain side effects (Strimzi `KafkaTopic` CRDs in [`messaging/kafka/topics.yaml`](../messaging/kafka/topics.yaml))
3. Never access another service's database directly
4. Never share a database between two services
5. All `.proto` files live in [`proto/`](../proto/) — generated code goes into each service
6. Buf breaking-change check blocks merge on proto regressions ([`ci/jenkins/proto-breaking-check.Jenkinsfile`](../ci/jenkins/proto-breaking-check.Jenkinsfile))

---

## Database Assignment

| Language services | Primary DB | Notes |
|---|---|---|
| Go (most) | PostgreSQL | golang-migrate for schema migrations |
| Java / Kotlin | PostgreSQL | Flyway for schema migrations (11 domain schemas under [`databases/postgres/`](../databases/postgres/)) |
| Python analytics | Cassandra / ClickHouse | High-volume time-series |
| Node.js review/CMS | MongoDB | Nested document structure |
| Cart / session | Redis | Ephemeral, sub-millisecond |
| Search | Elasticsearch / Meilisearch | Full-text + faceted |
| ML / RAG | Weaviate / Dgraph | Vector embeddings + graph |
| Recommendations | Neo4j / Dgraph | Graph traversal |
| Geo-distributed SQL | CockroachDB / YugabyteDB | When ACID + multi-region required |

Dynamic Postgres credentials per service are issued by Vault (1h TTL, 24h max) — see
[`security/vault/bootstrap/02-secret-engines.sh`](../security/vault/bootstrap/02-secret-engines.sh).

---

## Ports

### gRPC Port Ranges

| Domain | Range |
|---|---|
| Platform | 50051–50059, 50210–50214, 50352–50359 |
| Identity | 50060–50069, 50215–50217, 50345 |
| Catalog | 50070–50079, 50180–50181, 50218–50220, 50370, 50375–50377 |
| Commerce | 50080–50099, 50183–50185, 50221–50233 |
| Supply Chain | 50100–50109, 50193–50194, 50226–50229, 50372, 50378–50379 |
| Financial | 50110–50119, 50191–50192, 50230–50233, 50360–50364 |
| Customer Experience | 50120–50129, 50182, 50186–50188, 50234–50238, 50371 |
| Communications | 50130–50139 |
| Content | 50140–50150, 50240 |
| Analytics & AI | 50150–50159, 50190 |
| B2B | 50160–50169, 50241–50244 |
| Integrations | 50170–50179, 50195–50196, 50244 |
| Affiliate | 50200–50209, 50248–50249 |
| Marketplace | 50250–50257 |
| Gamification | 50260–50266 |
| Developer Platform | 50270–50271 |
| Compliance | 50280–50286 |
| Sustainability | 50290–50295 |
| Events & Ticketing | 50300–50306 |
| Auction | 50310–50314 |
| Rental | 50320–50324 |

### HTTP Ports (external-facing)

| Service | Port |
|---|---|
| api-gateway | 8080 |
| web-bff | 8081 |
| mobile-bff | 8082 |
| partner-bff | 8083 |
| admin-portal | 8085 |
| graphql-gateway | 8086 |
| tenant-service | 8087 |
| reports-portal-service | 8219 |
| graphql-federation-service | 8220 |
| product-feed-service | 8221 |
| zapier-connector-service | 8222 |
| make-connector-service | 8223 |
| sdk-generator-service | 8224 |
| api-changelog-service | 8225 |
| punchout-service | 8226 |

---

## Building a Single Service

```bash
# Go
cd src/platform/api-gateway && make build && make test && make docker
# Java / Kotlin
cd src/commerce/order-service && make build && make test
# Python
cd src/analytics-ai/recommendation-service && pip install -r requirements.txt && make test
# Node.js
cd src/communications/notification-orchestrator && npm ci && npm test
```

---

## Building All Services

```bash
make build-all          # all images via Earthly + Ko
make test-all           # all language test suites
make push-all HARBOR_REGISTRY=harbor.shopos.internal IMAGE_TAG=v1.0.0
```

---

## Local Development

See [GETTING_STARTED.md](../GETTING_STARTED.md) for Docker Compose, Skaffold, and Tilt
instructions. For cluster-attached dev see [`dev/devspace/`](../dev/devspace/) and
[`dev/coder/`](../dev/coder/).
