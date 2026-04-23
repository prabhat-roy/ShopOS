.PHONY: build test run docker-build docker-run

SERVICE_NAME=financial-rules-engine
IMAGE_NAME=enterprise/$(SERVICE_NAME)
VERSION?=latest

build:
	stack build

test:
	stack test

run:
	stack run

docker-build:
	docker build -t $(IMAGE_NAME):$(VERSION) .

docker-run:
	docker run --rm -p 50360:50360 --env-file .env $(IMAGE_NAME):$(VERSION)
