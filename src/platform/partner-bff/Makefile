.PHONY: build test tidy lint docker-build run

SERVICE = partner-bff
IMAGE   = shopos/$(SERVICE):latest

build:
	go build -o bin/$(SERVICE) .

test:
	go test ./... -race -count=1

tidy:
	go mod tidy

lint:
	golangci-lint run ./...

docker-build:
	docker build -t $(IMAGE) .

run:
	go run .
