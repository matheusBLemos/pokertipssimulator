.PHONY: up down build logs backend-logs frontend-logs mongo-logs test lint

up:
	docker-compose up -d

up-build:
	docker-compose up -d --build

down:
	docker-compose down

build:
	docker-compose build

logs:
	docker-compose logs -f

backend-logs:
	docker-compose logs -f backend

frontend-logs:
	docker-compose logs -f frontend

mongo-logs:
	docker-compose logs -f mongo

test:
	cd backend && go test ./...

lint:
	cd backend && go vet ./...

clean:
	docker-compose down -v
