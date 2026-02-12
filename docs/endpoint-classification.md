# Endpoint Classification

129 API endpoints classified by implementation mechanism.

## Implementation Types

| Type | Count | Description |
|------|-------|-------------|
| JS API call | 62 | Calls `window.TradingViewApi` methods directly via `evalOnChart` / `evalOnAnyChart` |
| Webpack internal | 15 | Accesses webpack module cache via `webpackChunktradingview.push()` trick |
| Keyboard shortcut | 14 | CDP `Input.dispatchKeyEvent` key presses |
| JS internal REST | 9 | Injects `fetch()` from page context against TradingView's private REST endpoints |
| Mixed | 7 | Combines 2+ techniques (DOM + click, keyboard + text input, JS + navigation) |
| File I/O | 4 | Local snapshot storage reads/writes |
| Static | 2 | Hardcoded responses (health, docs) |
| DOM manipulation | 2 | Reads DOM elements directly |
| CDP protocol | 1 | Raw CDP command (`Page.reload`) |
| CDP target listing | 1 | Enumerates browser targets via CDP |
| JS API + File I/O | 1 | JS captures canvas data, Go writes to disk |

## Endpoints by Feature Area

### Infrastructure

| # | Method | Path | Type | Mechanism |
|---|--------|------|------|-----------|
| 1 | GET | `/health` | Static | Returns `{"status":"ok"}` |
| 2 | GET | `/docs` | Static | Returns Swagger UI HTML |

### Charts (tag: Charts)

| # | Method | Path | Type | Mechanism |
|---|--------|------|------|-----------|
| 3 | GET | `/api/v1/charts` | CDP target listing | `cdp.listTargets()` — enumerates browser tabs, filters by URL, extracts chart IDs |
| 4 | GET | `/api/v1/charts/active` | JS API call | `api.chartsCount()`, `api.activeChartIndex()` |

### Symbol (tag: Symbol)

| # | Method | Path | Type | Mechanism |
|---|--------|------|------|-----------|
| 5 | GET | `/api/v1/chart/{id}/symbol` | JS API call | `chart.symbol()` |
| 6 | PUT | `/api/v1/chart/{id}/symbol` | JS API call | `chart.setSymbol()` |
| 7 | GET | `/api/v1/chart/{id}/symbol/info` | JS API call | `chart.symbolExt()` |

### Resolution (tag: Resolution)

| # | Method | Path | Type | Mechanism |
|---|--------|------|------|-----------|
| 8 | GET | `/api/v1/chart/{id}/resolution` | JS API call | `chart.resolution()` |
| 9 | PUT | `/api/v1/chart/{id}/resolution` | JS API call | `chart.setResolution()` |

### Action (tag: Action)

| # | Method | Path | Type | Mechanism |
|---|--------|------|------|-----------|
| 10 | POST | `/api/v1/chart/{id}/action` | JS API call | `api.executeActionById()` |

### Studies (tag: Studies)

| # | Method | Path | Type | Mechanism |
|---|--------|------|------|-----------|
| 11 | GET | `/api/v1/chart/{id}/studies` | JS API call | `chart.getAllStudies()` |
| 12 | POST | `/api/v1/chart/{id}/studies` | JS API call | `await chart.createStudy()` |
| 13 | DELETE | `/api/v1/chart/{id}/studies/{sid}` | JS API call | `chart.removeEntity()` |
| 14 | GET | `/api/v1/chart/{id}/studies/{sid}` | JS API call | `chart.getStudyById().getInputValues()` |
| 15 | PATCH | `/api/v1/chart/{id}/studies/{sid}` | JS API call | `study.mergeUp()` / `study.setInputValues()` |

### Watchlists (tag: Watchlists)

