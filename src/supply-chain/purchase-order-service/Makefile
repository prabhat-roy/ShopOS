SERVICE_NAME  := purchase-order-service
IMAGE_NAME    := shopos/$(SERVICE_NAME)
IMAGE_TAG     ?= latest
GRPC_PORT     ?= 50101
HTTP_PORT     ?= 8201

.PHONY: build test run clean docker-build docker-run help

help:
	@echo "Usage: make <target>"
	@echo ""
	@echo "Targets:"
	@echo "  build         Compile and package the service (skip tests)"
	@echo "  test          Run unit tests"
	@echo "  run           Run the service locally (requires Postgres)"
	@echo "  clean         Remove build artifacts"
	@echo "  docker-build  Build the Docker image"
	@echo "  docker-run    Run the service in Docker"

build:
	./gradlew bootJar

test:
	./gradlew test

run:
	./gradlew bootRun

clean:
	./gradlew clean

docker-build:
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .

docker-run:
	docker run --rm \
		--name $(SERVICE_NAME) \
		-p $(HTTP_PORT):$(HTTP_PORT) \
		--env-file .env \
		$(IMAGE_NAME):$(IMAGE_TAG)
