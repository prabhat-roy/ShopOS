# ADR-003: Polyglot Programming — Language per Domain

Status: Accepted  
Date: 2024-01-20  
Deciders: Platform Architecture Team, Domain Leads

---

## Context

ShopOS is built by multiple domain teams with different performance requirements, ecosystem needs, and existing expertise. We chose between enforcing a single language (operational simplicity) vs. polyglot (optimal tool per domain).

Forcing Go onto ML pipelines (poor ML ecosystem) or Python onto latency-sensitive auth (poor concurrency) would be suboptimal at enterprise scale.

---

## Decision

Each service uses the language best suited to its domain, fixed in the service registry and enforced by code review.

| Language | Version | Domains / Rationale |
|---|---|---|
| Go | 1.24 | Platform, API gateways, supply chain — high concurrency, low memory, fast startup |
| Java | 21 LTS | Financial, B2B, integrations — Spring ecosystem, JPA, transaction management |
| Kotlin | 2.1 | Orders, financial workflows — JVM with concise syntax and coroutines |
| Python | 3.13 | All ML/AI, analytics, data pipelines — PyTorch, scikit-learn, no equivalent elsewhere |
| Node.js | 22 LTS | BFF layers, communications — non-blocking I/O, event-driven patterns |
| C# / .NET | 9 | Cart, returns — team expertise, high-performance collections |
| Rust | 1.82+ | Auth, shipping — memory safety without GC, cryptographic operations |
| Scala | 3.x | Batch reporting — Spark integration, functional stream processing |

---

## Rationale

Go suits platform infrastructure: goroutines handle thousands of concurrent connections with minimal memory. The API gateway, rate limiter, and saga orchestrator all benefit from this.

Rust is used only where memory safety and zero-cost abstractions are critical. The auth-service handles cryptographic token signing and validation — Rust's borrow checker eliminates entire classes of security vulnerabilities without a GC pause.

Python is unavoidable for ML/AI. PyTorch, scikit-learn, and the broader ML ecosystem have no Go or Java equivalents at the same maturity level.

Java/Kotlin suit the financial domain: Spring's `@Transactional`, JPA for complex queries, and the JVM's mature JIT for long-running batch processes.

---

## Consequences

Positive: Each team uses optimal tooling; Rust provides memory safety for security-critical paths; Python enables state-of-the-art ML without compromising other services.

Negative: CI must support 8 toolchains; onboarding requires breadth; dependency scanning covers 8 package ecosystems.

Mitigations: Backstage surfaces each service's language; per-language Dockerfile patterns are standardised; CI auto-selects the right image based on detected language; gRPC contracts make language choice invisible at the service boundary.
