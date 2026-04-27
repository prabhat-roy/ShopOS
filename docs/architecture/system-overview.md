# System Overview â€” ShopOS

ShopOS is an enterprise-grade, cloud-native commerce platform comprising 224 microservices + 6 frontend apps (230 total) across 19 domains, written in 13 programming languages. Services communicate via gRPC (synchronous reads/commands) and Apache Kafka (asynchronous domain events). Every service owns its own dedicated database â€” no shared data stores.

---

## Full System Architecture

```
â•”â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•—
â•‘                              EXTERNAL CLIENTS                                        â•‘
â•‘                                                                                      â•‘
â•‘   â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”  â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”  â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”  â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â” â•‘
â•‘   â”‚  Web Browser  â”‚  â”‚  Mobile App  â”‚  â”‚  Partner API  â”‚  â”‚  Admin / Back-Office   â”‚ â•‘
â•‘   â””â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”˜  â””â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”˜  â””â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”˜  â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜ â•‘
â•šâ•â•â•â•â•â•â•â•â•â•â•ªâ•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•ªâ•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•ªâ•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•ªâ•â•â•â•â•â•â•â•â•â•â•â•â•
           â”‚                  â”‚                  â”‚                        â”‚
           â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”´â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”´â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜
                                       â”‚ HTTPS / WSS
                          â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â–¼â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”
                          â”‚     TRAEFIK EDGE ROUTER   â”‚  TLS termination (cert-manager)
                          â”‚     + Coraza WAF           â”‚  OWASP Core Rule Set â€” SQLi, XSS,
                          â”‚     + rate-limiter-service â”‚  path traversal blocked at edge
                          â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜
                                       â”‚
                 â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â–¼â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”
                 â”‚              API GATEWAY  [Go :8080]         â”‚
                 â”‚  JWT validation Â· request routing Â· tracing  â”‚
                 â”‚  GraphQL Gateway [Go :8086] â€” unified query  â”‚
                 â””â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”€â”˜
                        â”‚              â”‚              â”‚
              â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â–¼â”€â”€â”  â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â–¼â”€â”€â”€â”  â”Œâ”€â”€â”€â”€â”€â”€â–¼â”€â”€â”€â”€â”€â”€â”
              â”‚  Web BFF    â”‚  â”‚ Mobile BFF  â”‚  â”‚ Partner BFF  â”‚
              â”‚  [Go :8081] â”‚  â”‚[Node :8082] â”‚  â”‚ [Go :8083]  â”‚
              â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”˜  â””â”€â”€â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”˜  â””â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”˜
                        â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¼â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜
                                       â”‚
              â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¼â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€
                    gRPC / Protobuf     â”‚    (Istio mTLS, SPIFFE SVID certs)
              â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¼â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€
                                       â”‚
       â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¼â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”
       â”‚                               â”‚                                      â”‚
       â–¼                               â–¼                                      â–¼
â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”               â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”                       â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”
â”‚  IDENTITY   â”‚               â”‚   CATALOG    â”‚                       â”‚   COMMERCE   â”‚
â”‚  8 services â”‚               â”‚  12 services â”‚                       â”‚  23 services â”‚
â”‚             â”‚               â”‚              â”‚                       â”‚              â”‚
â”‚ auth        â”‚               â”‚ product      â”‚â—„â”€â”€CDC (Debezium)â”€â”€â”€â”€â”€â”€â”‚ cart         â”‚
â”‚ user        â”‚               â”‚ category     â”‚   MongoDBâ†’Kafka       â”‚ checkout     â”‚
â”‚ session     â”‚               â”‚ brand        â”‚      â†“                â”‚ order        â”‚
â”‚ permission  â”‚               â”‚ pricing      â”‚  search-service       â”‚ payment      â”‚
â”‚ mfa         â”‚               â”‚ inventory    â”‚  (Elasticsearch)      â”‚ fraud-detect â”‚
â”‚ gdpr        â”‚               â”‚ search       â”‚                       â”‚ promotions   â”‚
â”‚ api-key     â”‚               â”‚ seo          â”‚                       â”‚ flash-sale   â”‚
â”‚ device-fp   â”‚               â”‚ ...          â”‚                       â”‚ bnpl, ...    â”‚
â””â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”˜               â””â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”˜                       â””â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”˜
       â”‚                             â”‚                                       â”‚
       â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¬â”˜                                      â”‚
                                    â”‚                                        â”‚
              â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•ªâ•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•ªâ•â•â•â•â•â•â•
                                    â”‚   KAFKA EVENT BUS (Confluent 7.7.1)   â”‚
              â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•ªâ•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•ªâ•â•â•â•â•â•â•
                                    â”‚   + Schema Registry (Avro / backward   â”‚
                                    â”‚     compatible schemas in events/)     â”‚
              â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¼â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜
              â”‚                     â”‚
    â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”´â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”   â”Œâ”€â”€â”€â”€â”€â”€â”´â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”
    â”‚                    â”‚   â”‚                                                     â”‚
    â–¼                    â–¼   â–¼                                                     â–¼
â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”    â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”    â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”    â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”
â”‚  SUPPLY  â”‚    â”‚  FINANCIAL   â”‚    â”‚ COMMUNICATIONS â”‚    â”‚    ANALYTICS & AI      â”‚
â”‚  CHAIN   â”‚    â”‚  11 services â”‚    â”‚   9 services   â”‚    â”‚      13 services       â”‚
â”‚ 13 svc   â”‚    â”‚              â”‚    â”‚                â”‚    â”‚                        â”‚
â”‚          â”‚    â”‚ invoice      â”‚    â”‚ notification-  â”‚    â”‚ analytics    reporting  â”‚
â”‚ vendor   â”‚    â”‚ accounting   â”‚    â”‚ orchestrator   â”‚    â”‚ recommendation  ml-fs  â”‚
â”‚ warehouseâ”‚    â”‚ payout       â”‚    â”‚ email  sms     â”‚    â”‚ personalization clv    â”‚
â”‚ fulfill  â”‚    â”‚ reconcile    â”‚    â”‚ push  in-app   â”‚    â”‚ price-opt ad attributionâ”‚
â”‚ tracking â”‚    â”‚ kyc-aml      â”‚    â”‚ whatsapp       â”‚    â”‚                        â”‚
â”‚ cold-ch  â”‚    â”‚ chargeback   â”‚    â”‚ chatbot        â”‚    â”‚ â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”   â”‚
â”‚ ...      â”‚    â”‚ ...          â”‚    â”‚                â”‚    â”‚ â”‚  Apache Flink    â”‚   â”‚
â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜    â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜    â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜    â”‚ â”‚  order-analytics â”‚   â”‚
                                                           â”‚ â”‚  fraud-detection â”‚   â”‚
                                                           â”‚ â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜   â”‚
                                                           â”‚   ClickHouse Â· Weaviate â”‚
                                                           â”‚   Neo4j Â· TimescaleDB   â”‚
                                                           â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜

       â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”
       â”‚                 REMAINING DOMAINS                               â”‚
       â”‚                                                                â”‚
       â”‚  â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”  â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”  â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”        â”‚
       â”‚  â”‚   CUSTOMER   â”‚  â”‚   CONTENT    â”‚  â”‚     B2B      â”‚        â”‚
       â”‚  â”‚  EXPERIENCE  â”‚  â”‚  8 services  â”‚  â”‚  7 services  â”‚        â”‚
       â”‚  â”‚  14 services â”‚  â”‚              â”‚  â”‚              â”‚        â”‚
       â”‚  â”‚ reviews  qa  â”‚  â”‚ media  cms   â”‚  â”‚ org  contractâ”‚        â”‚
       â”‚  â”‚ wishlist     â”‚  â”‚ video  i18n  â”‚  â”‚ quote  edi   â”‚        â”‚
       â”‚  â”‚ support chat â”‚  â”‚ documents    â”‚  â”‚ approval     â”‚        â”‚
       â”‚  â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜  â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜  â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜        â”‚
       â”‚                                                                â”‚
       â”‚  â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”  â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”      â”‚
       â”‚  â”‚ INTEGRATIONS â”‚  â”‚           AFFILIATE               â”‚      â”‚
       â”‚  â”‚  10 services â”‚  â”‚           4 services              â”‚      â”‚
       â”‚  â”‚              â”‚  â”‚                                   â”‚      â”‚
       â”‚  â”‚ erp  crm     â”‚  â”‚ affiliate  referral               â”‚      â”‚
       â”‚  â”‚ marketplace  â”‚  â”‚ influencer  commission-payout     â”‚      â”‚
       â”‚  â”‚ cdp  pim     â”‚  â”‚                                   â”‚      â”‚
       â”‚  â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜  â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜      â”‚
       â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜
```

