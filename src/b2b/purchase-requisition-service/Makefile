SERVICE_NAME := purchase-requisition-service
IMAGE        := shopos/$(SERVICE_NAME):latest

.PHONY: build test docker-build docker-push

build:
	gradle bootJar --no-daemon

test:
	gradle test --no-daemon

docker-build:
	docker build -t $(IMAGE) .

docker-push:
	docker push $(IMAGE)