| # | Method | Path | Type | Mechanism |
|---|--------|------|------|-----------|
| 16 | GET | `/api/v1/watchlists` | JS internal REST | `fetch("/api/v1/symbols_list/all/")` from page context |
| 17 | GET | `/api/v1/watchlists/active` | JS internal REST | `fetch("/api/v1/symbols_list/active/")` |
| 18 | PUT | `/api/v1/watchlists/active` | JS internal REST | `fetch("/api/v1/symbols_list/active/{id}/", {method:"POST"})` |
| 19 | POST | `/api/v1/watchlists` | JS internal REST | `fetch("/api/v1/symbols_list/custom/", {method:"POST"})` |
| 20 | GET | `/api/v1/watchlist/{wid}` | JS internal REST | `fetch("/api/v1/symbols_list/all/")` then filter |
| 21 | PATCH | `/api/v1/watchlist/{wid}` | JS internal REST | `fetch("/api/v1/symbols_list/custom/{id}/rename/", {method:"POST"})` |
| 22 | DELETE | `/api/v1/watchlist/{wid}` | JS internal REST | `fetch("/api/v1/symbols_list/custom/{id}/", {method:"DELETE"})` |
| 23 | POST | `/api/v1/watchlist/{wid}/symbols` | JS internal REST | `fetch("/api/v1/symbols_list/custom/{id}/append/", {method:"POST"})` |
| 24 | DELETE | `/api/v1/watchlist/{wid}/symbols` | JS internal REST | `fetch("/api/v1/symbols_list/custom/{id}/remove/", {method:"POST"})` |
| 25 | POST | `/api/v1/watchlist/{wid}/flag` | Mixed (DOM + React fiber) **[EXPERIMENTAL]** | Finds `[data-name='symbol-list-wrap']`, walks `__reactFiber` tree, calls `markSymbol()`. Fragile — `__reactFiber` key scanning + `markSymbol` prop walk breaks on React version upgrades. No REST alternative exists. |

### Navigation (tag: Navigation)

| # | Method | Path | Type | Mechanism |
|---|--------|------|------|-----------|
| 26 | POST | `/api/v1/chart/{id}/zoom` | JS API call | `api.executeActionById("chartZoomIn"/"chartZoomOut")` |
| 27 | POST | `/api/v1/chart/{id}/scroll` | JS API call | `chart.scrollChartByBar()` / iterated `executeActionById` |
| 28 | POST | `/api/v1/chart/{id}/scroll/to-realtime` | JS API call | `chart.scrollToRealtime()` |
| 29 | POST | `/api/v1/chart/{id}/go-to-date` | JS API call | `chart.goToDate()` / `chart.setVisibleRange()` |
| 30 | GET | `/api/v1/chart/{id}/visible-range` | JS API call | `chart.getVisibleRange()` |
| 31 | PUT | `/api/v1/chart/{id}/visible-range` | JS API call | `await chart.setVisibleRange()` |
| 32 | POST | `/api/v1/chart/{id}/reset-scales` | JS API call | `chart.resetScales()` |

### ChartAPI (tag: ChartAPI)

| # | Method | Path | Type | Mechanism |
|---|--------|------|------|-----------|
| 33 | GET | `/api/v1/chart/{id}/chart-api/probe` | JS API call | Introspects `api.chartApi()` singleton |
| 34 | GET | `/api/v1/chart/{id}/chart-api/probe/deep` | JS API call | Deep introspection of `chartApi()` methods and state |
| 35 | GET | `/api/v1/chart/{id}/chart-api/resolve-symbol` | JS API call | `capi.resolveSymbol(sessionId, requestId, symbol, callback)` |
| 36 | PUT | `/api/v1/chart/{id}/chart-api/timezone` | JS API call | `chart.setTimezone()` |

### Replay (tag: Replay)

