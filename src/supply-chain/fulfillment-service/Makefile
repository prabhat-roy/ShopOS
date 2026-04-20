SERVICE   := fulfillment-service
IMAGE     := shopos/$(SERVICE)
TAG       ?= latest

.PHONY: build test run docker-build docker-run lint tidy

## build: compile the binary to ./bin/fulfillment-service
build:
	mkdir -p bin
	go build -trimpath -ldflags="-s -w" -o bin/$(SERVICE) ./main.go

## test: run all unit tests with race detector
test:
	go test -race -count=1 ./...

## run: run the service locally (requires DATABASE_URL in environment)
run:
	go run ./main.go

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

## docker-run: run the service in Docker (requires DATABASE_URL)
docker-run:
	docker run --rm \
	  -e DATABASE_URL="$(DATABASE_URL)" \
	  -e HTTP_PORT=8203 \
	  -e GRPC_PORT=50103 \
	  -p 8203:8203 \
	  -p 50103:50103 \
	  $(IMAGE):$(TAG)
