.PHONY: dev dev-backend dev-frontend test lint clean \
       build-frontend embed-frontend build-all build-windows build-mac build-linux \
       wails-dev wails-build wails-build-mac wails-build-windows wails-build-linux

WAILS := $(shell go env GOPATH)/bin/wails

# ── Development ──────────────────────────────────────────────

dev: dev-backend dev-frontend

dev-backend:
	go run ./cmd/server

dev-frontend:
	cd frontend && npm run dev

wails-dev:
	$(WAILS) dev

test:
	go test ./...

lint:
	go vet ./...

clean:
	rm -rf build/windows build/mac build/linux build/bin

# ── Wails Build (Desktop App) ───────────────────────────────

wails-build:
	rm -rf build/bin
	$(WAILS) build

wails-build-mac:
	rm -rf build/bin
	dot_clean . 2>/dev/null || true
	$(WAILS) build -platform darwin/arm64 2>&1 || ( \
		xattr -cr build/bin && \
		codesign --force --deep --sign - build/bin/pokertips.app \
	)

wails-build-windows:
	$(WAILS) build -platform windows/amd64

wails-build-linux:
	$(WAILS) build -platform linux/amd64

# ── Headless Build (Server Only) ────────────────────────────

BINARY_NAME := pokertips

build-frontend:
	cd frontend && npm ci && npm run build

embed-frontend: build-frontend
	rm -rf internal/frontend/dist
	cp -r frontend/dist internal/frontend/dist

build-all: embed-frontend
	@$(MAKE) build-windows-only
	@$(MAKE) build-mac-only
	@$(MAKE) build-linux-only
	@echo "Build complete! Check the build/ directory."

build-windows: embed-frontend build-windows-only
build-mac: embed-frontend build-mac-only
build-linux: embed-frontend build-linux-only

build-windows-only:
	mkdir -p build/windows
	GOOS=windows GOARCH=amd64 go build -o build/windows/$(BINARY_NAME).exe ./cmd/server

build-mac-only:
	mkdir -p build/mac
	GOOS=darwin GOARCH=arm64 go build -o build/mac/$(BINARY_NAME) ./cmd/server

build-linux-only:
	mkdir -p build/linux
	GOOS=linux GOARCH=amd64 go build -o build/linux/$(BINARY_NAME) ./cmd/server
