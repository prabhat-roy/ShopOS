SERVICE_NAME := etl-service
IMAGE        := shopos/$(SERVICE_NAME):latest

.PHONY: install run test docker-build docker-push

install:
	pip install -r requirements.txt

run:
	python main.py

test:
	pytest

docker-build:
	docker build -t $(IMAGE) .

docker-push:
	docker push $(IMAGE)
