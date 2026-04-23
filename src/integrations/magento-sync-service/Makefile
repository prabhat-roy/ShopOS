.PHONY: build test run docker-build docker-run

SERVICE_NAME=magento-sync-service
IMAGE_NAME=enterprise/$(SERVICE_NAME)
VERSION?=latest

build:
	composer install --optimize-autoloader

test:
	./vendor/bin/phpunit

run:
	php -S 0.0.0.0:8215 -t public

docker-build:
	docker build -t $(IMAGE_NAME):$(VERSION) .

docker-run:
	docker run --rm -p 8215:8215 --env-file .env $(IMAGE_NAME):$(VERSION)
