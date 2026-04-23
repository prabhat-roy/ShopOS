.PHONY: build test run docker-build docker-run

SERVICE_NAME=auction-settlement-service
IMAGE_NAME=enterprise/$(SERVICE_NAME)
VERSION?=latest

build:
	mvn package -DskipTests -B

test:
	mvn test -B

run:
	mvn spring-boot:run

docker-build:
	docker build -t $(IMAGE_NAME):$(VERSION) .

docker-run:
	docker run --rm -p 50313:50313 --env-file .env $(IMAGE_NAME):$(VERSION)
