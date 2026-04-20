SERVICE   := audit-service
IMAGE     := shopos/$(SERVICE)
TAG       ?= latest

.PHONY: build test package docker-build docker-push clean

## Compile the project
build:
	mvn -B compile

## Run unit tests
test:
	mvn -B test

## Package into a JAR (skipping tests — run `make test` first)
package:
	mvn -B package -DskipTests

## Build the Docker image
docker-build: package
	docker build -t $(IMAGE):$(TAG) .

## Push the Docker image to the registry
docker-push: docker-build
	docker push $(IMAGE):$(TAG)

## Remove build artifacts
clean:
	mvn -B clean
