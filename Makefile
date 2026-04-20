.PHONY: build test lint docker-build docker-push proto-gen \
        up down logs ps \
        deploy-local undeploy-local deploy-svc \
        migrate clean help

REGISTRY ?= shopos
TAG      ?= latest

# ── Service lists by domain ────────────────────────────────────────────────────

PLATFORM_SERVICES := api-gateway web-bff mobile-bff partner-bff config-service \
  feature-flag-service rate-limiter-service health-check-service saga-orchestrator \
  event-store-service cache-warming-service webhook-service scheduler-service \
  worker-job-queue audit-service load-generator admin-portal graphql-gateway \
  dead-letter-service geolocation-service event-replay-service tenant-service

IDENTITY_SERVICES := auth-service user-service session-service permission-service \
  mfa-service gdpr-service api-key-service device-fingerprint-service

CATALOG_SERVICES := product-catalog-service category-service brand-service \
  pricing-service inventory-service bundle-service configurator-service \
  subscription-product-service search-service seo-service product-import-service \
  price-list-service

COMMERCE_SERVICES := cart-service checkout-service order-service payment-service \
  shipping-service currency-service tax-service promotions-service loyalty-service \
  return-refund-service subscription-billing-service fraud-detection-service \
  wallet-service ab-testing-service gift-card-service address-validation-service \
  digital-goods-service voucher-service pre-order-service backorder-service \
  waitlist-service flash-sale-service bnpl-service

SUPPLY_CHAIN_SERVICES := vendor-service purchase-order-service warehouse-service \
  fulfillment-service tracking-service label-service carrier-integration-service \
  demand-forecast-service customs-duties-service returns-logistics-service \
  supplier-portal-service cold-chain-service supplier-rating-service

FINANCIAL_SERVICES := invoice-service accounting-service payout-service \
  reconciliation-service tax-reporting-service expense-management-service \
  credit-service kyc-aml-service budget-service chargeback-service \
  revenue-recognition-service

CX_SERVICES := review-rating-service qa-service wishlist-service compare-service \
  recently-viewed-service support-ticket-service live-chat-service \
  consent-management-service age-verification-service survey-service \
  feedback-service price-alert-service back-in-stock-service gift-registry-service

COMMUNICATIONS_SERVICES := notification-orchestrator email-service sms-service \
  push-notification-service template-service in-app-notification-service \
  digest-service whatsapp-service chatbot-service

CONTENT_SERVICES := media-asset-service image-processing-service document-service \
  cms-service video-service sitemap-service i18n-l10n-service data-export-service

ANALYTICS_AI_SERVICES := analytics-service reporting-service recommendation-service \
  sentiment-analysis-service price-optimization-service ml-feature-store \
  personalization-service data-pipeline-service ad-service event-tracking-service \
  attribution-service clv-service search-analytics-service

B2B_SERVICES := organization-service contract-service quote-rfq-service \
  approval-workflow-service b2b-credit-limit-service edi-service \
  marketplace-seller-service

INTEGRATIONS_SERVICES := erp-integration-service marketplace-connector-service \
  social-commerce-service crm-integration-service payment-gateway-integration \
  logistics-provider-integration tax-provider-integration pim-integration-service \
  cdp-integration-service accounting-integration-service

AFFILIATE_SERVICES := affiliate-service referral-service influencer-service \
  commission-payout-service

ALL_SERVICES := \
  $(addprefix src/platform/,$(PLATFORM_SERVICES)) \
  $(addprefix src/identity/,$(IDENTITY_SERVICES)) \
  $(addprefix src/catalog/,$(CATALOG_SERVICES)) \
  $(addprefix src/commerce/,$(COMMERCE_SERVICES)) \
  $(addprefix src/supply-chain/,$(SUPPLY_CHAIN_SERVICES)) \
  $(addprefix src/financial/,$(FINANCIAL_SERVICES)) \
  $(addprefix src/customer-experience/,$(CX_SERVICES)) \
  $(addprefix src/communications/,$(COMMUNICATIONS_SERVICES)) \
  $(addprefix src/content/,$(CONTENT_SERVICES)) \
  $(addprefix src/analytics-ai/,$(ANALYTICS_AI_SERVICES)) \
  $(addprefix src/b2b/,$(B2B_SERVICES)) \
  $(addprefix src/integrations/,$(INTEGRATIONS_SERVICES)) \
  $(addprefix src/affiliate/,$(AFFILIATE_SERVICES))

# ── Build ──────────────────────────────────────────────────────────────────────

## build: Build all services (runs make build in each service directory)
build:
	@for svc in $(ALL_SERVICES); do \
	  echo "==> Building $$svc"; \
	  $(MAKE) -C $$svc build || exit 1; \
	done

## test: Run all tests
test:
	@for svc in $(ALL_SERVICES); do \
	  echo "==> Testing $$svc"; \
	  $(MAKE) -C $$svc test || exit 1; \
	done

