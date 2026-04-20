SERVICE     := marketplace-connector-service
IMAGE       := shopos/$(SERVICE)
TAG         := latest
BINARY      := bin/$(SERVICE)
MODULE      := github.com/shopos/$(SERVICE)

.PHONY: all build run test lint docker-build docker-run clean tidy

all: build

## build: compile the binary locally
build:
	@mkdir -p bin
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o $(BINARY) ./main.go
	@echo "built $(BINARY)"

## run: run the service with default env
run: build
	HTTP_PORT=8901 ./$(BINARY)

## test: run all tests
test:
	go test ./... -v -count=1 -timeout 60s

## lint: run golangci-lint
lint:
	golangci-lint run ./...

## tidy: tidy go modules
tidy:
	go mod tidy

## docker-build: build Docker image
docker-build:
	docker build -t $(IMAGE):$(TAG) .

## docker-run: run Docker container locally
docker-run:
	docker run --rm \
		-p 8901:8901 \
		-e HTTP_PORT=8901 \
		-e KAFKA_BROKERS=host.docker.internal:9092 \
		-e KAFKA_TOPIC=marketplace.sync.completed \
		$(IMAGE):$(TAG)

## clean: remove build artefacts
clean:
	rm -rf bin/