---

## Infrastructure Layers

```
â•”â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•—
â•‘                    KUBERNETES CLUSTER (EKS / GKE / AKS)                 â•‘
â•‘                                                                          â•‘
â•‘  â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â” â•‘
â•‘  â”‚                     PLATFORM DOMAIN (22 svc)                       â”‚ â•‘
â•‘  â”‚  api-gateway  Â·  graphql-gateway  Â·  web/mobile/partner-bff        â”‚ â•‘
â•‘  â”‚  saga-orchestrator  Â·  event-store  Â·  audit-service               â”‚ â•‘
â•‘  â”‚  scheduler  Â·  worker-job-queue  Â·  dead-letter-service            â”‚ â•‘
â•‘  â”‚  tenant-service  Â·  webhook-service  Â·  config-service             â”‚ â•‘
â•‘  â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜ â•‘
â•‘                                                                          â•‘
â•‘  â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â” â•‘
â•‘  â”‚              WORKFLOW ORCHESTRATION                                  â”‚ â•‘
â•‘  â”‚  Temporal 1.24 â€” durable sagas, retry-safe multi-step flows        â”‚ â•‘
â•‘  â”‚  Argo Workflows â€” CI builds, ML training pipelines                 â”‚ â•‘
â•‘  â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜ â•‘
â•‘                                                                          â•‘
â•‘  â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â” â•‘
â•‘  â”‚              DATA & MESSAGING INFRASTRUCTURE                         â”‚ â•‘
â•‘  â”‚                                                                      â”‚ â•‘
â•‘  â”‚  Kafka (Confluent 7.7.1) + ZooKeeper + Schema Registry             â”‚ â•‘
â•‘  â”‚  RabbitMQ 3.13 (AMQP task queues + delayed messages)               â”‚ â•‘
â•‘  â”‚  NATS JetStream 2.10 (real-time pub/sub)                           â”‚ â•‘
â•‘  â”‚                                                                      â”‚ â•‘
â•‘  â”‚  PostgreSQL 16  Â·  MongoDB 8.0  Â·  Redis 7  Â·  Memcached 1.6      â”‚ â•‘
â•‘  â”‚  Cassandra 5.0  Â·  TimescaleDB 2.15  Â·  Elasticsearch 8.15        â”‚ â•‘
â•‘  â”‚  ClickHouse 24.8  Â·  Weaviate 1.26  Â·  Neo4j 5.23                 â”‚ â•‘
â•‘  â”‚  OpenSearch 2.17  Â·  MinIO  Â·  etcd 3.5                           â”‚ â•‘
â•‘  â”‚                                                                      â”‚ â•‘
â•‘  â”‚  Debezium 2.7 (CDC: Postgres + MongoDB â†’ Kafka)                    â”‚ â•‘
â•‘  â”‚  Apache Flink 1.20 (stream processing: fraud, order analytics)     â”‚ â•‘
â•‘  â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜ â•‘
â•‘                                                                          â•‘
â•‘  â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â” â•‘
â•‘  â”‚              NETWORKING & SERVICE MESH                               â”‚ â•‘
â•‘  â”‚  Istio (mTLS, traffic management, circuit breaking)                â”‚ â•‘
â•‘  â”‚  Cilium eBPF CNI (L3/L4 policies, Hubble observability)           â”‚ â•‘
â•‘  â”‚  Traefik 3.1 (edge router, TLS, service discovery)                â”‚ â•‘
â•‘  â”‚  Consul 1.19 (service discovery, health check, K/V)               â”‚ â•‘
â•‘  â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜ â•‘
â•‘                                                                          â•‘
â•‘  â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â” â•‘
â•‘  â”‚              OBSERVABILITY STACK                                     â”‚ â•‘
â•‘  â”‚                                                                      â”‚ â•‘
â•‘  â”‚  OpenTelemetry (all 13 languages auto-instrumented)                 â”‚ â•‘
â•‘  â”‚  Prometheus + Thanos + VictoriaMetrics (metrics)                   â”‚ â•‘
â•‘  â”‚  Grafana (dashboards) + Alertmanager (routing to PagerDuty/Slack)  â”‚ â•‘
â•‘  â”‚  Jaeger + Grafana Tempo + Zipkin (distributed tracing)             â”‚ â•‘
â•‘  â”‚  Loki + Fluent Bit + Fluentd + ELK + OpenSearch (logs)            â”‚ â•‘
â•‘  â”‚  Sentry OSS + GlitchTip (error tracking)                          â”‚ â•‘
â•‘  â”‚  Grafana Pyroscope (continuous profiling)                          â”‚ â•‘
â•‘  â”‚  Pyrra + Sloth + Uptime Kuma (SLO management)                     â”‚ â•‘
â•‘  â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜ â•‘
â•‘                                                                          â•‘
â•‘  â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â” â•‘
â•‘  â”‚              SECURITY STACK                                          â”‚ â•‘
â•‘  â”‚                                                                      â”‚ â•‘
â•‘  â”‚  Keycloak 25.0 (SSO/OIDC) Â· SPIFFE/SPIRE (workload identity)      â”‚ â•‘
â•‘  â”‚  HashiCorp Vault (dynamic secrets) Â· cert-manager (TLS)            â”‚ â•‘
â•‘  â”‚  OPA/Gatekeeper Â· Kyverno Â· Kubewarden (admission policy)          â”‚ â•‘
â•‘  â”‚  OpenFGA (relationship-based authz)                                â”‚ â•‘
â•‘  â”‚  Falco + Tetragon + Tracee + Wazuh (runtime security/SIEM)        â”‚ â•‘
â•‘  â”‚  Cosign + Syft + Rekor + Fulcio (supply chain)                    â”‚ â•‘
â•‘  â”‚  Trivy + Grype + Semgrep + SonarQube + ZAP (scanning)             â”‚ â•‘
â•‘  â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜ â•‘
â•‘                                                                          â•‘
â•‘  â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â” â•‘
â•‘  â”‚              GITOPS & CI/CD                                          â”‚ â•‘
â•‘  â”‚                                                                      â”‚ â•‘
â•‘  â”‚  Jenkins + Drone CI + Dagger (build / test / scan / publish)       â”‚ â•‘
â•‘  â”‚  ArgoCD (App-of-Apps â†’ 230 Applications) + Flux CD                â”‚ â•‘
â•‘  â”‚  Argo Rollouts (canary 10%â†’25%â†’50%â†’100%)                          â”‚ â•‘
â•‘  â”‚  Argo Events (GitHub webhook â†’ pipeline trigger)                   â”‚ â•‘
â•‘  â”‚  Harbor + Nexus + Gitea (registry / artifact / git)                â”‚ â•‘
â•‘  â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜ â•‘
â•‘                                                                          â•‘
â•‘  â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â” â•‘
â•‘  â”‚              RESILIENCE & TESTING                                    â”‚ â•‘
â•‘  â”‚                                                                      â”‚ â•‘
â•‘  â”‚  KEDA (Kafka consumer lag + Redis list HPA autoscaling)            â”‚ â•‘
â•‘  â”‚  Velero (daily cluster backups â†’ MinIO/S3)                         â”‚ â•‘
â•‘  â”‚  Pod Disruption Budgets (all stateful / critical services)         â”‚ â•‘
â•‘  â”‚                                                                      â”‚ â•‘
â•‘  â”‚  Chaos Mesh (PodChaos, NetworkChaos, StressChaos, IOChaos,        â”‚ â•‘
â•‘  â”‚              HTTPChaos, TimeChaos â€” game-day schedule Sat 02:00)   â”‚ â•‘
â•‘  â”‚  LitmusChaos (Argo Workflow pipelines + SLO probe validation)      â”‚ â•‘
â•‘  â”‚                                                                      â”‚ â•‘
â•‘  â”‚  k6 (smoke / load / spike / soak â€” CI gate)                       â”‚ â•‘
â•‘  â”‚  Locust (exploratory load, web UI, distributed)                    â”‚ â•‘
â•‘  â”‚  Gatling (JVM high-concurrency, HTML reports)                      â”‚ â•‘
â•‘  â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜ â•‘
â•šâ•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•â•
```

