SERVICE   := price-optimization-service
IMAGE     := shopos/$(SERVICE)
TAG       ?= latest
HTTP_PORT ?= 8704

.PHONY: install run test lint build push clean

install:
	pip install -r requirements.txt

run:
	python main.py

test:
	pytest tests/ -v --asyncio-mode=auto

lint:
	python -m py_compile main.py optimizer/*.py

build:
	docker build -t $(IMAGE):$(TAG) .

push:
	docker push $(IMAGE):$(TAG)

clean:
	find . -type d -name __pycache__ -exec rm -rf {} + 2>/dev/null || true
	find . -name "*.pyc" -delete 2>/dev/null || true
