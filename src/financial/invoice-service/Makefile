SERVICE   := invoice-service
IMAGE     := shopos/$(SERVICE)
TAG       ?= latest
REGISTRY  ?= ghcr.io

.PHONY: build test clean docker-build docker-push run help

help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
	  awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Compile the service (skip tests)
	mvn package -DskipTests -B --no-transfer-progress

test: ## Run all unit tests
	mvn test -B --no-transfer-progress

clean: ## Remove build artifacts
	mvn clean -B --no-transfer-progress

docker-build: ## Build Docker image
	docker build -t $(IMAGE):$(TAG) .

docker-push: docker-build ## Push image to registry
	docker tag $(IMAGE):$(TAG) $(REGISTRY)/$(IMAGE):$(TAG)
	docker push $(REGISTRY)/$(IMAGE):$(TAG)

run: ## Run locally (requires DB_URL, DB_USER, DB_PASSWORD env vars)
	mvn spring-boot:run -Dspring-boot.run.jvmArguments="-Dspring.profiles.active=local"
