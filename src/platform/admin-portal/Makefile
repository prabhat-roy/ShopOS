.PHONY: build test run docker-build docker-run clean

SERVICE_NAME := admin-portal
IMAGE_TAG    := $(SERVICE_NAME):latest

build:
	mvn -q package -DskipTests

test:
	mvn test

run:
	mvn spring-boot:run

docker-build:
	docker build -t $(IMAGE_TAG) .

docker-run:
	docker run --rm \
		-p 8085:8085 \
		-e ADMIN_API_KEY=admin-secret \
		-e TENANT_SERVICE_URL=http://tenant-service:8087 \
		$(IMAGE_TAG)

clean:
	mvn -q clean
