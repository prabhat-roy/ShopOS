# ADR-001: gRPC for All Synchronous Internal Communication

Status: Accepted  
Date: 2024-01-15  
Deciders: Platform Architecture Team

---

## Context

ShopOS comprises 154 microservices across 13 domains. Services frequently make synchronous calls to each other â€” `checkout-service` calls `cart-service`, `inventory-service`, `payment-service`, `tax-service`, and `shipping-service` within a single request.

We evaluated three options:

| Option | Pros | Cons |
|---|---|---|
| REST/HTTP+JSON | Familiar, easy to curl | Verbose payloads, no schema enforcement, slower serialisation |
| gRPC/Protobuf | Binary encoding, strong contracts, streaming, codegen | Requires protoc toolchain, harder to curl directly |
| GraphQL | Flexible queries | Complex resolvers, overkill for service-to-service |

---

## Decision

gRPC with Protocol Buffers for all synchronous service-to-service communication.

- All `.proto` files live in `proto/` at the repository root
- Generated stubs are checked in per service so builds don't require protoc
- All gRPC services implement `grpc.health.v1.Health`
- REST is exposed only at BFF layer (web-bff, mobile-bff, partner-bff) for external clients

---

## Rationale

1. Schema enforcement â€” Protobuf contracts are machine-checked at compile time; breaking changes are caught before deployment.
2. Performance â€” Binary serialisation is 3â€“10Ã— smaller than JSON and 2â€“5Ã— faster to parse â€” critical for high-frequency checkout flows.
3. Code generation â€” `protoc` generates client and server stubs for all 8 languages used in ShopOS.
4. Streaming â€” gRPC supports server-streaming, client-streaming, and bidirectional streaming (e.g., live-chat-service).
5. Polyglot parity â€” Every language in ShopOS has a mature, well-maintained gRPC library.

---

## Consequences

Positive: Strong contracts at build time, smaller payloads, automatic client generation, single source of truth for API shapes.

Negative: Cannot `curl` services directly without `grpcurl`; proto changes require regenerating stubs across consumer services.

Mitigations: Buf CLI enforces backward compatibility in CI (`buf breaking`); the GraphQL gateway provides a human-friendly query interface; stubs are checked in so consumers don't need protoc installed.
