SERVICE_NAME = seo-service
IMAGE_NAME   = shopos/$(SERVICE_NAME)
IMAGE_TAG   ?= latest

.PHONY: install start dev test lint docker-build docker-run clean

## Install Node.js dependencies
install:
	npm install

## Start the service (production mode)
start:
	node index.js

## Start the service in development mode (hot-reload via nodemon)
dev:
	npm run dev

## Run tests with coverage
test:
	npm test

## Lint the source files
lint:
	npx eslint "src/**/*.js" "index.js" --max-warnings 0

## Build Docker image
docker-build:
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .

## Run the Docker container locally
docker-run:
	docker run --rm \
		--env-file .env.example \
		-p 8118:8118 \
		$(IMAGE_NAME):$(IMAGE_TAG)

## Remove node_modules and coverage artifacts
clean:
	rm -rf node_modules coverage
