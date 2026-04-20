SERVICE   := pricing-service
IMAGE     := shopos/$(SERVICE)
TAG       ?= latest
MVN       := mvn

.PHONY: build test run docker-build docker-run clean

## build — compile and package the JAR (skip tests)
build:
	$(MVN) package -DskipTests

## test — run unit tests
test:
	$(MVN) test

## run — run the service locally (requires Postgres + .env)
run:
	$(MVN) spring-boot:run

## docker-build — build the Docker image
docker-build:
	docker build -t $(IMAGE):$(TAG) .

## docker-run — run the Docker image locally
docker-run:
	docker run --rm \
		--env-file .env \
		-p 8113:8113 \
		$(IMAGE):$(TAG)

## clean — remove compiled artifacts
clean:
	$(MVN) clean
