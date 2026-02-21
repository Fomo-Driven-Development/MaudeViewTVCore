# Controller Runbook

## Purpose

Run the active control API that uses:

- API (Huma)
- CDP
- in-page TradingView JS APIs

## Start

1. Start browser with CDP and authenticated profile:

```bash
just start-browser
```

2. Start controller:

```bash
just run-tv-controller
```

3. Open docs (forced dark mode):

- `http://127.0.0.1:8188/docs`

## Ports

Default bind address: `127.0.0.1:8188` (set via `CONTROLLER_BIND_ADDR`).

If the port is busy, startup fails with an explicit error.

## Key Environment Variables

- `CHROMIUM_CDP_ADDRESS`
- `CHROMIUM_CDP_PORT`
- `CONTROLLER_BIND_ADDR`
- `CONTROLLER_TAB_URL_FILTER`
- `CONTROLLER_EVAL_TIMEOUT_MS`
- `CONTROLLER_LOG_LEVEL`
- `CONTROLLER_LOG_FILE`
- `SNAPSHOT_DIR`

## Logs

- Controller logs write to stdout and `logs/tv_controller.log` by default.
- For deeper diagnostics, set:

```bash
CONTROLLER_LOG_LEVEL=debug
```

- Follow logs live:

```bash
tail -f logs/tv_controller.log
```

## Quick Curl

```bash
curl -s http://127.0.0.1:8188/health
curl -s http://127.0.0.1:8188/api/v1/charts
```

For full endpoint documentation (184 endpoints), see [`dev/implementation-status.md`](dev/implementation-status.md).
