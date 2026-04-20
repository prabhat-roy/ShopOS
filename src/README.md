# Services — ShopOS

154 microservices across 13 business domains. Each service is a self-contained unit with
its own codebase, database, Dockerfile, Helm chart, and CI pipeline.

---

## Domain Overview

| # | Domain | Directory | Services | Languages |
|---|---|---|---|---|
| 1 | Platform | `platform/` | 22 | Go, Java, Python, Node.js |
| 2 | Identity | `identity/` | 8 | Go, Java, Rust |
| 3 | Catalog | `catalog/` | 12 | Go, Java, Kotlin, Python, Node.js |
| 4 | Commerce | `commerce/` | 23 | Go, Java, Kotlin, Python, C#, Rust, Node.js |
| 5 | Supply Chain | `supply-chain/` | 13 | Go, Java, Kotlin, Python, Node.js |
| 6 | Financial | `financial/` | 11 | Go, Java, Kotlin |
| 7 | Customer Experience | `customer-experience/` | 14 | Go, Java, Node.js |
| 8 | Communications | `communications/` | 9 | Go, Python, Node.js |
| 9 | Content | `content/` | 8 | Go, Java, Python, Node.js |
| 10 | Analytics & AI | `analytics-ai/` | 13 | Python, Java, Scala |
| 11 | B2B | `b2b/` | 7 | Go, Java, Kotlin |
| 12 | Integrations | `integrations/` | 10 | Go, Java, Node.js |
| 13 | Affiliate | `affiliate/` | 4 | Go |
| | **Total** | | **154** | **8 languages** |

---

## Service Structure

Every service follows the same layout regardless of language:

```
src/{domain}/{service-name}/
├── Dockerfile                  ← Multi-stage build, non-root user, minimal base
├── Makefile                    ← build, test, lint, run targets
├── .env.example                ← All environment variables documented
├── README.md                   ← Service-specific docs
│
├── (Go service)
│   ├── main.go
│   ├── go.mod / go.sum
│   └── internal/
│
├── (Java/Kotlin service)
│   ├── pom.xml / build.gradle.kts
│   └── src/main/java|kotlin/com/enterprise/{pkg}/
│       └── Application.java|kt
│
├── (Python service)
│   ├── main.py
│   └── requirements.txt
│
├── (Node.js service)
│   ├── index.js
│   └── package.json
│
├── (C# service)
│   ├── Program.cs
│   └── {Service}.csproj
│
├── (Rust service)
│   ├── src/main.rs
│   └── Cargo.toml
│
└── (Scala service)
    ├── src/main/scala/com/enterprise/{pkg}/Main.scala
    └── build.sbt
```

---

## Service Contracts

Every service exposes:

| Endpoint | Purpose |
|---|---|
| `GET /healthz` | Returns `{"status":"ok"}` — used by Kubernetes liveness/readiness probes |
| `GET /metrics` | Prometheus metrics (Phase 4 instrumentation) |
| gRPC health check | `grpc.health.v1.Health/Check` |

---

## Communication Rules

1. **Synchronous**: gRPC for reads and commands that need a response
2. **Asynchronous**: Kafka events for cross-domain side effects
3. **Never** access another service's database directly
4. **Never** share a database between two services
5. All `.proto` files live in `proto/` — generated code goes into each service

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
| Platform | 50051–50059 |
| Identity | 50060–50069 |
| Catalog | 50070–50079 |
| Commerce | 50080–50099 |
| Supply Chain | 50100–50109 |
| Financial | 50110–50119 |
| Customer Experience | 50120–50129 |
| Communications | 50130–50139 |
| Content | 50140–50149 |
| Analytics & AI | 50150–50159 |
| B2B | 50160–50169 |
| Integrations | 50170–50179 |
| Affiliate | 50200–50209 |

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
