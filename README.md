# Poker Tips Simulator

A real-time poker chips simulator where players join rooms, manage chip stacks, and play rounds with blinds, bets, and side pots — all from the browser.

## Tech Stack

| Layer    | Technology                                              |
|----------|---------------------------------------------------------|
| Backend  | Go 1.25, Fiber v2, SQLite, JWT, WebSocket               |
| Frontend | React 19, TypeScript, Vite, TailwindCSS v4, Zustand     |

## Prerequisites

- [Go 1.25+](https://go.dev/dl/)
- [Node.js 20+](https://nodejs.org/) (for frontend development)
- [Docker](https://www.docker.com/) (optional, only if using MongoDB)

## Quick Start (Development)

### 1. Backend

```bash
cd backend
go run ./cmd/server
```

The API server starts on `http://localhost:8080` by default.

### 2. Frontend

```bash
cd frontend
npm install
npm run dev
```

The Vite dev server starts on `http://localhost:5173` and proxies `/api` and `/ws` requests to the backend.

### Both at once

```bash
# Terminal 1
make dev-backend

# Terminal 2
make dev-frontend
```

## Environment Variables

Create a `.env` file in the `backend/` directory (or project root):

| Variable     | Default                            | Description               |
|--------------|------------------------------------|---------------------------|
| `PORT`       | `8080`                             | Server listen port        |
| `DB_PATH`    | `pokertips.db`                     | SQLite database file path |
| `JWT_SECRET` | `dev-secret-change-in-production`  | Secret for signing JWTs   |

## Production Build

The production build compiles the frontend into the Go binary using `go:embed`, producing a single self-contained executable.

```bash
# Build for all platforms (windows, mac, linux)
make build-all

# Or build for a specific platform
make build-mac
make build-linux
make build-windows
```

The binaries are output to `build/<platform>/`. Run the binary directly — no Node.js, no separate frontend server needed:

```bash
./build/mac/pokertips
```

### Build Steps (manual)

```bash
# 1. Build the frontend
cd frontend && npm ci && npm run build && cd ..

# 2. Copy dist into the backend embed directory
rm -rf backend/internal/frontend/dist
cp -r frontend/dist backend/internal/frontend/dist

# 3. Compile the Go binary
cd backend && go build -o pokertips ./cmd/server
```

## Makefile Targets

| Target           | Description                                |
|------------------|--------------------------------------------|
| `dev-backend`    | Run backend with `go run`                  |
| `dev-frontend`   | Run frontend Vite dev server               |
| `test`           | Run backend tests                          |
| `lint`           | Run `go vet` on backend                    |
| `build-frontend` | Build frontend production bundle           |
| `embed-frontend` | Build frontend and copy into backend embed |
| `build-all`      | Build binaries for all platforms            |
| `build-mac`      | Build macOS arm64 binary                   |
| `build-linux`    | Build Linux amd64 binary                   |
| `build-windows`  | Build Windows amd64 binary                 |
| `clean`          | Stop Docker services and remove volumes    |

## Running Tests

```bash
make test
# or
cd backend && go test ./...
```

## Project Structure

```
backend/
  cmd/server/           Entry point
  internal/
    domain/entity/      Domain models (room, player, round, pot, blinds, etc.)
    domain/event/       Domain events
    application/        Use cases (room, game, action, blind timer)
      dto/              Request/response DTOs
      port/             Repository and broadcaster interfaces
    adapter/
      handler/          HTTP/WS handlers and routes
      repository/       SQLite repository implementation
      ws/               WebSocket hub, client, messages
    infrastructure/     Config, auth (JWT), database (SQLite)
    frontend/           Embedded frontend build (go:embed)
  pkg/                  Shared utilities (env loader, validator)

frontend/
  src/
    pages/              Home, Lobby, Table pages
    components/         UI components (lobby, table, host, room)
    store/              Zustand state management
    services/           API client, WebSocket client
    hooks/              Custom React hooks
    types/              TypeScript type definitions
    utils/              Helpers (constants, formatting, token)
```

## API Overview

| Method | Endpoint                        | Auth     | Description            |
|--------|---------------------------------|----------|------------------------|
| POST   | `/api/v1/rooms`                 | Public   | Create a new room      |
| POST   | `/api/v1/rooms/join`            | Public   | Join an existing room  |
| *      | `/api/v1/rooms/:roomId/*`       | Bearer   | Room-specific actions  |
| GET    | `/ws?token=<jwt>`               | Token    | WebSocket connection   |


