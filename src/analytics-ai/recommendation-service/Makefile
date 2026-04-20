SERVICE   := recommendation-service
IMAGE     := shopos/$(SERVICE)
TAG       := latest
PORT      := 8702
GRPC_PORT := 50150

.PHONY: help install lint test test-cov run docker-build docker-run docker-push clean

help:
	@echo "Usage: make <target>"
	@echo ""
	@echo "Targets:"
	@echo "  install      Install Python dependencies"
	@echo "  lint         Run flake8 linter"
	@echo "  test         Run unit tests"
	@echo "  test-cov     Run tests with coverage report"
	@echo "  run          Start the service locally"
	@echo "  docker-build Build Docker image"
	@echo "  docker-run   Run Docker container locally"
	@echo "  docker-push  Push Docker image to registry"
	@echo "  clean        Remove build artefacts"

install:
	pip install --upgrade pip
	pip install -r requirements.txt

lint:
	flake8 . --max-line-length=100 --exclude=__pycache__,.venv

test:
	PYTHONPATH=. pytest tests/ -v

test-cov:
	PYTHONPATH=. pytest tests/ -v --cov=recommender --cov-report=term-missing --cov-report=html

run:
	python main.py

docker-build:
	docker build -t $(IMAGE):$(TAG) .

docker-run:
	docker run --rm \
		--name $(SERVICE) \
		-p $(PORT):$(PORT) \
		-p $(GRPC_PORT):$(GRPC_PORT) \
		--env-file .env.example \
		$(IMAGE):$(TAG)

docker-push:
	docker push $(IMAGE):$(TAG)

clean:
	find . -type d -name __pycache__ -exec rm -rf {} + 2>/dev/null || true
	find . -type f -name "*.pyc" -delete
	rm -rf .pytest_cache htmlcov .coverage
