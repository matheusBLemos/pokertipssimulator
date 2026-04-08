.PHONY: dev dev-backend dev-frontend test lint clean \
       build-frontend embed-frontend build-all build-windows build-mac build-linux

# ── Development ──────────────────────────────────────────────

dev: dev-backend dev-frontend

dev-backend:
	go run ./cmd/server

dev-frontend:
	cd frontend && npm run dev

test:
	go test ./...

lint:
	go vet ./...

clean:
	rm -rf build/windows/$(BINARY_NAME).exe build/mac/$(BINARY_NAME) build/linux/$(BINARY_NAME)

# ── Build ────────────────────────────────────────────────────

BINARY_NAME := pokertips

build-frontend:
	cd frontend && npm ci && npm run build

embed-frontend: build-frontend
	rm -rf internal/frontend/dist
	cp -r frontend/dist internal/frontend/dist

build-all: embed-frontend build-windows build-mac build-linux
	cp build/env.example build/windows/.env
	cp build/env.example build/mac/.env
	cp build/env.example build/linux/.env
	@echo "Build complete! Check the build/ directory."

build-windows:
	mkdir -p build/windows
	GOOS=windows GOARCH=amd64 go build -o build/windows/$(BINARY_NAME).exe ./cmd/server

build-mac:
	mkdir -p build/mac
	GOOS=darwin GOARCH=arm64 go build -o build/mac/$(BINARY_NAME) ./cmd/server

build-linux:
	mkdir -p build/linux
	GOOS=linux GOARCH=amd64 go build -o build/linux/$(BINARY_NAME) ./cmd/server