| # | Method | Path | Type | Mechanism |
|---|--------|------|------|-----------|
| 37 | GET | `/api/v1/chart/{id}/replay/scan` | JS API call | Inspects `api._replayApi` methods and DOM replay button |
| 38 | POST | `/api/v1/chart/{id}/replay/activate` | JS API call | `await rapi.selectDate(date)` |
| 39 | POST | `/api/v1/chart/{id}/replay/activate/auto` | JS API call | `await rapi.selectFirstAvailableDate()` |
| 40 | POST | `/api/v1/chart/{id}/replay/deactivate` | JS API call | `rapi.stopReplay()` / `rapi.leaveReplay()` |
| 41 | GET | `/api/v1/chart/{id}/replay/probe` | JS API call | Introspects `api._replayApi` singleton |
| 42 | GET | `/api/v1/chart/{id}/replay/probe/deep` | JS API call | Deep introspection of `_replayApi` (34 method targets) |
| 43 | GET | `/api/v1/chart/{id}/replay/status` | JS API call | `rapi.isReplayAvailable()`, `rapi.isReplayStarted()`, WatchedValues |
| 44 | POST | `/api/v1/chart/{id}/replay/start` | JS API call | `await rapi.selectDate(point)` |
| 45 | POST | `/api/v1/chart/{id}/replay/stop` | JS API call | `rapi.stopReplay()` |
| 46 | POST | `/api/v1/chart/{id}/replay/step` | JS API call | `rapi.doStep()` |
| 47 | POST | `/api/v1/chart/{id}/replay/autoplay/start` | JS API call | `rapi.toggleAutoplay()` (if not running) |
| 48 | POST | `/api/v1/chart/{id}/replay/autoplay/stop` | JS API call | `rapi.toggleAutoplay()` (if running) |
| 49 | POST | `/api/v1/chart/{id}/replay/reset` | JS API call | `rapi.goToRealtime()` |
| 50 | PUT | `/api/v1/chart/{id}/replay/autoplay/delay` | JS API call | `rapi.changeAutoplayDelay(delay)` |

### Strategy (tag: Strategy)

| # | Method | Path | Type | Mechanism |
|---|--------|------|------|-----------|
| 51 | GET | `/api/v1/chart/{id}/strategy/probe` | JS API call | Introspects `api._backtestingStrategyApi` singleton |
| 52 | GET | `/api/v1/chart/{id}/strategy/list` | JS API call | `bsa.allStrategies` WatchedValue |
| 53 | GET | `/api/v1/chart/{id}/strategy/active` | JS API call | `bsa.activeStrategy`, `bsa.activeStrategyInputsValues` WatchedValues |
| 54 | PUT | `/api/v1/chart/{id}/strategy/active` | JS API call | `bsa.setActiveStrategy(id)` |
| 55 | PUT | `/api/v1/chart/{id}/strategy/input` | JS API call | `bsa.setStrategyInput(name, value)` |
| 56 | GET | `/api/v1/chart/{id}/strategy/report` | JS API call | `bsa.activeStrategyReportData` WatchedValue |
| 57 | GET | `/api/v1/chart/{id}/strategy/date-range` | JS API call | `bsa.getChartDateRange()` |
| 58 | POST | `/api/v1/chart/{id}/strategy/goto` | JS API call | `bsa.gotoDate(ts, belowBar)` |

### Alerts (tag: Alerts)

| # | Method | Path | Type | Mechanism |
|---|--------|------|------|-----------|
| 59 | GET | `/api/v1/chart/{id}/alerts/scan` | Webpack internal | Extracts `__webpack_require__` via `webpackChunktradingview.push()`, scans module cache for `getAlertsRestApi` |
| 60 | GET | `/api/v1/chart/{id}/alerts/probe` | Webpack internal | Introspects `getAlertsRestApi()` singleton via webpack cache |
| 61 | GET | `/api/v1/chart/{id}/alerts/probe/deep` | Webpack internal | Deep introspection of 18 method targets on singleton |
| 62 | GET | `/api/v1/alerts` | Webpack internal | `await aapi.listAlerts()` |
| 63 | GET | `/api/v1/alerts/{aid}` | Webpack internal | `await aapi.getAlerts({alert_ids: ids})` |
| 64 | POST | `/api/v1/alerts` | Webpack internal | `await aapi.createAlert(params)` |
| 65 | PUT | `/api/v1/alerts/{aid}` | Webpack internal | `await aapi.modifyRestartAlert(params)` |
| 66 | DELETE | `/api/v1/alerts` | Webpack internal | `await aapi.deleteAlerts({alert_ids: ids})` |
| 67 | POST | `/api/v1/alerts/stop` | Webpack internal | `await aapi.stopAlerts({alert_ids: ids})` |
| 68 | POST | `/api/v1/alerts/restart` | Webpack internal | `await aapi.restartAlerts({alert_ids: ids})` |
| 69 | POST | `/api/v1/alerts/clone` | Webpack internal | `await aapi.cloneAlerts({alert_ids: ids})` |
| 70 | GET | `/api/v1/alerts/fires` | Webpack internal | `await aapi.listFires()` |
| 71 | DELETE | `/api/v1/alerts/fires` | Webpack internal | `await aapi.deleteFires({fire_ids: ids})` |
| 72 | DELETE | `/api/v1/alerts/fires/all` | Webpack internal | `await aapi.deleteAllFires()` |

