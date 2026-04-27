# Kafka Event Schemas â€” ShopOS

This directory contains Avro schemas for all Kafka events published across the ShopOS
platform. Every event is schema-validated via the Confluent Schema Registry before
production and consumption, ensuring contracts between domains are explicit and versioned.

---

## Directory Structure

```
events/
â”œâ”€â”€ identity/
â”‚   â”œâ”€â”€ user-registered.avsc
â”‚   â””â”€â”€ user-deleted.avsc
â”œâ”€â”€ commerce/
â”‚   â”œâ”€â”€ order-placed.avsc
â”‚   â”œâ”€â”€ order-cancelled.avsc
â”‚   â”œâ”€â”€ order-fulfilled.avsc
â”‚   â”œâ”€â”€ payment-processed.avsc
â”‚   â”œâ”€â”€ payment-failed.avsc
â”‚   â””â”€â”€ cart-abandoned.avsc
â”œâ”€â”€ supply-chain/
â”‚   â”œâ”€â”€ shipment-created.avsc
â”‚   â”œâ”€â”€ shipment-updated.avsc
â”‚   â”œâ”€â”€ inventory-low.avsc
â”‚   â””â”€â”€ inventory-restocked.avsc
â”œâ”€â”€ notifications/
â”‚   â”œâ”€â”€ email-requested.avsc
â”‚   â”œâ”€â”€ sms-requested.avsc
â”‚   â””â”€â”€ push-requested.avsc
â”œâ”€â”€ analytics/
â”‚   â”œâ”€â”€ page-viewed.avsc
â”‚   â”œâ”€â”€ product-clicked.avsc
â”‚   â””â”€â”€ search-performed.avsc
â””â”€â”€ security/
    â”œâ”€â”€ fraud-detected.avsc
    â””â”€â”€ login-failed.avsc
```

Total: 20 Avro schema files.

---

## Naming Convention

Topic names follow the pattern:

```
{domain}.{entity}.{event}
```

All components are lowercase, separated by dots.

| Component | Examples |
|---|---|
| `domain` | `identity`, `commerce`, `supplychain`, `analytics`, `security` |
| `entity` | `user`, `order`, `payment`, `shipment`, `inventory`, `page` |
| `event` | past-tense verb â€” `registered`, `placed`, `processed`, `viewed` |

---

## All 20 Kafka Topics

### Identity Domain

| Topic | Publisher | Consumers | Description |
|---|---|---|---|
| `identity.user.registered` | `user-service` | `email-service`, `notification-orchestrator`, `audit-service`, `analytics-service` | Fired when a new user account is created. Contains user ID, email, registration timestamp, and registration source. |
| `identity.user.deleted` | `user-service` | `gdpr-service`, `session-service`, `loyalty-service`, `analytics-service` | Fired on account deletion (GDPR erasure request). Triggers cascade cleanup across services. |

### Commerce Domain

| Topic | Publisher | Consumers | Description |
|---|---|---|---|
| `commerce.order.placed` | `order-service` | `fulfillment-service`, `inventory-service`, `notification-orchestrator`, `analytics-service`, `loyalty-service`, `accounting-service` | Fired when an order transitions to `PLACED` state. Contains order ID, line items, totals, shipping address, and customer ID. |
| `commerce.order.cancelled` | `order-service` | `return-refund-service`, `inventory-service`, `notification-orchestrator`, `analytics-service` | Fired when an order is cancelled (by customer or system). Triggers inventory restock and refund initiation. |
| `commerce.order.fulfilled` | `fulfillment-service` | `notification-orchestrator`, `analytics-service`, `loyalty-service`, `invoice-service` | Fired when all items in an order have shipped. Contains tracking numbers and carrier information. |
| `commerce.payment.processed` | `payment-service` | `order-service`, `notification-orchestrator`, `accounting-service`, `analytics-service` | Fired on successful payment capture. Contains payment ID, amount, currency, and payment method type. |
| `commerce.payment.failed` | `payment-service` | `order-service`, `notification-orchestrator`, `fraud-detection-service` | Fired on payment failure. Contains failure reason code and retry eligibility flag. |
| `commerce.cart.abandoned` | `cart-service` | `notification-orchestrator`, `analytics-service`, `personalization-service` | Fired when a cart with items has been inactive for 1 hour. Triggers abandoned cart email sequence. |

### Supply Chain Domain

| Topic | Publisher | Consumers | Description |
|---|---|---|---|
| `supplychain.shipment.created` | `fulfillment-service` | `tracking-service`, `notification-orchestrator`, `analytics-service` | Fired when a shipping label is generated and a shipment record is created. Contains carrier, tracking number, and estimated delivery date. |
| `supplychain.shipment.updated` | `tracking-service` | `notification-orchestrator`, `order-service`, `analytics-service` | Fired on each carrier status update (picked up, in transit, out for delivery, delivered). |
| `supplychain.inventory.low` | `inventory-service` | `notification-orchestrator`, `purchase-order-service`, `demand-forecast-service` | Fired when stock for a SKU falls below the reorder threshold. Triggers purchase order creation workflow. |
| `supplychain.inventory.restocked` | `warehouse-service` | `inventory-service`, `notification-orchestrator`, `search-service` | Fired when a previously out-of-stock SKU is restocked. Triggers back-in-stock notifications to wishlisted customers. |

