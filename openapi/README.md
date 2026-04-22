# OpenAPI Specifications

OpenAPI 3.1 specs for ShopOS public and internal APIs.

| File | Service | Port | Audience |
|---|---|---|---|
| `api-gateway.yaml` | api-gateway | 8080 | Public (storefront, mobile) |
| `admin-api.yaml` | admin-portal | 8085 | Internal (admin-dashboard) |
| `developer-platform-api.yaml` | api-management-service | 8206 | Developers (partner integrations) |

## Viewing

```bash
# Using Swagger UI via Docker
docker run -p 8090:8080 -e SWAGGER_JSON=/specs/api-gateway.yaml \
  -v $(pwd)/openapi:/specs swaggerapi/swagger-ui

# Using Redoc
npx @redocly/cli preview-docs openapi/api-gateway.yaml

# Validate
npx @redocly/cli lint openapi/*.yaml
```

## Generating clients

```bash
# TypeScript client for storefront
npx openapi-typescript openapi/api-gateway.yaml -o src/web/storefront/src/lib/api-types.ts

# Go client for internal services
oapi-codegen -package api -generate types,client openapi/api-gateway.yaml > internal/api/client.gen.go
```
