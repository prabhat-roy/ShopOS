SERVICE   := affiliate-service
IMAGE     := shopos/$(SERVICE)
TAG       ?= latest
GRPC_PORT ?= 50200

.PHONY: all build run test lint docker-build docker-run tidy

all: build

## build: compile the binary to ./bin/affiliate-service
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
	  -e GRPC_PORT=$(GRPC_PORT) \
	  -e DATABASE_URL="$(DATABASE_URL)" \
	  -p $(GRPC_PORT):$(GRPC_PORT) \
	  --env-file .env.example \
	  $(IMAGE):$(TAG)