---

## Domain Summary

| # | Domain | Services | Languages | Core Responsibility |
|---|---|---|---|---|
| 1 | Platform | 27 | Go, Java, Python | API gateways, BFFs, event store, saga orchestration, Temporal workflows |
| 2 | Identity | 11 | Rust, Java, Go | Auth, users, sessions, MFA, GDPR, SSO, bot detection |
| 3 | Catalog | 15 | Go, Java, Kotlin, Python, Node.js | Products, pricing, inventory, variants, search |
| 4 | Commerce | 28 | Go, Kotlin, Java, C#, Rust, Python, Node.js | Cart, checkout, orders, payments, promotions, BNPL, flash sales |
| 5 | Supply Chain | 17 | Go, Java, Kotlin, Python, Node.js | Vendors, warehouses, fulfilment, tracking, cold chain, routing |
| 6 | Financial | 15 | Java, Kotlin, Go | Invoicing, accounting, payouts, compliance, forex, dunning |
| 7 | Customer Experience | 17 | Go, Java, Node.js | Reviews, support, wishlists, price alerts, gift registry, loyalty tiers |
| 8 | Communications | 12 | Node.js, Python, Go | Email, SMS, push, in-app, WhatsApp, Telegram, voice, chatbot |
| 9 | Content | 9 | Go, Java, Python, Node.js | Media, CMS, documents, i18n, video, A/B content |
| 10 | Analytics & AI | 13 | Python, Scala, Java | Events, ML, recommendations, attribution, CLV, Flink stream jobs |
| 11 | B2B | 10 | Java, Kotlin, Go | Organisations, contracts, procurement, EDI, RFP, vendor onboarding |
| 12 | Integrations | 14 | Java, Go, Python | ERP, CRM, marketplace, payment gateway, CDP, ETL, iPaaS |
| 13 | Affiliate | 6 | Go | Affiliate, referral, influencer, commissions, fraud prevention |
| 14 | Marketplace | 8 | Go, Java, Node.js | Seller registration, listing approval, commissions, disputes |
| 15 | Gamification | 6 | Go | Points, badges, leaderboards, challenges, streaks |
| 16 | Developer Platform | 6 | Go, Node.js | API management, sandboxes, OAuth, developer analytics |
| 17 | Compliance | 5 | Go, Java | Data retention, consent audit, privacy requests, data lineage |
| 18 | Sustainability | 5 | Go | Carbon tracking, eco scoring, green shipping, offset |
| 19 | Web (Frontends) | 6 | TS/Next.js, React, Vue, Angular, React Native | Storefront, Admin, Seller, Partner, Mobile, Dev portals |
| Total | 19 domains | 230 | 13 languages | |

