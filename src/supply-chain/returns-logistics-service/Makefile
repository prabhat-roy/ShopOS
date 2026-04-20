SERVICE      := returns-logistics-service
MODULE       := github.com/shopos/returns-logistics-service
IMAGE        := shopos/$(SERVICE)
TAG          ?= latest
HTTP_PORT    ?= 8208
GRPC_PORT    ?= 50108
DATABASE_URL ?= postgres://postgres:postgres@localhost:5432/returns_db?sslmode=disable

.PHONY: all build test lint run docker-build docker-run migrate-up migrate-down clean tidy

all: build

## build: compile the binary
build:
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/$(SERVICE) ./main.go

## test: run all unit tests (no DB required — uses mocks)
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
	HTTP_PORT=$(HTTP_PORT) GRPC_PORT=$(GRPC_PORT) DATABASE_URL='$(DATABASE_URL)' go run ./main.go

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

## migrate-up: apply all pending migrations (requires migrate CLI)
migrate-up:
	migrate -path db/migrations -database '$(DATABASE_URL)' up

## migrate-down: roll back the last migration
migrate-down:
	migrate -path db/migrations -database '$(DATABASE_URL)' down 1

## clean: remove build artefacts
clean:
	rm -rf bin/
