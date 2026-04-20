SERVICE   = bundle-service
IMAGE     = shopos/$(SERVICE)
TAG      ?= latest

.PHONY: build test lint run docker-build docker-push migrate-up migrate-down tidy

build:
	go build -o bin/$(SERVICE) ./...

test:
	go test ./... -v -race -count=1

lint:
	golangci-lint run ./...

run:
	go run main.go

tidy:
	go mod tidy

docker-build:
	docker build -t $(IMAGE):$(TAG) .

docker-push:
	docker push $(IMAGE):$(TAG)

migrate-up:
	migrate -path ./migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path ./migrations -database "$(DATABASE_URL)" down
