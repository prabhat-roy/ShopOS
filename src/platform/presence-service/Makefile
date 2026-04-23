.PHONY: build test run docker-build docker-run deps

SERVICE_NAME=presence-service
IMAGE_NAME=enterprise/$(SERVICE_NAME)
VERSION?=latest

deps:
	mix deps.get

build: deps
	MIX_ENV=prod mix release

test: deps
	mix test

run: deps
	mix run --no-halt

docker-build:
	docker build -t $(IMAGE_NAME):$(VERSION) .

docker-run:
	docker run --rm -p 8212:8212 -p 50350:50350 --env-file .env $(IMAGE_NAME):$(VERSION)
