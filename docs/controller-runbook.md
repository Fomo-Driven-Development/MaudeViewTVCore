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

### Strategy Endpoints (tag: Strategy)

Access to `_backtestingStrategyApi` — the backtesting/strategy facade on the TradingView API. Requires a Pine strategy to be loaded on the chart for data to be populated.

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/chart/{chart_id}/strategy/probe` | Probe backtesting strategy API (methods, state) |
| GET | `/api/v1/chart/{chart_id}/strategy/list` | List all loaded strategies |
| GET | `/api/v1/chart/{chart_id}/strategy/active` | Get active strategy with inputs and metadata |
| PUT | `/api/v1/chart/{chart_id}/strategy/active` | Set active strategy by entity ID |
| PUT | `/api/v1/chart/{chart_id}/strategy/input` | Set a strategy input parameter |
| GET | `/api/v1/chart/{chart_id}/strategy/report` | Get backtest report data (P&L, trades, etc.) |
| GET | `/api/v1/chart/{chart_id}/strategy/date-range` | Get backtest date range |
| POST | `/api/v1/chart/{chart_id}/strategy/goto` | Navigate chart to a specific trade/bar timestamp |

#### Prerequisites

- A Pine strategy must be added to the chart (e.g., "MACD Strategy", or a custom Pine script with `strategy()` calls)
- Without a loaded strategy, `list` returns empty, `active` returns nulls, and `report` returns null

#### Quick Curl

```bash
# Probe the API
curl -s http://127.0.0.1:8188/api/v1/chart/{id}/strategy/probe | jq

# List strategies on chart
curl -s http://127.0.0.1:8188/api/v1/chart/{id}/strategy/list | jq

# Get active strategy + inputs
curl -s http://127.0.0.1:8188/api/v1/chart/{id}/strategy/active | jq

# Set a strategy input
curl -X PUT http://127.0.0.1:8188/api/v1/chart/{id}/strategy/input \
  -H 'Content-Type: application/json' -d '{"name":"Fast Length","value":10}'

# Get backtest report
curl -s http://127.0.0.1:8188/api/v1/chart/{id}/strategy/report | jq
```

### Alerts Endpoints (tag: Alerts)

Access to `getAlertsRestApi()` — a webpack-internal singleton wrapping `pricealerts.tradingview.com`. Uses `{alert_ids: [...]}` / `{fire_ids: [...]}` payload format.

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/chart/{chart_id}/alerts/scan` | Scan for alerts API access paths |
| GET | `/api/v1/chart/{chart_id}/alerts/probe` | Probe getAlertsRestApi() singleton |
| GET | `/api/v1/chart/{chart_id}/alerts/probe/deep` | Deep probe with method signatures |
| GET | `/api/v1/alerts` | List all alerts |
| GET | `/api/v1/alerts/{alert_id}` | Get specific alert(s) |
| POST | `/api/v1/alerts` | Create alert |
| PUT | `/api/v1/alerts/{alert_id}` | Modify alert |
| DELETE | `/api/v1/alerts` | Delete alerts |
| POST | `/api/v1/alerts/stop` | Stop alerts |
| POST | `/api/v1/alerts/restart` | Restart alerts |
| POST | `/api/v1/alerts/clone` | Clone alerts |
| GET | `/api/v1/alerts/fires` | List fires (triggers) |
| DELETE | `/api/v1/alerts/fires` | Delete fires |
| DELETE | `/api/v1/alerts/fires/all` | Delete all fires |

### Not Implemented

#### Strategy API — Skipped Methods

These `_backtestingStrategyApi` methods were not exposed as endpoints:

| Method | Reason |
|--------|--------|
| `addAlert()` | Opens UI dialog (alert editor) — not automatable via REST |
| `openSettingDialog()` | Opens UI dialog (strategy settings) — not automatable via REST |
| `openReplaySettingDialog()` | Opens UI dialog (replay properties) — not automatable via REST |
| `requestDeepBacktestingData(e, t)` | Requires active deep backtesting WebSocket session; use `setReportDataSource()` first |
| `copyReportToDeepBacktestingReport()` | Internal plumbing between facade and deep backtesting manager |
| `resetDeepBacktestingReportData()` | Internal plumbing — resets deep backtesting state |
| `setReportDataSource(isDeep)` | Toggles deep backtesting mode — requires WebSocket session infrastructure |
| `formatDWMTimestamp(e)` | Passthrough function — no value as endpoint |
| `destroy()` | Cleanup — would break the API |

