SERVICE_NAME := accessibility-service
IMAGE        := shopos/$(SERVICE_NAME):latest

.PHONY: install start test docker-build docker-push

install:
	npm install

start:
	node index.js

test:
	npm test

docker-build:
	docker build -t $(IMAGE) .

docker-push:
	docker push $(IMAGE)
