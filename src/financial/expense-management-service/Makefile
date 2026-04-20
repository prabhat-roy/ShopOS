SERVICE      := expense-management-service
MODULE       := github.com/shopos/expense-management-service
IMAGE        := shopos/$(SERVICE)
VERSION      ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
HTTP_PORT    := 8305
GRPC_PORT    := 50114

.PHONY: all build test lint fmt docker-build docker-push migrate-up migrate-down clean run

all: build

## build: compile the binary
build:
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/$(SERVICE) ./main.go

## run: run the service locally (requires .env or exported env vars)
run: build
	./bin/$(SERVICE)

## test: run all unit tests
test:
	go test -v -race -count=1 ./...

## lint: run golangci-lint
lint:
	golangci-lint run ./...

## fmt: format all Go source files
fmt:
	gofmt -w .

## docker-build: build the Docker image
docker-build:
	docker build -t $(IMAGE):$(VERSION) -t $(IMAGE):latest .

## docker-push: push the Docker image to the registry
docker-push: docker-build
	docker push $(IMAGE):$(VERSION)
	docker push $(IMAGE):latest

## migrate-up: apply database migrations (requires golang-migrate CLI)
migrate-up:
	migrate -path db/migrations -database "$(DATABASE_URL)" up

## migrate-down: rollback last database migration
migrate-down:
	migrate -path db/migrations -database "$(DATABASE_URL)" down 1

## clean: remove build artifacts
clean:
	rm -rf bin/

## help: display this help
help:
	@grep -E '^## ' Makefile | sed 's/## //'
