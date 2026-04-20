SERVICE := supplier-portal-service
IMAGE   := shopos/$(SERVICE)
TAG     ?= latest
MVN     := ./mvnw

.PHONY: help build test run clean docker-build docker-push

help: ## Show this help
	@awk 'BEGIN{FS=":.*##"} /^[a-zA-Z_-]+:.*##/{printf "  \033[36m%-18s\033[0m %s\n",$$1,$$2}' $(MAKEFILE_LIST)

build: ## Compile (skip tests)
	mvn package -DskipTests -B --no-transfer-progress

test: ## Run unit tests
	mvn test -B --no-transfer-progress

run: ## Run locally via Spring Boot plugin
	mvn spring-boot:run

clean: ## Remove build artifacts
	mvn clean

docker-build: ## Build Docker image
	docker build -t $(IMAGE):$(TAG) .

docker-push: docker-build ## Push Docker image to registry
	docker push $(IMAGE):$(TAG)
