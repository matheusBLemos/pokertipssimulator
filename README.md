# Poker Application

A desktop poker application with two modes:

- **Poker Tips** -- Digital chip simulator for live games. Track chips, manage blinds, and transfer chips between players.
- **Poker With Friends** -- Full online poker game with rounds, betting actions, streets, and settlements.

## How It Works

The host's machine **is** the server. When you create a room, a local HTTP + WebSocket server starts and displays your LAN and public IP addresses. Friends connect by entering your IP and room code — no cloud infrastructure, no accounts, no Docker.

```
Host creates room  →  Server starts on host machine  →  Shares IP:Port
Friends open app   →  Enter host IP:Port + room code →  Join via network
```

All players (including the host) communicate with the same Fiber server over HTTP/WebSocket.

## Tech Stack

| Layer    | Technology                                              |
|----------|---------------------------------------------------------|
| Desktop  | Wails v2 (native window with WebView)                   |
| Backend  | Go 1.25, Fiber v2, SQLite, JWT, WebSocket, UPnP        |
| Frontend | React 19, TypeScript, Vite, TailwindCSS v4, Zustand     |

## Prerequisites

- [Go 1.25+](https://go.dev/dl/)
- [Node.js 20+](https://nodejs.org/)
- [Wails CLI](https://wails.io/docs/gettingstarted/installation): `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- macOS: Xcode Command Line Tools
- Linux: `libgtk-3-dev libwebkit2gtk-4.0-dev`

Verify your setup:

```bash
wails doctor
```

## Quick Start (Development)

### Desktop App (Wails)

```bash
wails dev
```

This starts the Go backend, Vite frontend with hot-reload, and opens a native desktop window.

### Headless Server (no desktop window)

Run the backend and frontend in two separate terminals:

```bash
# Terminal 1 — Backend (from project root)
make dev-backend

# Terminal 2 — Frontend
make dev-frontend
```

- The Go API server starts on `http://localhost:8080`
- The Vite dev server starts on `http://localhost:5173` and proxies `/api` and `/ws` to the backend
- On startup, the server prints your LAN and public IP so friends can connect

## Production Build

### Desktop App (recommended)

Wails builds a native executable with the frontend embedded:

```bash
# Build for your current platform
make wails-build

# Or for a specific platform
make wails-build-mac
make wails-build-windows
make wails-build-linux
```

Output: `build/bin/pokertips` (or `.exe` on Windows). Double-click to launch.

### Headless Server

For running without a desktop window (e.g., on a VPS or headless machine):

```bash
make build-mac    # or build-linux, build-windows, build-all
```

Output: `build/<platform>/pokertips`. Run and open `http://localhost:8080` in a browser.

## Network Connectivity

### LAN (same Wi-Fi)

Works out of the box. The host shares their LAN IP (shown in the lobby) and friends connect.

### Internet (different networks)

The app attempts **UPnP automatic port mapping** on the host's router. If successful, friends can connect using the public IP shown in the lobby.

If UPnP is unavailable:
1. Forward TCP port 8080 (or your chosen port) on your router to your machine
2. Share your public IP and port with friends

### VPN (Tailscale, ZeroTier, Hamachi)

Treated as LAN. No extra configuration needed — just use the VPN-assigned IP.

## Environment Variables

Create a `.env` file in the project root (optional — defaults work out of the box):

| Variable     | Default                            | Description               |
|--------------|------------------------------------|---------------------------|
| `PORT`       | `8080`                             | Server listen port        |
| `DB_PATH`    | `pokertips.db`                     | SQLite database file path |
| `JWT_SECRET` | `dev-secret-change-in-production`  | Secret for signing JWTs   |

## Makefile Targets

| Target              | Description                                              |
|---------------------|----------------------------------------------------------|
| `wails-dev`         | Development mode with Wails (hot-reload + native window) |
| `wails-build`       | Build desktop app for current platform                   |
| `wails-build-mac`   | Build desktop app for macOS arm64                        |
| `wails-build-windows` | Build desktop app for Windows amd64                    |
| `wails-build-linux` | Build desktop app for Linux amd64                        |
| `dev-backend`       | Run headless backend with `go run`                       |
| `dev-frontend`      | Run frontend Vite dev server                             |
| `test`              | Run all Go tests                                         |
| `lint`              | Run `go vet`                                             |
| `build-all`         | Headless build for all platforms                          |
| `build-mac`         | Headless build for macOS arm64                            |
| `build-linux`       | Headless build for Linux amd64                            |
| `build-windows`     | Headless build for Windows amd64                          |

## Running Tests

```bash
make test
# or
go test ./...
```

## Project Structure

```
main.go                       Wails entry point (desktop app)
app.go                        Wails bindings (StartServer, GetConnectionInfo)
wails.json                    Wails configuration
cmd/server/                   Headless server entry point
internal/
  server/                     Server lifecycle (Start/Stop)
  domain/entity/              Domain models (room, player, round, pot, blinds)
  domain/event/               Domain events
  application/                Use cases (room, game, action, tips, blind timer)
    dto/                      Request/response DTOs
    port/                     Repository, broadcaster, and network interfaces
  adapter/
    handler/                  HTTP/WS handlers and routes
    repository/               SQLite repository implementation
    ws/                       WebSocket hub, client, messages
    network/                  IP detection, UPnP port mapping
  infrastructure/             Config, auth (JWT), database (SQLite)
  frontend/                   Embedded frontend build (go:embed)
pkg/                          Shared utilities (env loader, ID gen, validator)

frontend/
  src/
    pages/
      MainMenuPage            Mode selection (Tips vs Game)
      tips/                   TipsHome, TipsLobby, TipsTable
      game/                   GameHome, GameLobby, GameTable
    components/
      shared/                 ConnectionInfo
      lobby/                  SeatPicker, PlayerList, GameSettings
      table/                  PokerTable, ActionBar, BlindTimer, PlayerSeat
      host/                   HostControls, SettlementModal
      room/                   RebuyModal
    store/                    Zustand stores (appStore, roomStore, gameStore)
    services/                 API client, WebSocket client, Wails bindings
    hooks/                    Custom React hooks
    types/                    TypeScript type definitions
    utils/                    Helpers (constants, formatting, token)
```

## API Overview

### Game Mode (`/api/v1/game`)

| Method | Endpoint                                  | Auth     | Description            |
|--------|-------------------------------------------|----------|------------------------|
| POST   | `/api/v1/game/rooms`                      | Public   | Create a game room     |
| POST   | `/api/v1/game/rooms/join`                 | Public   | Join a game room       |
| GET    | `/api/v1/game/rooms/:roomId/`             | Bearer   | Get room state         |
| PUT    | `/api/v1/game/rooms/:roomId/config`       | Bearer   | Update config (host)   |
| POST   | `/api/v1/game/rooms/:roomId/rounds/start` | Bearer   | Start round (host)     |
| POST   | `/api/v1/game/rooms/:roomId/action`       | Bearer   | Perform action         |
| POST   | `/api/v1/game/rooms/:roomId/rounds/settle`| Bearer   | Settle round (host)    |

### Tips Mode (`/api/v1/tips`)

| Method | Endpoint                                      | Auth     | Description             |
|--------|-----------------------------------------------|----------|-------------------------|
| POST   | `/api/v1/tips/rooms`                          | Public   | Create a tips room      |
| POST   | `/api/v1/tips/rooms/join`                     | Public   | Join a tips room        |
| GET    | `/api/v1/tips/rooms/:roomId/`                 | Bearer   | Get room state          |
| POST   | `/api/v1/tips/rooms/:roomId/chips/transfer`   | Bearer   | Transfer chips          |
| POST   | `/api/v1/tips/rooms/:roomId/blinds/advance`   | Bearer   | Advance blind (host)    |
| POST   | `/api/v1/tips/rooms/:roomId/pause`            | Bearer   | Pause/resume (host)     |

### Shared

| Method | Endpoint                    | Auth     | Description                  |
|--------|-----------------------------|----------|------------------------------|
| GET    | `/api/v1/connection-info`   | Public   | Get server IPs, port, UPnP   |
| GET    | `/ws?token=jwt`             | Token    | Real-time state updates      |
| GET    | `/health`                   | Public   | Health check                 |
