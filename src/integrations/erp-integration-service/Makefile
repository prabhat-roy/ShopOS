SERVICE_NAME   := erp-integration-service
IMAGE_NAME     := shopos/$(SERVICE_NAME)
IMAGE_TAG      ?= latest
CONTAINER_PORT := 8900
GRPC_PORT      := 50170
JAVA_OPTS      ?= -XX:+UseContainerSupport -XX:MaxRAMPercentage=75.0

.PHONY: all build test run docker-build docker-run docker-stop clean lint help

all: build

## build: Compile and package the JAR
build:
	mvn clean package -DskipTests

## test: Run all unit and integration tests
test:
	mvn test

## run: Start the service locally (requires JDK 21)
run:
	mvn spring-boot:run -Dspring-boot.run.jvmArguments="$(JAVA_OPTS)"

## docker-build: Build Docker image
docker-build:
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .

## docker-run: Run the Docker image locally
docker-run:
	docker run --rm \
	  --name $(SERVICE_NAME) \
	  -p $(CONTAINER_PORT):$(CONTAINER_PORT) \
	  --env-file .env.example \
	  $(IMAGE_NAME):$(IMAGE_TAG)

## docker-stop: Stop a running container
docker-stop:
	docker stop $(SERVICE_NAME) || true

## clean: Remove build artifacts
clean:
	mvn clean

## lint: Check code style with Maven checkstyle
lint:
	mvn checkstyle:check

## help: Print this help
help:
	@grep -E '^##' Makefile | sed 's/## //'
