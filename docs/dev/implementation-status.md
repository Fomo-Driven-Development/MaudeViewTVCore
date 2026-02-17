# Implementation Status

177 controller API endpoints across 9 feature areas, built on CDP browser automation with in-page JavaScript evaluation.

![Coverage Map](chart_coverage.png)

## Summary

| Feature Area | File | Endpoints |
|---|---|---|
| Charts | `server_chart.go` | 28 |
| Misc (health, strategy, snapshots, currency, hotlists) | `server_misc.go` | 28 |
| Pine Editor | `server_pine.go` | 21 |
| Layout | `server_layout.go` | 19 |
| Drawings | `server_drawing.go` | 20 |
| Studies & Indicators | `server_study.go` | 15 |
| Watchlists | `server_watchlist.go` | 15 |
| Replay | `server_replay.go` | 14 |
| Alerts | `server_alert.go` | 14 |
| **Total** | | **174** |

Note: 3 additional endpoints (health, docs at root level) bring the total to 177.

## Endpoints by Feature Area

### Infrastructure

| Method | Path | Type | Mechanism |
|--------|------|------|-----------|
| GET | `/health` | Static | Returns `{"status":"ok"}` |
| GET | `/api/v1/health/deep` | JS API call | Browser connection, tab state, chart readiness checks |

### Charts

| Method | Path | Type | Mechanism |
|--------|------|------|-----------|
| GET | `/api/v1/charts` | CDP target listing | `cdp.listTargets()` — enumerates browser tabs, filters by URL |
| GET | `/api/v1/charts/active` | JS API call | `api.chartsCount()`, `api.activeChartIndex()` |

### Symbol

| Method | Path | Type | Mechanism |
|--------|------|------|-----------|
| GET | `/api/v1/chart/{id}/symbol` | JS API call | `chart.symbol()` |
| PUT | `/api/v1/chart/{id}/symbol` | JS API call | `chart.setSymbol()` |
| GET | `/api/v1/chart/{id}/symbol/info` | JS API call | `chart.symbolExt()` |

### Resolution

| Method | Path | Type | Mechanism |
|--------|------|------|-----------|
| GET | `/api/v1/chart/{id}/resolution` | JS API call | `chart.resolution()` |
| PUT | `/api/v1/chart/{id}/resolution` | JS API call | `chart.setResolution()` |

### Chart Type

| Method | Path | Type | Mechanism |
|--------|------|------|-----------|
| GET | `/api/v1/chart/{id}/chart-type` | JS API call | Reads current chart type |
| PUT | `/api/v1/chart/{id}/chart-type` | JS API call | Switch between candles, bars, line, area, Heikin Ashi, etc. |

### Currency & Unit

| Method | Path | Type | Mechanism |
|--------|------|------|-----------|
| GET | `/api/v1/chart/{id}/currency` | JS API call | Get price denomination currency |
| PUT | `/api/v1/chart/{id}/currency` | JS API call | Set price denomination currency |
| GET | `/api/v1/chart/{id}/currency/available` | JS API call | List available currencies |
| GET | `/api/v1/chart/{id}/unit` | JS API call | Get display unit |
| PUT | `/api/v1/chart/{id}/unit` | JS API call | Set display unit |
| GET | `/api/v1/chart/{id}/unit/available` | JS API call | List available units |

### Action

| Method | Path | Type | Mechanism |
|--------|------|------|-----------|
| POST | `/api/v1/chart/{id}/action` | JS API call | `api.executeActionById()` |

### Navigation

| Method | Path | Type | Mechanism |
|--------|------|------|-----------|
| POST | `/api/v1/chart/{id}/zoom` | JS API call | `api.executeActionById("chartZoomIn"/"chartZoomOut")` |
| POST | `/api/v1/chart/{id}/scroll` | JS API call | `chart.scrollChartByBar()` |
| POST | `/api/v1/chart/{id}/reset-view` | CDP keyboard | Alt+R (Reset chart view) |
| POST | `/api/v1/chart/{id}/go-to-date` | JS API call | `chart.goToDate()` / `chart.setVisibleRange()` |
| GET | `/api/v1/chart/{id}/visible-range` | JS API call | `chart.getVisibleRange()` |
| PUT | `/api/v1/chart/{id}/visible-range` | JS API call | `await chart.setVisibleRange()` |
| PUT | `/api/v1/chart/{id}/timeframe` | JS API call | `chart.setTimeFrame()` with period presets (1D, 5D, 1M, YTD, 1Y, etc.) |
| POST | `/api/v1/chart/{id}/reset-scales` | JS API call | `chart.resetScales()` |
| POST | `/api/v1/chart/{id}/undo` | CDP keyboard | Ctrl+Z |
| POST | `/api/v1/chart/{id}/redo` | CDP keyboard | Ctrl+Y |

