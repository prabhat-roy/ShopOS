SERVICE   := logistics-provider-integration
IMAGE     := shopos/$(SERVICE)
TAG       ?= latest
HTTP_PORT ?= 8905
GRPC_PORT ?= 50174

.PHONY: build run test lint docker-build docker-run clean tidy

## build: compile the binary
build:
	go build -ldflags="-s -w" -o bin/$(SERVICE) ./main.go

## run: run locally using .env.example values
run:
	export $$(grep -v '^#' .env.example | xargs) && go run ./main.go

## test: run all unit tests with race detector
test:
	go test -race -count=1 ./...

## lint: run go vet
lint:
	go vet ./...

## tidy: tidy and vendor modules
tidy:
	go mod tidy

## docker-build: build Docker image
docker-build:
	docker build -t $(IMAGE):$(TAG) .

## docker-run: run Docker image with env vars
docker-run:
	docker run --rm \
		-p $(HTTP_PORT):$(HTTP_PORT) \
		-p $(GRPC_PORT):$(GRPC_PORT) \
		--env-file .env.example \
		$(IMAGE):$(TAG)

## clean: remove build artefacts
clean:
	rm -rf bin/
