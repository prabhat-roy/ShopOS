SERVICE_NAME  := review-rating-service
IMAGE_NAME    := shopos/$(SERVICE_NAME)
IMAGE_TAG     ?= latest
CONTAINER_PORT := 8400
GRPC_PORT     := 50120

.PHONY: install dev start test build docker-build docker-run docker-stop lint clean

install:
	npm ci

dev:
	npm run dev

start:
	npm run start

test:
	npm test

test-coverage:
	npx jest --testPathPattern=tests/ --coverage --forceExit --runInBand

build: install

docker-build:
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .

docker-run:
	docker run --rm \
		--name $(SERVICE_NAME) \
		-p $(CONTAINER_PORT):8400 \
		-p $(GRPC_PORT):50120 \
		--env-file .env \
		$(IMAGE_NAME):$(IMAGE_TAG)

docker-stop:
	docker stop $(SERVICE_NAME) || true

lint:
	npx eslint src/ index.js --ext .js

clean:
	rm -rf node_modules coverage

help:
	@echo "Available targets:"
	@echo "  install        Install dependencies"
	@echo "  dev            Start with hot-reload (nodemon)"
	@echo "  start          Start production server"
	@echo "  test           Run tests"
	@echo "  test-coverage  Run tests with coverage report"
	@echo "  build          Install production dependencies"
	@echo "  docker-build   Build Docker image"
	@echo "  docker-run     Run Docker container"
	@echo "  docker-stop    Stop Docker container"
	@echo "  clean          Remove build artifacts"
