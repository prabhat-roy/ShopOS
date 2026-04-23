.PHONY: build test run docker-build docker-run

SERVICE_NAME=mobile-flutter-service
IMAGE_NAME=enterprise/$(SERVICE_NAME)
VERSION?=latest

build:
	flutter pub get
	flutter build web --release

test:
	flutter test

run:
	flutter run

docker-build:
	docker build -t $(IMAGE_NAME):$(VERSION) .

docker-run:
	docker run --rm -p 3005:3005 --env-file .env $(IMAGE_NAME):$(VERSION)
