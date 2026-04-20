SERVICE   := search-service
IMAGE     := shopos/$(SERVICE)
TAG       := latest
PORT      := 8121

.PHONY: install run test lint docker-build docker-run clean help

## install       Install Python dependencies into a local venv
install:
	python -m venv .venv
	.venv/bin/pip install --upgrade pip
	.venv/bin/pip install -r requirements.txt

## run           Start the service locally (requires Elasticsearch running)
run:
	.venv/bin/uvicorn main:app --host 0.0.0.0 --port $(PORT) --reload

## test          Run the unit-test suite
test:
	.venv/bin/pytest tests/ -v --tb=short

## lint          Run ruff linter (install separately: pip install ruff)
lint:
	.venv/bin/ruff check . || true

## docker-build  Build the Docker image
docker-build:
	docker build -t $(IMAGE):$(TAG) .

## docker-run    Run the Docker image locally
docker-run:
	docker run --rm \
		-p $(PORT):$(PORT) \
		--env-file .env.example \
		--name $(SERVICE) \
		$(IMAGE):$(TAG)

## clean         Remove __pycache__ and .venv
clean:
	find . -type d -name __pycache__ -exec rm -rf {} + 2>/dev/null || true
	rm -rf .venv .pytest_cache

## help          Show this help message
help:
	@grep -E '^##' Makefile | sed 's/## //'
