.PHONY: build test run docker-build docker-run

SERVICE_NAME=lease-service
IMAGE_NAME=enterprise/$(SERVICE_NAME)
VERSION?=latest

build:
	gradle bootJar --no-daemon -x test

test:
	gradle test --no-daemon

run:
	gradle bootRun --no-daemon

docker-build:
	docker build -t $(IMAGE_NAME):$(VERSION) .

docker-run:
	docker run --rm -p 50321:50321 --env-file .env $(IMAGE_NAME):$(VERSION)
