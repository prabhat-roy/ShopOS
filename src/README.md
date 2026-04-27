# Services â€” ShopOS

263 services (256 backend microservices + 7 frontend) across 22 business domains. Each service
is a self-contained unit with its own codebase, database, Dockerfile, Helm chart, and CI pipeline.

---

## Domain Overview

| # | Domain | Directory | Services | Languages |
|---|---|---|---|---|
| 1 | Platform | `platform/` | 34 | Go, Java, Python, Node.js, Elixir, Clojure, Swift, Zig |
| 2 | Identity | `identity/` | 11 | Go, Java, Rust |
| 3 | Catalog | `catalog/` | 16 | Go, Java, Kotlin, Python, Node.js |
| 4 | Commerce | `commerce/` | 28 | Go, Java, Kotlin, Python, C#, Rust, Node.js |
| 5 | Supply Chain | `supply-chain/` | 18 | Go, Java, Kotlin, Python, Node.js |
| 6 | Financial | `financial/` | 18 | Go, Java, Kotlin, Haskell |
| 7 | Customer Experience | `customer-experience/` | 18 | Go, Java, Node.js |
| 8 | Communications | `communications/` | 12 | Go, Python, Node.js |
| 9 | Content | `content/` | 11 | Go, Java, Python, Node.js, Ruby, Crystal |
| 10 | Analytics & AI | `analytics-ai/` | 13 | Python, Java, Scala |
| 11 | B2B | `b2b/` | 10 | Go, Java, Kotlin |
| 12 | Integrations | `integrations/` | 16 | Go, Java, Node.js, PHP |
| 13 | Affiliate | `affiliate/` | 6 | Go |
| 14 | Marketplace | `marketplace/` | 8 | Go, Java, Node.js |
| 15 | Gamification | `gamification/` | 6 | Go |
| 16 | Developer Platform | `developer-platform/` | 6 | Go, Node.js |
| 17 | Compliance | `compliance/` | 5 | Go, Java |
| 18 | Sustainability | `sustainability/` | 5 | Go |
| 19 | Events & Ticketing | `events-ticketing/` | 6 | Go, Elixir |
| 20 | Auction | `auction/` | 4 | Go, Java, Elixir |
| 21 | Rental | `rental/` | 4 | Go, Kotlin |
| 22 | Web (Frontend) | `web/` | 7 | Next.js, React, Vue.js, Angular, React Native, Flutter, Dart |
| | Total | | 263 | 19 languages |

---

## Service Structure

Every service follows the same layout regardless of language:

```
src/{domain}/{service-name}/
â”œâ”€â”€ Dockerfile                  â† Multi-stage build, non-root user, minimal base
â”œâ”€â”€ Makefile                    â† build, test, lint, run targets
â”œâ”€â”€ .env.example                â† All environment variables documented
â”œâ”€â”€ README.md                   â† Service-specific docs
â”‚
â”œâ”€â”€ (Go service)
â”‚   â”œâ”€â”€ main.go
â”‚   â”œâ”€â”€ go.mod / go.sum
â”‚   â””â”€â”€ internal/
â”‚
â”œâ”€â”€ (Java/Kotlin service)
â”‚   â”œâ”€â”€ pom.xml / build.gradle.kts
â”‚   â””â”€â”€ src/main/java|kotlin/com/enterprise/{pkg}/
â”‚       â””â”€â”€ Application.java|kt
â”‚
â”œâ”€â”€ (Python service)
â”‚   â”œâ”€â”€ main.py
â”‚   â””â”€â”€ requirements.txt
â”‚
â”œâ”€â”€ (Node.js service)
â”‚   â”œâ”€â”€ index.js
â”‚   â””â”€â”€ package.json
â”‚
â”œâ”€â”€ (C# service)
â”‚   â”œâ”€â”€ Program.cs
â”‚   â””â”€â”€ {Service}.csproj
â”‚
â”œâ”€â”€ (Rust service)
â”‚   â”œâ”€â”€ src/main.rs
â”‚   â””â”€â”€ Cargo.toml
â”‚
â””â”€â”€ (Scala service)
    â”œâ”€â”€ src/main/scala/com/enterprise/{pkg}/Main.scala
    â””â”€â”€ build.sbt
```

---

## Service Contracts

Every service exposes:

| Endpoint | Purpose |
|---|---|
| `GET /healthz` | Returns `{"status":"ok"}` â€” used by Kubernetes liveness/readiness probes |
| `GET /metrics` | Prometheus metrics (Phase 4 instrumentation) |
| gRPC health check | `grpc.health.v1.Health/Check` |

---

## Communication Rules

1. Synchronous: gRPC for reads and commands that need a response
2. Asynchronous: Kafka events for cross-domain side effects
3. Never access another service's database directly
4. Never share a database between two services
5. All `.proto` files live in `proto/` â€” generated code goes into each service

---

## Database Assignment

| Language services | Primary DB | Notes |
|---|---|---|
| Go (most) | PostgreSQL | golang-migrate for schema migrations |
| Java / Kotlin | PostgreSQL | Flyway for schema migrations |
| Python analytics | Cassandra / ClickHouse | High-volume time-series |
| Node.js review/CMS | MongoDB | Nested document structure |
| Cart / session | Redis | Ephemeral, sub-millisecond |
| Search | Elasticsearch | Full-text + faceted |
| ML / RAG | Weaviate | Vector embeddings |
| Recommendations | Neo4j | Graph traversal |

---

## Ports

### gRPC Port Ranges

| Domain | Range |
|---|---|
| Platform | 50051â€“50059 |
| Identity | 50060â€“50069 |
| Catalog | 50070â€“50079 |
| Commerce | 50080â€“50099 |
| Supply Chain | 50100â€“50109 |
| Financial | 50110â€“50119 |
| Customer Experience | 50120â€“50129 |
| Communications | 50130â€“50139 |
| Content | 50140â€“50149 |
| Analytics & AI | 50150â€“50159 |
| B2B | 50160â€“50169 |
| Integrations | 50170â€“50179 |
| Affiliate | 50200â€“50209 |

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

---

## Building a Single Service

```bash
# Go service
cd src/platform/api-gateway
make build         # go build
make test          # go test ./...
make docker-build  # docker build

# Java/Kotlin service
cd src/commerce/order-service
make build         # mvn package / gradle build
make test          # mvn test / gradle test

# Python service
cd src/analytics-ai/recommendation-service
pip install -r requirements.txt
make test          # pytest

# Node.js service
cd src/communications/notification-orchestrator
npm ci
npm test
```

---

## Building All Services

```bash
# Build all Docker images
make build-all

# Run all tests
make test-all

# Push all images to Harbor
make push-all HARBOR_REGISTRY=harbor.shopos.internal IMAGE_TAG=v1.0.0
```

---

## Local Development

See [GETTING_STARTED.md](../GETTING_STARTED.md) for full local dev setup including
Docker Compose, Skaffold, and Tilt instructions.
