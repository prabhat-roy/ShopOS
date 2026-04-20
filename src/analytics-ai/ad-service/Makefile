.PHONY: build test run docker-build docker-run clean

SERVICE_NAME := ad-service
IMAGE_NAME   := shopos/$(SERVICE_NAME)
IMAGE_TAG    := latest

build:
	mvn clean package -DskipTests

test:
	mvn test

run:
	mvn spring-boot:run

docker-build:
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .

docker-run:
	docker run --rm \
		--env-file .env \
		-p 8708:8708 \
		--name $(SERVICE_NAME) \
		$(IMAGE_NAME):$(IMAGE_TAG)

clean:
	mvn clean

lint:
	mvn checkstyle:check