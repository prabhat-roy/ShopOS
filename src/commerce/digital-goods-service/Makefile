SERVICE   := digital-goods-service
IMAGE     := shopos/$(SERVICE)
TAG       ?= latest

.PHONY: run build test lint docker-build docker-run tidy

run:
	go run .

build:
	go build -o bin/$(SERVICE) .

test:
	go test ./... -v

lint:
	go vet ./...

tidy:
	go mod tidy

docker-build:
	docker build -t $(IMAGE):$(TAG) .

docker-run:
	docker run --rm \
	  --env-file .env.example \
	  -p 8147:8147 \
	  -p 50096:50096 \
	  $(IMAGE):$(TAG)
