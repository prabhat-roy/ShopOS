SERVICE   := email-service
IMAGE     := shopos/$(SERVICE)
TAG       := $(shell git rev-parse --short HEAD 2>/dev/null || echo "dev")
PORT      := 8503

.PHONY: help build run test lint fmt docker-build docker-run clean

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Install Python dependencies
	pip install -r requirements.txt

run: ## Run the service locally (requires Kafka)
	python main.py

test: ## Run pytest suite
	PYTHONPATH=. pytest tests/ -v --tb=short

lint: ## Run ruff linter
	ruff check .

fmt: ## Auto-format with ruff
	ruff format .

docker-build: ## Build Docker image
	docker build -t $(IMAGE):$(TAG) -t $(IMAGE):latest .

docker-run: ## Run container locally
	docker run --rm \
		--name $(SERVICE) \
		-p $(PORT):$(PORT) \
		--env-file .env.example \
		$(IMAGE):latest

clean: ## Remove .pyc / __pycache__
	find . -type f -name "*.pyc" -delete
	find . -type d -name "__pycache__" -exec rm -rf {} +
	find . -type d -name ".pytest_cache" -exec rm -rf {} +