### Drawings (tag: Drawings)

| # | Method | Path | Type | Mechanism |
|---|--------|------|------|-----------|
| 73 | GET | `/api/v1/chart/{id}/drawings` | JS API call | `chart.getAllShapes()` |
| 74 | GET | `/api/v1/chart/{id}/drawings/{sid}` | JS API call | `chart.getShapeById(id).getProperties()` |
| 75 | POST | `/api/v1/chart/{id}/drawings` | JS API call | `await chart.createShape(point, opts)` |
| 76 | POST | `/api/v1/chart/{id}/drawings/multipoint` | JS API call | `await chart.createMultipointShape(points, opts)` |
| 77 | POST | `/api/v1/chart/{id}/drawings/{sid}/clone` | JS API call | `chart.cloneLineTool(id)` |
| 78 | DELETE | `/api/v1/chart/{id}/drawings/{sid}` | JS API call | `chart.removeEntity(id, opts)` |
| 79 | DELETE | `/api/v1/chart/{id}/drawings` | JS API call | `chart.removeAllShapes()` |
| 80 | GET | `/api/v1/chart/{id}/drawings/toggles` | JS API call | `api.hideAllDrawingTools()`, `api.lockAllDrawingTools()`, `api.magnetEnabled()` WatchedValues |
| 81 | PUT | `/api/v1/chart/{id}/drawings/toggles/hide` | JS API call | `api.hideAllDrawingTools().setValue(val)` |
| 82 | PUT | `/api/v1/chart/{id}/drawings/toggles/lock` | JS API call | `api.lockAllDrawingTools().setValue(val)` |
| 83 | PUT | `/api/v1/chart/{id}/drawings/toggles/magnet` | JS API call | `api.magnetEnabled().setValue()` + `api.magnetMode().setValue()` |
| 84 | PUT | `/api/v1/chart/{id}/drawings/{sid}/visibility` | JS API call | `chart.setEntityVisibility(id, vis)` |
| 85 | GET | `/api/v1/chart/{id}/drawings/tool` | JS API call | `api.selectedLineTool()` WatchedValue |
| 86 | PUT | `/api/v1/chart/{id}/drawings/tool` | JS API call | `await api.selectLineTool(tool)` |
| 87 | POST | `/api/v1/chart/{id}/drawings/{sid}/z-order` | JS API call | `shape.bringToFront()` / `.sendToBack()` |
| 88 | GET | `/api/v1/chart/{id}/drawings/state` | JS API call | `chart.getLineToolsState()` |
| 89 | PUT | `/api/v1/chart/{id}/drawings/state` | JS API call | `await chart.applyLineToolsState(dto)` |

### Snapshots (tag: Snapshots)

