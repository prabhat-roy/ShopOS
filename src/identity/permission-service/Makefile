SERVICE   := permission-service
MODULE    := github.com/shopos/$(SERVICE)
IMAGE     := shopos/$(SERVICE)
TAG       ?= latest

.PHONY: all build test lint vet fmt docker-build docker-push migrate-up migrate-down clean

all: build

## build: compile the service binary
build:
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/$(SERVICE) ./main.go

## test: run all unit tests
test:
	go test ./... -v -race -count=1

## lint: run golangci-lint (requires golangci-lint installed)
lint:
	golangci-lint run ./...

## vet: run go vet
vet:
	go vet ./...

## fmt: format all Go source files
fmt:
	gofmt -w .

## docker-build: build the container image
docker-build:
	docker build -t $(IMAGE):$(TAG) .

## docker-push: push the container image to the registry
docker-push:
	docker push $(IMAGE):$(TAG)

## migrate-up: apply database migrations (requires DATABASE_URL to be set)
migrate-up:
	@echo "Applying migrations..."
	psql "$(DATABASE_URL)" -f db/migrations/001_create_permissions.sql

## migrate-down: tear down database schema
migrate-down:
	psql "$(DATABASE_URL)" -c "DROP TABLE IF EXISTS user_roles; DROP TABLE IF EXISTS roles;"

## clean: remove build artifacts
clean:
	rm -rf bin/

## run: run the service locally (requires .env.example to be copied to .env and sourced)
run:
	go run ./main.go

## help: print this help message
help:
	@grep -E '^##' Makefile | sed 's/## /  /'
