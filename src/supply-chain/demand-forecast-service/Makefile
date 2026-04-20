SERVICE        := demand-forecast-service
IMAGE          := shopos/$(SERVICE)
TAG            ?= latest
PYTHON         := python3
PORT           ?= 8209

.PHONY: help install run test lint build push clean

help: ## Show this help
	@awk 'BEGIN{FS=":.*##"} /^[a-zA-Z_-]+:.*##/{printf "  \033[36m%-18s\033[0m %s\n",$$1,$$2}' $(MAKEFILE_LIST)

install: ## Install Python dependencies
	$(PYTHON) -m pip install --upgrade pip
	$(PYTHON) -m pip install -r requirements.txt

run: ## Run locally with uvicorn
	$(PYTHON) main.py

test: ## Run unit tests
	$(PYTHON) -m pytest tests/ -v --tb=short

lint: ## Lint with ruff (if installed)
	$(PYTHON) -m ruff check . || true

build: ## Build Docker image
	docker build -t $(IMAGE):$(TAG) .

push: build ## Push Docker image to registry
	docker push $(IMAGE):$(TAG)

clean: ## Remove __pycache__ and .pytest_cache
	find . -type d -name "__pycache__" -exec rm -rf {} + 2>/dev/null || true
	find . -type d -name ".pytest_cache" -exec rm -rf {} + 2>/dev/null || true
	find . -name "*.pyc" -delete 2>/dev/null || true
