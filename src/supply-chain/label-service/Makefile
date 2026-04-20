.PHONY: install test run docker-build docker-run clean

SERVICE_NAME := label-service
IMAGE_TAG    := $(SERVICE_NAME):latest
PORT         := 8205

install:
	pip install -r requirements.txt

test:
	pytest tests/ -v --tb=short

run:
	python main.py

docker-build:
	docker build -t $(IMAGE_TAG) .

docker-run:
	docker run --rm -p $(PORT):$(PORT) \
	  --env-file .env.example \
	  $(IMAGE_TAG)

clean:
	find . -type d -name __pycache__ -exec rm -rf {} + 2>/dev/null; true
	find . -type d -name .pytest_cache -exec rm -rf {} + 2>/dev/null; true
	find . -name "*.pyc" -delete 2>/dev/null; true
