SERVICE  := payment-service
IMAGE    := shopos/$(SERVICE)
VERSION  := 1.0.0
PORT     := 8136

.PHONY: build test run clean docker-build docker-run docker-push

build:
	mvn package -DskipTests

test:
	mvn test

run:
	mvn spring-boot:run

clean:
	mvn clean

docker-build:
	docker build -t $(IMAGE):$(VERSION) -t $(IMAGE):latest .

docker-run:
	docker run --rm \
		--env-file .env.example \
		-p $(PORT):$(PORT) \
		$(IMAGE):latest

docker-push:
	docker push $(IMAGE):$(VERSION)
	docker push $(IMAGE):latest

lint:
	mvn checkstyle:check

format:
	mvn fmt:format
