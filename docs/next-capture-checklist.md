# Next Capture Checklist (Coverage Lift Pass)

Use this checklist to raise runtime evidence for static-only candidate areas.

## Goal

Increase validated capability coverage for:

- `alerts`
- `broker`
- `order`
- `drawing`
- `strategy`
- `position`
- `dom`

## Pre-Run Setup

- [x] Ensure `.env` points to your active CDP browser (`CHROMIUM_CDP_ADDRESS`, `CHROMIUM_CDP_PORT`).
- [x] Start browser with your authenticated profile:
  - `just start-browser`
- [x] Start passive capture:
  - `just run-researcher`
- [x] Open TradingView chart tab and keep it active during flows.

## Guided Action Flows

Perform each flow slowly and visibly so network/runtime events are emitted.

### 1) Alerts Flow

- [x] Open Alerts panel/dialog.
- [x] Create a new alert for current symbol.
- [x] Edit alert condition/value/expiration.
- [x] Enable/disable alert.
- [x] Delete one alert.
- [x] Switch alert tabs/filters if present.

### 2) Order Ticket Flow

- [ ] Open trading/order ticket UI.
- [ ] Toggle order type (market/limit/stop if available).
- [ ] Change quantity, TP/SL fields.
- [ ] Open confirmation/review dialog.
- [ ] Cancel before final submission if needed for safety.

### 3) Broker Flow

- [ ] Open broker connection/integration panel.
- [ ] Navigate broker list/details.
- [ ] Open connect/disconnect UI (no destructive confirmation needed).
- [ ] Switch between broker-related subviews.

### 4) Drawing Tools Flow

- [x] Add at least 3 drawing objects (line, trendline, rectangle/shape).
- [x] Edit style/settings for one drawing.
- [x] Move/resize drawing objects.
- [x] Delete a drawing object.
- [x] Open drawing manager/tool panel if available.

### 5) Position/DOM Flow

- [ ] Open position widget/panel.
- [ ] Expand/collapse position rows or details.
- [ ] Open DOM panel.
- [ ] Interact with DOM depth view controls (zoom/aggregation/settings).

### 6) Strategy/Replay Flow

- [x] Open replay controls.
- [x] Start/stop replay and change speed/step.
- [ ] Open strategy/backtesting settings.
- [ ] Modify one strategy parameter.
- [ ] Apply/save parameter changes (if safe).

## Session Quality Boosters

- [x] Change symbols at least 2 times during session.
- [x] Change timeframe at least 3 times.
- [x] Toggle at least 2 chart layouts/panels.
- [ ] Keep session running for at least 10-15 minutes.

## Stop + Process

- [ ] Stop researcher with `Ctrl+C` after flows complete.
- [ ] Review captured data in `research_data/<date>/` directories.

## Safety Notes

- Keep all interactions passive/observational where possible.
- Avoid submitting real orders unless explicitly intended.
- Use cancellation/review screens for trading-related flows.
