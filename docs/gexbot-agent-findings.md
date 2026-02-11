# gexbot-agent Findings

## Repository Summary

`gexbot-agent` is a Go-based automation layer specialized for `gexbot.com`, with:

- a REST control API (`cmd/gexbot-agent`)
- an MCP server that wraps the REST API (`cmd/mcp-server`)
- a passive network researcher (`cmd/researcher`)

Primary focus: route-aware SPA navigation, chart manipulation, playback control, annotations, and snapshots.

## Runtime Components

- `cmd/gexbot-agent/main.go`
  - attaches to Chromium tab via CDP
  - initializes navigator + snapshot store
  - starts REST API (default `:8088`)
- `cmd/mcp-server/main.go`
  - exposes streamable HTTP MCP endpoint (default `:8089`)
  - MCP tools proxy to REST API
- `cmd/researcher/main.go`
  - passive HTTP/WebSocket/resource capture to disk

## MCP Capabilities

MCP surface is intentionally narrower and maps closely to API endpoints:

- route tools: list/switch/info
- diagnostics: page status, reload
- snapshot capture
- chart discovery/scales/duration/full-day/reset
- annotation CRUD
- playback/history controls (load/clear/play/pause/seek/step/speed/live)

Tool registration: `internal/mcptools/*.go`.

## REST API Capabilities

Core endpoint groups (`internal/api/*.go`):

- `/api/v1/route`, `/api/v1/route/info`
- `/api/v1/page-status`, `/api/v1/reload`
- `/api/v1/snapshot` (+ raw PNG fetch by ID)
- `/api/v1/chart/*` for scales/duration/discovery/reset
- `/api/v1/chart/annotations*`
- `/api/v1/playback/*`

API is built with Huma + Chi (`internal/api/server.go`).

## Domain Modeling and Validation

Strong route configuration model:

- embedded `routes.yaml` defines valid routes, tickers, and category hash templates
- validation helpers enforce route/ticker/category constraints
- hash parsing maps current URL state back to canonical category names

Key files:

- `internal/api/validate.go`
- `internal/api/routes.yaml`
- `internal/api/info.go` + `routeinfo.yaml` for contextual route documentation

## CDP/Navigator Design

`Navigator` encapsulates browser control:

- attach to target tab filtered by domain
- SPA navigation via pushState + popstate
- two-step cross-route navigation for reliability
- bootstrap injection to expose SPA stores on `window.__gexbot__`
- screenshot capture and page diagnostics

Key file: `internal/cdp/navigator.go`.

## Capture and Storage (Researcher)

Researcher mode mirrors the passive-capture pattern:

- HTTP/WS capture filtered by `RESEARCHER_DOMAIN_FILTER` (default `gexbot.com`)
- JSONL output by route path segment + data type
- static resources saved as raw files

Key files:

- `internal/capture/http.go`
- `internal/capture/websocket.go`
- `internal/storage/*.go`

## Practical Strengths

- focused scope with lower operational complexity than `tv_scraper`
- clean route/ticker/category validation model
- strong fit for agent workflows that need deterministic page context

## Risks / Complexity Hotspots

- still coupled to site internals (React/Zustand/uPlot/Chart.js hooks)
- playback behavior depends on route-specific client-side state mechanics
- no DB analytics layer by default, so advanced persistence/reporting requires extra design