---

## Data Architecture

```
â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”
â”‚                     DATA FLOW ARCHITECTURE                               â”‚
â”‚                                                                          â”‚
â”‚  OPERATIONAL DATABASES (source of truth per service)                    â”‚
â”‚  â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â” â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â” â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â” â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â” â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”   â”‚
â”‚  â”‚ Postgres  â”‚ â”‚ MongoDB  â”‚ â”‚  Redis   â”‚ â”‚Cassandra â”‚ â”‚  MinIO   â”‚   â”‚
â”‚  â”‚  (ACID)   â”‚ â”‚  (docs)  â”‚ â”‚(cache)   â”‚ â”‚(timeseriesâ”‚ â”‚(objects) â”‚   â”‚
â”‚  â””â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”˜ â””â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”˜ â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜ â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜ â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜   â”‚
â”‚        â”‚             â”‚                                                   â”‚
â”‚        â””â”€â”€â”€â”€â”€â”€â”€â”€ Debezium CDC â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”      â”‚
â”‚                       â”‚                                          â”‚      â”‚
â”‚                       â–¼                                          â”‚      â”‚
â”‚              â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”                                 â”‚      â”‚
â”‚              â”‚   KAFKA TOPICS   â”‚                                 â”‚      â”‚
â”‚              â”‚  *.cdc (changes) â”‚                                 â”‚      â”‚
â”‚              â””â”€â”€â”€â”€â”€â”€â”€â”€â”¬â”€â”€â”€â”€â”€â”€â”€â”€â”˜                                 â”‚      â”‚
â”‚                       â”‚                                          â”‚      â”‚
â”‚          â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”¼â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”                 â”‚      â”‚
â”‚          â–¼            â–¼                        â–¼                 â”‚      â”‚
â”‚  â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â” â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”   â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”       â”‚      â”‚
â”‚  â”‚ Elasticsearchâ”‚ â”‚  ClickHouse  â”‚   â”‚   OpenSearch     â”‚       â”‚      â”‚
â”‚  â”‚ (product     â”‚ â”‚ (OLAP reportsâ”‚   â”‚  (log analytics  â”‚       â”‚      â”‚
â”‚  â”‚  search)     â”‚ â”‚  revenue MV) â”‚   â”‚   audit search)  â”‚       â”‚      â”‚
â”‚  â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜ â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜   â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜       â”‚      â”‚
â”‚                       â–²                                          â”‚      â”‚
â”‚          â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜                                          â”‚      â”‚
â”‚          â”‚                                                       â”‚      â”‚
â”‚  â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”   â”‚      â”‚
â”‚  â”‚              APACHE FLINK (stream processing)             â”‚   â”‚      â”‚
â”‚  â”‚  order-analytics job:  Kafka â†’ windowed revenue â†’ CH     â”‚â—„â”€â”€â”˜      â”‚
â”‚  â”‚  fraud-detection job:  Kafka â†’ velocity check â†’ Kafka    â”‚          â”‚
â”‚  â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜          â”‚
â”‚                                                                          â”‚
â”‚  SPECIALIST STORES                                                       â”‚
â”‚  â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â” â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â” â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â” â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”                 â”‚
â”‚  â”‚ Weaviate â”‚ â”‚  Neo4j   â”‚ â”‚TimescaleDBâ”‚ â”‚Memcached â”‚                 â”‚
â”‚  â”‚(vectors  â”‚ â”‚(product  â”‚ â”‚(time-seriesâ”‚ â”‚(hot read â”‚                 â”‚
â”‚  â”‚ ANN)     â”‚ â”‚ graph)   â”‚ â”‚ metrics)  â”‚ â”‚ cache)   â”‚                 â”‚
â”‚  â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜ â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜ â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜ â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜                 â”‚
â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜
```

