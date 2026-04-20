SERVICE  = product-catalog-service
MODULE   = github.com/shopos/$(SERVICE)
IMAGE    = shopos/$(SERVICE)
TAG     ?= latest

.PHONY: all build test lint fmt vet tidy docker-build docker-run clean help

all: build

## build: Compile the service binary
build:
	go build -o bin/$(SERVICE) ./...

## test: Run all unit tests with race detector
test:
	go test -race -count=1 -timeout 120s ./...

## test-verbose: Run tests with verbose output
test-verbose:
	go test -race -count=1 -timeout 120s -v ./...

## lint: Run golangci-lint (requires golangci-lint installed)
lint:
	golangci-lint run ./...

## fmt: Format source files
fmt:
	gofmt -w .

## vet: Run go vet
vet:
	go vet ./...

## tidy: Tidy and vendor go modules
tidy:
	go mod tidy

## docker-build: Build the Docker image
docker-build:
	docker build -t $(IMAGE):$(TAG) .

## docker-run: Run the service in Docker (requires .env file)
docker-run:
	docker run --rm \
		--env-file .env \
		-p 8110:8110 \
		-p 50070:50070 \
		$(IMAGE):$(TAG)

## clean: Remove build artifacts
clean:
	rm -rf bin/

## help: Show this help
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## //'
