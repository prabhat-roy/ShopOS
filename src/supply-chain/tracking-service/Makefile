.PHONY: install test start docker-build docker-run clean

SERVICE_NAME := tracking-service
IMAGE_TAG    := $(SERVICE_NAME):latest
PORT         := 8204

install:
	npm install

test:
	npm test

start:
	npm start

dev:
	npm run dev

docker-build:
	docker build -t $(IMAGE_TAG) .

docker-run:
	docker run --rm -p $(PORT):$(PORT) \
	  --env-file .env.example \
	  $(IMAGE_TAG)

clean:
	rm -rf node_modules
