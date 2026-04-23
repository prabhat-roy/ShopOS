.PHONY: build test run docker-build docker-run

SERVICE_NAME=rate-limiter-core
IMAGE_NAME=enterprise/$(SERVICE_NAME)
VERSION?=latest

build:
	zig build -Doptimize=ReleaseSafe

test:
	zig build test

run:
	zig build run

docker-build:
	docker build -t $(IMAGE_NAME):$(VERSION) .

docker-run:
	docker run --rm -p 50355:50355 --env-file .env $(IMAGE_NAME):$(VERSION)
