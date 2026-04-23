.PHONY: build test run docker-build docker-run

SERVICE_NAME=ios-push-gateway-service
IMAGE_NAME=enterprise/$(SERVICE_NAME)
VERSION?=latest

build:
	swift build -c release

test:
	swift test

run:
	swift run

docker-build:
	docker build -t $(IMAGE_NAME):$(VERSION) .

docker-run:
	docker run --rm -p 8214:8214 --env-file .env $(IMAGE_NAME):$(VERSION)
