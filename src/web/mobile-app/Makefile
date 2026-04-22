APP_NAME = mobile-app
IMAGE    = shopos/mobile-app
TAG     ?= latest

.PHONY: start android ios docker-build docker-push

start:
	npm start

android:
	npm run android

ios:
	npm run ios

docker-build:
	docker build -t $(IMAGE):$(TAG) .

docker-push:
	docker push $(IMAGE):$(TAG)