---

## Request Lifecycle â€” Full Purchase Journey

```
Browser                 Traefik          API Gateway         Services
   â”‚                       â”‚                  â”‚                  â”‚
   â”‚â”€â”€ POST /checkout â”€â”€â”€â”€â”€â–ºâ”‚                  â”‚                  â”‚
   â”‚                       â”‚â”€â”€ TLS â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â–ºâ”‚                  â”‚
   â”‚                       â”‚  WAF check        â”‚â”€â”€ JWT validate â”€â–ºâ”‚ auth-service
   â”‚                       â”‚                  â”‚â—„â”€ token valid â”€â”€â”€â”‚
   â”‚                       â”‚                  â”‚                  â”‚
   â”‚                       â”‚                  â”‚â”€â”€ gRPC â”€â”€â”€â”€â”€â”€â”€â”€â”€â–ºâ”‚ cart-service
   â”‚                       â”‚                  â”‚â—„â”€ cart contents â”€â”‚
   â”‚                       â”‚                  â”‚                  â”‚
   â”‚                       â”‚                  â”‚â”€â”€ gRPC â”€â”€â”€â”€â”€â”€â”€â”€â”€â–ºâ”‚ inventory-service
   â”‚                       â”‚                  â”‚â”€â”€ gRPC â”€â”€â”€â”€â”€â”€â”€â”€â”€â–ºâ”‚ tax-service
   â”‚                       â”‚                  â”‚â”€â”€ gRPC â”€â”€â”€â”€â”€â”€â”€â”€â”€â–ºâ”‚ shipping-service
   â”‚                       â”‚                  â”‚â”€â”€ gRPC â”€â”€â”€â”€â”€â”€â”€â”€â”€â–ºâ”‚ promotions-service
   â”‚                       â”‚                  â”‚                  â”‚
   â”‚                       â”‚                  â”‚â”€â”€ Temporal â”€â”€â”€â”€â”€â–ºâ”‚ saga-orchestrator
   â”‚                       â”‚                  â”‚  (order saga)    â”‚â”€â”€ gRPC â”€â”€â–º payment-service
   â”‚                       â”‚                  â”‚                  â”‚â”€â”€ gRPC â”€â”€â–º order-service
   â”‚                       â”‚                  â”‚                  â”‚
   â”‚                       â”‚                  â”‚  order created   â”‚â”€â”€ Kafka â”€â”€â–º fulfillment
   â”‚                       â”‚                  â”‚                  â”‚â”€â”€ Kafka â”€â”€â–º loyalty
   â”‚                       â”‚                  â”‚                  â”‚â”€â”€ Kafka â”€â”€â–º notifications
   â”‚                       â”‚                  â”‚                  â”‚â”€â”€ Kafka â”€â”€â–º analytics
   â”‚                       â”‚                  â”‚                  â”‚â”€â”€ Kafka â”€â”€â–º fraud-detection
   â”‚                       â”‚                  â”‚                  â”‚
   â”‚â—„â”€â”€ 201 Created â”€â”€â”€â”€â”€â”€â”€â”‚â—„â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”‚                  â”‚
   â”‚   { orderId: "..." }  â”‚                  â”‚                  â”‚
```

