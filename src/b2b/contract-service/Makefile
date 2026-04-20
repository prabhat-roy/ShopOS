SERVICE_NAME=contract-service
IMAGE_NAME=shopos/$(SERVICE_NAME)
IMAGE_TAG?=latest

.PHONY: build test run clean docker-build docker-run

build:
	./gradlew bootJar --no-daemon

test:
	./gradlew test --no-daemon

run:
	./gradlew bootRun --no-daemon

clean:
	./gradlew clean --no-daemon

docker-build:
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .

docker-run:
	docker run --rm \
		--name $(SERVICE_NAME) \
		-p 8801:8801 \
		--env-file .env.example \
		$(IMAGE_NAME):$(IMAGE_TAG)
