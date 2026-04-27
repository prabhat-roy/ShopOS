# ShopOS â€” Pact Contract Tests

[Pact](https://pact.io/) is a contract testing framework that ensures consumer and provider
services agree on their API contract without requiring both to be live simultaneously.

## Contracts in this directory

| File | Consumer | Provider | Interactions |
|------|----------|----------|-------------|
| `consumer/order-service-cart-service.pact.json` | order-service | cart-service | GET /cart/{userId}, DELETE /cart/{userId}, GET /cart/{userId}/totals |
| `consumer/checkout-service-payment-service.pact.json` | checkout-service | payment-service | POST /payments, GET /payments/{id}, POST /payments/{id}/refunds |

All contracts use Pact Specification v3.

## Tools

| Language | Library | Version |
|---|---|---|
| Go | [pact-go](https://github.com/pact-foundation/pact-go) | v2.x |
| Java/Kotlin | [pact-jvm](https://github.com/pact-foundation/pact-jvm) | 4.6.x |
| Node.js | [@pact-foundation/pact](https://github.com/pact-foundation/pact-js) | 12.x |
| Python | [pact-python](https://github.com/pact-foundation/pact-python) | 2.x |

## Running consumer tests (Go example)

```bash
cd src/commerce/order-service
go test ./internal/pact/... -v -run TestPact
```

## Running provider verification

Provider verification is run against the live (or containerised) provider service:

```bash
# Go provider verification
cd src/commerce/cart-service
go test ./internal/pact/... -v -run TestCartServiceProviderVerification \
  -pact-broker-url=http://pact-broker:9292 \
  -pact-provider-version=$(git rev-parse --short HEAD)
```

## Publishing contracts to Pact Broker

After running consumer tests, publish the generated Pact files to the Pact Broker:

```bash
# Using the pact-broker CLI
pact-broker publish ./testing/pact/consumer \
  --broker-base-url=http://pact-broker:9292 \
  --consumer-app-version=$(git rev-parse --short HEAD) \
  --branch=$(git rev-parse --abbrev-ref HEAD)
```

Or with Docker:

```bash
docker run --rm \
  -v "$(pwd)/testing/pact/consumer:/pacts" \
  pactfoundation/pact-cli:latest \
  broker publish /pacts \
    --broker-base-url=http://pact-broker:9292 \
    --consumer-app-version=$(git rev-parse --short HEAD) \
    --branch=$(git rev-parse --abbrev-ref HEAD)
```

## Can I Deploy?

Before deploying a service, check whether it is compatible with all deployed providers/consumers:

```bash
pact-broker can-i-deploy \
  --pacticipant order-service \
  --version=$(git rev-parse --short HEAD) \
  --to-environment=production \
  --broker-base-url=http://pact-broker:9292
```

## Pact Broker

The ShopOS Pact Broker runs at `http://pact-broker:9292` (internal) or via Traefik at
`https://pact.internal.shopos.dev`.

It is deployed as part of `docker-compose.yml` under the `testing` profile:

```bash
docker-compose --profile testing up pact-broker pact-broker-db
```

## Adding new contracts

1. Write the consumer test using your language's Pact library.
2. Run the consumer test â€” Pact generates a `.pact.json` file in the output directory.
3. Copy/symlink the generated file to `testing/pact/consumer/`.
4. Publish to the Pact Broker using the command above.
5. Add provider verification test to the provider service.

## Contract naming convention

```
{consumer-service-name}-{provider-service-name}.pact.json
```

Examples:
- `order-service-cart-service.pact.json`
- `checkout-service-payment-service.pact.json`
- `notification-orchestrator-template-service.pact.json`
