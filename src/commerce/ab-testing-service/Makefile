SERVICE   := ab-testing-service
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
	  -p 8144:8144 \
	  -p 50093:50093 \
	  $(IMAGE):$(TAG)
