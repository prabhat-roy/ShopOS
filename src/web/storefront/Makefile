APP_NAME = storefront
IMAGE    = shopos/storefront
TAG     ?= latest

.PHONY: dev build start docker-build docker-push

dev:
	npm run dev

build:
	npm run build

start:
	npm start

docker-build:
	docker build -t $(IMAGE):$(TAG) .

docker-push:
	docker push $(IMAGE):$(TAG)
