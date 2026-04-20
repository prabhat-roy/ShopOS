SERVICE   := session-service
IMAGE     := shopos/$(SERVICE)
VERSION   ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
PORT_HTTP := 8100
PORT_GRPC := 50062

.PHONY: all build run test lint fmt vet tidy docker-build docker-run clean

## Default target
all: build

## Compile the binary
build:
	go build -ldflags="-w -s" -o bin/$(SERVICE) ./main.go

## Run locally (requires Redis reachable at REDIS_ADDR)
run:
	go run ./main.go

## Run all unit tests with race detector
test:
	go test -race -count=1 ./...

## Run tests and emit coverage report
cover:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

## Lint (requires golangci-lint)
lint:
	golangci-lint run ./...

## Format source
fmt:
	gofmt -w .

## Vet source
vet:
	go vet ./...

## Tidy module dependencies
tidy:
	go mod tidy

## Build Docker image
docker-build:
	docker build -t $(IMAGE):$(VERSION) -t $(IMAGE):latest .

## Run Docker image (requires REDIS_ADDR env var)
docker-run:
	docker run --rm \
		-p $(PORT_HTTP):$(PORT_HTTP) \
		-p $(PORT_GRPC):$(PORT_GRPC) \
		-e REDIS_ADDR=$(REDIS_ADDR) \
		-e REDIS_PASSWORD=$(REDIS_PASSWORD) \
		-e SESSION_TTL=$(SESSION_TTL) \
		$(IMAGE):latest

## Remove build artifacts
clean:
	rm -rf bin/ coverage.out coverage.html
