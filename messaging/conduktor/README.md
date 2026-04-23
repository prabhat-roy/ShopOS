# Conduktor Gateway

## What is Conduktor Gateway?

Conduktor Gateway is a **Kafka proxy** that sits transparently between all Kafka
producers/consumers and the broker cluster. Producers and consumers connect to the
gateway on port `9099` instead of directly to Kafka on `9092`. All traffic is
forwarded to the real brokers — the gateway intercepts the Kafka wire protocol in-flight.

No client code changes are required. The gateway is completely transparent to
producers and consumers: they use the standard Kafka client library, just pointed
at a different bootstrap address.

---

## Architecture

```
┌──────────────────┐        ┌─────────────────────────┐        ┌──────────────┐
│  Kafka Producers │        │                         │        │              │
│  (224 services)  │──9099─►│  Conduktor Gateway      │──9092─►│  Kafka       │
│                  │        │                         │        │  Brokers     │
│  Kafka Consumers │◄─9099──│  Interceptor Chain      │◄─9092──│  (3-node)    │
└──────────────────┘        │  1. Schema Validation   │        └──────────────┘
                            │  2. Rate Limiting       │
                            │  3. PII Data Masking    │
                            │  4. Audit Logging       │
                            └─────────────────────────┘
                                        │
                                        ▼
                            ┌─────────────────────────┐
                            │  gateway-audit  topic   │
                            │  (Kafka internal)       │
                            └─────────────────────────┘
```

---

## Interceptor Chain

### 1. Schema Validation (priority 1)

Every produce request is validated against the registered Avro schema in the
Confluent Schema Registry (`http://schema-registry:8081`).

- Records that do not conform to the registered schema are **rejected immediately**.
- Producers receive a `POLICY_VIOLATION` error.
- Invalid records **never reach the broker** — no poison pills in the topic.

### 2. Rate Limiting (priority 2)

Each producer `client-id` is allowed at most **100 messages/second**.

- Excess produce requests are rejected with `POLICY_VIOLATION`.
- The limit is enforced per gateway instance; if the gateway is scaled horizontally,
  the state is synchronised via the `gateway-rate-limit-state` compacted topic.
- Tune `messagesPerSecond` per use-case — bulk importers should use a dedicated client-id
  that can be allowlisted.

### 3. PII Data Masking (priority 3)

Applied to all topics matching `identity.user.*`, `commerce.order.*`, and
`customer-experience.*`.

| Field | Masking Strategy |
|---|---|
| `email` | Last 10 characters replaced with `*` |
| `phone` | Last 7 digits replaced with `*` |
| `firstName` | First 3 characters replaced with `*` |
| `lastName` | Entire value replaced with `*` |
| `dateOfBirth` | Replaced with `0000-00-00` |
| `ipAddress` | Last 6 characters replaced with `*` |
| `creditCardNumber` | Last 12 digits replaced with `*` |
| `address.street` | Entire value replaced with `*` |
| `taxId` | Entire value replaced with `*` |

Consumer groups `compliance-consumer`, `kyc-aml-consumer`, and `gdpr-consumer` receive
unmasked data for regulatory purposes.

### 4. Audit Logging (priority 4)

Every produce and consume operation generates a structured JSON audit event written to
the `gateway-audit` topic.

Audit event fields:

```json
{
  "timestamp": "2026-04-23T10:00:00Z",
  "operation": "PRODUCE",
  "clientId": "order-service",
  "consumerGroup": null,
  "topic": "commerce.order.placed",
  "partition": 2,
  "offset": 184729,
  "decision": "ALLOW",
  "interceptor": "schema-validation",
  "gatewayNode": "conduktor-gateway-0"
}
```

The audit topic retains 7 days of data for compliance review.

---

## Deployment

### Helm install

```bash
helm repo add conduktor https://helm.conduktor.io
helm repo update

helm upgrade --install conduktor-gateway conduktor/conduktor-gateway \
  --namespace messaging \
  --values messaging/conduktor/gateway-config.yaml
```

### Connecting producers/consumers

Change your service's `KAFKA_BOOTSTRAP_SERVERS` from:
```
kafka:9092
```
to:
```
conduktor-gateway:9099
```

That's the only change needed. All interceptors apply automatically.

---

## Monitoring

The gateway exposes Prometheus metrics at `http://conduktor-gateway:8888/metrics`.

Key metrics:

| Metric | Description |
|---|---|
| `gateway_produced_records_total` | Total records forwarded to broker |
| `gateway_blocked_records_total` | Records blocked by any interceptor |
| `gateway_rate_limit_blocked_total` | Records blocked by rate limiter |
| `gateway_schema_validation_errors_total` | Schema validation rejections |
| `gateway_latency_ms_p99` | Gateway overhead p99 latency (should be < 2ms) |

Import the Conduktor Gateway Grafana dashboard from `observability/grafana/dashboards/`.

---

## Virtual Clusters

Virtual clusters provide topic prefix isolation between domains. A producer in the
`commerce` virtual cluster publishes to `my-topic` and it is transparently mapped to
`commerce.my-topic` on the real broker. Consumers in the same virtual cluster only
see their own prefix — providing strong multi-tenancy without separate Kafka clusters.
