SERVICE_NAME := edi-service
IMAGE_NAME   := shopos/$(SERVICE_NAME)
IMAGE_TAG    ?= latest
PORT         := 8805

.PHONY: build test run docker-build docker-run clean lint

## Build the JAR
build:
	mvn clean package -DskipTests

## Run all unit tests
test:
	mvn test

## Run the service locally (requires Java 21)
run:
	mvn spring-boot:run

## Build the Docker image
docker-build:
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .

## Run the Docker container
docker-run:
	docker run --rm \
		--name $(SERVICE_NAME) \
		-p $(PORT):$(PORT) \
		--env-file .env \
		$(IMAGE_NAME):$(IMAGE_TAG)

## Remove compiled artifacts
clean:
	mvn clean

## Lint / static analysis via Maven checkstyle
lint:
	mvn checkstyle:check

## Show help
help:
	@grep -E '^##' Makefile | sed 's/## //'
