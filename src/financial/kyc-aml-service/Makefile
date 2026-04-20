SERVICE_NAME   := kyc-aml-service
IMAGE_REGISTRY := ghcr.io/shopos
IMAGE_TAG      ?= latest
IMAGE_FULL     := $(IMAGE_REGISTRY)/$(SERVICE_NAME):$(IMAGE_TAG)
CONTAINER_PORT := 8307

.PHONY: build test clean package docker-build docker-push run fmt lint help

## build: compile the service (skip tests)
build:
	mvn compile -q

## test: run unit tests
test:
	mvn test

## package: produce the fat-jar (skip tests)
package:
	mvn package -DskipTests -q

## clean: remove build artifacts
clean:
	mvn clean -q

## docker-build: build Docker image
docker-build:
	docker build -t $(IMAGE_FULL) .

## docker-push: push Docker image to registry
docker-push: docker-build
	docker push $(IMAGE_FULL)

## run: start the service locally (requires Postgres at localhost:5432)
run:
	mvn spring-boot:run

## lint: check code style with Checkstyle (if configured)
lint:
	mvn checkstyle:check -q

## fmt: format source code with Spotless (if configured)
fmt:
	mvn spotless:apply -q

## help: list all available targets
help:
	@grep -E '^## ' Makefile | sed 's/## //'
