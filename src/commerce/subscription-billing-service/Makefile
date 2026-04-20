SERVICE_NAME := subscription-billing-service
IMAGE_NAME   := shopos/$(SERVICE_NAME)
IMAGE_TAG    ?= latest

.PHONY: build test run docker-build docker-run clean

build:
	go build -ldflags="-s -w" -o $(SERVICE_NAME) .

test:
	go test ./...

run:
	go run .

docker-build:
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .

docker-run:
	docker run --rm \
		--env-file .env \
		-p 8135:8135 \
		-p 50090:50090 \
		$(IMAGE_NAME):$(IMAGE_TAG)

clean:
	rm -f $(SERVICE_NAME)
