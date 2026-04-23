.PHONY: build test run docker-build docker-run

SERVICE_NAME=content-webhook-service
IMAGE_NAME=enterprise/$(SERVICE_NAME)
VERSION?=latest

build:
	shards install
	crystal build src/main.cr --release -o bin/$(SERVICE_NAME)

test:
	crystal spec

run:
	crystal run src/main.cr

docker-build:
	docker build -t $(IMAGE_NAME):$(VERSION) .

docker-run:
	docker run --rm -p 8218:8218 --env-file .env $(IMAGE_NAME):$(VERSION)
