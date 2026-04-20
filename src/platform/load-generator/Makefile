.PHONY: run test lint build docker-build docker-run

SERVICE_NAME := load-generator
IMAGE_NAME   := shopos/$(SERVICE_NAME):latest
PORT         ?= 8089

run:
	uvicorn main:app --host 0.0.0.0 --port $(PORT) --reload

test:
	pytest tests/ -v

lint:
	ruff check . || true

build:
	pip install -r requirements.txt

docker-build:
	docker build -t $(IMAGE_NAME) .

docker-run:
	docker run --rm -p $(PORT):$(PORT) --env-file .env $(IMAGE_NAME)
