SERVICE_NAME   := document-service
IMAGE_NAME     := shopos/$(SERVICE_NAME)
IMAGE_TAG      ?= latest
CONTAINER_PORT := 8602

.PHONY: build test clean run docker-build docker-run lint

build:
	mvn -B package -DskipTests

test:
	mvn -B test

clean:
	mvn -B clean

run:
	mvn spring-boot:run

lint:
	mvn -B checkstyle:check

docker-build:
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .

docker-run:
	docker run --rm \
	  --env-file .env \
	  -p $(CONTAINER_PORT):$(CONTAINER_PORT) \
	  $(IMAGE_NAME):$(IMAGE_TAG)

migrate:
	mvn flyway:migrate

migrate-info:
	mvn flyway:info
