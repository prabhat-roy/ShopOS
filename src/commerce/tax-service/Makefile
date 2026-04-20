SERVICE   := tax-service
IMAGE     := shopos/$(SERVICE)
TAG       ?= latest
REGISTRY  ?= ghcr.io

.PHONY: all build test lint fmt vet docker-build docker-push clean run

all: build

## build: compile the service binary
build:
	go build -trimpath -ldflags="-s -w" -o bin/$(SERVICE) ./main.go

## run: run the service locally
run:
	go run ./main.go

## test: run all tests with race detector
test:
	go test -race -count=1 -timeout 60s ./...

## lint: run golangci-lint (must be installed)
lint:
	golangci-lint run ./...

## fmt: format all Go source
fmt:
	gofmt -w .

## vet: run go vet
vet:
	go vet ./...

## docker-build: build the Docker image
docker-build:
	docker build -t $(REGISTRY)/$(IMAGE):$(TAG) .

## docker-push: push the Docker image
docker-push: docker-build
	docker push $(REGISTRY)/$(IMAGE):$(TAG)

## clean: remove build artefacts
clean:
	rm -rf bin/
