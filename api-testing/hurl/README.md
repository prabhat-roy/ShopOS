# ShopOS — Hurl API Tests

[Hurl](https://hurl.dev/) is a CLI tool that runs API tests defined in plain `.hurl` files.

## Prerequisites

Install Hurl:
```bash
# macOS
brew install hurl

# Linux
curl -LO https://github.com/Orange-OpenSource/hurl/releases/latest/download/hurl-x86_64-unknown-linux-gnu.tar.gz
tar -xzf hurl-*.tar.gz && sudo mv hurl /usr/local/bin/

# Windows (Scoop)
scoop install hurl
```

## Running the tests

```bash
# Run all test files
hurl --test api-testing/hurl/*.hurl

# Run a specific file
hurl --test api-testing/hurl/health-checks.hurl

# Run with verbose output
hurl --test --verbose api-testing/hurl/health-checks.hurl

# Run against a different host
hurl --test --variable base_url=https://staging.shopos.dev api-testing/hurl/*.hurl

# Run and output JUnit XML (for CI)
hurl --test --report-junit results.xml api-testing/hurl/*.hurl

# Run and output HTML report
hurl --test --report-html hurl-report/ api-testing/hurl/*.hurl

# Run with a specific number of retries
hurl --test --retry 3 --retry-interval 500 api-testing/hurl/health-checks.hurl
```

## Test files

| File | Description |
|---|---|
| `health-checks.hurl` | GET /healthz on all BFF/gateway services (ports 8080–8091) |
| `auth-flow.hurl` | Register → login → refresh token → profile → logout |
| `catalog-flow.hurl` | List products → get by ID → search → categories |
| `checkout-flow.hurl` | Login → cart → add items → apply coupon → checkout → order |

## CI integration

In Jenkins / GitHub Actions:

```bash
hurl --test \
  --report-junit api-testing/hurl/results.xml \
  --report-html api-testing/hurl/report/ \
  --variable base_url=http://api-gateway:8080 \
  api-testing/hurl/*.hurl
```

The JUnit XML is compatible with most CI dashboards (Jenkins Test Results, GitHub Actions summary).

## Environment variables

Override the default `localhost` target with Hurl variables:

```bash
# Run against staging
hurl --test \
  --variable api_host=https://api-staging.shopos.dev \
  api-testing/hurl/health-checks.hurl
```

## Notes

- Tests run sequentially within a file; captures from earlier entries are available in later ones.
- The auth-flow test creates a real user in the database — ensure the test environment supports cleanup or use a dedicated test tenant.
- `checkout-flow.hurl` depends on seed data (`prod-001`, `prod-002`) and a valid coupon `WELCOME10` existing in the environment.
