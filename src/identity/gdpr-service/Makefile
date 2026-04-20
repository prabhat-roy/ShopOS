SERVICE     := gdpr-service
IMAGE       := shopos/$(SERVICE)
TAG         ?= latest
GO          := go
GOFLAGS     ?=

.PHONY: all build run test lint fmt vet tidy docker-build docker-run migrate-up migrate-down clean

all: build

## build: compile the service binary
build:
	$(GO) build $(GOFLAGS) -o bin/$(SERVICE) ./...

## run: run the service locally (requires .env or exported env vars)
run:
	$(GO) run ./main.go

## test: run all unit tests
test:
	$(GO) test ./... -v -count=1

## lint: run golangci-lint
lint:
	golangci-lint run ./...

## fmt: format all Go source files
fmt:
	$(GO) fmt ./...

## vet: run go vet
vet:
	$(GO) vet ./...

## tidy: tidy and vendor go modules
tidy:
	$(GO) mod tidy

## docker-build: build the Docker image
docker-build:
	docker build -t $(IMAGE):$(TAG) .

## docker-run: run the Docker image (reads from .env)
docker-run:
	docker run --env-file .env -p 8103:8103 -p 50065:50065 $(IMAGE):$(TAG)

## migrate-up: apply all pending migrations (requires DATABASE_URL)
migrate-up:
	migrate -path db/migrations -database "$(DATABASE_URL)" up

## migrate-down: roll back the last migration (requires DATABASE_URL)
migrate-down:
	migrate -path db/migrations -database "$(DATABASE_URL)" down 1

## clean: remove build artefacts
clean:
	rm -rf bin/