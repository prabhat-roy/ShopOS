APP_NAME = partner-portal
IMAGE    = shopos/partner-portal
TAG     ?= latest

.PHONY: start build docker-build docker-push

start:
	npm start

build:
	npm run build

docker-build:
	docker build -t $(IMAGE):$(TAG) .

docker-push:
	docker push $(IMAGE):$(TAG)
