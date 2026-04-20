.PHONY: build test run docker-build docker-run clean

SERVICE_NAME := reporting-service
IMAGE_NAME   := shopos/$(SERVICE_NAME)
IMAGE_TAG    := latest

build:
	sbt assembly

test:
	sbt test

run:
	sbt run

docker-build:
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .

docker-run:
	docker run --rm \
		--env-file .env \
		-p 8701:8701 \
		--name $(SERVICE_NAME) \
		$(IMAGE_NAME):$(IMAGE_TAG)

clean:
	sbt clean

lint:
	sbt scalafmtCheck

format:
	sbt scalafmt