| # | Method | Path | Type | Mechanism |
|---|--------|------|------|-----------|
| 90 | POST | `/api/v1/chart/{id}/snapshot` | JS API + File I/O | JS: `await api.takeClientScreenshot()` + `canvas.toDataURL()`. Go: decode base64, write to `snapshot.Store` |
| 91 | GET | `/api/v1/snapshots` | File I/O | Read metadata from `snapshot.Store` |
| 92 | GET | `/api/v1/snapshots/{sid}` | File I/O | Read single metadata from `snapshot.Store` |
| 93 | GET | `/api/v1/snapshots/{sid}/image` | File I/O | Serve raw image bytes from `snapshot.Store` |
| 94 | DELETE | `/api/v1/snapshots/{sid}` | File I/O | Delete file from `snapshot.Store` |

### Page (tag: Page)

| # | Method | Path | Type | Mechanism |
|---|--------|------|------|-----------|
| 95 | POST | `/api/v1/page/reload` | CDP protocol | `Page.reload` with `ignoreCache`, polls `document.readyState`, refreshes tab registry |

### Pine Editor (tag: Pine Editor)

| # | Method | Path | Type | Mechanism |
|---|--------|------|------|-----------|
| 96 | POST | `/api/v1/pine/toggle` | Mixed (DOM + CDP trusted click) | JS locates button coords via DOM query, CDP `Input.dispatchMouseEvent` clicks at coords, JS polls for state change |
| 97 | GET | `/api/v1/pine/status` | DOM manipulation | Checks `.monaco-editor` visibility, `.view-lines` children, `.tv-spinner--shown` overlay |
| 98 | GET | `/api/v1/pine/source` | Mixed (Webpack + DOM) | Primary: webpack-discovered Monaco `model.getValue()`. Fallback: reads `.view-line` DOM elements |
| 99 | PUT | `/api/v1/pine/source` | Webpack internal | Monaco via webpack: `model.setValue(source)` |
| 100 | POST | `/api/v1/pine/save` | Mixed (DOM + Keyboard) | JS focuses Monaco textarea, CDP sends Ctrl+S (`Input.dispatchKeyEvent`, modifiers=2, keyCode=83) |
| 101 | POST | `/api/v1/pine/add-to-chart` | Mixed (DOM + Keyboard + DOM click) | JS focuses textarea, CDP sends Ctrl+Enter, JS polls for confirmation dialog, clicks "Save and add to chart" |
| 102 | GET | `/api/v1/pine/console` | DOM manipulation | Reads `[class*="console"] [class*="message"]` selectors |
| 103 | POST | `/api/v1/pine/undo` | Keyboard shortcut | Ctrl+Z (key="z", keyCode=90, modifiers=2) |
| 104 | POST | `/api/v1/pine/redo` | Keyboard shortcut | Ctrl+Shift+Z (key="Z", keyCode=90, modifiers=10) |
| 105 | POST | `/api/v1/pine/new-indicator` | Keyboard shortcut | Chord: Ctrl+K then Ctrl+I |
| 106 | POST | `/api/v1/pine/new-strategy` | Keyboard shortcut | Chord: Ctrl+K then Ctrl+S |
| 107 | POST | `/api/v1/pine/open-script` | Mixed (Keyboard + text + DOM + CDP click) | Ctrl+O opens dialog, `Input.insertText` types name, JS clicks result item, CDP click dismisses |
| 108 | POST | `/api/v1/pine/find-replace` | Webpack internal | Monaco via webpack: `model.findMatches()` + `model.pushEditOperations()` |
| 109 | POST | `/api/v1/pine/go-to-line` | Mixed (Keyboard + text) | Ctrl+G opens dialog, `Input.insertText` types line number, Enter confirms |
| 110 | POST | `/api/v1/pine/delete-line` | Keyboard shortcut | Ctrl+Shift+K (key="K", keyCode=75, modifiers=10), repeatable |
| 111 | POST | `/api/v1/pine/move-line` | Keyboard shortcut | Alt+ArrowUp/Down (modifiers=1), repeatable |
| 112 | POST | `/api/v1/pine/toggle-comment` | Keyboard shortcut | Ctrl+/ (key="/", keyCode=191, modifiers=2) |
| 113 | POST | `/api/v1/pine/toggle-console` | Keyboard shortcut | Ctrl+` (key="`", keyCode=192, modifiers=2) |
| 114 | POST | `/api/v1/pine/insert-line` | Keyboard shortcut | Ctrl+Shift+Enter (keyCode=13, modifiers=10) |
| 115 | POST | `/api/v1/pine/new-tab` | Keyboard shortcut | Shift+Alt+T (key="T", keyCode=84, modifiers=9) |
| 116 | POST | `/api/v1/pine/command-palette` | Keyboard shortcut | F1 (key="F1", keyCode=112, modifiers=0) |

