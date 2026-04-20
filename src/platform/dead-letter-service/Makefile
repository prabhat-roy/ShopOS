SERVICE   := dead-letter-service
IMAGE     := shopos/$(SERVICE)
TAG       ?= latest

.PHONY: build test lint docker-build docker-push migrate-up migrate-down run tidy

## build: compile the service binary
build:
	go build -o bin/$(SERVICE) ./main.go

## test: run all unit tests
test:
	go test ./... -race -timeout 60s

## lint: run golangci-lint
lint:
	golangci-lint run ./...

## tidy: tidy go.mod and go.sum
tidy:
	go mod tidy

## docker-build: build the Docker image
docker-build:
	docker build -t $(IMAGE):$(TAG) .

## docker-push: push the Docker image to the registry
docker-push:
	docker push $(IMAGE):$(TAG)

## migrate-up: apply all up migrations (requires DATABASE_URL)
migrate-up:
	migrate -path db/migrations -database "$(DATABASE_URL)" up

## migrate-down: roll back the last migration
migrate-down:
	migrate -path db/migrations -database "$(DATABASE_URL)" down 1

## run: run the service locally using .env.example defaults
run:
	@export $$(grep -v '^#' .env.example | xargs) && go run ./main.go
