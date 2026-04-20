SERVICE_NAME := subscription-product-service
IMAGE_NAME   := shopos/$(SERVICE_NAME)
IMAGE_TAG    ?= latest

.PHONY: build test run docker-build docker-run clean

build:
	./gradlew bootJar

test:
	./gradlew test

run:
	./gradlew bootRun

docker-build:
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .

docker-run:
	docker run --rm \
		--env-file .env \
		-p 8119:8119 \
		$(IMAGE_NAME):$(IMAGE_TAG)

clean:
	./gradlew clean
