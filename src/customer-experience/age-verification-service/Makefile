SERVICE_NAME = age-verification-service
IMAGE_NAME   = shopos/$(SERVICE_NAME)
IMAGE_TAG   ?= latest

.PHONY: build test run docker-build docker-run clean help

build:
	go build -o bin/$(SERVICE_NAME) .

test:
	go test ./... -v -race -count=1

run:
	go run .

docker-build:
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .

docker-run:
	docker run --rm \
		--env-file .env \
		-p 8407:8407 \
		$(IMAGE_NAME):$(IMAGE_TAG)

clean:
	rm -rf bin/

help:
	@echo "Targets: build | test | run | docker-build | docker-run | clean"
