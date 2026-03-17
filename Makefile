.PHONY: dev dev-backend dev-frontend mongo mongo-stop test lint clean \
       build-frontend embed-frontend build-all build-windows build-mac build-linux

# ── Development ──────────────────────────────────────────────

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

# ── Build ────────────────────────────────────────────────────

BINARY_NAME := pokertips

# Build frontend production bundle
build-frontend:
	cd frontend && npm ci && npm run build

# Copy frontend dist into backend for go:embed
embed-frontend: build-frontend
	rm -rf backend/internal/frontend/dist
	cp -r frontend/dist backend/internal/frontend/dist

# Build for all platforms
build-all: embed-frontend build-windows build-mac build-linux
	cp build/env.example build/windows/.env
	cp build/env.example build/mac/.env
	cp build/env.example build/linux/.env
	@echo "Build complete! Check the build/ directory."

# Build for Windows (amd64)
build-windows:
	mkdir -p build/windows
	cd backend && GOOS=windows GOARCH=amd64 go build -o ../build/windows/$(BINARY_NAME).exe ./cmd/server

# Build for macOS (arm64)
build-mac:
	mkdir -p build/mac
	cd backend && GOOS=darwin GOARCH=arm64 go build -o ../build/mac/$(BINARY_NAME) ./cmd/server

# Build for Linux (amd64)
build-linux:
	mkdir -p build/linux
	cd backend && GOOS=linux GOARCH=amd64 go build -o ../build/linux/$(BINARY_NAME) ./cmd/server
