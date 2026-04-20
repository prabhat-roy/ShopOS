SERVICE   := fraud-detection-service
IMAGE     := shopos/$(SERVICE)
TAG       ?= latest

.PHONY: run build test lint docker-build docker-run

run:
	uvicorn main:app --host 0.0.0.0 --port 8143 --reload

build:
	pip install -r requirements.txt

test:
	pytest tests/ -v

lint:
	python -m py_compile main.py fraud/config.py fraud/models.py fraud/scorer.py fraud/store.py handler/api.py

docker-build:
	docker build -t $(IMAGE):$(TAG) .

docker-run:
	docker run --rm \
	  --env-file .env.example \
	  -p 8143:8143 \
	  $(IMAGE):$(TAG)
