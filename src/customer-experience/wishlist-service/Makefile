SERVICE   := wishlist-service
IMAGE     := shopos/$(SERVICE)
TAG       ?= latest
HTTP_PORT ?= 8402

.PHONY: all build test lint docker-build docker-run migrate-up migrate-down clean

all: build

## build: compile the service binary
build:
	go build -trimpath -ldflags="-s -w" -o bin/$(SERVICE) .

## test: run all unit tests with race detector
test:
	go test -race -count=1 -timeout 60s ./...

## lint: run golangci-lint (requires golangci-lint in PATH)
lint:
	golangci-lint run ./...

## docker-build: build the Docker image
docker-build:
	docker build -t $(IMAGE):$(TAG) .

## docker-run: run the service in Docker (requires DATABASE_URL env var)
docker-run:
	docker run --rm -p $(HTTP_PORT):$(HTTP_PORT) \
		--env-file .env.example \
		$(IMAGE):$(TAG)

## migrate-up: apply all up migrations (requires migrate CLI)
migrate-up:
	migrate -path db/migrations -database "$(DATABASE_URL)" up

## migrate-down: roll back the last migration
migrate-down:
	migrate -path db/migrations -database "$(DATABASE_URL)" down 1

## clean: remove build artifacts
clean:
	rm -rf bin/
