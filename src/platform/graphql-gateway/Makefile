.PHONY: build run test lint docker-build docker-run clean

SERVICE     := graphql-gateway
IMAGE       := enterprise/$(SERVICE)
IMAGE_TAG   ?= latest
BINARY      := bin/$(SERVICE)

## Build the service binary
build:
	go build -o $(BINARY) ./...

## Run the service locally (reads .env.example or real env vars)
run: build
	./$(BINARY)

## Run all unit tests with race detector
test:
	go test -race -count=1 ./...

## Run go vet + staticcheck (install staticcheck separately if needed)
lint:
	go vet ./...

## Build the Docker image
docker-build:
	docker build -t $(IMAGE):$(IMAGE_TAG) .

## Run the Docker image with env vars from .env.example
docker-run:
	docker run --rm --env-file .env.example -p 8086:8086 $(IMAGE):$(IMAGE_TAG)

## Remove build artefacts
clean:
	rm -rf bin/
