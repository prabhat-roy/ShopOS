SERVICE   := carrier-integration-service
MODULE    := github.com/shopos/carrier-integration-service
IMAGE     := shopos/$(SERVICE)
TAG       ?= latest
HTTP_PORT ?= 8206
GRPC_PORT ?= 50106

.PHONY: all build test lint run docker-build docker-run clean tidy

all: build

## build: compile the binary
build:
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/$(SERVICE) ./main.go

## test: run all unit tests
test:
	go test -v -race -count=1 ./...

## lint: run go vet
lint:
	go vet ./...

## tidy: tidy module dependencies
tidy:
	go mod tidy

## run: run the service locally
run:
	HTTP_PORT=$(HTTP_PORT) GRPC_PORT=$(GRPC_PORT) go run ./main.go

## docker-build: build the Docker image
docker-build:
	docker build -t $(IMAGE):$(TAG) .

## docker-run: run the Docker image
docker-run:
	docker run --rm \
		-p $(HTTP_PORT):$(HTTP_PORT) \
		-p $(GRPC_PORT):$(GRPC_PORT) \
		--env-file .env.example \
		$(IMAGE):$(TAG)

## clean: remove build artefacts
clean:
	rm -rf bin/