### Chart Toggles

| Method | Path | Type | Mechanism |
|--------|------|------|-----------|
| GET | `/api/v1/chart/{id}/toggles` | JS API call | Read log scale, auto scale, extended hours WatchedValues |
| POST | `/api/v1/chart/{id}/toggles/log-scale` | CDP keyboard | Alt+L |
| POST | `/api/v1/chart/{id}/toggles/auto-scale` | CDP keyboard | Alt+A |
| POST | `/api/v1/chart/{id}/toggles/extended-hours` | CDP keyboard | Alt+E |

### ChartAPI (Introspection)

| Method | Path | Type | Mechanism |
|--------|------|------|-----------|
| GET | `/api/v1/chart/{id}/chart-api/probe` | JS API call | Introspects `api.chartApi()` singleton |
| GET | `/api/v1/chart/{id}/chart-api/probe/deep` | JS API call | Deep introspection of `chartApi()` methods and state |
| GET | `/api/v1/chart/{id}/chart-api/resolve-symbol` | JS API call | `capi.resolveSymbol()` |
| PUT | `/api/v1/chart/{id}/chart-api/timezone` | JS API call | `chart.setTimezone()` |
| POST | `/api/v1/chart/{id}/data-window/probe` | JS API call | Discover data window state |

### Multi-Pane

| Method | Path | Type | Mechanism |
|--------|------|------|-----------|
| GET | `/api/v1/chart/{id}/panes` | JS API call | List chart panes |
| POST | `/api/v1/chart/next` | Keyboard shortcut | Tab |
| POST | `/api/v1/chart/prev` | Keyboard shortcut | Shift+Tab |
| POST | `/api/v1/chart/maximize` | Keyboard shortcut | Alt+Enter |
| POST | `/api/v1/chart/activate` | JS API call | `api.setActiveChart(index)` |

### Studies & Indicators

| Method | Path | Type | Mechanism |
|--------|------|------|-----------|
| GET | `/api/v1/chart/{id}/studies` | JS API call | `chart.getAllStudies()` |
| POST | `/api/v1/chart/{id}/studies` | JS API call | `await chart.createStudy()` |
| GET | `/api/v1/chart/{id}/studies/{sid}` | JS API call | `chart.getStudyById().getInputValues()` |
| PATCH | `/api/v1/chart/{id}/studies/{sid}` | JS API call | `study.mergeUp()` / `study.setInputValues()` |
| DELETE | `/api/v1/chart/{id}/studies/{sid}` | JS API call | `chart.removeEntity()` |
| POST | `/api/v1/chart/{id}/compare` | JS API call | Add overlay/compare symbol |
| GET | `/api/v1/chart/{id}/compare` | JS API call | List active comparisons |
| DELETE | `/api/v1/chart/{id}/compare/{sid}` | JS API call | Remove comparison |
| POST | `/api/v1/chart/{id}/indicators/search` | JS API call | Search indicator library |
| POST | `/api/v1/chart/{id}/indicators/add` | JS API call | Add indicator by name |
| GET | `/api/v1/chart/{id}/indicators/favorites` | JS API call | List favorite indicators |
| POST | `/api/v1/chart/{id}/indicators/favorite` | JS API call | Toggle indicator favorite |
| GET | `/api/v1/indicators/probe-dom` | DOM manipulation | Probe indicator dialog DOM |
| GET | `/api/v1/study-templates` | JS internal REST | List saved study templates |
| GET | `/api/v1/study-templates/{id}` | JS internal REST | Get study template detail |

### Watchlists

