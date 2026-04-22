APP_NAME = admin-dashboard
IMAGE    = shopos/admin-dashboard
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
