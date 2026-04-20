SERVICE   := credit-service
IMAGE     := shopos/$(SERVICE)
TAG       ?= latest
GOFLAGS   := CGO_ENABLED=0

.PHONY: all build test lint vet fmt docker-build docker-push migrate-up migrate-down clean

all: build

## build: compile the binary to ./bin/credit-service
build:
	$(GOFLAGS) go build -ldflags="-s -w" -trimpath -o bin/$(SERVICE) .

## test: run all unit tests with race detector
test:
	go test -race -cover ./...

## lint: run golangci-lint (must be installed)
lint:
	golangci-lint run ./...

## vet: run go vet
vet:
	go vet ./...

## fmt: format all Go source files
fmt:
	gofmt -w .

## docker-build: build the Docker image
docker-build:
	docker build -t $(IMAGE):$(TAG) .

## docker-push: push the Docker image to the registry
docker-push:
	docker push $(IMAGE):$(TAG)

## migrate-up: apply all database migrations using golang-migrate
migrate-up:
	migrate -path db/migrations -database "$$DATABASE_URL" up

## migrate-down: roll back the last database migration
migrate-down:
	migrate -path db/migrations -database "$$DATABASE_URL" down 1

## clean: remove build artefacts
clean:
	rm -rf bin/