### Notifications Domain

| Topic | Publisher | Consumers | Description |
|---|---|---|---|
| `notification.email.requested` | `notification-orchestrator` | `email-service` | Standardised envelope for email dispatch requests. Contains recipient, template ID, and template variables. |
| `notification.sms.requested` | `notification-orchestrator` | `sms-service` | Standardised envelope for SMS dispatch. Contains phone number, message body, and message type. |
| `notification.push.requested` | `notification-orchestrator` | `push-notification-service` | Standardised envelope for mobile push notifications. Contains device token, title, body, and deep-link URL. |

### Analytics Domain

| Topic | Publisher | Consumers | Description |
|---|---|---|---|
| `analytics.page.viewed` | `web-bff`, `mobile-bff` | `analytics-service`, `event-tracking-service`, `personalization-service` | Fired on every page/screen view. Contains user ID (or anonymous ID), page URL, referrer, and device type. |
| `analytics.product.clicked` | `web-bff`, `mobile-bff` | `analytics-service`, `recommendation-service`, `personalization-service` | Fired when a user clicks on a product card. Contains product ID, position in listing, and source page. |
| `analytics.search.performed` | `search-service` | `analytics-service`, `recommendation-service`, `event-tracking-service` | Fired on every search query. Contains query string, result count, applied filters, and whether results were found. |

### Security Domain

| Topic | Publisher | Consumers | Description |
|---|---|---|---|
| `security.fraud.detected` | `fraud-detection-service` | `order-service`, `payment-service`, `notification-orchestrator`, `audit-service` | Fired when the fraud model scores a transaction above the risk threshold. Contains risk score, triggering signals, and recommended action. |
| `security.login.failed` | `auth-service` | `mfa-service`, `notification-orchestrator`, `audit-service`, `fraud-detection-service` | Fired on failed authentication attempts. Contains user ID (if known), IP address, user agent, and failure reason. Triggers lockout after 5 failures. |

---

## Schema Example

```json
// events/commerce/order-placed.avsc
{
  "type": "record",
  "name": "OrderPlaced",
  "namespace": "com.enterprise.commerce.events.v1",
  "doc": "Fired when a customer order transitions to PLACED state.",
  "fields": [
    { "name": "event_id",      "type": "string",  "doc": "UUID v4 â€” idempotency key" },
    { "name": "event_time",    "type": "long",    "logicalType": "timestamp-millis" },
    { "name": "schema_version","type": "string",  "default": "1.0.0" },
    { "name": "order_id",      "type": "string" },
    { "name": "customer_id",   "type": "string" },
    { "name": "total_amount",  "type": { "type": "record", "name": "Money",
        "fields": [
          { "name": "amount",   "type": "string", "doc": "Decimal string â€” no float" },
          { "name": "currency", "type": "string", "doc": "ISO 4217 â€” e.g. USD" }
        ]
      }
    },
    { "name": "line_items", "type": {
        "type": "array",
        "items": {
          "type": "record", "name": "LineItem",
          "fields": [
            { "name": "product_id", "type": "string" },
            { "name": "sku",        "type": "string" },
            { "name": "quantity",   "type": "int" },
            { "name": "unit_price", "type": "string" }
          ]
        }
      }
    },
    { "name": "shipping_address", "type": "string", "doc": "JSON-encoded Address" }
  ]
}
```

---

## Schema Registry

All schemas are registered with the Confluent Schema Registry (or its open-source
equivalent Apicurio Registry) before production use.

```bash
# Register a schema
curl -X POST http://schema-registry:8081/subjects/commerce.order.placed-value/versions \
  -H "Content-Type: application/vnd.schemaregistry.v1+json" \
  -d "{\"schema\": $(cat events/commerce/order-placed.avsc | jq -Rs .)}"

# List all subjects
curl http://schema-registry:8081/subjects

# Get latest version of a schema
curl http://schema-registry:8081/subjects/commerce.order.placed-value/versions/latest
```

---

## Compatibility Policy

All schemas use BACKWARD compatibility by default:

- New optional fields (with defaults) are allowed
- Removing fields requires a schema version bump and a 2-sprint deprecation window
- Renaming fields is a breaking change â€” use `aliases` instead
- `FULL_TRANSITIVE` compatibility is enforced for security domain topics

---

## References

- [Apache Avro Specification](https://avro.apache.org/docs/current/specification/)
- [Confluent Schema Registry](https://docs.confluent.io/platform/current/schema-registry/)
- [Apicurio Registry](https://www.apicur.io/registry/)
- [ShopOS Kafka Topic Naming](../README.md)
- [ShopOS Proto Definitions](../proto/README.md)
