SERVICE_NAME = feedback-service
IMAGE_NAME   = shopos/$(SERVICE_NAME)
IMAGE_TAG    ?= latest
HTTP_PORT    = 8410

.PHONY: install dev start test test-coverage build docker-build docker-run docker-push migrate lint clean help

help:
	@echo "Usage: make <target>"
	@echo ""
	@echo "Targets:"
	@echo "  install        Install dependencies"
	@echo "  dev            Run in development mode with hot reload"
	@echo "  start          Run in production mode"
	@echo "  test           Run test suite"
	@echo "  test-coverage  Run tests with coverage report"
	@echo "  build          Verify build (node syntax check)"
	@echo "  docker-build   Build Docker image"
	@echo "  docker-run     Run Docker container locally"
	@echo "  docker-push    Push Docker image to registry"
	@echo "  migrate        Run database migrations"
	@echo "  lint           Lint code"
	@echo "  clean          Remove node_modules and coverage"

install:
	npm ci

dev:
	npm run dev

start:
	npm start

test:
	npm test

test-coverage:
	npm run test:coverage

build:
	node --check index.js && node --check src/app.js
	@echo "Build check passed"

docker-build:
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .

docker-run:
	docker run --rm \
		--env-file .env \
		-p $(HTTP_PORT):$(HTTP_PORT) \
		--name $(SERVICE_NAME) \
		$(IMAGE_NAME):$(IMAGE_TAG)

docker-push:
	docker push $(IMAGE_NAME):$(IMAGE_TAG)

migrate:
	@echo "Running database migrations..."
	psql "$(DATABASE_URL)" -f src/migrations/001_create_feedback.sql
	@echo "Migrations complete"

lint:
	npx eslint src/ index.js --ext .js

clean:
	rm -rf node_modules coverage
