SERVICE   := cdp-integration-service
IMAGE     := shopos/$(SERVICE)
TAG       ?= latest
HTTP_PORT ?= 8080

.PHONY: all build run test lint docker-build docker-run tidy

all: build

## build: compile the binary to ./bin/cdp-integration-service
build:
	mkdir -p bin
	go build -trimpath -ldflags="-s -w" -o bin/$(SERVICE) ./main.go

## run: run the service locally
run:
	go run ./main.go

## test: run all unit tests with race detector
test:
	go test -race -count=1 ./...

## lint: run golangci-lint (must be installed separately)
lint:
	golangci-lint run ./...

## tidy: tidy and verify go modules
tidy:
	go mod tidy
	go mod verify

## docker-build: build the Docker image
docker-build:
	docker build -t $(IMAGE):$(TAG) .

## docker-run: run the service in Docker
docker-run: docker-build
	docker run --rm \
	  -e CDP_WEBHOOK_URL="$(CDP_WEBHOOK_URL)" \
	  -e CDP_WRITE_KEY="$(CDP_WRITE_KEY)" \
	  -e KAFKA_BROKERS="$(KAFKA_BROKERS)" \
	  -p $(HTTP_PORT):$(HTTP_PORT) \
	  --env-file .env.example \
	  $(IMAGE):$(TAG)
