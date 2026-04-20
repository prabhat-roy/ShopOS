SERVICE   := quote-rfq-service
IMAGE     := shopos/$(SERVICE)
TAG       ?= latest
REGISTRY  ?= ghcr.io

.PHONY: build run test lint docker-build docker-push migrate-up migrate-down tidy

build:
	go build -ldflags="-s -w" -o bin/$(SERVICE) ./main.go

run:
	go run ./main.go

test:
	go test ./... -v -race -count=1

lint:
	golangci-lint run ./...

tidy:
	go mod tidy

docker-build:
	docker build -t $(REGISTRY)/$(IMAGE):$(TAG) .

docker-push:
	docker push $(REGISTRY)/$(IMAGE):$(TAG)

migrate-up:
	migrate -path db/migrations -database "$$DATABASE_URL" up

migrate-down:
	migrate -path db/migrations -database "$$DATABASE_URL" down 1

.DEFAULT_GOAL := build
