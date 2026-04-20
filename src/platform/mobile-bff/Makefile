.PHONY: install test run lint docker-build docker-run

install:
	npm install

test:
	npm test

coverage:
	npm run test:coverage

run:
	npm start

dev:
	npm run dev

docker-build:
	docker build -t shopos/mobile-bff:latest .

docker-run:
	docker run --rm -p 8082:8082 --env-file .env shopos/mobile-bff:latest
