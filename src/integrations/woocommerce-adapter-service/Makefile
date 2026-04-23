.PHONY: build test run docker-build docker-run

SERVICE_NAME=woocommerce-adapter-service
IMAGE_NAME=enterprise/$(SERVICE_NAME)
VERSION?=latest

build:
	composer install --optimize-autoloader

test:
	./vendor/bin/phpunit

run:
	php -S 0.0.0.0:8216 -t public

docker-build:
	docker build -t $(IMAGE_NAME):$(VERSION) .

docker-run:
	docker run --rm -p 8216:8216 --env-file .env $(IMAGE_NAME):$(VERSION)
