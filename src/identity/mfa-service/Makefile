SERVICE   := mfa-service
MODULE    := github.com/shopos/mfa-service
IMAGE     := shopos/$(SERVICE)
TAG       ?= latest
HTTP_PORT ?= 8102

.PHONY: all build run test lint fmt vet tidy docker-build docker-run migrate-up migrate-down clean

all: build

## build: compile the binary to ./bin/mfa-service
build:
	@mkdir -p bin
	go build -o bin/$(SERVICE) ./...

## run: start the service locally (requires .env or exported env vars)
run:
	go run ./main.go

## test: run all unit tests with race detector
test:
	go test -race -cover ./...

## lint: run golangci-lint (install separately)
lint:
	golangci-lint run ./...

## fmt: gofmt all source files
fmt:
	gofmt -w .

## vet: run go vet
vet:
	go vet ./...

## tidy: tidy and verify go.mod / go.sum
tidy:
	go mod tidy
	go mod verify

## docker-build: build the Docker image
docker-build:
	docker build -t $(IMAGE):$(TAG) .

## docker-run: run the Docker image (requires DATABASE_URL env var)
docker-run:
	docker run --rm \
	  -e DATABASE_URL="$(DATABASE_URL)" \
	  -e HTTP_PORT=$(HTTP_PORT) \
	  -p $(HTTP_PORT):$(HTTP_PORT) \
	  $(IMAGE):$(TAG)

## migrate-up: apply DB migrations with golang-migrate
migrate-up:
	migrate -path db/migrations -database "$(DATABASE_URL)" up

## migrate-down: roll back the last DB migration
migrate-down:
	migrate -path db/migrations -database "$(DATABASE_URL)" down 1

## clean: remove build artifacts
clean:
	rm -rf bin/
