.PHONY: build test run docker-build docker-run

SERVICE_NAME=event-transform-service
IMAGE_NAME=enterprise/$(SERVICE_NAME)
VERSION?=latest

build:
	lein uberjar

test:
	lein test

run:
	lein run

docker-build:
	docker build -t $(IMAGE_NAME):$(VERSION) .

docker-run:
	docker run --rm -p 50354:50354 --env-file .env $(IMAGE_NAME):$(VERSION)
