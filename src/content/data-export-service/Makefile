SERVICE      := data-export-service
IMAGE        := shopos/$(SERVICE)
TAG          ?= latest
HTTP_PORT    ?= 8607
GRPC_PORT    ?= 50147

.PHONY: install run test lint build docker-build docker-run docker-push clean

install:
	pip install -r requirements.txt

run:
	python main.py

test:
	pytest tests/ -v --tb=short

lint:
	python -m py_compile main.py exporter/*.py exporter/exporters/*.py

build: install

docker-build:
	docker build -t $(IMAGE):$(TAG) .

docker-run:
	docker run --rm \
		-p $(HTTP_PORT):$(HTTP_PORT) \
		-p $(GRPC_PORT):$(GRPC_PORT) \
		--name $(SERVICE) \
		$(IMAGE):$(TAG)

docker-push:
	docker push $(IMAGE):$(TAG)

clean:
	find . -type d -name __pycache__ -exec rm -rf {} + 2>/dev/null || true
	find . -name "*.pyc" -delete 2>/dev/null || true
	rm -rf .pytest_cache
