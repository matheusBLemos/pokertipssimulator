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
| Backend  | Go 1.25, Fiber v2, SQLite, JWT, WebSocket               |
| Frontend | React 19, TypeScript, Vite, TailwindCSS v4, Zustand     |

## Prerequisites

- [Go 1.25+](https://go.dev/dl/)
- [Node.js 20+](https://nodejs.org/) (for frontend development only)

## Quick Start (Development)

Run the backend and frontend in two separate terminals:

```bash
# Terminal 1 — Backend (from project root)
go run ./cmd/server

# Terminal 2 — Frontend
cd frontend && npm install && npm run dev
```

Or use the Makefile:

```bash
# Terminal 1
make dev-backend

# Terminal 2
make dev-frontend
```

- The Go API server starts on `http://localhost:8080`
- The Vite dev server starts on `http://localhost:5173` and proxies `/api` and `/ws` to the backend
- On startup, the server prints your LAN and public IP so friends can connect

## Production Build (Single Binary)

The production build compiles the React frontend into the Go binary using `go:embed`, producing a **single self-contained executable**. No Node.js, no separate frontend server, no database setup — just run the binary.

```bash
# Build for all platforms
make build-all

# Or build for a specific platform
make build-mac
make build-linux
make build-windows
```

Binaries are output to `build/<platform>/`. Run directly:

```bash
./build/mac/pokertips
```

The binary serves both the API and the frontend on the same port. Open `http://localhost:8080` in a browser, or share the IP shown in the terminal with friends.

## Environment Variables

Create a `.env` file in the project root (optional — defaults work out of the box):

| Variable     | Default                            | Description               |
|--------------|------------------------------------|---------------------------|
| `PORT`       | `8080`                             | Server listen port        |
| `DB_PATH`    | `pokertips.db`                     | SQLite database file path |
| `JWT_SECRET` | `dev-secret-change-in-production`  | Secret for signing JWTs   |

## Makefile Targets

| Target           | Description                                |
|------------------|--------------------------------------------|
| `dev-backend`    | Run backend with `go run`                  |
| `dev-frontend`   | Run frontend Vite dev server               |
| `test`           | Run all Go tests                           |
| `lint`           | Run `go vet`                               |
| `build-frontend` | Build frontend production bundle           |
| `embed-frontend` | Build frontend and copy into embed dir     |
| `build-all`      | Build binaries for all platforms            |
| `build-mac`      | Build macOS arm64 binary                   |
| `build-linux`    | Build Linux amd64 binary                   |
| `build-windows`  | Build Windows amd64 binary                 |

## Running Tests

```bash
make test
# or
go test ./...
```

## Project Structure

```
cmd/server/               Entry point
internal/
  domain/entity/           Domain models (room, player, round, pot, blinds)
  domain/event/            Domain events
  application/             Use cases (room, game, action, tips, blind timer)
    dto/                   Request/response DTOs
    port/                  Repository, broadcaster, and network interfaces
  adapter/
    handler/               HTTP/WS handlers and routes (game + tips groups)
    repository/            SQLite repository implementation
    ws/                    WebSocket hub, client, messages
  infrastructure/          Config, auth (JWT), database (SQLite)
  frontend/                Embedded frontend build (go:embed)
pkg/                       Shared utilities (env loader, ID gen, validator)

frontend/
  src/
    pages/
      MainMenuPage         Mode selection (Tips vs Game)
      tips/                TipsHome, TipsLobby, TipsTable
      game/                GameHome, GameLobby, GameTable
    components/
      shared/              ChipTransfer (tips mode)
      lobby/               SeatPicker, PlayerList, GameSettings
      table/               PokerTable, ActionBar, BlindTimer, PlayerSeat
      host/                HostControls, SettlementModal
      room/                RebuyModal
    store/                 Zustand stores (appStore, roomStore, gameStore)
    services/              API clients (gameApi, tipsApi), WebSocket client
    hooks/                 Custom React hooks
    types/                 TypeScript type definitions
    utils/                 Helpers (constants, formatting, token)
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

### WebSocket (shared)

| Endpoint            | Description              |
|---------------------|--------------------------|
| `GET /ws?token=jwt` | Real-time state updates  |
