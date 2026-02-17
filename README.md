<p align="center">
  <img src="assets/banner.jpg" alt="MaudeViewTVCore" width="600">
</p>

# MaudeViewTVCore

Browser automation for TradingView via Chrome DevTools Protocol.

Exposes a REST API for TradingView chart control — symbol, resolution, studies, drawings, alerts, watchlists, Pine editor, replay, snapshots, and layouts. Connects to a Chromium browser over CDP and evaluates JavaScript directly in TradingView pages. A separate researcher daemon passively captures network traffic for offline analysis.

**[Full Documentation](https://fomo-driven-development.github.io/MaudeViewTvDocs/)** — architecture, quickstart, all 177 endpoints, building agents, configuration, security.

## Prerequisites

- Go 1.24+
- Chromium (or Chrome) with remote debugging enabled
- [`just`](https://github.com/casey/just) command runner

## Quick Start

```bash
cp example.env .env            # configure CDP port, bind address, etc.
just run-tv-controller-with-browser  # launch browser + REST API server
```

Open <http://127.0.0.1:8188/docs> for interactive API documentation.

```bash
# List detected chart tabs
curl -s http://127.0.0.1:8188/api/v1/charts | jq
```

See the [Quick Start guide](https://fomo-driven-development.github.io/MaudeViewTvDocs/quickstart/) for the full walkthrough including agent setup.

## Security Warning

**CDP and the REST API have no authentication.** Always bind to `127.0.0.1` (the default). Use a dedicated browser profile — the controller can execute arbitrary JavaScript in any matched tab. Do not expose the CDP port or API to untrusted networks.

## Running Tests

```bash
go test ./...                                          # unit tests
just test-integration                                  # integration tests (requires running browser + controller)
```

## Documentation

- **[MaudeView Docs](https://fomo-driven-development.github.io/MaudeViewTvDocs/)** — architecture, quickstart, API reference, building agents, configuration, security
- [`docs/dev/implementation-status.md`](docs/dev/implementation-status.md) — all endpoints with mechanism details and coverage
- [`docs/dev/js-internals.md`](docs/dev/js-internals.md) — TradingView JS manager singletons reference
