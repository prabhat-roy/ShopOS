SERVICE   := address-validation-service
IMAGE     := shopos/$(SERVICE)
TAG       ?= latest

.PHONY: run build test lint docker-build docker-run

run:
	go run .

build:
	go build -o bin/$(SERVICE) .

test:
	go test ./... -v

lint:
	go vet ./...

docker-build:
	docker build -t $(IMAGE):$(TAG) .

docker-run:
	docker run --rm \
	  --env-file .env.example \
	  -p 8145:8145 \
	  -p 50095:50095 \
	  $(IMAGE):$(TAG)
