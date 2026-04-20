SERVICE   := recently-viewed-service
IMAGE     := shopos/$(SERVICE)
TAG       ?= latest
HTTP_PORT ?= 8404

.PHONY: all build test lint docker-build docker-run clean

all: build

## build: compile the service binary
build:
	go build -trimpath -ldflags="-s -w" -o bin/$(SERVICE) .

## test: run all unit tests with race detector
test:
	go test -race -count=1 -timeout 60s ./...

## lint: run golangci-lint
lint:
	golangci-lint run ./...

## docker-build: build the Docker image
docker-build:
	docker build -t $(IMAGE):$(TAG) .

## docker-run: run the service in Docker
docker-run:
	docker run --rm -p $(HTTP_PORT):$(HTTP_PORT) \
		--env-file .env.example \
		$(IMAGE):$(TAG)

## clean: remove build artifacts
clean:
	rm -rf bin/
