SERVICE   := api-key-service
IMAGE     := shopos/$(SERVICE)
TAG       ?= latest

.PHONY: build run test lint docker-build docker-run migrate-up migrate-down tidy

## build: compile the binary
build:
	go build -o bin/$(SERVICE) ./...

## run: run the service locally (requires .env variables exported)
run:
	go run ./main.go

## test: run all unit tests
test:
	go test ./... -v -count=1

## lint: run golangci-lint
lint:
	golangci-lint run ./...

## tidy: tidy go modules
tidy:
	go mod tidy

## docker-build: build the Docker image
docker-build:
	docker build -t $(IMAGE):$(TAG) .

## docker-run: run the containerised service
docker-run:
	docker run --rm \
		--env-file .env.example \
		-p 8104:8104 \
		-p 50066:50066 \
		$(IMAGE):$(TAG)

## migrate-up: apply database migrations (uses golang-migrate)
migrate-up:
	migrate -path db/migrations -database "$(DATABASE_URL)" up

## migrate-down: roll back all database migrations
migrate-down:
	migrate -path db/migrations -database "$(DATABASE_URL)" down

## help: list available targets
help:
	@grep -E '^##' Makefile | sed 's/## //'
