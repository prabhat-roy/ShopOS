SERVICE_NAME   := social-commerce-service
IMAGE_NAME     := shopos/$(SERVICE_NAME)
IMAGE_TAG      ?= latest
CONTAINER_PORT := 8902
GRPC_PORT      := 50171

.PHONY: all install build test run docker-build docker-run docker-stop clean help

all: install

## install: Install npm dependencies
install:
	npm ci

## build: Alias for install (no compile step for Node.js)
build: install

## test: Run the Jest test suite
test:
	npm test

## test-coverage: Run tests with coverage report
test-coverage:
	npm run test:coverage

## run: Start the service in production mode
run:
	npm start

## dev: Start the service with hot-reload via nodemon
dev:
	npm run dev

## docker-build: Build the Docker image
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

## clean: Remove node_modules and coverage artefacts
clean:
	rm -rf node_modules coverage

## help: Print this help
help:
	@grep -E '^##' Makefile | sed 's/## //'