# TradingView Functionality Catalog

Generated from latest artifacts in `research_data/2026-02-11`.

## Snapshot

- JS bundles indexed: `375`
- JS bundles parsed successfully: `297`
- JS parse errors: `78`
- Runtime trace events: `13`
- Runtime trace sessions: `5`
- Correlated capabilities: `6`
- Validated capability rows: `6`

## Validated Capabilities

These are capabilities with runtime evidence and correlation records.

| Capability | Confidence | Recommended Control Method | Recommended Path |
| --- | --- | --- | --- |
| `chart` | `0.85` | `network_drivable.http_api` | `runtime.api_request_path` |
| `replay` | `0.85` | `network_drivable.http_api` | `runtime.api_request_path` |
| `studies` | `0.85` | `network_drivable.http_api` | `runtime.api_request_path` |
| `trading` | `0.85` | `network_drivable.http_api` | `runtime.api_request_path` |
| `watchlist` | `0.85` | `network_drivable.http_api` | `runtime.api_request_path` |
| `widget` | `0.15` | `dom_drivable.entry_module` | `static.domain_chunk_entry_path` |

## Runtime Coverage Observed

Trace profiles captured:

- `chart_interaction`
- `replay_actions`
- `study_add_remove`
- `trading_panel_interactions`
- `watchlist_edits`

Runtime event surfaces observed:

- `trace_profile`: 5
- `xhr`: 2
- `websocket`: 2
- `fetch`: 2
- `event_bus`: 2

## Static Coverage Signals

Domain hints found in static graph:

- `trading`: 296
- `chart`: 43
- `widget`: 28
- `studies`: 18
- `replay`: 7
- `watchlist`: 5

Signal anchors found in static extraction:

- `action_event`: 91
- `api_route`: 76
- `feature_flag`: 42
- `websocket_channel`: 29

## Deferred REST API Endpoints (Captured in Traffic)

Internal TradingView REST API endpoints observed in HTTP captures, grouped under their validated capability category. These are ready for controller implementation using the same `fetch()` pattern as watchlist endpoints.

### `watchlist` — Colored Symbol Lists

Colored lists (Red, Orange, Green, Purple, Blue) appear in the watchlist listing with `type: "colored"`. These endpoints manage them separately from custom watchlists.

| Method | TradingView Path | Description |
|--------|-----------------|-------------|
| GET | `/api/v1/symbols_list/colored/` | List all colored lists with symbols |
| POST | `/api/v1/symbols_list/colored/{color}/replace/` | Replace entire color list |
| POST | `/api/v1/symbols_list/colored/{color}/append/` | Add symbols to color list |
| POST | `/api/v1/symbols_list/colored/{color}/remove/` | Remove symbols from color list |
| POST | `/api/v1/symbols_list/colored/bulk_remove/` | Remove symbols from all colored lists |

Response shape (list):
```json
[
  {
    "id": 16656903, "type": "colored", "name": "Green list",
    "symbols": ["TVC:VIX", "CBOE:UVXY"],
    "active": true, "shared": false, "color": "green",
    "description": null, "created": null, "modified": "2026-02-06T00:21:32Z"
  }
]
```

Colors observed: `red`, `orange`, `green`, `purple`, `blue`.

### `studies` — Study Templates

Template CRUD for saved study (indicator) configurations.

| Method | TradingView Path | Description |
|--------|-----------------|-------------|
| GET | `/api/v1/study-templates` | List all templates (custom/standard/fundamentals) |
| GET | `/api/v1/study-templates/{id}` | Get individual template detail |

Response shape (list):
```json
{
  "custom": [
    {"id": 62699809, "name": "0DTE", "meta_info": {"indicators": [{"id": "...", "description": "..."}], "interval": null}, "favorite_date": "2024-08-25T18:33:22Z"}
  ],
  "standard": [
    {"id": 1, "name": "Bill Williams' 3 Lines", "meta_info": {"indicators": [...], "interval": null}, "favorite_date": null}
  ],
  "fundamentals": [
    {"id": 14, "name": "Buyback and dividend yield", "meta_info": {"indicators": [...], "interval": null}, "favorite_date": null}
  ]
}
```

## Candidate Functionality (Static-Only So Far)

These areas were detected in chunk names/paths but are not yet validated as standalone capabilities in the matrix.

- `alerts` (7 chunks)
  - `alerts-error-presenter`
  - `alerts-fires-focus-handler`
  - `alerts-get-watchlists-states`
  - `alerts-rest-api`
  - `alerts-session`
- `broker` (2 chunks)
  - `modular-broker`
  - `replay-broker`
- `order` (2 chunks)
  - `open-payment-order-dialog-on-load`
  - `trading-order-ticket`
- `drawing` (1 chunk)
  - `drawing-toolbar`
- `strategy` (1 chunk)
  - `backtesting-replay-strategy-facade`
- `position` (1 chunk)
  - `position-widget`
- `dom` (1 chunk)
  - `dom-panel`

## Capability Status Model

- `validated`: present in `capability-matrix.jsonl` with runtime evidence and confidence.
- `candidate`: present in static analysis only; not yet elevated to capability matrix rows.

## Primary Gaps

- Runtime sample is thin (`13` events total), so coverage confidence is still narrow.
- `widget` remains low-confidence (`0.15`).
- `alerts`, `broker`, `order`, `drawing`, `strategy`, and `position` are static-only candidates.
- `78` bundle parse failures still reduce static breadth.

## Recommendation On Step 5

Yes, run step 5.

Run another focused capture pass that targets the static-only candidate areas and intentionally increases runtime surface activity.

Suggested focused flows for the next pass:

- Alerts creation/edit/delete and alert manager navigation.
- Order ticket workflows (open, modify fields, submit/cancel where safe).
- Broker-related panel interactions.
- Drawing tools creation/edit/remove.
- Position widget and DOM panel interactions.
- Strategy/backtesting configuration interactions.

Then rerun:

- `just mapper-full`
- `just mapper-validate`

## Source Artifacts

- `research_data/2026-02-11/mapper/static-analysis/js-bundle-index.jsonl`
- `research_data/2026-02-11/mapper/static-analysis/js-bundle-analysis.jsonl`
- `research_data/2026-02-11/mapper/static-analysis/js-bundle-analysis-errors.jsonl`
- `research_data/2026-02-11/mapper/static-analysis/js-bundle-dependency-graph.jsonl`
- `research_data/2026-02-11/mapper/runtime-probes/runtime-trace.jsonl`
- `research_data/2026-02-11/mapper/runtime-probes/trace-sessions.jsonl`
- `research_data/2026-02-11/mapper/correlation/capability-correlations.jsonl`
- `research_data/2026-02-11/mapper/reporting/capability-matrix.jsonl`
- `research_data/2026-02-11/mapper/reporting/capability-matrix-summary.md`
