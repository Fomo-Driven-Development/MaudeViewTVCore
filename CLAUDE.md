# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**tv_agent** is a browser automation system for TradingView. It captures network traffic via Chrome DevTools Protocol (CDP) for passive research, and provides a REST API for active chart control. Module: `github.com/dgnsrekt/tv_agent`, Go 1.24.

## Build & Run Commands

All commands use `just` (justfile). The justfile auto-loads `.env` via `set dotenv-load`.

```bash
# Prerequisites: copy example.env to .env, start Chromium with CDP enabled
just start-browser          # Launch Chromium with remote debugging on CDP port

# Two main binaries
just run-researcher         # Build + run passive traffic capture
just run-tv-controller      # Build + run Huma REST API server
```

Manual build: `go build -o ./bin/<name> ./cmd/<name>` where name is `researcher` or `tv_controller`.

## Running Tests

```bash
go test ./...                              # All tests
go test ./internal/netutil/                # Single package
go test -run TestPortCheck ./internal/netutil/  # Single test
```

Key test files exist in: `internal/netutil/`, `internal/api/`, and `test/integration/`.

## Architecture

### Two Entry Points (`cmd/`)

1. **researcher** — Passive capture daemon. Attaches to browser tabs matching a URL filter, intercepts HTTP requests/responses, WebSocket frames, and static resources. Writes JSONL to `research_data/`.
2. **tv_controller** — Huma REST API for active TradingView chart control (set symbol/resolution, manage studies, execute actions). Evaluates JavaScript in browser tabs via CDP.

### Internal Packages (`internal/`)

| Package | Purpose |
|---------|---------|
| `cdp` | CDP connection manager + tab registry. Attaches to page targets, enables network/page domains, registers event handlers. |
| `capture` | HTTP and WebSocket event handlers. Intercepts `network.EventRequestWillBeSent`, `ResponseReceived`, `LoadingFinished`, and WebSocket lifecycle events. |
| `storage` | `WriterRegistry` → `JSONLWriter`. Async channel-based JSONL writing with date-partitioned directories and lumberjack log rotation (25MB, 10 backups). |
| `types` | Shared structs: `HTTPCapture`, `HTTPRequest`, `HTTPResponse`, `WebSocketCapture`, `TabInfo`. Body handling with base64 + SHA256. |
| `config` | Env-var driven configuration for researcher and controller with sensible defaults. |
| `api` | Huma REST API server with chi router. Defines operations via `Service` interface. |
| `controller` | Service layer wrapping `cdpcontrol.Client` with input validation. |
| `cdpcontrol` | Active CDP client for TradingView JS evaluation. Per-chart mutex locking, configurable eval timeout. |
| `netutil` | Port availability checking with fallback candidates. |

### Data Flow

```
Chromium (CDP) → researcher → capture handlers → WriterRegistry → research_data/{date}/{chart}/
                                                                    ├── http/
                                                                    ├── websocket/
                                                                    └── resources/js/

Chromium (CDP) ← tv_controller API ← cdpcontrol (JS eval in TradingView pages)
```

## Key Patterns

- **Configuration**: Environment variables with defaults, loaded from `.env` via godotenv. See `example.env` for all options.
- **Thread safety**: `sync.RWMutex` on shared state (tab registry, chart locks, writer registry). Double-checked locking in `WriterRegistry`.
- **Async I/O**: Buffered channels for JSONL writes. `JSONLWriter` drains pending writes on close with 5-second timeout.
- **JSONL everywhere**: All capture artifacts are line-delimited JSON for incremental processing.
- **Date partitioning**: `research_data/` organized by `YYYY-MM-DD` directories.
- **Context cancellation**: All long-running operations respect `ctx.Done()`.
- **Error wrapping**: Typed error codes via `newError` pattern in cdpcontrol.

## Configuration

Copy `example.env` → `.env`. Key settings:
- `CHROMIUM_CDP_PORT` (default 9220) — CDP port for browser connection
- `RESEARCHER_TAB_URL_FILTER` (default `tradingview.com`) — which tabs to attach to
- `CONTROLLER_BIND_ADDR` (default `127.0.0.1:8188`) — API server bind address
- `CONTROLLER_EVAL_TIMEOUT_MS` (default 5000) — JS evaluation timeout

## Documentation

- `docs/controller-runbook.md` — Controller API usage guide
- `docs/functionality-catalog.md` — Latest capability matrix snapshot
- `docs/next-capture-checklist.md` — Planned capture expansion