---

## Key Architectural Patterns

| Pattern | Implementation | Why |
|---|---|---|
| API Gateway | `api-gateway` (Go) | Single ingress; JWT validation, rate limiting, and routing in one place |
| BFF | `web-bff`, `mobile-bff`, `partner-bff` | Different clients need different response shapes; BFF tailors the API per client |
| CQRS | `order-service`, `accounting-service` | High-read reporting queries must not compete with write transactions |
| Event Sourcing | `event-store-service` (Postgres, append-only) | Full audit trail; enables replay and temporal queries |
| Saga | `saga-orchestrator` (choreography) + Temporal (orchestration) | Distributed transactions across services without 2PC; each step is compensatable |
| Database-per-service | 230 service-owned databases | No coupling between services; each chooses the optimal store for its access pattern |
| Outbox pattern | Services write to outbox table â†’ Debezium CDC â†’ Kafka | Guarantees at-least-once event delivery even if Kafka is briefly unavailable |
| CQRS read models | Debezium â†’ ClickHouse / Elasticsearch / OpenSearch | Reporting queries run against optimised read stores, not transactional databases |
| Stream processing | Apache Flink (Kafka â†’ stateful jobs â†’ Kafka/ClickHouse) | Fraud detection and analytics require windowed aggregations across millions of events |
| Circuit breaker | All gRPC clients (exponential backoff + jitter, 3 attempts) | Prevents cascade failure when a downstream service is slow or unavailable |
| Progressive delivery | Argo Rollouts canary (10%â†’25%â†’50%â†’100%) | Reduce blast radius of a broken release; auto-rollback on metric degradation |

