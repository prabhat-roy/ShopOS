SERVICE_NAME  := accounting-service
IMAGE_NAME    := shopos/$(SERVICE_NAME)
IMAGE_TAG     ?= latest
CONTAINER_REG ?= ghcr.io

.PHONY: build test run docker-build docker-push clean

## Build the JAR
build:
	./gradlew bootJar

## Run all tests
test:
	./gradlew test

## Run locally (requires .env or exported vars)
run:
	./gradlew bootRun

## Build Docker image
docker-build:
	docker build -t $(CONTAINER_REG)/$(IMAGE_NAME):$(IMAGE_TAG) .

## Push Docker image
docker-push: docker-build
	docker push $(CONTAINER_REG)/$(IMAGE_NAME):$(IMAGE_TAG)

## Clean build artifacts
clean:
	./gradlew clean

## Start local dependencies (Postgres)
deps-up:
	docker compose -f ../../docker-compose.yml up -d postgres

## Stop local dependencies
deps-down:
	docker compose -f ../../docker-compose.yml down

## Run with local env file
dev:
	set -a && . ./.env.example && set +a && ./gradlew bootRun