### Layout (tag: Layout)

| # | Method | Path | Type | Mechanism |
|---|--------|------|------|-----------|
| 117 | GET | `/api/v1/layouts` | JS API call | `api._loadChartService.state().value().chartList` |
| 118 | GET | `/api/v1/layout/status` | JS API call | `api.layoutName()`, `api.layoutId()`, `api.layout()`, `api.chartsCount()`, `api.activeChartIndex()`, fullscreen/maximize WatchedValues, `hasChanges()` |
| 119 | POST | `/api/v1/layout/switch` | Mixed (JS API + Navigation) | JS resolves URL from `_loadChartService`, sets `window.location.href`, Go invalidates sessions, polls `document.readyState` |
| 120 | POST | `/api/v1/layout/save` | JS API call | `await api.getSaveChartService().saveChartSilently()` |
| 121 | POST | `/api/v1/layout/clone` | JS API call | `await api.getSaveChartService()._saveAsController._doCloneCurrentLayout(name)` |
| 122 | POST | `/api/v1/layout/rename` | JS API call | `api._chartWidgetCollection.metaInfo.name.setValue(name)` then `saveChartSilently()` |
| 123 | POST | `/api/v1/layout/grid` | JS API call | `api.setLayout(template)` |
| 124 | POST | `/api/v1/layout/fullscreen` | JS API call | `api.startFullscreen()` / `api.exitFullscreen()` |
| 125 | POST | `/api/v1/layout/dismiss-dialog` | Keyboard shortcut | Escape via CDP `Input.dispatchKeyEvent` (keyCode=27, modifiers=0) |

### Chart Navigation (tag: Chart Navigation)

| # | Method | Path | Type | Mechanism |
|---|--------|------|------|-----------|
| 126 | POST | `/api/v1/chart/next` | Keyboard shortcut | Tab (keyCode=9, modifiers=0), then JS reads active chart |
| 127 | POST | `/api/v1/chart/prev` | Keyboard shortcut | Shift+Tab (keyCode=9, modifiers=8), then JS reads active chart |
| 128 | POST | `/api/v1/chart/maximize` | Keyboard shortcut | Alt+Enter (keyCode=13, modifiers=1), then JS reads layout status |
| 129 | POST | `/api/v1/chart/activate` | JS API call | `api.setActiveChart(index)` |

## Fragility Notes

| Type | Fragility | Affected Endpoints |
|------|-----------|-------------------|
| JS API call | Low — `window.TradingViewApi` is a stable public-facing charting library API | 62 endpoints |
| Webpack internal | High — module IDs and internal singletons can change on any TradingView deploy | Alerts (59-72), Pine source (98-99), Pine find-replace (108) |
| JS internal REST | Medium — TradingView's private REST paths could change but are versioned (`/api/v1/`) | Watchlists (16-24) |
| Keyboard shortcut | Low — standard keyboard shortcuts rarely change | Pine shortcuts (103-116), Chart nav (126-128), Dismiss (125) |
| DOM manipulation | High — CSS class names and DOM structure change frequently | Pine status (97), Pine console (102), Flag symbol (25) |
| CDP trusted click | Medium — requires correct button coordinates from DOM, but click mechanism is stable | Pine toggle (96), Pine open-script (107) |
| Navigation | Low — URL patterns (`/chart/{id}/`) are stable | Layout switch (119) |
| File I/O | None — local storage, no external dependencies | Snapshots (91-94) |
| CDP protocol | None — standard CDP commands | Page reload (95) |
