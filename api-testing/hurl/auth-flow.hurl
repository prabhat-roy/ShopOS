# ============================================================
# ShopOS — Auth Service Flow Tests
# Covers: register → login → refresh → profile → logout
# Run: hurl --test api-testing/hurl/auth-flow.hurl
# ============================================================

# ── 1. Register new user ──────────────────────────────────────────────────────
POST http://localhost:8080/api/v1/auth/register
Content-Type: application/json
{
  "email": "testuser@shopos.dev",
  "password": "T3stP@ssw0rd!",
  "firstName": "Test",
  "lastName": "User",
  "acceptedTerms": true
}
HTTP 201
[Asserts]
header "Content-Type" contains "application/json"
jsonpath "$.userId" isString
jsonpath "$.email" == "testuser@shopos.dev"
jsonpath "$.status" == "active"
[Captures]
registered_user_id: jsonpath "$.userId"

# ── 2. Login ──────────────────────────────────────────────────────────────────
POST http://localhost:8080/api/v1/auth/login
Content-Type: application/json
{
  "email": "testuser@shopos.dev",
  "password": "T3stP@ssw0rd!"
}
HTTP 200
[Asserts]
header "Content-Type" contains "application/json"
jsonpath "$.accessToken" isString
jsonpath "$.refreshToken" isString
jsonpath "$.tokenType" == "Bearer"
jsonpath "$.expiresIn" isInteger
jsonpath "$.userId" == {{registered_user_id}}
[Captures]
access_token: jsonpath "$.accessToken"
refresh_token: jsonpath "$.refreshToken"

# ── 3. Get current user profile (authenticated) ───────────────────────────────
GET http://localhost:8080/api/v1/users/me
Authorization: Bearer {{access_token}}
HTTP 200
[Asserts]
jsonpath "$.userId" == {{registered_user_id}}
jsonpath "$.email" == "testuser@shopos.dev"
jsonpath "$.firstName" == "Test"
jsonpath "$.lastName" == "User"

# ── 4. Refresh token ──────────────────────────────────────────────────────────
POST http://localhost:8080/api/v1/auth/refresh
Content-Type: application/json
{
  "refreshToken": "{{refresh_token}}"
}
HTTP 200
[Asserts]
jsonpath "$.accessToken" isString
jsonpath "$.tokenType" == "Bearer"
[Captures]
new_access_token: jsonpath "$.accessToken"

# ── 5. Access protected resource with new token ───────────────────────────────
GET http://localhost:8080/api/v1/users/me
Authorization: Bearer {{new_access_token}}
HTTP 200
[Asserts]
jsonpath "$.userId" == {{registered_user_id}}

# ── 6. Reject request with invalid token ──────────────────────────────────────
GET http://localhost:8080/api/v1/users/me
Authorization: Bearer invalid-token-12345
HTTP 401
[Asserts]
jsonpath "$.error" isString

# ── 7. Reject login with wrong password ───────────────────────────────────────
POST http://localhost:8080/api/v1/auth/login
Content-Type: application/json
{
  "email": "testuser@shopos.dev",
  "password": "WrongPassword!"
}
HTTP 401
[Asserts]
jsonpath "$.error" isString

# ── 8. Logout ─────────────────────────────────────────────────────────────────
POST http://localhost:8080/api/v1/auth/logout
Authorization: Bearer {{new_access_token}}
Content-Type: application/json
{
  "refreshToken": "{{refresh_token}}"
}
HTTP 204

# ── 9. Confirm token is invalidated after logout ──────────────────────────────
GET http://localhost:8080/api/v1/users/me
Authorization: Bearer {{new_access_token}}
HTTP 401
