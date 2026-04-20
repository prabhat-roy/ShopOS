SERVICE_NAME := product-import-service
IMAGE_NAME   := shopos/$(SERVICE_NAME)
IMAGE_TAG    ?= latest

.PHONY: build test run docker-build docker-run clean tidy

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
		-p 8120:8120 \
		-p 50180:50180 \
		$(IMAGE_NAME):$(IMAGE_TAG)

tidy:
	go mod tidy

clean:
	rm -rf bin/