#### Conditional APIs — Not Available

These APIs only exist when specific TradingView features are active:

| API | Condition Required | Status |
|-----|-------------------|--------|
| `_accountsManager` | Broker connected (Paper Trading or real broker like Alpaca/IBKR) | Not instantiated without broker |
| `currentAccountApi()` | Same as above — trading account must be active | Not instantiated without broker |
| `_deepBacktestingManager` | Active deep backtesting session (nested inside `_backtestingStrategyApi._deepBacktestingManager`) | Accessible as internal component, not top-level |

#### Alerts API — Skipped Methods

| Method | Reason |
|--------|--------|
| `deleteFiresByFilter(filter)` | Complex filter input — `deleteFires` + `deleteAllFires` suffice |
| `getOfflineFires()` / `clearOfflineFires()` | Notification plumbing, not core alerts |
| `getOfflineFireControls()` / `clearOfflineFireControls()` | Notification settings, not alerts |

### Drawings (tag: `Drawings`)

All under `/api/v1/chart/{chart_id}/drawings/...`. Uses `evalOnChart` (chart context required).

#### Shape CRUD

| Method | Path | Op ID | Notes |
|--------|------|-------|-------|
| GET | `/drawings` | `list-drawings` | `chart.getAllShapes()` |
| GET | `/drawings/{shape_id}` | `get-drawing` | `chart.getShapeById(id)` + properties |
| POST | `/drawings` | `create-drawing` | Body: `{point:{time,price}, options:{shape,...}}` |
| POST | `/drawings/multipoint` | `create-multipoint-drawing` | Body: `{points:[...], options:{shape,...}}`, `points >= 2` |
| POST | `/drawings/{shape_id}/clone` | `clone-drawing` | `chart.cloneLineTool(id)` |
| DELETE | `/drawings/{shape_id}` | `remove-drawing` | Query: `disable_undo`, `chart.removeEntity(id, opts)` |
| DELETE | `/drawings` | `remove-all-drawings` | `chart.removeAllShapes()` |

#### Drawing Toggles

| Method | Path | Op ID | Notes |
|--------|------|-------|-------|
| GET | `/drawings/toggles` | `get-drawing-toggles` | Reads WVs: hide, lock, magnet |
| PUT | `/drawings/toggles/hide` | `set-hide-drawings` | Body: `{value: bool}` |
| PUT | `/drawings/toggles/lock` | `set-lock-drawings` | Body: `{value: bool}` |
| PUT | `/drawings/toggles/magnet` | `set-magnet` | Body: `{enabled: bool, mode?: 0|1}` |
| PUT | `/drawings/{shape_id}/visibility` | `set-drawing-visibility` | Body: `{visible: bool}` |

#### Tool Selection

| Method | Path | Op ID | Notes |
|--------|------|-------|-------|
| GET | `/drawings/tool` | `get-drawing-tool` | `api.selectedLineTool()` |
| PUT | `/drawings/tool` | `set-drawing-tool` | Body: `{tool: string}`, `api.selectLineTool(tool)` |

#### Z-Order

| Method | Path | Op ID | Notes |
|--------|------|-------|-------|
| POST | `/drawings/{shape_id}/z-order` | `set-drawing-z-order` | Body: `{action: "bring_forward"|"bring_to_front"|"send_backward"|"send_to_back"}` |

#### State Export/Import

| Method | Path | Op ID | Notes |
|--------|------|-------|-------|
| GET | `/drawings/state` | `export-drawings-state` | `chart.getLineToolsState()` |
| PUT | `/drawings/state` | `import-drawings-state` | Body: `{state: any}`, `chart.applyLineToolsState(dto)` |

#### Drawings API — Skipped Methods

| Method | Reason |
|--------|--------|
| `createAnchoredShape` / `createExecutionShape` | Specialized shapes, add later |
| `copyEntityToClipboard` | UI-coupled clipboard |
| `drawOnAllCharts` | Multi-layout feature, edge case |
| Drawing groups (createGroupFromSelection, etc.) | Lower priority |
| Individual shape property editing | Needs probing with shapes on chart |

## Quick Curl

```bash
curl -s http://127.0.0.1:8188/health
curl -s http://127.0.0.1:8188/api/v1/charts
```
