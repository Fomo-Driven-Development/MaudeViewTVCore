# Mapper Pipeline Runbook

This runbook is for a new operator to run the full passive capture + mapper pipeline without inspecting code.

## Prerequisites

- Go 1.22+ installed (`go version`).
- `just` installed (or run the equivalent `scripts/*.sh` directly).
- Chromium or Chrome available on PATH (`chromium-browser`, `chromium`, or `google-chrome`) for CDP.
- TradingView session available in the launched browser profile (if authenticated captures are required).

## Environment Variables

Configuration is loaded from `.env` (copied from `example.env`).

Required for browser start:

- `CHROMIUM_CDP_ADDRESS` (example: `127.0.0.1`)
- `CHROMIUM_CDP_PORT` (example: `9220`)
- `CHROMIUM_START_URL` (example: `https://www.tradingview.com/`)

Required for pipeline output location:

- `RESEARCHER_DATA_DIR` (default: `./research_data`)

Commonly used researcher toggles:

- `RESEARCHER_TAB_URL_FILTER` (default: `tradingview.com`)
- `RESEARCHER_RELOAD_ON_ATTACH` (default: `true`)
- `RESEARCHER_CAPTURE_HTTP` (default: `true`)
- `RESEARCHER_CAPTURE_WS` (default: `true`)
- `RESEARCHER_CAPTURE_STATIC` (default: `true`)

## Command Order

Run in this order from repository root.

1. Start browser with CDP enabled (leave running):

```bash
just start-browser
```

2. Run passive researcher and perform the browsing/session activity you want to capture. Stop with `Ctrl+C` when done:

```bash
just run-researcher
```

3. Run mapper pipeline stages (static analysis -> runtime probes -> correlation -> reporting):

```bash
just mapper-full
```

Expected stdout order:

- `static-analysis: complete`
- `runtime-probes: complete`
- `correlation: complete`
- `reporting: complete`

4. Validate mapper artifacts:

```bash
just mapper-validate
```

Expected stdout:

- `validation: complete`

## Expected `research_data/` Layout

Output is date-partitioned (`YYYY-MM-DD`).

```text
research_data/
  2026-02-11/
    chart_<tabid8>/
      http/
        <tabid8>.jsonl
      websocket/
        <tabid8>.jsonl
      resources/
        js/
          *.js
        css|img|font|media|docs|manifest|other/
          *
    mapper/
      static-analysis/
        js-bundle-index.jsonl
        js-bundle-analysis.jsonl
        js-bundle-analysis-errors.jsonl
        js-bundle-dependency-graph.jsonl
      runtime-probes/
        runtime-trace.jsonl
        trace-sessions.jsonl
      correlation/
        capability-correlations.jsonl
      reporting/
        capability-matrix.jsonl
        capability-matrix.schema.json
        capability-matrix-summary.md
```

Notes:

- `chart_<tabid8>` names are generated from the first 8 chars of the CDP target ID.
- Mapper output files are written under each date root discovered in `RESEARCHER_DATA_DIR`.

## Passive-Only Guardrails

- The researcher captures network/resource traffic and writes artifacts; it does not submit orders or execute trading actions.
- Runtime probes only inject passive wrappers around `fetch`, `XMLHttpRequest`, and `WebSocket` to emit telemetry.
- No synthetic clicks, keyboard input, or navigation automation is performed by mapper stages.
- Keep `CHROMIUM_CDP_ADDRESS=127.0.0.1` so CDP is localhost-only.

## Known Limitations (Minified Bundle Analysis)

- Static extraction is regex-based (not full JS AST/control-flow), so minified/obfuscated bundles reduce semantic fidelity.
- Parsing requires balanced delimiters; malformed or truncated minified files are written to `js-bundle-analysis-errors.jsonl`.
- Only `.js` files under `/resources/js/` are indexed; inline scripts and source maps are not analyzed.
- Dependency resolution only links relative imports/requires (`./`, `../`) with simple `.js`/`index.js` matching.
- Domain hints and signal anchors are heuristic keyword/string-literal matches and can produce false positives/negatives.

## Troubleshooting

- `mapper failed: ... artifact missing`: ensure capture exists for that date and rerun `just run-researcher` before mapper stages.
- CDP connection errors in runtime probes: restart browser with `just start-browser` and verify `.env` CDP host/port match.
- Empty captures: confirm `RESEARCHER_TAB_URL_FILTER` matches the tab URL and that activity occurred before stopping researcher.
