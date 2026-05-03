# API Management — ShopOS

API gateways and management for external-facing APIs. The internal `api-gateway` service
under [`src/platform/api-gateway/`](../src/platform/api-gateway/) is the primary entry point;
the tools here are for partner and developer-platform-facing API surfaces.

## Layout

| Subdir | Tool | Role |
|---|---|---|
| [apisix/](apisix/) | Apache APISIX | Cloud-native API gateway with Lua plugin ecosystem; used for partner-bff edge plugins (rate-limit per partner key, request transformation, JWT validation) |
| [hasura/](hasura/) | Hasura | Instant GraphQL engine on top of Postgres — eliminates CRUD boilerplate for internal admin tooling |
| Tyk | Open-source API management with built-in developer portal — alternative deployment for the developer-platform external API products | (helm chart vendored) |

## Related

- Internal API gateway: [`src/platform/api-gateway/`](../src/platform/api-gateway/)
- Developer-platform services (sandbox, OAuth client, SDK gen, changelog): [`src/developer-platform/`](../src/developer-platform/)
- OpenAPI specs: [`openapi/`](../openapi/)
- API testing (Hurl + Spectral): [`api-testing/`](../api-testing/)
