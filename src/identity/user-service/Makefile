SERVICE_NAME  := user-service
IMAGE_NAME    := shopos/$(SERVICE_NAME)
IMAGE_TAG     ?= latest

.PHONY: build test run docker-build docker-run clean help

## build: compile and package the JAR (skip tests)
build:
	mvn package -DskipTests

## test: run all unit tests
test:
	mvn test

## run: start the service via Spring Boot Maven plugin
run:
	mvn spring-boot:run

## docker-build: build the Docker image
docker-build:
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .

## docker-run: run the service in Docker (requires Postgres running)
docker-run:
	docker run --rm \
		--env-file .env.example \
		-p 8099:8099 \
		$(IMAGE_NAME):$(IMAGE_TAG)

## clean: remove build artefacts
clean:
	mvn clean

## help: show this help message
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## //'
