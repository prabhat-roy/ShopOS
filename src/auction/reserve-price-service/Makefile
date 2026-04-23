.PHONY: build test run docker-build docker-run

SERVICE_NAME=reserve-price-service
IMAGE_NAME=enterprise/$(SERVICE_NAME)
VERSION?=latest

build:
	go build -o bin/$(SERVICE_NAME) .

test:
	go test ./...

run:
	go run .

docker-build:
	docker build -t $(IMAGE_NAME):$(VERSION) .

docker-run:
	docker run --rm -p 50312:50312 --env-file .env $(IMAGE_NAME):$(VERSION)
