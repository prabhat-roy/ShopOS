SERVICE     := device-fingerprint-service
MODULE      := github.com/shopos/device-fingerprint-service
IMAGE       := shopos/$(SERVICE)
TAG         ?= latest
HTTP_PORT   ?= 8105

.PHONY: all build run test lint fmt vet tidy docker-build docker-run clean

all: build

## build: compile the binary
build:
	go build -o bin/$(SERVICE) ./...

## run: run the service locally (requires Redis)
run:
	go run ./main.go

## test: run all unit tests with race detector
test:
	go test -race -count=1 -timeout 60s ./...

## lint: run golangci-lint (must be installed)
lint:
	golangci-lint run ./...

## fmt: format all Go source files
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

## docker-run: run the Docker image locally
docker-run:
	docker run --rm \
		-p $(HTTP_PORT):$(HTTP_PORT) \
		--env-file .env.example \
		--name $(SERVICE) \
		$(IMAGE):$(TAG)

## clean: remove build artefacts
clean:
	rm -rf bin/
