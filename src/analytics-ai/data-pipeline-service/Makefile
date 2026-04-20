.PHONY: install test run docker-build docker-run lint clean

SERVICE_NAME := data-pipeline-service
IMAGE_NAME   := shopos/$(SERVICE_NAME)
IMAGE_TAG    := latest

install:
	pip install -r requirements.txt

test:
	pytest tests/ -v --asyncio-mode=auto

run:
	python main.py

docker-build:
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .

docker-run:
	docker run --rm \
		--env-file .env \
		-p 8707:8707 \
		--name $(SERVICE_NAME) \
		$(IMAGE_NAME):$(IMAGE_TAG)

lint:
	flake8 pipeline/ main.py --max-line-length=120

clean:
	find . -type d -name __pycache__ -exec rm -rf {} + 2>/dev/null || true
	find . -name "*.pyc" -delete 2>/dev/null || true