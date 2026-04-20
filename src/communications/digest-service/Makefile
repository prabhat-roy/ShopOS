SERVICE      := digest-service
IMAGE_REPO   := ghcr.io/shopos
IMAGE_TAG    ?= latest
IMAGE        := $(IMAGE_REPO)/$(SERVICE):$(IMAGE_TAG)

.PHONY: build test lint docker-build docker-push run migrate-up migrate-down clean tidy

## build: compile the service binary
build:
	go build -ldflags="-s -w" -o bin/$(SERVICE) ./main.go

## test: run all unit tests
test:
	go test ./... -v -race -count=1

## lint: run golangci-lint
lint:
	golangci-lint run ./...

## tidy: tidy go modules
tidy:
	go mod tidy

## run: run the service locally (requires Postgres + Kafka)
run:
	go run ./main.go

## docker-build: build the Docker image
docker-build:
	docker build -t $(IMAGE) .

## docker-push: push the Docker image
docker-push:
	docker push $(IMAGE)

## migrate-up: apply database migrations (requires migrate CLI)
migrate-up:
	migrate -path db/migrations -database "$$DATABASE_URL" up

## migrate-down: rollback database migrations
migrate-down:
	migrate -path db/migrations -database "$$DATABASE_URL" down

## clean: remove build artefacts
clean:
	rm -rf bin/
