SERVICE   := brand-service
IMAGE     := shopos/$(SERVICE)
TAG       ?= latest
HTTP_PORT ?= 8112
GRPC_PORT ?= 50072

.PHONY: all build test lint docker-build docker-run migrate clean

all: build

## build: compile the binary for the current platform
build:
	go build -ldflags="-s -w" -o bin/$(SERVICE) .

## test: run all unit tests with race detector
test:
	go test -race -count=1 ./...

## lint: run go vet
lint:
	go vet ./...

## docker-build: build the Docker image
docker-build:
	docker build -t $(IMAGE):$(TAG) .

## docker-run: run the container locally (requires .env)
docker-run:
	docker run --rm \
		--env-file .env.example \
		-p $(HTTP_PORT):$(HTTP_PORT) \
		-p $(GRPC_PORT):$(GRPC_PORT) \
		$(IMAGE):$(TAG)

## migrate: apply DB migrations with golang-migrate
migrate:
	migrate -path ./migrations \
		-database "$(DATABASE_URL)" \
		up

## clean: remove compiled artifacts
clean:
	rm -rf bin/
