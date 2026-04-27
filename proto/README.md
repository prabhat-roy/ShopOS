# Protocol Buffers â€” ShopOS

All gRPC service contracts for ShopOS are defined as `.proto` files in this directory.
Using a centralised proto repository ensures consistent types, versioning, and a single
source of truth for inter-service contracts across all 8 programming languages.

---

## Directory Structure

```
proto/
â”œâ”€â”€ common/
â”‚   â”œâ”€â”€ money.proto                 â† Money, Currency types
â”‚   â”œâ”€â”€ address.proto               â† Address, GeoPoint types
â”‚   â”œâ”€â”€ pagination.proto            â† PageRequest, PageResponse
â”‚   â”œâ”€â”€ timestamp.proto             â† Consistent timestamp wrappers
â”‚   â””â”€â”€ error.proto                 â† Standard error details
â”‚
â”œâ”€â”€ platform/
â”‚   â”œâ”€â”€ gateway.proto               â† APIGateway service
â”‚   â”œâ”€â”€ config.proto                â† ConfigService
â”‚   â”œâ”€â”€ feature_flag.proto          â† FeatureFlagService
â”‚   â”œâ”€â”€ audit.proto                 â† AuditService
â”‚   â”œâ”€â”€ scheduler.proto             â† SchedulerService
â”‚   â”œâ”€â”€ webhook.proto               â† WebhookService
â”‚   â””â”€â”€ tenant.proto                â† TenantService
â”‚
â”œâ”€â”€ identity/
â”‚   â”œâ”€â”€ auth.proto                  â† AuthService (login, token, refresh)
â”‚   â”œâ”€â”€ user.proto                  â† UserService (CRUD, profile)
â”‚   â”œâ”€â”€ session.proto               â† SessionService
â”‚   â”œâ”€â”€ permission.proto            â† PermissionService
â”‚   â”œâ”€â”€ mfa.proto                   â† MFAService (TOTP, SMS)
â”‚   â”œâ”€â”€ api_key.proto               â† APIKeyService
â”‚   â””â”€â”€ device_fingerprint.proto    â† DeviceFingerprintService
â”‚
â”œâ”€â”€ catalog/
â”‚   â”œâ”€â”€ product.proto               â† ProductCatalogService
â”‚   â”œâ”€â”€ category.proto              â† CategoryService
â”‚   â”œâ”€â”€ brand.proto                 â† BrandService
â”‚   â”œâ”€â”€ pricing.proto               â† PricingService
â”‚   â”œâ”€â”€ inventory.proto             â† InventoryService
â”‚   â”œâ”€â”€ bundle.proto                â† BundleService
â”‚   â”œâ”€â”€ search.proto                â† SearchService
â”‚   â””â”€â”€ seo.proto                   â† SEOService
â”‚
â”œâ”€â”€ commerce/
â”‚   â”œâ”€â”€ cart.proto                  â† CartService
â”‚   â”œâ”€â”€ checkout.proto              â† CheckoutService
â”‚   â”œâ”€â”€ order.proto                 â† OrderService
â”‚   â”œâ”€â”€ payment.proto               â† PaymentService
â”‚   â”œâ”€â”€ shipping.proto              â† ShippingService
â”‚   â”œâ”€â”€ tax.proto                   â† TaxService
â”‚   â”œâ”€â”€ promotions.proto            â† PromotionsService
â”‚   â”œâ”€â”€ loyalty.proto               â† LoyaltyService
â”‚   â””â”€â”€ wallet.proto                â† WalletService
â”‚
â”œâ”€â”€ supply-chain/
â”‚   â”œâ”€â”€ vendor.proto                â† VendorService
â”‚   â”œâ”€â”€ warehouse.proto             â† WarehouseService
â”‚   â”œâ”€â”€ fulfillment.proto           â† FulfillmentService
â”‚   â”œâ”€â”€ tracking.proto              â† TrackingService
â”‚   â””â”€â”€ carrier.proto               â† CarrierIntegrationService
â”‚
â”œâ”€â”€ financial/
â”‚   â”œâ”€â”€ invoice.proto               â† InvoiceService
â”‚   â”œâ”€â”€ payout.proto                â† PayoutService
â”‚   â”œâ”€â”€ accounting.proto            â† AccountingService
â”‚   â””â”€â”€ credit.proto                â† CreditService
â”‚
â”œâ”€â”€ customer-experience/
â”‚   â”œâ”€â”€ review.proto                â† ReviewRatingService
â”‚   â”œâ”€â”€ wishlist.proto              â† WishlistService
â”‚   â”œâ”€â”€ support.proto               â† SupportTicketService
â”‚   â””â”€â”€ survey.proto                â† SurveyService
â”‚
â”œâ”€â”€ communications/
â”‚   â”œâ”€â”€ notification.proto          â† NotificationOrchestrator
â”‚   â””â”€â”€ template.proto              â† TemplateService
â”‚
â”œâ”€â”€ content/
â”‚   â”œâ”€â”€ media.proto                 â† MediaAssetService
â”‚   â”œâ”€â”€ document.proto              â† DocumentService
â”‚   â”œâ”€â”€ cms.proto                   â† CMSService
â”‚   â””â”€â”€ i18n.proto                  â† I18nL10nService
â”‚
â”œâ”€â”€ analytics-ai/
â”‚   â”œâ”€â”€ analytics.proto             â† AnalyticsService
â”‚   â”œâ”€â”€ recommendation.proto        â† RecommendationService
â”‚   â”œâ”€â”€ ad.proto                    â† AdService
â”‚   â””â”€â”€ personalization.proto       â† PersonalizationService
â”‚
â”œâ”€â”€ b2b/
â”‚   â”œâ”€â”€ organization.proto          â† OrganizationService
â”‚   â”œâ”€â”€ contract.proto              â† ContractService
â”‚   â””â”€â”€ quote.proto                 â† QuoteRFQService
â”‚
â”œâ”€â”€ integrations/
â”‚   â”œâ”€â”€ erp.proto                   â† ERPIntegrationService
â”‚   â”œâ”€â”€ crm.proto                   â† CRMIntegrationService
â”‚   â””â”€â”€ marketplace.proto           â† MarketplaceConnectorService
â”‚
â”œâ”€â”€ marketplace/
â”‚   â””â”€â”€ marketplace.proto           â† SellerRegistration, ListingApproval, Commission, Dispute
â”‚
â”œâ”€â”€ gamification/
â”‚   â””â”€â”€ gamification.proto          â† Points, Badge, Leaderboard, Challenge
â”‚
â”œâ”€â”€ developer-platform/
â”‚   â””â”€â”€ developer-platform.proto    â† OAuthClient, WebhookManagement
â”‚
â”œâ”€â”€ compliance/
â”‚   â””â”€â”€ compliance.proto            â† PrivacyRequest, ConsentAudit, DataLineage
â”‚
â””â”€â”€ sustainability/
    â””â”€â”€ sustainability.proto        â† CarbonTracker, EcoScore, Offset
```

Total: 63 `.proto` files across 18 directories.

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

- Service name: `{Entity}Service` â€” e.g., `OrderService`, `PaymentService`
- RPC methods: `PascalCase` verbs â€” `CreateOrder`, `GetOrder`, `ListOrders`, `CancelOrder`
- Request/Response messages: `{Method}Request` / `{Method}Response`

### Field Conventions

- Use `google.protobuf.Timestamp` for all timestamps â€” never string-encoded dates
- Use `common.Money` for all monetary amounts â€” never raw floats
- Use `common.Address` for all postal addresses
- Field numbers 1â€“15 are reserved for the most frequently used fields (1-byte encoding)
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
â”œâ”€â”€ go/
â”‚   â”œâ”€â”€ common/v1/
â”‚   â”œâ”€â”€ platform/v1/
â”‚   â”œâ”€â”€ commerce/v1/
â”‚   â””â”€â”€ ...
â”œâ”€â”€ java/com/enterprise/
â”œâ”€â”€ python/enterprise/
â”œâ”€â”€ node/enterprise/
â”œâ”€â”€ csharp/Enterprise/
â”œâ”€â”€ kotlin/com/enterprise/
â””â”€â”€ rust/enterprise/
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
