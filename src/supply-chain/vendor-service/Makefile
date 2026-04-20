SERVICE_NAME  := vendor-service
IMAGE_NAME    := shopos/$(SERVICE_NAME)
IMAGE_TAG     ?= latest
GRPC_PORT     ?= 50100
HTTP_PORT     ?= 8200

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
	mvn package -DskipTests

test:
	mvn test

run:
	mvn spring-boot:run

clean:
	mvn clean

docker-build:
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .

docker-run:
	docker run --rm \
		--name $(SERVICE_NAME) \
		-p $(HTTP_PORT):$(HTTP_PORT) \
		--env-file .env \
		$(IMAGE_NAME):$(IMAGE_TAG)
