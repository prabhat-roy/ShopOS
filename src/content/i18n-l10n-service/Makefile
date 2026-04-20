SERVICE_NAME   := i18n-l10n-service
IMAGE_NAME     := shopos/$(SERVICE_NAME)
IMAGE_TAG      ?= latest
MODULE         := github.com/shopos/i18n-l10n-service
HTTP_PORT      := 8606
GRPC_PORT      := 50146

.PHONY: build test lint run docker-build docker-run tidy migrate

build:
	go build -ldflags="-s -w" -o bin/$(SERVICE_NAME) ./main.go

test:
	go test ./... -v -race -cover

lint:
	go vet ./...

tidy:
	go mod tidy

run:
	go run ./main.go

docker-build:
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .

docker-run:
	docker run --rm \
	  --env-file .env \
	  -p $(HTTP_PORT):$(HTTP_PORT) \
	  -p $(GRPC_PORT):$(GRPC_PORT) \
	  $(IMAGE_NAME):$(IMAGE_TAG)

migrate-up:
	psql "$(DATABASE_URL)" -f db/migrations/001_init.up.sql

migrate-down:
	psql "$(DATABASE_URL)" -f db/migrations/001_init.down.sql
