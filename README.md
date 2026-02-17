<p align="center">
  <img src="assets/banner.jpg" alt="MaudeViewTVCore" width="600">
</p>

# MaudeViewTVCore

Browser automation for TradingView via Chrome DevTools Protocol.

Exposes a REST API for TradingView chart control — symbol, resolution, studies, drawings, alerts, watchlists, Pine editor, replay, snapshots, and layouts. Connects to a Chromium browser over CDP and evaluates JavaScript directly in TradingView pages. A separate researcher daemon passively captures network traffic for offline analysis.

## Prerequisites

- Go 1.24+
- Chromium (or Chrome) with remote debugging enabled
- [`just`](https://github.com/casey/just) command runner

## Quick Start

```bash
cp example.env .env            # configure CDP port, bind address, etc.
just start-browser             # launch Chromium with remote debugging
just run-tv-controller         # start the REST API server
```

Open <http://127.0.0.1:8188/docs> for interactive API documentation.

```bash
# List detected chart tabs
curl -s http://127.0.0.1:8188/api/v1/charts | jq
```

## Architecture

```
Chromium (CDP)  <──  tv_controller API  <──  cdpcontrol (JS eval in TradingView pages)
                         :8188
```

### Internal Packages

| Package | Purpose |
|---------|---------|
| `api` | Huma REST API server with chi router, operation definitions |
| `cdp` | CDP connection manager and tab registry |
| `cdpcontrol` | JS evaluation client for TradingView — per-chart mutex, eval timeout |
| `controller` | Service layer with input validation over cdpcontrol |
| `config` | Env-var driven configuration with defaults |
| `snapshot` | Chart screenshot capture and storage |
| `capture` | HTTP/WebSocket event handlers (researcher) |
| `storage` | Async JSONL writer with log rotation (researcher) |
| `types` | Shared structs for captures and tab info |

## Configuration

Copy `example.env` to `.env`. Key settings:

| Variable | Default | Description |
|----------|---------|-------------|
| `CHROMIUM_CDP_ADDRESS` | `127.0.0.1` | CDP listening address |
| `CHROMIUM_CDP_PORT` | `9220` | CDP listening port |
| `CONTROLLER_BIND_ADDR` | `127.0.0.1:8188` | API server bind address |
| `CONTROLLER_EVAL_TIMEOUT_MS` | `5000` | JS evaluation timeout (ms) |
| `SNAPSHOT_DIR` | `./snapshots` | Chart snapshot storage directory |

See [`example.env`](example.env) for the full list including researcher settings.

## Security Warning

**CDP and the REST API have no authentication.** Always bind to `127.0.0.1` (the default). Use a dedicated browser profile — the controller can execute arbitrary JavaScript in any matched tab. Do not expose the CDP port or API to untrusted networks.

See [`docs/security-guide.md`](docs/security-guide.md) for hardening details.

## Running Tests

```bash
go test ./...                                          # unit tests
just test-integration                                  # integration tests (requires running browser + controller)
```

## Documentation

- [`docs/security-guide.md`](docs/security-guide.md) — Security hardening for CDP browser automation
- [`docs/controller-runbook.md`](docs/controller-runbook.md) — Controller startup, configuration, and usage
- [`docs/dev/implementation-status.md`](docs/dev/implementation-status.md) — All endpoints with mechanism details and coverage
- [`docs/dev/js-internals.md`](docs/dev/js-internals.md) — TradingView JS manager singletons reference
- Interactive API docs at [`/docs`](http://127.0.0.1:8188/docs) when the controller is running

## Project Structure

```
cmd/
  tv_controller/       # REST API entry point
  researcher/          # Passive traffic capture daemon
internal/
  api/                 # Huma server + route definitions
  cdp/                 # CDP connection + tab registry
  cdpcontrol/          # JS eval client for TradingView
  controller/          # Service layer (validation, orchestration)
  config/              # Environment-based configuration
  snapshot/            # Chart screenshot capture
  capture/             # Network event handlers (researcher)
  storage/             # JSONL writer + rotation (researcher)
  types/               # Shared data structures
docs/                  # Security guide, runbook, dev docs
scripts/               # Browser launch + diagnostic scripts
```