| Method | Path | Type | Mechanism |
|--------|------|------|-----------|
| GET | `/api/v1/watchlists` | JS internal REST | `fetch("/api/v1/symbols_list/all/")` |
| GET | `/api/v1/watchlists/active` | JS internal REST | `fetch("/api/v1/symbols_list/active/")` |
| PUT | `/api/v1/watchlists/active` | JS internal REST | Set active watchlist |
| POST | `/api/v1/watchlists` | JS internal REST | Create watchlist |
| GET | `/api/v1/watchlist/{wid}` | JS internal REST | Get single watchlist |
| PATCH | `/api/v1/watchlist/{wid}` | JS internal REST | Rename watchlist |
| DELETE | `/api/v1/watchlist/{wid}` | JS internal REST | Delete watchlist |
| POST | `/api/v1/watchlist/{wid}/symbols` | JS internal REST | Add symbols |
| DELETE | `/api/v1/watchlist/{wid}/symbols` | JS internal REST | Remove symbols |
| POST | `/api/v1/watchlist/{wid}/flag` | Mixed (DOM + React fiber) | `markSymbol()` via `__reactFiber` tree walk **[EXPERIMENTAL]** |
| GET | `/api/v1/watchlists/colored` | JS internal REST | List colored watchlists |
| PUT | `/api/v1/watchlists/colored/{color}` | JS internal REST | Replace color list |
| POST | `/api/v1/watchlists/colored/{color}/append` | JS internal REST | Add symbols to color list |
| POST | `/api/v1/watchlists/colored/{color}/remove` | JS internal REST | Remove symbols from color list |
| POST | `/api/v1/watchlists/colored/bulk-remove` | JS internal REST | Remove symbols from all colored lists |

### Drawings

| Method | Path | Type | Mechanism |
|--------|------|------|-----------|
| GET | `/api/v1/chart/{id}/drawings` | JS API call | `chart.getAllShapes()` |
| GET | `/api/v1/chart/{id}/drawings/{sid}` | JS API call | `chart.getShapeById(id).getProperties()` |
| POST | `/api/v1/chart/{id}/drawings` | JS API call | `await chart.createShape(point, opts)` |
| POST | `/api/v1/chart/{id}/drawings/multipoint` | JS API call | `await chart.createMultipointShape(points, opts)` |
| POST | `/api/v1/chart/{id}/drawings/tweet` | Webpack internal | `createTweetLineToolByUrl(url, model)` — fetches tweet data from TV backend |
| POST | `/api/v1/chart/{id}/drawings/{sid}/clone` | JS API call | `chart.cloneLineTool(id)` |
| DELETE | `/api/v1/chart/{id}/drawings/{sid}` | JS API call | `chart.removeEntity(id, opts)` |
| DELETE | `/api/v1/chart/{id}/drawings` | JS API call | `chart.removeAllShapes()` |
| GET | `/api/v1/chart/{id}/drawings/toggles` | JS API call | Read hide/lock/magnet WatchedValues |
| PUT | `/api/v1/chart/{id}/drawings/toggles/hide` | JS API call | `api.hideAllDrawingTools().setValue(val)` |
| PUT | `/api/v1/chart/{id}/drawings/toggles/lock` | JS API call | `api.lockAllDrawingTools().setValue(val)` |
| PUT | `/api/v1/chart/{id}/drawings/toggles/magnet` | JS API call | `api.magnetEnabled().setValue()` + `api.magnetMode().setValue()` |
| PUT | `/api/v1/chart/{id}/drawings/{sid}/visibility` | JS API call | `chart.setEntityVisibility(id, vis)` |
| GET | `/api/v1/chart/{id}/drawings/tool` | JS API call | `api.selectedLineTool()` |
| PUT | `/api/v1/chart/{id}/drawings/tool` | JS API call | `await api.selectLineTool(tool)` |
| POST | `/api/v1/chart/{id}/drawings/{sid}/z-order` | JS API call | `shape.bringToFront()` / `.sendToBack()` |
| GET | `/api/v1/chart/{id}/drawings/state` | JS API call | `chart.getLineToolsState()` |
| PUT | `/api/v1/chart/{id}/drawings/state` | JS API call | `await chart.applyLineToolsState(dto)` |
| GET | `/api/v1/drawings/shapes` | JS API call | List available shape types |
| POST | `/api/v1/chart/{id}/tools/measure` | JS API call | Select measure tool |
| POST | `/api/v1/chart/{id}/tools/zoom` | JS API call | Select zoom tool |
| POST | `/api/v1/chart/{id}/tools/eraser` | JS API call | Select eraser tool |
| POST | `/api/v1/chart/{id}/tools/cursor` | JS API call | Select cursor tool |

