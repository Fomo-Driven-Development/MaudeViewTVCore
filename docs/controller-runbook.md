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
- `SNAPSHOT_DIR`

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
- `GET /api/v1/pine/probe`
- `POST /api/v1/pine/open`
- `GET /api/v1/pine/source`
- `PUT /api/v1/pine/source`
- `POST /api/v1/pine/add-to-chart`
- `POST /api/v1/pine/update-on-chart`
- `GET /api/v1/pine/scripts`
- `POST /api/v1/pine/scripts/open`
- `GET /api/v1/pine/console`

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

### Pine Editor (tag: `Pine Editor`)

Programmatic control of the TradingView Pine Script editor via `api.pineEditorApi()`. The editor panel is lazy-loaded — `openEditor()` initializes it. All endpoints are session-level (`evalOnAnyChart`).

| Method | Path | Op ID | Description |
|--------|------|-------|-------------|
| GET | `/api/v1/pine/probe` | `probe-pine-editor` | Probe Pine editor API availability + state |
| POST | `/api/v1/pine/open` | `open-pine-editor` | Open/initialize the Pine editor panel |
| GET | `/api/v1/pine/source` | `get-pine-source` | Get current editor source text + metadata |
| PUT | `/api/v1/pine/source` | `set-pine-source` | Set editor source text |
| POST | `/api/v1/pine/add-to-chart` | `add-pine-to-chart` | Compile + add current script to chart |
| POST | `/api/v1/pine/update-on-chart` | `update-pine-on-chart` | Update already-added script on chart |
| GET | `/api/v1/pine/scripts` | `list-pine-scripts` | List user's saved scripts |
| POST | `/api/v1/pine/scripts/open` | `open-pine-script` | Open a saved script by ID |
| GET | `/api/v1/pine/console` | `get-pine-console` | Get Pine console messages |

#### Full Workflow

```
1. Open editor  →  2. Set source  →  3. Add to chart
```

#### Quick Curl

```bash
# Probe Pine editor
curl -s http://127.0.0.1:8188/api/v1/pine/probe | jq

# Open editor
curl -s -X POST http://127.0.0.1:8188/api/v1/pine/open | jq

# Set Pine source code
curl -s -X PUT http://127.0.0.1:8188/api/v1/pine/source \
  -H 'Content-Type: application/json' \
  -d '{"source": "//@version=6\nindicator(\"Test\", overlay=true)\nplot(close, color=color.blue)"}' | jq

# Get current source
curl -s http://127.0.0.1:8188/api/v1/pine/source | jq

# Add to chart
curl -s -X POST http://127.0.0.1:8188/api/v1/pine/add-to-chart | jq

# Update on chart (after editing source)
curl -s -X POST http://127.0.0.1:8188/api/v1/pine/update-on-chart | jq

# List saved scripts
curl -s http://127.0.0.1:8188/api/v1/pine/scripts | jq

# Open a saved script
curl -s -X POST http://127.0.0.1:8188/api/v1/pine/scripts/open \
  -H 'Content-Type: application/json' \
  -d '{"script_id_part": "USER;abc123", "version": "8.0"}' | jq

# Get console messages
curl -s http://127.0.0.1:8188/api/v1/pine/console | jq
```

#### Pine Editor API — Skipped Methods

| Method | Reason |
|--------|--------|
| `saveScript()` | Cloud save — side effects, add later if needed |
| `focusEditor()` | UI focus — minimal value as REST endpoint |
| `openNewScript()` | Redundant — `openEditor()` + `setEditorText()` achieves the same |
| `addOpenedScript()` / `removeOpenedScript()` | Internal tab management |
| `destroy()` | Cleanup — would break the API |
| `pineLibApi().saveNew()` / `saveNext()` | Library publishing — add later |
| `pineLibApi().requestBuiltinScripts()` | Read-only list — add later if needed |

### Snapshots (tag: `Snapshots`)

Takes chart screenshots via `api.takeClientScreenshot()`, stores them with UUID filenames, and serves them back. Metadata includes symbol, exchange, resolution, theme from `api._chartWidgetCollection.images()`.

| Method | Path | Op ID | Description |
|--------|------|-------|-------------|
| POST | `/api/v1/chart/{chart_id}/snapshot` | `take-snapshot` | Take snapshot, store, return metadata + URL |
| GET | `/api/v1/snapshots` | `list-snapshots` | List all snapshots (newest first) |
| GET | `/api/v1/snapshots/{snapshot_id}` | `get-snapshot` | Get snapshot metadata |
| GET | `/api/v1/snapshots/{snapshot_id}/image` | *(raw chi)* | Serve raw image bytes |
| DELETE | `/api/v1/snapshots/{snapshot_id}` | `delete-snapshot` | Delete snapshot + file |

#### Storage

Flat directory (configurable via `SNAPSHOT_DIR`, default `./snapshots`):

```
snapshots/
  {uuid}.png     <- image file
  {uuid}.json    <- metadata sidecar
```

#### Quick Curl

```bash
# Take a PNG snapshot
curl -s -X POST http://127.0.0.1:8188/api/v1/chart/{id}/snapshot \
  -H 'Content-Type: application/json' -d '{"format":"png"}' | jq

# Take a JPEG snapshot with quality
curl -s -X POST http://127.0.0.1:8188/api/v1/chart/{id}/snapshot \
  -H 'Content-Type: application/json' -d '{"format":"jpeg","quality":"0.85"}' | jq

# List snapshots
curl -s http://127.0.0.1:8188/api/v1/snapshots | jq

# Get metadata
curl -s http://127.0.0.1:8188/api/v1/snapshots/{uuid} | jq

# Fetch raw image
curl -s http://127.0.0.1:8188/api/v1/snapshots/{uuid}/image -o test.png

# Delete
curl -s -X DELETE http://127.0.0.1:8188/api/v1/snapshots/{uuid}
```

## Quick Curl

```bash
curl -s http://127.0.0.1:8188/health
curl -s http://127.0.0.1:8188/api/v1/charts
```
