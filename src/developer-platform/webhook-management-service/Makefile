SERVICE_NAME := webhook-management-service
IMAGE        := shopos/$(SERVICE_NAME):latest

.PHONY: build run test lint docker-build docker-push

build:
	go build -o $(SERVICE_NAME) .

run:
	go run .

test:
	go test ./...

lint:
	golangci-lint run ./...

docker-build:
	docker build -t $(IMAGE) .

docker-push:
	docker push $(IMAGE)
