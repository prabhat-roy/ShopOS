SERVICE_NAME  := analytics-service
IMAGE_NAME    := shopos/$(SERVICE_NAME)
IMAGE_TAG     ?= latest
PORT          := 8700

.PHONY: help install test lint run docker-build docker-run clean

help:
	@echo "Usage: make <target>"
	@echo ""
	@echo "  install      Install Python dependencies"
	@echo "  test         Run test suite"
	@echo "  lint         Run ruff linter"
	@echo "  run          Start the service locally"
	@echo "  docker-build Build Docker image"
	@echo "  docker-run   Run Docker container"
	@echo "  clean        Remove __pycache__ and .pytest_cache"

install:
	pip install -r requirements.txt

test:
	STORE_BACKEND=memory pytest tests/ -v --asyncio-mode=auto

lint:
	ruff check .

run:
	STORE_BACKEND=memory uvicorn main:app --host 0.0.0.0 --port $(PORT) --reload

docker-build:
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .

docker-run:
	docker run --rm \
		--env-file .env \
		-p $(PORT):$(PORT) \
		$(IMAGE_NAME):$(IMAGE_TAG)

clean:
	find . -type d -name "__pycache__" -exec rm -rf {} + 2>/dev/null || true
	find . -type d -name ".pytest_cache" -exec rm -rf {} + 2>/dev/null || true
	find . -name "*.pyc" -delete 2>/dev/null || true