## lint: Lint all services
lint:
	@for svc in $(ALL_SERVICES); do \
	  echo "==> Linting $$svc"; \
	  $(MAKE) -C $$svc lint 2>/dev/null || true; \
	done

# ── Docker ─────────────────────────────────────────────────────────────────────

## docker-build: Build Docker images for all services
docker-build:
	@for svc in $(ALL_SERVICES); do \
	  name=$$(basename $$svc); \
	  echo "==> docker build $$name"; \
	  docker build -t $(REGISTRY)/$$name:$(TAG) $$svc || exit 1; \
	done

## docker-push: Push all images to registry
docker-push:
	@for svc in $(ALL_SERVICES); do \
	  name=$$(basename $$svc); \
	  docker push $(REGISTRY)/$$name:$(TAG) || exit 1; \
	done

## docker-build-domain: Build images for a specific domain: make docker-build-domain DOMAIN=commerce
docker-build-domain:
	@for svc in $(ALL_SERVICES); do \
	  echo $$svc | grep -q "src/$(DOMAIN)/" || continue; \
	  name=$$(basename $$svc); \
	  echo "==> docker build $$name"; \
	  docker build -t $(REGISTRY)/$$name:$(TAG) $$svc || exit 1; \
	done

# ── Protobuf ───────────────────────────────────────────────────────────────────

## proto-gen: Generate gRPC code from all .proto files
proto-gen:
	@find proto -name '*.proto' | while read f; do \
	  echo "==> protoc $$f"; \
	  protoc --proto_path=proto \
	    --go_out=. --go_opt=paths=source_relative \
	    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
	    --java_out=. \
	    --python_out=. \
	    --grpc-gateway_out=. --grpc-gateway_opt=paths=source_relative \
	    $$f; \
	done

# ── Local Docker Compose ───────────────────────────────────────────────────────

## up: Start full local stack (all services + infra)
up:
	docker compose up -d

## up-infra: Start infrastructure only (databases, brokers, tools)
up-infra:
	docker compose up -d postgres mongodb redis cassandra elasticsearch minio \
	  zookeeper kafka etcd rabbitmq nats clickhouse weaviate neo4j opensearch \
	  scylladb memcached temporal temporal-ui debezium flink-jobmanager flink-taskmanager \
	  traefik consul keycloak mlflow pyroscope zipkin uptime-kuma nexus gitea sonarqube

## down: Stop and remove all containers
down:
	docker compose down

## down-v: Stop containers and remove volumes
down-v:
	docker compose down -v

## ps: Show running containers
ps:
	docker compose ps

## logs: Tail logs for a service: make logs SVC=order-service
logs:
	docker compose logs -f $(SVC)

## restart: Restart a service: make restart SVC=order-service
restart:
	docker compose restart $(SVC)

# ── Kubernetes / Helm ──────────────────────────────────────────────────────────

## deploy-local: Deploy all services to local Kubernetes via per-service Helm charts
deploy-local:
	@for svc in $(ALL_SERVICES); do \
	  name=$$(basename $$svc); \
	  echo "==> helm upgrade $$name"; \
	  helm upgrade --install $$name helm/charts/$$name \
	    --namespace $$name --create-namespace \
	    --values helm/charts/$$name/values.yaml \
	    --set image.tag=$(TAG) 2>/dev/null || true; \
	done

## undeploy-local: Remove all service Helm releases
undeploy-local:
	@for svc in $(ALL_SERVICES); do \
	  name=$$(basename $$svc); \
	  helm uninstall $$name --namespace $$name 2>/dev/null || true; \
	done

## deploy-svc: Deploy a single service: make deploy-svc SVC=order-service NS=order-service
deploy-svc:
	helm upgrade --install $(SVC) helm/charts/$(SVC) \
	  --namespace $(or $(NS),$(SVC)) --create-namespace \
	  --values helm/charts/$(SVC)/values.yaml \
	  --set image.tag=$(TAG)

# ── Database Migrations ────────────────────────────────────────────────────────

## migrate: Run DB migrations for all services that have them
migrate:
	@for svc in $(ALL_SERVICES); do \
	  if [ -d "$$svc/db/migrations" ]; then \
	    echo "==> migrating $$svc"; \
	    psql "$(DATABASE_URL)" -f $$svc/db/migrations/*.sql 2>/dev/null || true; \
	  fi; \
	done

# ── Cleanup ────────────────────────────────────────────────────────────────────

## clean: Remove build artifacts from all services
clean:
	@for svc in $(ALL_SERVICES); do \
	  rm -rf $$svc/bin $$svc/target $$svc/build $$svc/dist 2>/dev/null || true; \
	done

## clean-images: Remove all locally built Docker images
clean-images:
	@for svc in $(ALL_SERVICES); do \
	  name=$$(basename $$svc); \
	  docker rmi $(REGISTRY)/$$name:$(TAG) 2>/dev/null || true; \
	done

# ── Help ───────────────────────────────────────────────────────────────────────

help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/^## //'
