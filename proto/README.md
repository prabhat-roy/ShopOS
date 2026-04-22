# Protocol Buffers — ShopOS

All gRPC service contracts for ShopOS are defined as `.proto` files in this directory.
Using a centralised proto repository ensures consistent types, versioning, and a single
source of truth for inter-service contracts across all 8 programming languages.

---

## Directory Structure

```
proto/
├── common/
│   ├── money.proto                 ← Money, Currency types
│   ├── address.proto               ← Address, GeoPoint types
│   ├── pagination.proto            ← PageRequest, PageResponse
│   ├── timestamp.proto             ← Consistent timestamp wrappers
│   └── error.proto                 ← Standard error details
│
├── platform/
│   ├── gateway.proto               ← APIGateway service
│   ├── config.proto                ← ConfigService
│   ├── feature_flag.proto          ← FeatureFlagService
│   ├── audit.proto                 ← AuditService
│   ├── scheduler.proto             ← SchedulerService
│   ├── webhook.proto               ← WebhookService
│   └── tenant.proto                ← TenantService
│
├── identity/
│   ├── auth.proto                  ← AuthService (login, token, refresh)
│   ├── user.proto                  ← UserService (CRUD, profile)
│   ├── session.proto               ← SessionService
│   ├── permission.proto            ← PermissionService
│   ├── mfa.proto                   ← MFAService (TOTP, SMS)
│   ├── api_key.proto               ← APIKeyService
│   └── device_fingerprint.proto    ← DeviceFingerprintService
│
├── catalog/
│   ├── product.proto               ← ProductCatalogService
│   ├── category.proto              ← CategoryService
│   ├── brand.proto                 ← BrandService
│   ├── pricing.proto               ← PricingService
│   ├── inventory.proto             ← InventoryService
│   ├── bundle.proto                ← BundleService
│   ├── search.proto                ← SearchService
│   └── seo.proto                   ← SEOService
│
├── commerce/
│   ├── cart.proto                  ← CartService
│   ├── checkout.proto              ← CheckoutService
│   ├── order.proto                 ← OrderService
│   ├── payment.proto               ← PaymentService
│   ├── shipping.proto              ← ShippingService
│   ├── tax.proto                   ← TaxService
│   ├── promotions.proto            ← PromotionsService
│   ├── loyalty.proto               ← LoyaltyService
│   └── wallet.proto                ← WalletService
│
├── supply-chain/
│   ├── vendor.proto                ← VendorService
│   ├── warehouse.proto             ← WarehouseService
│   ├── fulfillment.proto           ← FulfillmentService
│   ├── tracking.proto              ← TrackingService
│   └── carrier.proto               ← CarrierIntegrationService
│
├── financial/
│   ├── invoice.proto               ← InvoiceService
│   ├── payout.proto                ← PayoutService
│   ├── accounting.proto            ← AccountingService
│   └── credit.proto                ← CreditService
│
├── customer-experience/
│   ├── review.proto                ← ReviewRatingService
│   ├── wishlist.proto              ← WishlistService
│   ├── support.proto               ← SupportTicketService
│   └── survey.proto                ← SurveyService
│
├── communications/
│   ├── notification.proto          ← NotificationOrchestrator
│   └── template.proto              ← TemplateService
│
├── content/
│   ├── media.proto                 ← MediaAssetService
│   ├── document.proto              ← DocumentService
│   ├── cms.proto                   ← CMSService
│   └── i18n.proto                  ← I18nL10nService
│
├── analytics-ai/
│   ├── analytics.proto             ← AnalyticsService
│   ├── recommendation.proto        ← RecommendationService
│   ├── ad.proto                    ← AdService
│   └── personalization.proto       ← PersonalizationService
│
├── b2b/
│   ├── organization.proto          ← OrganizationService
│   ├── contract.proto              ← ContractService
│   └── quote.proto                 ← QuoteRFQService
│
├── integrations/
│   ├── erp.proto                   ← ERPIntegrationService
│   ├── crm.proto                   ← CRMIntegrationService
│   └── marketplace.proto           ← MarketplaceConnectorService
│
├── marketplace/
│   └── marketplace.proto           ← SellerRegistration, ListingApproval, Commission, Dispute
│
├── gamification/
│   └── gamification.proto          ← Points, Badge, Leaderboard, Challenge
│
├── developer-platform/
│   └── developer-platform.proto    ← OAuthClient, WebhookManagement
│
├── compliance/
│   └── compliance.proto            ← PrivacyRequest, ConsentAudit, DataLineage
│
└── sustainability/
    └── sustainability.proto        ← CarbonTracker, EcoScore, Offset
```

**Total: 63 `.proto` files across 18 directories.**

