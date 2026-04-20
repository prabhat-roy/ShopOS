SERVICE   := bnpl-service
IMAGE     := shopos/$(SERVICE)
TAG       ?= latest
GRPC_PORT ?= 50185

.PHONY: all build run test lint docker-build docker-run migrate-up migrate-down tidy

all: build

## build: compile the binary
build:
	go build -trimpath -ldflags="-s -w" -o bin/$(SERVICE) ./...

## run: run the service locally (requires .env to be sourced)
run:
	go run ./...

## test: run all tests
test:
	go test ./... -v -race -count=1

## lint: run go vet
lint:
	go vet ./...

## docker-build: build the container image
docker-build:
	docker build -t $(IMAGE):$(TAG) .

## docker-run: run the container locally
docker-run: docker-build
	docker run --rm \
		-p $(GRPC_PORT):$(GRPC_PORT) \
		--env-file .env \
		$(IMAGE):$(TAG)

## migrate-up: apply database migrations (requires golang-migrate)
migrate-up:
	migrate -path migrations -database "$(DATABASE_URL)" up

## migrate-down: rollback last migration
migrate-down:
	migrate -path migrations -database "$(DATABASE_URL)" down 1

## tidy: tidy go modules
tidy:
	go mod tidy