---

## Observability Stack

```
â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”
â”‚                     OBSERVABILITY PIPELINE                               â”‚
â”‚                                                                          â”‚
â”‚  Service A  â†’  OpenTelemetry SDK  â†’  OTel Collector  â†’  â”¬â”€â”€ Jaeger      â”‚
â”‚                 (traces, metrics,                         â”œâ”€â”€ Tempo       â”‚
â”‚                  logs, profiles)                          â”œâ”€â”€ Prometheus  â”‚
â”‚                                                           â”œâ”€â”€ Loki        â”‚
â”‚                                                           â””â”€â”€ Pyroscope   â”‚
â”‚                                                                          â”‚
â”‚  Prometheus â†’ Thanos (long-term storage, dedup, global query)           â”‚
â”‚  Loki â† Fluent Bit (log shipping from all pods) â† Fluentd               â”‚
â”‚  OpenSearch â† Logstash â† application structured logs                    â”‚
â”‚                                                                          â”‚
â”‚  Grafana â† all of the above (unified dashboards)                        â”‚
â”‚  Alertmanager â† Prometheus alerts â†’ Slack / PagerDuty / email          â”‚
â”‚                                                                          â”‚
â”‚  Sentry OSS + GlitchTip â† application exception tracking               â”‚
â”‚  Uptime Kuma â† external HTTP health checks (public status page)         â”‚
â”‚  Pyrra + Sloth â† SLO burn rate recording rules                          â”‚
â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜
```

---

## Resilience Architecture

```
â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”
â”‚                    RESILIENCE LAYERS                                      â”‚
â”‚                                                                          â”‚
â”‚  APPLICATION LEVEL                                                       â”‚
â”‚  â”œâ”€â”€ Circuit breakers on all gRPC clients (grpc-go retry interceptor)  â”‚
â”‚  â”œâ”€â”€ Timeouts configured per RPC method (per proto service definition)  â”‚
â”‚  â”œâ”€â”€ Idempotency keys on all payment and order mutations                â”‚
â”‚  â””â”€â”€ Saga compensation handlers for every distributed transaction step  â”‚
â”‚                                                                          â”‚
â”‚  INFRASTRUCTURE LEVEL                                                    â”‚
â”‚  â”œâ”€â”€ Pod Disruption Budgets â€” minAvailable=1 on all stateful services   â”‚
â”‚  â”œâ”€â”€ KEDA autoscaling â€” Kafka consumer lag + Redis queue depth triggers â”‚
â”‚  â”œâ”€â”€ Argo Rollouts â€” canary with Prometheus metric gates                â”‚
â”‚  â””â”€â”€ Velero daily backup â€” namespace-level restore to clean cluster     â”‚
â”‚                                                                          â”‚
â”‚  CHAOS ENGINEERING                                                       â”‚
â”‚  â”œâ”€â”€ Chaos Mesh experiments (automated, namespace-scoped)               â”‚
â”‚  â”‚   â”œâ”€â”€ PodChaos    â€” api-gateway, order, payment, kafka (33%)        â”‚
â”‚  â”‚   â”œâ”€â”€ NetworkChaos â€” delay (100ms), loss (10%), partition, bw limit â”‚
â”‚  â”‚   â”œâ”€â”€ StressChaos  â€” CPU 80% on checkout, memory 256MB on recommend â”‚
â”‚  â”‚   â”œâ”€â”€ IOChaos      â€” disk delay 100ms on product-catalog            â”‚
â”‚  â”‚   â”œâ”€â”€ HTTPChaos    â€” 30% abort on payment, 3s delay on checkout     â”‚
â”‚  â”‚   â””â”€â”€ TimeChaos    â€” -10m clock skew on order-service               â”‚
â”‚  â”œâ”€â”€ Chaos Mesh Workflows â€” multi-phase resilience scenarios            â”‚
â”‚  â”œâ”€â”€ Game-day schedule    â€” Saturdays 02:00 UTC (automated)             â”‚
â”‚  â””â”€â”€ LitmusChaos          â€” Argo Workflow pipelines with SLO probes    â”‚
â”‚                                                                          â”‚
â”‚  LOAD TESTING (CI-gated SLO validation)                                 â”‚
â”‚  â”œâ”€â”€ k6 smoke test  â€” every deploy (1 VU, 2m, p95 < 2s)               â”‚
â”‚  â”œâ”€â”€ k6 load test   â€” nightly (50 VUs, checkout p95 < 5s)             â”‚
â”‚  â”œâ”€â”€ k6 spike test  â€” weekly (500 VUs burst, recovery rate > 90%)      â”‚
â”‚  â”œâ”€â”€ k6 soak test   â€” weekly overnight (30 VUs, 2h, no p95 drift)      â”‚
â”‚  â””â”€â”€ Gatling        â€” pre-release (JVM concurrency, HTML reports)       â”‚
â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜
```