---

## Proto Conventions

### Package Naming

```protobuf
// Format: enterprise.{domain}.v1
syntax = "proto3";
package enterprise.commerce.v1;

option go_package = "github.com/shopos/enterprise-platform/gen/go/commerce/v1;commercev1";
option java_package = "com.enterprise.commerce.v1";
option java_multiple_files = true;
```

### Service Naming

- Service name: `{Entity}Service` — e.g., `OrderService`, `PaymentService`
- RPC methods: `PascalCase` verbs — `CreateOrder`, `GetOrder`, `ListOrders`, `CancelOrder`
- Request/Response messages: `{Method}Request` / `{Method}Response`

### Field Conventions

- Use `google.protobuf.Timestamp` for all timestamps — never string-encoded dates
- Use `common.Money` for all monetary amounts — never raw floats
- Use `common.Address` for all postal addresses
- Field numbers 1–15 are reserved for the most frequently used fields (1-byte encoding)
- Reserved field numbers are documented with `reserved` keyword when removed

### Versioning

- Breaking changes require a new package version: `enterprise.commerce.v2`
- Non-breaking additions (new fields, new RPCs) are made in-place
- Deprecated fields use the `[deprecated = true]` option and are retained for 2 minor versions

---

## Generating Code

ShopOS uses [Buf](https://buf.build/) for proto linting, breaking-change detection, and
code generation.

### Setup

```bash
# Install Buf CLI
brew install bufbuild/buf/buf

# Or with Go
go install github.com/bufbuild/buf/cmd/buf@latest
```

### Buf Configuration (`buf.yaml`)

```yaml
# proto/buf.yaml
version: v2
modules:
  - path: .
lint:
  use:
    - DEFAULT
breaking:
  use:
    - FILE
```

### Code Generation (`buf.gen.yaml`)

```yaml
# proto/buf.gen.yaml
version: v2
plugins:
  # Go
  - plugin: buf.build/protocolbuffers/go
    out: ../gen/go
    opt: paths=source_relative
  - plugin: buf.build/grpc/go
    out: ../gen/go
    opt: paths=source_relative

  # Java
  - plugin: buf.build/protocolbuffers/java
    out: ../gen/java

  # Python
  - plugin: buf.build/protocolbuffers/python
    out: ../gen/python
  - plugin: buf.build/grpc/python
    out: ../gen/python

  # Node.js / TypeScript
  - plugin: buf.build/grpc/node
    out: ../gen/node
  - plugin: buf.build/bufbuild/ts
    out: ../gen/node

  # C# (.NET)
  - plugin: buf.build/protocolbuffers/csharp
    out: ../gen/csharp

  # Kotlin
  - plugin: buf.build/grpc/kotlin
    out: ../gen/kotlin

  # Rust
  - plugin: buf.build/community/neoeinstein-prost
    out: ../gen/rust
  - plugin: buf.build/community/neoeinstein-tonic
    out: ../gen/rust
```

### Running Code Generation

```bash
# Lint all proto files
buf lint proto/

# Check for breaking changes against main branch
buf breaking proto/ --against '.git#branch=main'

# Generate code for all languages
buf generate proto/

# Generate for a single domain
buf generate proto/commerce/

# Push to Buf Schema Registry (BSR)
buf push proto/ --tag v1.5.0
```

### Generated Code Location

```
gen/
├── go/
│   ├── common/v1/
│   ├── platform/v1/
│   ├── commerce/v1/
│   └── ...
├── java/com/enterprise/
├── python/enterprise/
├── node/enterprise/
├── csharp/Enterprise/
├── kotlin/com/enterprise/
└── rust/enterprise/
```

Each service imports generated code from `gen/{language}/`. Generated files are checked into
the repository so services can consume them without running `buf generate` locally.

---

## Health Check Protocol

Every gRPC service implements the standard `grpc.health.v1.Health` protocol:

```protobuf
// Imported from grpc/health/v1/health.proto (standard)
service Health {
  rpc Check(HealthCheckRequest) returns (HealthCheckResponse);
  rpc Watch(HealthCheckRequest) returns (stream HealthCheckResponse);
}
```

This is used by Kubernetes readiness probes (via `grpc_health_probe`) and by ArgoCD for
deployment health checks.

---

## References

- [Protocol Buffers Language Guide](https://protobuf.dev/programming-guides/proto3/)
- [Buf CLI Documentation](https://buf.build/docs/)
- [gRPC Health Checking Protocol](https://grpc.io/docs/guides/health-checking/)
- [Google APIs Design Guide](https://cloud.google.com/apis/design)
- [ShopOS gRPC Port Ranges](../README.md)
