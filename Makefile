.PHONY: dev dev-backend dev-frontend mongo mongo-stop test lint clean

# Start MongoDB, backend, and frontend
dev: mongo dev-backend dev-frontend

# Run backend with go run
dev-backend:
	cd backend && go run ./cmd/server

# Run frontend dev server
dev-frontend:
	cd frontend && npm run dev

# Start MongoDB in Docker
mongo:
	docker-compose up -d mongo

# Stop MongoDB
mongo-stop:
	docker-compose down

# Run backend tests
test:
	cd backend && go test ./...

# Lint backend
lint:
	cd backend && go vet ./...

# Stop MongoDB and remove volumes
clean:
	docker-compose down -v
