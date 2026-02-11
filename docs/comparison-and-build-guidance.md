# Comparison and Build Guidance

## Side-by-Side Snapshot

## Scope

- `tv_scraper`
  - broad TradingView automation platform
  - very large MCP/API surface
  - optional DB analytics + replay persistence + UW integration + A2A bridge
- `gexbot-agent`
  - focused gexbot.com automation/control layer
  - smaller, tighter MCP/API surface
  - strong route/ticker/category modeling

## Architecture Style

- `tv_scraper`: feature-rich platform architecture (many modules, optional subsystems)
- `gexbot-agent`: thin-control architecture (navigator + route model + direct control endpoints)

## Data Plane

- `tv_scraper`
  - passive capture + optional relational/time-series analytics model
  - materialized views and replay statistics
- `gexbot-agent`
  - passive capture only (no built-in DB analytics domain)

## Agent Integration

- both use MCP
- `tv_scraper` adds A2A orchestration skills and broader workflow tooling
- `gexbot-agent` keeps MCP close to API primitives (simpler behavior model)

## What to Reuse for a “Similar but Different” Build

## Keep

1. Three-process model:
   - control API
   - MCP wrapper
   - passive researcher
2. CDP abstraction boundary:
   - low-level browser ops in `internal/cdp`
   - API/MCP only call typed methods
3. Route/config-driven validation:
   - embed YAML for routes/entities/categories/tickers
4. Snapshot subsystem:
   - stable UUID file storage and optional inline/base64 returns

## Change

1. Reduce initial API/MCP surface:
   - ship minimal vertical slice first
   - add tool families incrementally
2. Separate “volatile UI hooks” from “stable domain contracts”:
   - put JS selectors/hooks behind adapter interfaces
   - isolate per-route/per-widget JS snippets
3. Add explicit capability negotiation:
   - endpoints that return which route/features are currently available
4. Add test harness early:
   - golden tests for route parsing/validation
   - smoke tests for endpoint-to-CDP flows
   - replayable fixtures for MCP tool IO

## Suggested Initial Target Architecture

## Services

1. `controller` (REST)
   - route context
   - chart discovery/control
   - snapshot and diagnostics
2. `mcp-gateway`
   - tool registration + policy constraints
3. `researcher`
   - passive HTTP/WS capture to JSONL/resources
4. `analytics` (optional phase 2)
   - persistence and derived metrics/reporting

## Modules

1. `internal/config`
2. `internal/domain` (typed route/ticker/category models)
3. `internal/cdp` (browser adapter)
4. `internal/api`
5. `internal/mcptools`
6. `internal/capture` + `internal/storage`
7. `internal/snapshot`

## Recommended Phase Plan

1. Phase 1 (minimal)
   - route list/switch
   - page status/reload
   - snapshot (file + inline option)
   - MCP wrappers for those endpoints
2. Phase 2
   - chart discovery/scales
   - annotations
   - deterministic validation model
3. Phase 3
   - playback/history controls
   - passive researcher hardening
4. Phase 4 (optional)
   - DB-backed analytics and workflow-level composite tools

## Key Design Principle

Prefer `gexbot-agent`’s tighter API/MCP contract discipline for the base system, then selectively add `tv_scraper`-style advanced workflow capabilities only after core control paths are stable.

