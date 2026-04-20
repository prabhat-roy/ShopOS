.PHONY: build test lint run docker-build docker-run clean

SERVICE   := geolocation-service
IMAGE     := shopos/$(SERVICE)
HTTP_PORT ?= 8093
GRPC_PORT ?= 50058

## build: compile the service binary
build:
	go build -o bin/$(SERVICE) ./...

## test: run all unit tests with race detector
test:
	go test -race -count=1 ./...

## lint: run go vet
lint:
	go vet ./...

## run: build then execute the service locally
run: build
	HTTP_PORT=$(HTTP_PORT) GRPC_PORT=$(GRPC_PORT) ./bin/$(SERVICE)

## docker-build: build the production Docker image
docker-build:
	docker build -t $(IMAGE):latest .

## docker-run: run the service inside Docker
docker-run: docker-build
	docker run --rm \
		-e HTTP_PORT=$(HTTP_PORT) \
		-e GRPC_PORT=$(GRPC_PORT) \
		-p $(HTTP_PORT):$(HTTP_PORT) \
		$(IMAGE):latest

## clean: remove build artefacts
clean:
	rm -rf bin/
