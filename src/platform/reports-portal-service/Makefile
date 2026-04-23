SERVICE  := reports-portal-service
PORT     := 8300
IMAGE    := $(SERVICE)

.PHONY: build run docker-build docker-run

build:
	go build -o $(SERVICE) .

run: build
	PORT=$(PORT) ./$(SERVICE)

docker-build:
	docker build -t $(IMAGE):local .

docker-run:
	docker run --rm -p $(PORT):$(PORT) \
		-e GRAFANA_URL=http://localhost:3000 \
		-e JENKINS_URL=http://localhost:8080 \
		$(IMAGE):local
