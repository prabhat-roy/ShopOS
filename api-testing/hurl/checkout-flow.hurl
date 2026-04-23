# ============================================================
# ShopOS — Checkout Flow Tests
# Covers: login → add to cart → view cart → checkout → order
# Run: hurl --test api-testing/hurl/checkout-flow.hurl
# ============================================================

# ── 1. Authenticate ───────────────────────────────────────────────────────────
POST http://localhost:8080/api/v1/auth/login
Content-Type: application/json
{
  "email": "testuser@shopos.dev",
  "password": "T3stP@ssw0rd!"
}
HTTP 200
[Captures]
access_token: jsonpath "$.accessToken"
user_id: jsonpath "$.userId"

# ── 2. Create a new cart (or retrieve existing) ───────────────────────────────
POST http://localhost:8080/api/v1/cart
Authorization: Bearer {{access_token}}
Content-Type: application/json
{
  "currency": "USD"
}
HTTP 200
[Asserts]
jsonpath "$.cartId" isString
jsonpath "$.userId" == {{user_id}}
jsonpath "$.items" isCollection
jsonpath "$.currency" == "USD"
[Captures]
cart_id: jsonpath "$.cartId"

# ── 3. Add first item to cart ─────────────────────────────────────────────────
POST http://localhost:8080/api/v1/cart/{{cart_id}}/items
Authorization: Bearer {{access_token}}
Content-Type: application/json
{
  "productId": "prod-001",
  "variantId": "var-001-blue-M",
  "quantity": 2,
  "attributes": {
    "color": "blue",
    "size": "M"
  }
}
HTTP 200
[Asserts]
jsonpath "$.cartId" == {{cart_id}}
jsonpath "$.items" isCollection
jsonpath "$.items[0].productId" == "prod-001"
jsonpath "$.items[0].quantity" == 2
jsonpath "$.totals.subtotal" isFloat
[Captures]
line_item_id: jsonpath "$.items[0].lineItemId"

# ── 4. Add second item to cart ────────────────────────────────────────────────
POST http://localhost:8080/api/v1/cart/{{cart_id}}/items
Authorization: Bearer {{access_token}}
Content-Type: application/json
{
  "productId": "prod-002",
  "variantId": "var-002-default",
  "quantity": 1
}
HTTP 200
[Asserts]
jsonpath "$.items" count >= 2

# ── 5. Get cart ───────────────────────────────────────────────────────────────
GET http://localhost:8080/api/v1/cart/{{cart_id}}
Authorization: Bearer {{access_token}}
HTTP 200
[Asserts]
jsonpath "$.cartId" == {{cart_id}}
jsonpath "$.userId" == {{user_id}}
jsonpath "$.items" count >= 2
jsonpath "$.totals.subtotal" isFloat
jsonpath "$.totals.tax" isFloat
jsonpath "$.totals.total" isFloat
jsonpath "$.currency" == "USD"

# ── 6. Update item quantity ───────────────────────────────────────────────────
PATCH http://localhost:8080/api/v1/cart/{{cart_id}}/items/{{line_item_id}}
Authorization: Bearer {{access_token}}
Content-Type: application/json
{
  "quantity": 3
}
HTTP 200
[Asserts]
jsonpath "$.items[0].quantity" == 3

# ── 7. Apply coupon code ──────────────────────────────────────────────────────
POST http://localhost:8080/api/v1/cart/{{cart_id}}/coupons
Authorization: Bearer {{access_token}}
Content-Type: application/json
{
  "couponCode": "WELCOME10"
}
HTTP 200
[Asserts]
jsonpath "$.appliedCoupons" isCollection
jsonpath "$.totals.discount" isFloat

# ── 8. Get cart summary (totals only) ────────────────────────────────────────
GET http://localhost:8080/api/v1/cart/{{cart_id}}/totals
Authorization: Bearer {{access_token}}
HTTP 200
[Asserts]
jsonpath "$.subtotal" isFloat
jsonpath "$.discount" isFloat
jsonpath "$.tax" isFloat
jsonpath "$.shipping" isFloat
jsonpath "$.total" isFloat

# ── 9. Initiate checkout ──────────────────────────────────────────────────────
POST http://localhost:8080/api/v1/checkout
Authorization: Bearer {{access_token}}
Content-Type: application/json
{
  "cartId": "{{cart_id}}",
  "shippingAddress": {
    "firstName": "Test",
    "lastName": "User",
    "addressLine1": "123 Commerce St",
    "city": "San Francisco",
    "state": "CA",
    "postalCode": "94102",
    "country": "US",
    "phone": "+14155551234"
  },
  "billingAddress": {
    "sameAsShipping": true
  },
  "shippingMethod": "standard",
  "paymentMethod": {
    "type": "card",
    "token": "tok_visa_test"
  },
  "customerEmail": "testuser@shopos.dev"
}
HTTP 201
[Asserts]
jsonpath "$.checkoutId" isString
jsonpath "$.status" == "pending_payment"
jsonpath "$.cartId" == {{cart_id}}
jsonpath "$.totals.total" isFloat
[Captures]
checkout_id: jsonpath "$.checkoutId"
order_total: jsonpath "$.totals.total"

# ── 10. Get checkout session ──────────────────────────────────────────────────
GET http://localhost:8080/api/v1/checkout/{{checkout_id}}
Authorization: Bearer {{access_token}}
HTTP 200
[Asserts]
jsonpath "$.checkoutId" == {{checkout_id}}
jsonpath "$.status" isString
jsonpath "$.shippingAddress.country" == "US"

# ── 11. Confirm payment and place order ───────────────────────────────────────
POST http://localhost:8080/api/v1/checkout/{{checkout_id}}/confirm
Authorization: Bearer {{access_token}}
Content-Type: application/json
{
  "paymentConfirmation": {
    "paymentIntentId": "pi_test_1234567890",
    "status": "succeeded"
  }
}
HTTP 201
[Asserts]
jsonpath "$.orderId" isString
jsonpath "$.status" == "confirmed"
jsonpath "$.checkoutId" == {{checkout_id}}
jsonpath "$.totalAmount" isFloat
[Captures]
order_id: jsonpath "$.orderId"

# ── 12. Get placed order ──────────────────────────────────────────────────────
GET http://localhost:8080/api/v1/orders/{{order_id}}
Authorization: Bearer {{access_token}}
HTTP 200
[Asserts]
jsonpath "$.orderId" == {{order_id}}
jsonpath "$.status" == "confirmed"
jsonpath "$.userId" == {{user_id}}
jsonpath "$.items" isCollection
jsonpath "$.shippingAddress" isCollection

# ── 13. Cart should be empty/cleared after order ─────────────────────────────
GET http://localhost:8080/api/v1/cart/{{cart_id}}
Authorization: Bearer {{access_token}}
HTTP 200
[Asserts]
jsonpath "$.items" count == 0

# ── 14. Remove item — cannot remove from fulfilled cart (idempotent 404/200) ──
DELETE http://localhost:8080/api/v1/cart/{{cart_id}}/items/{{line_item_id}}
Authorization: Bearer {{access_token}}
HTTP *
