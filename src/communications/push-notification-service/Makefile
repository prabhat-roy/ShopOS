SERVICE   := push-notification-service
IMAGE     := shopos/$(SERVICE)
TAG       := $(shell git rev-parse --short HEAD 2>/dev/null || echo "dev")
PORT      := 8504
MODULE    := github.com/shopos/push-notification-service

.PHONY: help build run test lint vet fmt docker-build docker-run clean

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Compile the service binary
	go build -ldflags="-s -w" -o bin/$(SERVICE) .

run: build ## Build and run locally (requires Kafka)
	./bin/$(SERVICE)

test: ## Run all tests
	go test ./... -v -race -timeout 30s

lint: ## Run staticcheck
	staticcheck ./...

vet: ## Run go vet
	go vet ./...

fmt: ## Format Go source files
	gofmt -w .

tidy: ## Tidy go.mod / go.sum
	go mod tidy

docker-build: ## Build Docker image
	docker build -t $(IMAGE):$(TAG) -t $(IMAGE):latest .

docker-run: ## Run container locally
	docker run --rm \
		--name $(SERVICE) \
		-p $(PORT):$(PORT) \
		--env-file .env.example \
		$(IMAGE):latest

clean: ## Remove build artefacts
	rm -rf bin/
