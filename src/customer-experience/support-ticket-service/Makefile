SERVICE_NAME = support-ticket-service
IMAGE_NAME   = shopos/$(SERVICE_NAME)
IMAGE_TAG   ?= latest

.PHONY: build test run docker-build docker-run clean

build:
	mvn -q package -DskipTests

test:
	mvn test

run:
	mvn spring-boot:run

docker-build:
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .

docker-run:
	docker run --rm \
		--env-file .env \
		-p 8405:8405 \
		$(IMAGE_NAME):$(IMAGE_TAG)

clean:
	mvn -q clean

lint:
	mvn checkstyle:check

help:
	@echo "Targets: build | test | run | docker-build | docker-run | clean"
