SERVICE_NAME=organization-service
IMAGE_NAME=shopos/$(SERVICE_NAME)
IMAGE_TAG?=latest

.PHONY: build test run clean docker-build docker-run

build:
	mvn package -DskipTests -q

test:
	mvn test

run:
	mvn spring-boot:run

clean:
	mvn clean

docker-build:
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .

docker-run:
	docker run --rm \
		--name $(SERVICE_NAME) \
		-p 8800:8800 \
		--env-file .env.example \
		$(IMAGE_NAME):$(IMAGE_TAG)
