SERVICE_NAME=sms-service
IMAGE_NAME=shopos/$(SERVICE_NAME)
IMAGE_TAG?=latest

.PHONY: install dev start test test-coverage lint build push clean

install:
	npm ci

dev:
	npm run dev

start:
	npm run start

test:
	npm test

test-coverage:
	npm run test:coverage

build:
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .

push:
	docker push $(IMAGE_NAME):$(IMAGE_TAG)

clean:
	rm -rf node_modules coverage

logs:
	docker logs -f $(SERVICE_NAME)
