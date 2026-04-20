SERVICE_NAME = live-chat-service
IMAGE_NAME   = shopos/$(SERVICE_NAME)
IMAGE_TAG   ?= latest

.PHONY: build test run docker-build docker-run tidy clean help

build:
	go build -o bin/$(SERVICE_NAME) .

test:
	go test ./... -v -race -count=1

run:
	go run .

tidy:
	go mod tidy

docker-build:
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .

docker-run:
	docker run --rm \
		--env-file .env \
		-p 8406:8406 \
		$(IMAGE_NAME):$(IMAGE_TAG)

clean:
	rm -rf bin/

help:
	@echo "Targets: build | test | run | tidy | docker-build | docker-run | clean"