### Replay

| Method | Path | Type | Mechanism |
|--------|------|------|-----------|
| GET | `/api/v1/chart/{id}/replay/scan` | JS API call | Inspects `api._replayApi` methods |
| POST | `/api/v1/chart/{id}/replay/activate` | JS API call | `await rapi.selectDate(date)` |
| POST | `/api/v1/chart/{id}/replay/activate/auto` | JS API call | `await rapi.selectFirstAvailableDate()` |
| POST | `/api/v1/chart/{id}/replay/deactivate` | JS API call | `rapi.stopReplay()` / `rapi.leaveReplay()` |
| GET | `/api/v1/chart/{id}/replay/probe` | JS API call | Introspects `api._replayApi` singleton |
| GET | `/api/v1/chart/{id}/replay/probe/deep` | JS API call | Deep introspection (34 method targets) |
| GET | `/api/v1/chart/{id}/replay/status` | JS API call | `isReplayAvailable()`, `isReplayStarted()`, WatchedValues |
| POST | `/api/v1/chart/{id}/replay/start` | JS API call | `await rapi.selectDate(point)` |
| POST | `/api/v1/chart/{id}/replay/stop` | JS API call | `rapi.stopReplay()` |
| POST | `/api/v1/chart/{id}/replay/step` | JS API call | `rapi.doStep()` |
| POST | `/api/v1/chart/{id}/replay/autoplay/start` | JS API call | `rapi.toggleAutoplay()` (if not running) |
| POST | `/api/v1/chart/{id}/replay/autoplay/stop` | JS API call | `rapi.toggleAutoplay()` (if running) |
| POST | `/api/v1/chart/{id}/replay/reset` | JS API call | `rapi.goToRealtime()` |
| PUT | `/api/v1/chart/{id}/replay/autoplay/delay` | JS API call | `rapi.changeAutoplayDelay(delay)` |

### Strategy

| Method | Path | Type | Mechanism |
|--------|------|------|-----------|
| GET | `/api/v1/chart/{id}/strategy/probe` | JS API call | Introspects `api._backtestingStrategyApi` |
| GET | `/api/v1/chart/{id}/strategy/list` | JS API call | `bsa.allStrategies` WatchedValue |
| GET | `/api/v1/chart/{id}/strategy/active` | JS API call | `bsa.activeStrategy`, inputs WatchedValues |
| PUT | `/api/v1/chart/{id}/strategy/active` | JS API call | `bsa.setActiveStrategy(id)` |
| PUT | `/api/v1/chart/{id}/strategy/input` | JS API call | `bsa.setStrategyInput(name, value)` |
| GET | `/api/v1/chart/{id}/strategy/report` | JS API call | `bsa.activeStrategyReportData` WatchedValue |
| GET | `/api/v1/chart/{id}/strategy/date-range` | JS API call | `bsa.getChartDateRange()` |
| POST | `/api/v1/chart/{id}/strategy/goto` | JS API call | `bsa.gotoDate(ts, belowBar)` |

### Alerts

| Method | Path | Type | Mechanism |
|--------|------|------|-----------|
| GET | `/api/v1/chart/{id}/alerts/scan` | Webpack internal | Extracts `__webpack_require__` via `webpackChunktradingview.push()` |
| GET | `/api/v1/chart/{id}/alerts/probe` | Webpack internal | Introspects `getAlertsRestApi()` singleton |
| GET | `/api/v1/chart/{id}/alerts/probe/deep` | Webpack internal | Deep introspection of 18 method targets |
| GET | `/api/v1/alerts` | Webpack internal | `await aapi.listAlerts()` |
| GET | `/api/v1/alerts/{aid}` | Webpack internal | `await aapi.getAlerts({alert_ids: ids})` |
| POST | `/api/v1/alerts` | Webpack internal | `await aapi.createAlert(params)` |
| PUT | `/api/v1/alerts/{aid}` | Webpack internal | `await aapi.modifyRestartAlert(params)` |
| DELETE | `/api/v1/alerts` | Webpack internal | `await aapi.deleteAlerts({alert_ids: ids})` |
| POST | `/api/v1/alerts/stop` | Webpack internal | `await aapi.stopAlerts({alert_ids: ids})` |
| POST | `/api/v1/alerts/restart` | Webpack internal | `await aapi.restartAlerts({alert_ids: ids})` |
| POST | `/api/v1/alerts/clone` | Webpack internal | `await aapi.cloneAlerts({alert_ids: ids})` |
| GET | `/api/v1/alerts/fires` | Webpack internal | `await aapi.listFires()` |
| DELETE | `/api/v1/alerts/fires` | Webpack internal | `await aapi.deleteFires({fire_ids: ids})` |
| DELETE | `/api/v1/alerts/fires/all` | Webpack internal | `await aapi.deleteAllFires()` |

