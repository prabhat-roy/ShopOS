SERVICE   := sitemap-service
IMAGE     := shopos/$(SERVICE)
TAG       ?= latest
HTTP_PORT ?= 8605

.PHONY: all build test lint docker-build docker-run clean tidy

all: build

## build: compile the service binary
build:
	go build -trimpath -ldflags="-s -w" -o bin/$(SERVICE) .

## test: run all tests with race detector
test:
	go test -race -v ./...

## lint: run go vet
lint:
	go vet ./...

## tidy: tidy and verify go modules
tidy:
	go mod tidy
	go mod verify

## docker-build: build the Docker image
docker-build:
	docker build -t $(IMAGE):$(TAG) .

## docker-run: run the service in Docker
docker-run:
	docker run --rm \
		--env-file .env.example \
		-p $(HTTP_PORT):$(HTTP_PORT) \
		$(IMAGE):$(TAG)

## clean: remove build artefacts
clean:
	rm -rf bin/
