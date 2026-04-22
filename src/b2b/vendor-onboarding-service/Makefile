SERVICE_NAME := vendor-onboarding-service
IMAGE        := shopos/$(SERVICE_NAME):latest

.PHONY: build test docker-build docker-push

build:
	mvn package -DskipTests

test:
	mvn test

docker-build:
	docker build -t $(IMAGE) .

docker-push:
	docker push $(IMAGE)
