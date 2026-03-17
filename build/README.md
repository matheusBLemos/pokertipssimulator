# Poker Tips Simulator - Local Build

A self-contained poker chips simulator. No external database or dependencies required.

## Quick Start

1. (Optional) Edit `.env` to change port, database path, or JWT secret
2. Run the executable:
   - **Windows**: double-click `pokertips.exe` or run it from a terminal
   - **macOS**: `./pokertips` (you may need to allow it in System Settings > Privacy & Security)
   - **Linux**: `./pokertips`
3. Open `http://localhost:8080` in your browser

## Configuration

| Variable     | Default              | Description                        |
|-------------|----------------------|------------------------------------|
| `PORT`      | `8080`               | HTTP server port                   |
| `DB_PATH`   | `pokertips.db`       | SQLite database file path          |
| `JWT_SECRET`| `change-me-...`      | Secret key for JWT authentication  |

The SQLite database file (`pokertips.db`) is created automatically on first run in the current directory.
