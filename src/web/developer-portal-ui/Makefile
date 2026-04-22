APP_NAME = developer-portal-ui
IMAGE    = shopos/developer-portal-ui
TAG     ?= latest

.PHONY: dev build docker-build docker-push

dev:
	npm run dev

build:
	npm run build

docker-build:
	docker build -t $(IMAGE):$(TAG) .

docker-push:
	docker push $(IMAGE):$(TAG)