---

## CI/CD & GitOps Pipeline

```
â”Œâ”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”
â”‚                    CI/CD PIPELINE                                         â”‚
â”‚                                                                          â”‚
â”‚  1. CODE PUSH â†’ GitHub / Gitea                                          â”‚
â”‚     â””â”€â”€ Argo Events (GitHub webhook EventSource â†’ Sensor â†’ trigger)    â”‚
â”‚                                                                          â”‚
â”‚  2. CI BUILD (Jenkins / Drone CI / Dagger)                              â”‚
â”‚     â”œâ”€â”€ Checkout + compile (language-specific)                          â”‚
â”‚     â”œâ”€â”€ Unit + integration tests                                        â”‚
â”‚     â”œâ”€â”€ SonarQube SAST (quality gate)                                   â”‚
â”‚     â”œâ”€â”€ Semgrep custom security rules                                   â”‚
â”‚     â”œâ”€â”€ Checkov / KICS IaC scan (if infra changed)                     â”‚
â”‚     â”œâ”€â”€ Docker multi-stage build (non-root, minimal image)             â”‚
â”‚     â”œâ”€â”€ Trivy scan â†’ block on CRITICAL CVE                             â”‚
â”‚     â”œâ”€â”€ Grype scan â†’ block on CRITICAL CVE (second opinion)            â”‚
â”‚     â”œâ”€â”€ Syft SBOM generation (CycloneDX + SPDX)                        â”‚
â”‚     â”œâ”€â”€ Cosign keyless sign (Fulcio cert + Rekor log entry)            â”‚
â”‚     â””â”€â”€ Push to Harbor registry                                         â”‚
â”‚                                                                          â”‚
â”‚  3. GITOPS SYNC (ArgoCD â€” pull model)                                   â”‚
â”‚     â”œâ”€â”€ ArgoCD detects Helm chart change in Git                         â”‚
â”‚     â”œâ”€â”€ Pre-flight checks (cluster health, secret existence, PDB)      â”‚
â”‚     â”œâ”€â”€ Kyverno admission: verify Cosign signature before pod starts   â”‚
â”‚     â””â”€â”€ Argo Rollouts canary: 10% â†’ metrics check â†’ 25% â†’ ... â†’ 100%  â”‚
â”‚                                                                          â”‚
â”‚  4. POST-DEPLOY VALIDATION                                               â”‚
â”‚     â”œâ”€â”€ k6 smoke test against new deployment                           â”‚
â”‚     â”œâ”€â”€ OWASP ZAP DAST scan (nightly on staging)                       â”‚
â”‚     â””â”€â”€ Nuclei CVE template scan (nightly on staging)                  â”‚
â””â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”˜
```

---

## Infrastructure as Code

| Tool | Provider | Location |
|---|---|---|
| Terraform | AWS EKS, GCP GKE, Azure AKS, Jenkins EC2 | `infra/terraform/` |
| OpenTofu | AWS, GCP, Azure (open-source Terraform fork) | `infra/opentofu/` |
| Crossplane | K8s-native IaC (compositions + claims) | `infra/crossplane/` |
| Ansible | K8s node bootstrapping, OS configuration | `infra/ansible/` |

---

## References

- [Communication Patterns](communication-patterns.md)
- [Database Strategy](database-strategy.md)
- [Domain Map](domain-map.md)
- [Security Model](security-model.md)
- [ADR-001: gRPC for Internal Communication](../adr/001-grpc-for-internal-communication.md)
- [ADR-002: Kafka for Async Events](../adr/002-kafka-for-async-events.md)
- [ADR-005: Database-per-Service](../adr/005-database-per-service.md)
- [ADR-006: GitOps with ArgoCD](../adr/006-gitops-with-argocd.md)
- [Chaos Engineering Runbook](../../chaos/README.md)
- [Load Testing Guide](../../load-testing/README.md)
