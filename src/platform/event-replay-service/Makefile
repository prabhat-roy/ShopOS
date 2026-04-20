SERVICE   := event-replay-service
IMAGE     := shopos/$(SERVICE)
TAG       ?= latest

.PHONY: build test lint docker-build docker-push migrate-up migrate-down clean

build:
	go build -o bin/$(SERVICE) ./...

test:
	go test ./... -v -race -count=1

lint:
	golangci-lint run ./...

docker-build:
	docker build -t $(IMAGE):$(TAG) .

docker-push:
	docker push $(IMAGE):$(TAG)

migrate-up:
	@echo "Applying migrations..."
	psql "$(DATABASE_URL)" -f db/migrations/001_create_replays.sql

migrate-down:
	@echo "Rolling back migrations..."
	psql "$(DATABASE_URL)" -c "DROP TABLE IF EXISTS replay_jobs;"

clean:
	rm -rf bin/
