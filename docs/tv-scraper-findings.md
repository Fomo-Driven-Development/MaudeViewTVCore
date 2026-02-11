# tv_scraper Findings

## Repository Summary

`tv_scraper` is a Go-based TradingView automation and data-capture platform with:

- a REST API service (`cmd/scraper`)
- an MCP server (`cmd/mcp-server`)
- optional database-backed replay analytics (Timescale/Postgres)
- optional Unusual Whales options data integration
- an A2A bridge server in Python (`a2a-server/`)

Primary focus: deep TradingView chart control + replay workflows + persistent analytics.

## Runtime Components

- `cmd/scraper/main.go`
  - connects to Chromium CDP
  - runs passive HTTP/WebSocket capture
  - starts REST API server
- `cmd/mcp-server/main.go`
  - connects to Chromium CDP in API-only mode
  - registers MCP tool suites
  - serves stdio or streamable HTTP transport
- `a2a-server/`
  - wraps MCP tool usage into higher-level “skills”
  - can attach chart PNG artifacts to A2A responses

## MCP Capabilities

MCP server registers broad tool coverage across navigation, viewport, data extraction, drawings, studies, layouts, trading overlays, alerts, watchlists, replay, DB persistence, and options data.

Tool families are implemented in `internal/mcp/tools_*.go`.

## REST API Capabilities

The scraper API exposes a broad `/api/v1/...` surface for chart operations, including:

- chart list/symbol/resolution/timezone/type
- viewport controls and date-range navigation
- studies/drawings/action APIs
- bars/sessions/market status/crosshair/readiness
- replay mode controls
- layout, panes, selection, theme, connection, sync
- order/position/execution shape overlays

Implementation: `internal/api/server.go`.

## Data Capture and Storage

Passive capture pipeline records:

- HTTP request/response traffic to JSONL
- WebSocket events to JSONL
- static assets (script/css/image/etc.) to raw files

Key files:

- `internal/capture/http.go`
- `internal/capture/websocket.go`
- `internal/storage/jsonl.go`
- `internal/storage/registry.go`

Writer organization: date-based directories + per-path/per-type writers with browser ID naming.

## Database and Analytics

Optional DB mode provides:

- watchlist ingestion tables + snapshots (`init-scripts/001_create_watchlists.sql`)
- replay analytics schema (`init-scripts/002_create_replay_analysis.sql`)
  - sessions/runs/checkpoints/filter rules
  - Timescale hypertables + compression policies
  - materialized views for leaderboard and win-rate matrices
  - refresh function `refresh_replay_views()`

DB persistence tools are exposed through MCP (`internal/mcp/tools_replay_db.go`).

## Notable Engineering Patterns

- clear separation: CDP primitives (`internal/cdp`) vs API/MCP interfaces
- composite MCP tools for multi-step workflows (`tools_composite.go`)
- optional integrations loaded conditionally (DB, UW client)
- extensive operational scripts and `justfile` recipes

## Practical Strengths

- very high automation coverage for TradingView
- supports both stateless control and stateful analytics workflows
- built for agent integration (MCP + A2A)

## Risks / Complexity Hotspots

- large tool surface increases maintenance and testing burden
- UI-coupled CDP/JS hooks can break when upstream site internals change
- replay and persistence logic spans many moving parts (browser state + DB state + workflow orchestration)

