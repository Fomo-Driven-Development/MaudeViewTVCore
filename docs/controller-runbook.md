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
just run-controller
```

3. Open docs (forced dark mode):

- `http://127.0.0.1:8188/docs`

## Ports

Defaults:

- Preferred bind: `127.0.0.1:8188`
- Auto fallback candidates: `127.0.0.1:8188,127.0.0.1:8189,127.0.0.1:8190,127.0.0.1:8191`

Behavior:

- If preferred port is busy and fallback is enabled, controller picks the first free candidate.
- If no candidate is free, startup fails with an explicit error.

## Key Environment Variables

- `CHROMIUM_CDP_ADDRESS`
- `CHROMIUM_CDP_PORT`
- `CONTROLLER_BIND_ADDR`
- `CONTROLLER_PORT_CANDIDATES`
- `CONTROLLER_PORT_AUTO_FALLBACK`
- `CONTROLLER_TAB_URL_FILTER`
- `CONTROLLER_EVAL_TIMEOUT_MS`
- `CONTROLLER_LOG_LEVEL`
- `CONTROLLER_LOG_FILE`

## Logs

- Controller logs write to stdout and `logs/controller.log` by default.
- For deeper diagnostics, set:

```bash
CONTROLLER_LOG_LEVEL=debug
```

- Follow logs live:

```bash
tail -f logs/controller.log
```

## Endpoints

- `GET /health`
- `GET /api/v1/charts`
- `GET /api/v1/chart/{chart_id}/symbol`
- `PUT /api/v1/chart/{chart_id}/symbol?symbol=NASDAQ:AAPL`
- `GET /api/v1/chart/{chart_id}/resolution`
- `PUT /api/v1/chart/{chart_id}/resolution?resolution=60`
- `POST /api/v1/chart/{chart_id}/action` with `{"action_id":"..."}`
- `GET /api/v1/chart/{chart_id}/studies`
- `POST /api/v1/chart/{chart_id}/studies`
- `DELETE /api/v1/chart/{chart_id}/studies/{study_id}`

## Quick Curl

```bash
curl -s http://127.0.0.1:8188/health
curl -s http://127.0.0.1:8188/api/v1/charts
```
