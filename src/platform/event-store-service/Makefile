.PHONY: build test tidy lint docker-build run migrate

SERVICE = event-store-service
IMAGE   = shopos/$(SERVICE):latest

build:
	go build -o bin/$(SERVICE) .

test:
	go test ./... -race -count=1

tidy:
	go mod tidy

lint:
	golangci-lint run ./...

docker-build:
	docker build -t $(IMAGE) .

run:
	go run .

migrate:
	psql "$(DATABASE_URL)" -f db/migrations/001_create_events.sql