### Snapshots

| Method | Path | Type | Mechanism |
|--------|------|------|-----------|
| POST | `/api/v1/chart/{id}/snapshot` | JS API + File I/O | `api.takeClientScreenshot()` + `canvas.toDataURL()`, Go writes to disk |
| POST | `/api/v1/browser_screenshot` | CDP protocol | Full browser viewport screenshot |
| GET | `/api/v1/snapshots` | File I/O | List snapshot metadata |
| GET | `/api/v1/snapshots/{sid}/metadata` | File I/O | Get single snapshot metadata |
| GET | `/api/v1/snapshots/{sid}/image` | File I/O | Serve raw image bytes |
| DELETE | `/api/v1/snapshots/{sid}` | File I/O | Delete snapshot + file |

### Page

| Method | Path | Type | Mechanism |
|--------|------|------|-----------|
| POST | `/api/v1/page/reload` | CDP protocol | `Page.reload` with `ignoreCache` |

### Layout

| Method | Path | Type | Mechanism |
|--------|------|------|-----------|
| GET | `/api/v1/layouts` | JS API call | `api._loadChartService.state().value().chartList` |
| GET | `/api/v1/layout/status` | JS API call | Layout name, ID, chart count, fullscreen/maximize state |
| POST | `/api/v1/layout/switch` | Mixed (JS + Navigation) | Resolves URL, sets `window.location.href`, invalidates sessions |
| POST | `/api/v1/layout/save` | JS API call | `await api.getSaveChartService().saveChartSilently()` |
| POST | `/api/v1/layout/clone` | JS API call | `_doCloneCurrentLayout(name)` |
| POST | `/api/v1/layout/rename` | JS API call | `metaInfo.name.setValue(name)` then save |
| DELETE | `/api/v1/layout/{id}` | JS API call | Delete layout |
| POST | `/api/v1/layout/grid` | JS API call | `api.setLayout(template)` |
| POST | `/api/v1/layout/fullscreen` | JS API call | `api.startFullscreen()` / `api.exitFullscreen()` |
| POST | `/api/v1/layout/dismiss-dialog` | Keyboard shortcut | Escape |
| POST | `/api/v1/layouts/batch-delete` | JS API call | Batch delete multiple layouts |
| POST | `/api/v1/layout/preview` | JS API call | Get layout preview |
| GET | `/api/v1/layout/favorite` | JS API call | Get favorite status |
| POST | `/api/v1/layout/favorite/toggle` | JS API call | Toggle favorite |

### Pine Editor

