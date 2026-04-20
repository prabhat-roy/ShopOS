SERVICE     := crm-integration-service
IMAGE       := shopos/$(SERVICE)
TAG         := latest
BINARY      := bin/$(SERVICE)

.PHONY: all build run test lint docker-build docker-run clean tidy

all: build

## build: compile the binary locally
build:
	@mkdir -p bin
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o $(BINARY) ./main.go
	@echo "built $(BINARY)"

## run: run the service with default env
run: build
	HTTP_PORT=8903 GRPC_PORT=50172 ./$(BINARY)

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
		-p 8903:8903 \
		-p 50172:50172 \
		-e HTTP_PORT=8903 \
		-e GRPC_PORT=50172 \
		$(IMAGE):$(TAG)

## clean: remove build artefacts
clean:
	rm -rf bin/
