SERVICE_NAME   := cms-service
IMAGE_NAME     := shopos/$(SERVICE_NAME)
IMAGE_TAG      ?= latest
CONTAINER_PORT := 8603

.PHONY: install build test lint run docker-build docker-run

install:
	npm ci

test:
	npm test

lint:
	npx eslint src/ index.js --ext .js

run:
	npm start

dev:
	npm run dev

docker-build:
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .

docker-run:
	docker run --rm \
	  --env-file .env \
	  -p $(CONTAINER_PORT):$(CONTAINER_PORT) \
	  $(IMAGE_NAME):$(IMAGE_TAG)
