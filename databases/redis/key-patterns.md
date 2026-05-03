# Redis Key Patterns — ShopOS

All keys follow: `{service}:{entity}:{id}[:{field}]`

TTLs are enforced via `EXPIRE` or set at key creation. No key is stored without a TTL unless it's a persistent data structure.

---

## Platform Domain

| Key Pattern | Type | TTL | Service | Description |
|---|---|---|---|---|
| `session:{session_id}` | Hash | 24h | session-service | JWT session data (user_id, roles, device_id) |
| `ratelimit:{service}:{ip}:{window}` | String (counter) | 60s | rate-limiter-service | Request count per IP per window |
| `ratelimit:{api_key}:{window}` | String (counter) | 60s | rate-limiter-service | API key rate limiting |
| `cache:config:{key}` | String | 5m | config-service | Feature flag / config values |
| `feature:{flag}:{user_id}` | String | 5m | feature-flag-service | Per-user flag override |
| `idempotency:{service}:{key}` | String | 24h | idempotency-service | Dedup key → response payload |
| `circuit:{service}:{endpoint}` | Hash | 30s | circuit-breaker-service | Circuit state (open/closed/half-open) |
| `correlationid:{id}` | String | 1h | correlation-id-service | Trace correlation mapping |

## Identity Domain

| Key Pattern | Type | TTL | Service | Description |
|---|---|---|---|---|
| `auth:refresh:{token_hash}` | String | 7d | auth-service | Refresh token → user_id lookup |
| `auth:blacklist:{jti}` | String | = token exp | auth-service | Revoked JTI blacklist |
| `mfa:otp:{user_id}:{type}` | String | 5m | mfa-service | OTP code (TOTP/SMS/email) |
| `mfa:attempts:{user_id}` | String (counter) | 15m | mfa-service | Failed MFA attempt count |
| `device:{device_id}` | Hash | 30d | device-fingerprint-service | Device fingerprint data |
| `botdetect:{ip}:{window}` | String (counter) | 1m | bot-detection-service | Bot detection signals per IP |

## Catalog Domain

| Key Pattern | Type | TTL | Service | Description |
|---|---|---|---|---|
| `product:{product_id}` | Hash | 10m | product-catalog-service | Product detail cache |
| `category:{category_id}:products` | ZSet | 5m | category-service | Product IDs sorted by position |
| `inventory:{product_id}:{warehouse_id}` | String | 30s | inventory-service | Current stock level |
| `stock:reserve:{product_id}:{order_id}` | String | 15m | stock-reservation-service | Temporary stock hold |
| `price:{product_id}:{price_list_id}` | Hash | 5m | pricing-service | Price with currency + tier |
| `search:suggest:{prefix}` | ZSet | 1h | search-service | Search autocomplete suggestions |

## Commerce Domain

| Key Pattern | Type | TTL | Service | Description |
|---|---|---|---|---|
| `cart:{user_id}` | Hash | 7d | cart-service | Cart items (product_id → qty+price JSON) |
| `cart:guest:{session_id}` | Hash | 24h | cart-service | Guest cart (pre-login) |
| `flash:sale:{sale_id}:stock` | String (counter) | = sale end | flash-sale-service | Remaining flash-sale units |
| `waitlist:{product_id}` | List | no TTL | waitlist-service | User IDs waiting for restock |
| `checkout:{checkout_id}` | Hash | 30m | checkout-service | Checkout session state |

## Customer Experience Domain

| Key Pattern | Type | TTL | Service | Description |
|---|---|---|---|---|
| `compare:{session_id}` | Set | 24h | compare-service | Product IDs in comparison |
| `recently_viewed:{user_id}` | ZSet | 30d | recently-viewed-service | Product IDs, score=timestamp |
| `price_alert:{product_id}:users` | Set | no TTL | price-alert-service | User IDs watching this product |
| `back_in_stock:{product_id}` | Set | no TTL | back-in-stock-service | User IDs to notify on restock |

## Communications Domain

| Key Pattern | Type | TTL | Service | Description |
|---|---|---|---|---|
| `notification:in_app:{user_id}` | List (max 100) | 30d | in-app-notification-service | Unread notifications |
| `notification:unread:{user_id}` | String (counter) | 30d | in-app-notification-service | Unread count |
| `notification:ws:{user_id}` | String | 5m | in-app-notification-service | WebSocket connection ID |

## Analytics/AI Domain

| Key Pattern | Type | TTL | Service | Description |
|---|---|---|---|---|
| `recommendations:{user_id}` | List | 1h | recommendation-service | Top-N recommended product IDs |
| `trending:products:{window}` | ZSet | 15m | analytics-service | Trending product IDs by view count |
| `ab:variant:{experiment_id}:{user_id}` | String | 30d | ab-testing-service | User's A/B variant assignment |

## Gamification Domain

| Key Pattern | Type | TTL | Service | Description |
|---|---|---|---|---|
| `points:{user_id}` | String | no TTL | points-service | Current point balance |
| `leaderboard:global` | ZSet | 1h | leaderboard-service | User IDs sorted by points |
| `leaderboard:weekly` | ZSet | 1h | leaderboard-service | Weekly leaderboard |
| `streak:{user_id}` | Hash | no TTL | streak-service | current_count, last_activity_date |

---

## Naming Conventions

1. Snake_case for field names, colon (`:`) as separator between key parts
2. UUIDs never abbreviated — always full UUID
3. Counters use `INCR`/`INCRBY`; never read-modify-write in application code
4. Sets for membership checks (O(1) SISMEMBER), ZSets for ranked/sorted data
5. Hashes for multi-field objects when you access fields independently
6. List for queues and bounded logs (use `LPUSH` + `LTRIM` to cap size)

## Memory Budgeting

| Domain | Estimated Footprint | Notes |
|---|---|---|
| Sessions | ~1GB | 1M active users × 1KB each |
| Carts | ~500MB | 500K active carts × 1KB |
| Inventory cache | ~100MB | 500K SKUs × 200B |
| Recommendations | ~200MB | 100K users × 2KB |
| Rate limiting | ~50MB | ephemeral, auto-expires |
| Total | ~2GB | Allocate 4GB for headroom |
