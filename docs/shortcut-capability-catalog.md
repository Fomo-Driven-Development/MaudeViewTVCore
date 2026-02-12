# TradingView Shortcut Capability Catalog

> Derived from TradingView keyboard shortcuts page (captured 2026-02-11).
> Use as a feature planning checklist — check items as they are implemented or explicitly skipped.

## How to Use

- `[ ]` = not yet evaluated
- `[x]` = implemented or covered
- `[-]` = explicitly skipped / not applicable

---

## 1. Chart

- [ ] Quick search
- [ ] Open indicators
- [ ] Open data window
- [ ] Load Chart Layout
- [ ] Save Chart Layout
- [ ] Change symbol
- [ ] Change interval
- [ ] Switch between sessions
- [ ] Move chart 1 bar to the left
- [ ] Move chart 1 bar to the right
- [ ] Move further to the left
- [ ] Move further to the right
- [ ] Move chart to the first bar
- [ ] Move chart to the last bar
- [ ] Move chart to left/right
- [ ] Zoom in
- [ ] Zoom out
- [ ] Focused zoom
- [ ] Replay play/pause
- [ ] Replay step forward
- [ ] Reset chart view
- [ ] Toggle maximize chart
- [ ] Toggle maximize pane
- [ ] Go to date
- [ ] Enable/disable logarithmic series scale
- [ ] Enable/disable percent series scale
- [ ] Invert series scale

### Deferred — REST API Endpoints (Layout Storage)

Service: `charts-storage.tradingview.com`

| Method | Path | Description |
|--------|------|-------------|
| GET | `/charts-storage/get/layout/{layout_id}/sources` | Load chart layout |
| PUT | `/charts-storage/layout/{layout_id}/sources` | Save chart layout |

## 2. Indicators & Drawings

- [ ] Open indicators
- [ ] New indicator
- [ ] Clone a drawing
- [ ] Move a drawing horizontally or vertically
- [ ] Move a point
- [ ] Move selected drawing up
- [ ] Move selected drawing down
- [ ] Move selected drawing left
- [ ] Move selected drawing right
- [ ] Hide all drawings
- [ ] Lock all drawings
- [ ] Remove drawings
- [ ] Drawings multiselect
- [ ] Keep drawing mode
- [ ] Hold hotkey for temporary drawing
- [ ] Show Object Tree
- [ ] Show Hidden Tools
- [ ] Drawing a straight line at angles of 45
- [ ] Switch between cells in Table drawing object
- [ ] Magnet Mode

### Deferred — REST API Endpoints (Drawing Templates & Study Templates)

Service: `www.tradingview.com`

| Method | Path | Description |
|--------|------|-------------|
| GET | `/drawing-templates/{LineToolType}/` | Get drawing templates by type |
| GET | `/api/v1/study-templates` | List all study templates (custom/standard/fundamentals) |
| GET | `/api/v1/study-templates/{id}` | Get individual study template detail |

Drawing template types observed: `LineToolTrendLine`, `LineToolHorzLine`, `LineToolFibRetracement`, `LineToolFibChannel`, `LineToolRectangle`, `LineToolCrossLine`, `LineToolBrush`, `LineTool5PointsPattern`, `LineToolRiskRewardLong`.

## 3. Watchlist

- [ ] Change watchlist
- [ ] Next symbol
- [ ] Previous symbol
- [ ] Flag/unflag symbol
- [ ] Select all symbols
- [ ] Select next symbol
- [ ] Select previous symbol

### Deferred
- [ ] Get/set watchlist columns
- [ ] Import watchlist
- [ ] Export watchlist
- [ ] Reorder symbols
- [ ] Sort watchlist by criteria

## 4. Screener

- [ ] Add new filter
- [ ] Show/hide filters

## 5. Pine Script Editor

### Deferred — REST API Endpoints (Pine Facade)

Service: `pine-facade.tradingview.com`

| Method | Path | Description |
|--------|------|-------------|
| GET | `/pine-facade/list/` | List all available Pine Scripts |
| GET | `/pine-facade/get/{script_id}/{version}` | Fetch script source code |
| GET | `/pine-facade/get_script_info/` | Get script metadata |
| GET | `/pine-facade/is_auth_to_get/{script_id}/{version}` | Check script access authorization |
| GET | `/pine-facade/versions/{script_id}/last` | Get latest version info |

Service: `www.tradingview.com`

| Method | Path | Description |
|--------|------|-------------|
| POST | `/pubscripts-get/` | Get published public scripts |

### General

- [ ] Toggle console
- [ ] Open Pine Script reference
- [ ] Save script
- [ ] Open script
- [ ] Rename script
- [ ] Command Palette

### Basic Editing

- [ ] Cut
- [ ] Copy
- [ ] Paste
- [ ] Undo
- [ ] Redo
- [ ] Delete line
- [ ] Toggle line comment
- [ ] Add line comment
- [ ] Remove line comment

### Navigation

- [ ] Go to Line/Column

### Search and Replace

- [ ] Find
- [ ] Replace
- [ ] Select all occurrences of Find match

### Multi-cursor and Selection

- [ ] Add cursor above
- [ ] Add cursor below
- [ ] Select all occurrences

### Rich Languages Editing

- [ ] Fold (collapse) block
- [ ] Unfold (uncollapse) block
- [ ] Fold all
- [ ] Unfold all
- [ ] Go to definition

## 6. Trading

- [ ] Open/close Order Panel
- [ ] Open/close DOM
- [ ] Place limit order
- [ ] Place limit order to buy
- [ ] Place limit order to sell
- [ ] Click in DOM cell

## 7. Alerts

- [ ] Add alert
- [ ] Open Edit alert dialog
- [ ] Save changes in Edit alert dialog
- [ ] Remove alert without confirmation
- [ ] Select alert/event
- [ ] Select all alerts/events
- [ ] Next alert/event
- [ ] Previous alert/event

### Deferred — REST API Endpoints (Alerts)

Service: `pricealerts.tradingview.com`

| Method | Path | Description |
|--------|------|-------------|
| GET | `/list_alerts` | List user's active alerts |
| POST | `/create_alert` | Create new alert |
| POST | `/delete_alerts` | Delete alerts |
| POST | `/modify_restart_alert` | Update and restart alert |
| POST | `/stop_alerts` | Pause alerts |
| POST | `/restart_alerts` | Resume alerts |
| POST | `/list_fires` | Get triggered alert events |
| POST | `/get_offline_fires` | Get offline-triggered events |
| POST | `/get_offline_fire_controls` | Get notification controls |
| POST | `/clear_offline_fires` | Clear offline fire history |
| POST | `/delete_all_fires` | Clear all fire history |

Service: `crud-storage.tradingview.com`

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/alert_preset/` | Get alert presets |

## 8. News

> Not part of the original keyboard shortcuts page but discovered in captured traffic.

### Deferred — REST API Endpoints (News)

Service: `news-mediator.tradingview.com`

| Method | Path | Description |
|--------|------|-------------|
| GET | `/public/news-flow/v2/news` | Get news feed |
| GET | `/public/news/v1/story` | Get individual story details |

## 9. Desktop App

- [ ] New tab
- [ ] Close tab
- [ ] Reopen closed tab/window

---

## Source

Extracted from captured JS bundle `support-wizard-shortcut-page.01ac5180a38d793d2dd8.js` and English locale files in `research_data/2026-02-11/chart_rzWLrz7t/resources/js/`.