| Method | Path | Type | Mechanism |
|--------|------|------|-----------|
| POST | `/api/v1/pine/toggle` | Mixed (DOM + CDP click) | JS locates button coords, CDP `Input.dispatchMouseEvent` clicks |
| GET | `/api/v1/pine/status` | DOM manipulation | Checks `.monaco-editor` visibility, spinner overlay |
| GET | `/api/v1/pine/source` | Mixed (Webpack + DOM) | Monaco `model.getValue()`, fallback reads `.view-line` DOM |
| PUT | `/api/v1/pine/source` | Webpack internal | Monaco `model.setValue(source)` |
| POST | `/api/v1/pine/save` | Mixed (DOM + Keyboard) | Focus Monaco textarea, CDP sends Ctrl+S |
| POST | `/api/v1/pine/add-to-chart` | Mixed (DOM + Keyboard + click) | Ctrl+Enter, polls for dialog, clicks confirm |
| GET | `/api/v1/pine/console` | DOM manipulation | Reads `[class*="console"]` selectors |
| POST | `/api/v1/pine/undo` | Keyboard shortcut | Ctrl+Z |
| POST | `/api/v1/pine/redo` | Keyboard shortcut | Ctrl+Shift+Z |
| POST | `/api/v1/pine/new-indicator` | Keyboard shortcut | Ctrl+K then Ctrl+I |
| POST | `/api/v1/pine/new-strategy` | Keyboard shortcut | Ctrl+K then Ctrl+S |
| POST | `/api/v1/pine/open-script` | Mixed (Keyboard + text + DOM + click) | Ctrl+O, type name, click result |
| POST | `/api/v1/pine/find-replace` | Webpack internal | Monaco `findMatches()` + `pushEditOperations()` |
| POST | `/api/v1/pine/go-to-line` | Mixed (Keyboard + text) | Ctrl+G, type line number, Enter |
| POST | `/api/v1/pine/delete-line` | Keyboard shortcut | Ctrl+Shift+K |
| POST | `/api/v1/pine/move-line` | Keyboard shortcut | Alt+ArrowUp/Down |
| POST | `/api/v1/pine/toggle-comment` | Keyboard shortcut | Ctrl+/ |
| POST | `/api/v1/pine/toggle-console` | Keyboard shortcut | Ctrl+` |
| POST | `/api/v1/pine/insert-line` | Keyboard shortcut | Ctrl+Shift+Enter |
| POST | `/api/v1/pine/new-tab` | Keyboard shortcut | Shift+Alt+T |
| POST | `/api/v1/pine/command-palette` | Keyboard shortcut | F1 |

### Hotlists

| Method | Path | Type | Mechanism |
|--------|------|------|-----------|
| GET | `/api/v1/hotlists/probe` | Webpack internal | Probe `hotlistsManager()` singleton |
| GET | `/api/v1/hotlists/probe/deep` | Webpack internal | Deep introspection |
| GET | `/api/v1/hotlists/markets` | Webpack internal | `hotlistsManager().getMarkets()` |
| GET | `/api/v1/hotlists/exchanges` | Webpack internal | List exchanges with groups |
| GET | `/api/v1/hotlists/{exchange}/{group}` | Webpack internal | `hotlistsManager().getOneHotlist()` |

## Implementation Mechanisms

| Type | Count | Fragility | Notes |
|------|-------|-----------|-------|
| JS API call | ~90 | Low | `window.TradingViewApi` — stable public-facing charting library API |
| Webpack internal | ~21 | **High** | Module IDs and internal singletons change on TradingView deploys |
| Keyboard shortcut | ~16 | Low | Standard shortcuts rarely change |
| JS internal REST | ~14 | Medium | TradingView's private REST paths versioned (`/api/v1/`) but could change |
| Mixed | ~10 | Medium | Combines 2+ techniques (DOM + click, keyboard + text) |
| File I/O | ~6 | None | Local snapshot storage |
| DOM manipulation | ~4 | **High** | CSS class names and DOM structure change frequently |
| CDP protocol | ~3 | None | Standard CDP commands |

### High-fragility endpoints to monitor

- **Alerts** (14 endpoints) — webpack-internal `getAlertsRestApi()` singleton
- **Pine source read/write** — webpack-discovered Monaco namespace
- **Pine status/console** — DOM class selectors (`[class*="console"]`, `.tv-spinner--shown`)
- **Flag symbol** — React fiber tree walk (`__reactFiber`)
- **Pine toggle** — DOM button selectors for CDP trusted click coordinates
- **Hotlists** (5 endpoints) — webpack-internal `hotlistsManager()` singleton
- **Tweet drawing** — webpack-internal `createTweetLineToolByUrl()` + TradingView backend fetch

## Coverage Gaps

Features visible in the TradingView UI that are **not** covered by the controller API.

| Feature | Location | Difficulty | Notes |
|---------|----------|------------|-------|
| Templates | Top bar (grid icon) | Medium | Save/load chart appearance presets. Dialog-driven. |
| Settings Gear | Top bar (far right) | Hard | Full chart properties dialog — colors, grid, scales, trading, events. Multi-tab. |
| Trade | Top bar | Hard | Opens broker integration / order panel. Requires connected broker account. |
| News | Right sidebar | Hard | Financial news feed. Separate data source. |
| Economic Calendar | Bottom panel | Hard | Events feed (FOMC, CPI, etc.). Separate data source. |
| Screener | Bottom panel | Hard | Stock/crypto screener with filters. |
| Trading Panel | Bottom panel | Hard | Order placement, positions, P&L. Requires broker connection. |
| Chart Properties | Modal dialog | Hard | Full settings dialog — appearance, scales, trading, events. |
| Object Tree | Right sidebar | Medium | Lists all drawings and studies in tree view. Already covered by `GET /drawings` and `GET /studies`. |

## Feature Checklist

Derived from TradingView keyboard shortcuts. Tracks what's automatable via the controller API.

### Chart

- [x] Change symbol
- [x] Change interval
- [x] Zoom in / Zoom out
- [x] Replay play/pause
- [x] Replay step forward
- [x] Reset chart view
- [x] Toggle maximize chart
- [x] Go to date
- [x] Enable/disable logarithmic series scale
- [x] Enable/disable percent series scale
- [ ] Quick search
- [ ] Open data window
- [ ] Load Chart Layout
- [ ] Save Chart Layout
- [ ] Switch between sessions
- [ ] Move chart 1 bar to the left/right
- [ ] Move further to the left/right
- [ ] Move chart to the first/last bar
- [ ] Focused zoom
- [ ] Toggle maximize pane
- [ ] Invert series scale

### Indicators & Drawings

- [x] Clone a drawing
- [x] Hide all drawings
- [x] Lock all drawings
- [x] Remove drawings
- [x] Magnet Mode
- [ ] Open indicators
- [ ] New indicator
- [ ] Move a drawing horizontally or vertically
- [ ] Move a point
- [ ] Move selected drawing up/down/left/right
- [ ] Drawings multiselect
- [ ] Keep drawing mode
- [ ] Hold hotkey for temporary drawing
- [ ] Show Object Tree
- [ ] Show Hidden Tools
- [ ] Drawing a straight line at angles of 45
- [ ] Switch between cells in Table drawing object

### Watchlist

- [x] Change watchlist
- [x] Flag/unflag symbol
- [ ] Next/previous symbol
- [ ] Select all symbols

### Alerts

- [x] Add alert
- [x] Remove alert
- [x] Stop/restart alert
- [x] Clone alert
- [ ] Open Edit alert dialog
- [ ] Save changes in Edit alert dialog
- [ ] Select/navigate alerts

### Pine Script Editor

- [x] Toggle console
- [x] Save script
- [x] Open script
- [x] Command Palette
- [x] Cut / Copy / Paste (via keyboard)
- [x] Undo / Redo
- [x] Delete line
- [x] Toggle line comment
- [x] Go to Line/Column
- [x] Find and Replace
- [x] Move line up/down
- [ ] Rename script
- [ ] Add cursor above/below
- [ ] Select all occurrences
- [ ] Fold/Unfold blocks
- [ ] Go to definition

### Trading

- [ ] Open/close Order Panel
- [ ] Open/close DOM
- [ ] Place limit order
- [ ] Click in DOM cell

### Screener

- [ ] Add new filter
- [ ] Show/hide filters

## Deferred TradingView REST APIs

Internal TradingView REST endpoints observed in traffic but not yet wrapped by the controller.

### Layout Storage (`charts-storage.tradingview.com`)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/charts-storage/get/layout/{layout_id}/sources` | Load chart layout data |
| PUT | `/charts-storage/layout/{layout_id}/sources` | Save chart layout data |

### Drawing Templates (`www.tradingview.com`)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/drawing-templates/{LineToolType}/` | Get drawing templates by type |

Types observed: `LineToolTrendLine`, `LineToolHorzLine`, `LineToolFibRetracement`, `LineToolFibChannel`, `LineToolRectangle`, `LineToolCrossLine`, `LineToolBrush`, `LineTool5PointsPattern`, `LineToolRiskRewardLong`.

### News (`news-mediator.tradingview.com`)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/public/news-flow/v2/news` | Get news feed |
| GET | `/public/news/v1/story` | Get individual story details |

### Alert Presets (`crud-storage.tradingview.com`)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/alert_preset/` | Get alert presets |
