SERVICE   = tenant-service
IMAGE     = shopos/$(SERVICE)
TAG      ?= latest

.PHONY: build test lint run docker-build docker-push migrate-up migrate-down tidy

## Build the binary
build:
	go build -trimpath -ldflags="-s -w" -o bin/$(SERVICE) .

## Run all tests
test:
	go test ./... -v -race -count=1

## Lint (requires golangci-lint in PATH)
lint:
	golangci-lint run ./...

## Run locally (requires DATABASE_URL in environment)
run:
	go run .

## Build Docker image
docker-build:
	docker build -t $(IMAGE):$(TAG) .

## Push Docker image
docker-push:
	docker push $(IMAGE):$(TAG)

## Apply migrations (requires golang-migrate and DATABASE_URL)
migrate-up:
	migrate -path db/migrations -database "$(DATABASE_URL)" up

## Roll back migrations
migrate-down:
	migrate -path db/migrations -database "$(DATABASE_URL)" down

## Tidy go.mod / go.sum
tidy:
	go mod tidy
