# TradingView WebSocket Internals

Documented from live captures on 2026-02-21 using the tv_agent researcher daemon.

TradingView maintains three concurrent WebSocket connections per chart page.

## Connection Summary

| Name | URL | Traffic | Protocol | Purpose |
|------|-----|---------|----------|---------|
| **Chart Data** | `wss://prodata.tradingview.com/socket.io/websocket?from=chart%2F{id}%2F&date=...&type=chart` | Very high (~5k+ frames/min) | Custom `~m~` framing | Real-time OHLCV, quotes, studies, chart commands |
| **Private Feed** | `wss://pushstream.tradingview.com/message-pipe-ws/private_feed` | Low (~1-5 frames/min) | JSON | User-specific: alerts, fires, account events |
| **Public Feed** | `wss://pushstream.tradingview.com/message-pipe-ws/public` | Very low (<1 frame/min) | JSON | Broadcast: economic calendar events |

---

## 1. Chart Data Socket

**URL**: `wss://prodata.tradingview.com/socket.io/websocket?from=chart%2F{chartId}%2F&date={iso}&type=chart&auth={jwt}`

This is the main data pipeline for everything the chart renders. It uses a custom binary-ish framing protocol and carries the highest volume of all three connections.

### Wire Protocol: `~m~` Framing

Messages are framed as: `~m~<length>~m~<payload>` where `<length>` is the byte length of `<payload>` as a decimal string. Multiple messages can be concatenated in a single WebSocket frame:

```
~m~166~m~{"m":"qsd","p":["qs_session",{"n":"COINBASE:BTCUSD","s":"ok","v":{"lp":68483.99}}]}~m~161~m~{"m":"qsd","p":["qs_session",{"n":"NYSE:AAPL","s":"ok","v":{"lp":245.12}}]}
```

Heartbeats use the same framing: `~m~1~m~~h~` (keep-alive ping/pong).

### Message Format

All JSON payloads follow the same structure:

```json
{"m": "<message_type>", "p": [<session_id>, <data>]}
```

- `m` — message type string
- `p` — parameters array, first element is always the session ID

### Session Types

The chart socket multiplexes multiple sessions:

| Session Prefix | Purpose |
|----------------|---------|
| `cs_*` | Chart session — OHLCV series, studies, symbol resolution |
| `qs_multiplexer_full_*` | Full quote session — detailed tick data for active chart symbol |
| `qs_multiplexer_watchlist_*` | Watchlist quote session — streaming quotes for all watchlist symbols |
| `qs_multiplexer_simple_*` | Simple quote session — basic metadata for symbol lookups |

### Outgoing Messages (Client → Server)

| Message Type | Purpose | Example |
|---|---|---|
| `set_auth_token` | Authenticate with JWT | `{"m":"set_auth_token","p":["<jwt_token>"]}` |
| `set_locale` | Set language/region | `{"m":"set_locale","p":["en","US"]}` |
| `chart_create_session` | Create a chart session | `{"m":"chart_create_session","p":["cs_abc123",""]}` |
| `chart_delete_session` | Destroy a chart session | `{"m":"chart_delete_session","p":["cs_abc123"]}` |
| `switch_timezone` | Change chart timezone | `{"m":"switch_timezone","p":["cs_abc123","Etc/UTC"]}` |
| `resolve_symbol` | Resolve symbol to data source | `{"m":"resolve_symbol","p":["cs_abc123","sds_sym_1","={...}"]}` |
| `create_series` | Subscribe to OHLCV data | `{"m":"create_series","p":["cs_abc123","sds_1","s1","sds_sym_1","1D",300,""]}` |
| `modify_series` | Change resolution/range | `{"m":"modify_series","p":["cs_abc123","sds_1","s1","sds_sym_1","15",300,""]}` |
| `remove_series` | Unsubscribe from series | `{"m":"remove_series","p":["cs_abc123","sds_1"]}` |
| `create_study` | Add indicator to chart | `{"m":"create_study","p":["cs_abc123","st1","s1","sds_1","EMA@tv-basicstudies",{...}]}` |
| `remove_study` | Remove indicator | `{"m":"remove_study","p":["cs_abc123","st1"]}` |
| `quote_create_session` | Create a quote session | `{"m":"quote_create_session","p":["qs_abc123"]}` |
| `quote_set_fields` | Configure quote fields | `{"m":"quote_set_fields","p":["qs_abc123","ch","chp","lp","volume",...]}` |
| `quote_add_symbols` | Subscribe to symbol quotes | `{"m":"quote_add_symbols","p":["qs_abc123","COINBASE:BTCUSD"]}` |
| `quote_remove_symbols` | Unsubscribe from quotes | `{"m":"quote_remove_symbols","p":["qs_abc123","COINBASE:BTCUSD"]}` |
| `quote_fast_symbols` | Enable fast-path ticks | `{"m":"quote_fast_symbols","p":["qs_abc123","COINBASE:BTCUSD"]}` |
| `quote_hibernate_all` | Pause all quote streams | `{"m":"quote_hibernate_all","p":["qs_abc123"]}` |
| `request_more_tickmarks` | Load more historical ticks | `{"m":"request_more_tickmarks","p":["cs_abc123","sds_1",10]}` |
| `set_future_tickmarks_mode` | Enable/disable future bars | `{"m":"set_future_tickmarks_mode","p":["cs_abc123","sds_1",true]}` |

### Incoming Messages (Server → Client)

#### Quote Stream Data (`qsd`) — ~65% of all traffic

Real-time price updates for watched symbols. Each update is a sparse object with only changed fields:

```json
{
  "m": "qsd",
  "p": ["qs_multiplexer_watchlist_abc123", {
    "n": "COINBASE:BTCUSD",
    "s": "ok",
    "v": {
      "lp": 68483.99,
      "lp_time": 1771699596,
      "ch": 499.57,
      "chp": 0.73,
      "volume": 2498.63
    }
  }]
}
```

**Common `v` fields:**

| Field | Type | Description |
|-------|------|-------------|
| `lp` | float | Last price |
| `lp_time` | int (unix) | Last price timestamp |
| `ch` | float | Price change (absolute) |
| `chp` | float | Price change (percent) |
| `volume` | float | Current session volume |
| `bid` / `ask` | float | Best bid/ask |
| `bid_size` / `ask_size` | float | Bid/ask depth |
| `trade_loaded` | bool | Whether trade data is available |
| `short_name` | string | Ticker short name |
| `pro_name` | string | Full exchange:ticker |
| `type` | string | Asset type (stock, spot, commodity, cfd) |
| `typespecs` | string[] | Tags (crypto, common, cfd, defi) |
| `update_mode` | string | "streaming" or "delayed" |
| `provider_id` | string | Data provider (coinbase, ice, oanda) |

First update after subscription includes all metadata fields; subsequent updates are sparse (only changed values).

#### Data Update (`du`) — ~20% of all traffic

OHLCV bar updates for chart series. Sent on every tick that changes the current bar:

```json
{
  "m": "du",
  "p": ["cs_abc123", {
    "sds_1": {
      "s": [{
        "i": 301,
        "v": [1771699560.0, 68474.52, 68483.99, 68474.52, 68483.99, 0.35798]
      }],
      "ns": {"d": "", "indexes": "nochange"},
      "t": "s3",
      "lbs": {"bar_close_time": 1771699620}
    }
  }]
}
```

**Bar format** (`v` array): `[timestamp, open, high, low, close, volume]`

- `i` — bar index (position in the series)
- `lbs.bar_close_time` — unix timestamp when the current bar closes
- `t` — series type identifier
- `ns.indexes` — "nochange" means bar count unchanged (in-place update)

#### Other Incoming Messages

| Message Type | Frequency | Purpose |
|---|---|---|
| `quote_completed` | ~10% | Quote subscription acknowledgement or timeout (`{"s":"timeout"}`) |
| `symbol_resolved` | On symbol change | Full symbol metadata after `resolve_symbol` |
| `series_loading` | On series change | Series data is loading |
| `series_completed` | After loading | Series data fully loaded |
| `series_timeframe` | On timeframe change | New timeframe applied |
| `timescale_update` | On range change | Time axis updated with new marks |
| `study_loading` | On indicator add | Indicator computation started |
| `study_completed` | After computation | Indicator data ready |
| `study_deleted` | On indicator remove | Indicator removed confirmation |

---

## 2. Private Feed

**URL**: `wss://pushstream.tradingview.com/message-pipe-ws/private_feed`

User-specific notification channel. Low traffic, high value. This is where all alert lifecycle events flow. Messages are plain JSON (no `~m~` framing).

### Envelope Format

```json
{
  "id": <sequence_number>,
  "text": {
    "content": {
      "m": "<message_type>",
      "p": <payload>,
      "id": "<event_id>"
    }
  }
}
```

Or for some messages, `text.content` is a JSON string that must be parsed separately.

### Message Types — Alert Lifecycle

The complete lifecycle of an alert as seen on the private feed:

```
alerts_created → alerts_updated (active=true) → alert_running → alert_fired → event → fires_updated
                                                                                         ↓
                                                                           alerts_deleted / alert_deleted
```

#### `alerts_created`

Sent when a new alert is created. Contains full alert configuration.

```json
{
  "m": "alerts_created",
  "p": [{
    "alert_id": 4074332684,
    "message": "BTCUSD Crossing 68,580.27",
    "type": "price",
    "active": false,
    "symbol": "={\"symbol\":\"COINBASE:BTCUSD\",\"adjustment\":\"splits\",\"session\":\"us_regular\",\"currency-id\":\"USD\"}",
    "pro_symbol": "={...}",
    "resolution": "1",
    "complexity": "primitive",
    "create_time": "2026-02-21T18:54:24Z",
    "conditions": [{
      "type": "cross",
      "cross_interval": true,
      "frequency": "on_first_fire",
      "resolution": "1",
      "series": [
        {"type": "barset"},
        {"type": "value", "value": 68580.27}
      ]
    }],
    "condition": { ... },
    "email": false,
    "mobile_push": true,
    "popup": true,
    "web_hook": "https://...",
    "sound_file": "alert/nature/whale",
    "sound_duration": 3,
    "auto_deactivate": false,
    "expiration": null,
    "kinds": ["regular"],
    "name": null,
    "presentation_data": {
      "main_series": {
        "pricescale": 100,
        "formatter": "price",
        "type": "spot",
        "logo": {"logoid": "crypto/XTVCBTC", "logoid2": "country/US", "style": "pair"}
      }
    }
  }],
  "id": "6eci-322747263"
}
```

**Key fields:**
- `alert_id` — numeric ID (e.g. `4074332684`)
- `type` — `"price"` for line alerts, `"indicator"` for Pine-based alerts
- `conditions[].type` — `"cross"` for crossing alerts
- `conditions[].series` — what's being compared (barset = price, value = static level)
- `conditions[].frequency` — `"on_first_fire"`, `"every_time_per_bar"`, etc.
- Notification channels: `email`, `mobile_push`, `popup`, `web_hook`, `sound_file`

#### `alerts_updated`

Sent when an alert is modified or activated. Same shape as `alerts_created` but with updated fields. Notable: `active` changes from `false` → `true` after creation.

Also sent for indicator-based alerts that monitor multiple symbols via `symbolset_data`:

```json
{
  "m": "alerts_updated",
  "p": [{
    "alert_id": 2221250356,
    "name": "FDD Breakout 3D",
    "type": "indicator",
    "active": true,
    "symbolset_data": {
      "inactive_symbols": {
        "NYSE:QS": {"last_error": "study_error", "last_stop_reason": "error"},
        "NYSE:SPR": {"last_error": "study_error", "last_stop_reason": "error"}
      },
      "symbols": ["COINBASE:BTCUSD", "NASDAQ:NVDA", "NASDAQ:TSLA", ...]
    },
    "message": "BTCUSD Crossing 68,580.27"
  }]
}
```

#### `alert_running`

Sent when the server starts monitoring the alert. Contains operational metadata:

```json
{
  "m": "alert_running",
  "p": {
    "id": 4074332684,
    "exp": -62135596800,
    "inst_id": 1771700064287438,
    "user_id": 123456,
    "cross_int": true,
    "inf_exp": true,
    "deact": false,
    "active": true,
    "sym": "COINBASE:BTCUSD",
    "internal_sym": "={...}",
    "res": "1",
    "desc": "BTCUSD Crossing 68,580.27",
    "email": false,
    "sms": false,
    "telegram": false,
    "push": true,
    "snd": true,
    "popup": true,
    "ns_only_fire": false,
    "extra": "{...}"
  }
}
```

**Key fields:**
- `inst_id` — server-side instance ID for this running alert
- `inf_exp` — `true` if alert never expires
- `deact` — whether to auto-deactivate after firing
- `extra` — full alert condition definition as escaped JSON string (contains Pine state, band definitions, etc.)

#### `alert_fired` — THE TOAST TRIGGER

This is the message that causes the toast notification to appear. Sent the instant the alert condition is met:

```json
{
  "m": "alert_fired",
  "p": {
    "fire_id": 46723852620,
    "alert_id": 4074311184,
    "fire_time": "2026-02-21T18:46:44Z",
    "bar_time": "2026-02-21T18:46:00Z",
    "message": "BTCUSD Crossing 68,521.11",
    "symbol": "={\"symbol\":\"COINBASE:BTCUSD\",\"adjustment\":\"splits\",\"session\":\"us_regular\",\"currency-id\":\"USD\"}",
    "pro_symbol": "={...}",
    "resolution": "1",
    "kinds": ["regular"],
    "cross_interval": true,
    "popup": true,
    "sound_file": "alert/nature/whale",
    "sound_duration": 3,
    "name": null
  },
  "id": "pp74-39525540"
}
```

**Key fields:**
- `fire_id` — unique ID for this specific trigger event
- `alert_id` — which alert fired
- `fire_time` — exact UTC timestamp of the fire
- `bar_time` — timestamp of the bar that triggered it
- `message` — the user-visible alert message (shown in toast)
- `popup` — `true` means show the toast notification
- `sound_file` — which sound to play (path under TradingView's sound assets)
- `kinds` — `["regular"]` for standard alerts

#### `event`

A compact fire notification sent alongside `alert_fired`. Uses abbreviated field names:

```json
{
  "m": "event",
  "p": {
    "id": 46723852620,
    "aid": 4074311184,
    "fire_time": 1771699604,
    "bar_time": 1771699560,
    "sym": "COINBASE:BTCUSD",
    "res": "1",
    "desc": "BTCUSD Crossing 68,521.11",
    "snd_file": "alert/nature/whale",
    "snd": true,
    "popup": true,
    "cross_int": true,
    "name": null,
    "snd_duration": 3
  }
}
```

**Field mapping:** `id`=fire_id, `aid`=alert_id, `desc`=message, `sym`=symbol, `res`=resolution, `snd`=play_sound, `snd_file`=sound_file

Note: timestamps here are **unix integers**, not ISO strings.

#### `fires_updated`

Sent after the fire is processed. Includes webhook delivery results:

```json
{
  "m": "fires_updated",
  "p": [{
    "fire_id": 46723852620,
    "alert_id": 4074311184,
    "fire_time": "2026-02-21T18:46:44Z",
    "bar_time": "2026-02-21T18:46:00Z",
    "message": "BTCUSD Crossing 68,521.11",
    "symbol": "={...}",
    "pro_symbol": "={...}",
    "resolution": "1",
    "kinds": ["regular"],
    "cross_interval": true,
    "popup": true,
    "webhook": {"http_code": 400},
    "sound_file": null,
    "sound_duration": 0
  }],
  "m": "fires_updated",
  "id": "vwze-37203429"
}
```

**Notable:**
- `webhook.http_code` — HTTP status from the webhook POST delivery (400 = bad request, 200 = success)
- `sound_file` becomes `null` after processing (sound already played)
- Arrives ~1s after `alert_fired`

#### `alerts_deleted` / `alert_deleted`

Sent when alerts are deleted. Two variants:

```json
// Batch deletion
{"m": "alerts_deleted", "p": [4074311184], "id": "3krq-298250710"}

// Single deletion
{"m": "alert_deleted", "p": {"id": 4074311184}}
```

---

## 3. Public Feed

**URL**: `wss://pushstream.tradingview.com/message-pipe-ws/public`

Broadcast channel for all users. Very low traffic. Currently observed carrying only economic calendar events.

### Envelope Format

```json
{
  "id": 10899,
  "channel": "public",
  "text": {
    "content": { ... },
    "channel": "economic-calendar"
  }
}
```

### Economic Calendar Events

```json
{
  "source": "Statistics Canada",
  "source_url": "https://www.statcan.gc.ca",
  "ticker": "CACA",
  "title": "Current Account",
  "actual": null,
  "actualRaw": null,
  "category": "trd",
  "comment": "Current Account is the sum of the balance of trade...",
  "country": "CA",
  "currency": "CAD",
  "forecast": -5.4,
  "forecastRaw": -5400000000,
  "importance": 0,
  "indicator": "Current Account",
  "unit": "C$",
  "id": "411597",
  "scale": "B",
  "previous": -9.7,
  "previousRaw": -9700000000,
  "referenceDate": "2025-12-31T00:00:00Z",
  "period": "Q4",
  "date": "2026-02-26T13:30:00.000Z"
}
```

**Key fields:**
- `importance` — 0 (low), 1 (medium), 2 (high impact)
- `actual` / `forecast` / `previous` — formatted values
- `actualRaw` / `forecastRaw` / `previousRaw` — raw numeric values
- `date` — scheduled release time
- `category` — economic category code (trd=trade, emp=employment, etc.)
- `country` — ISO 2-letter country code

---

## Implications for Alert Watching

### Option A: Toast Manager Subscription (JS eval)

Use `toastManager.subscribeOnToast(callback)` in the browser context. The toast manager consumes the same `private_feed` messages internally. Toast groups: `"alerts"`, `"alertsFireControl"`, `"alertsWatchlistFireControl"`.

**Pros:** Sub-100ms latency, direct callback, no parsing needed.
**Cons:** Requires active JS eval, controller must be running.

### Option C: WebSocket Frame Parsing (researcher)

Filter captured `private_feed` frames for `m: "alert_fired"` messages. The researcher already captures these passively to JSONL.

**Pros:** Passive, no browser-side injection, full history in JSONL.
**Cons:** Requires post-processing, latency depends on flush interval.

### Option D: SSE Relay via Controller (implemented)

The controller's WebSocket relay (`internal/relay/`) listens for CDP `Network.webSocketCreated/FrameReceived/Closed` events, matches connections against feed configs in `config/relay.yaml`, filters by message type (`"m"` field), and publishes to an SSE endpoint at `GET /api/v1/relay/events`.

```
Browser WS traffic → CDP Network events → Relay engine (filter) → SSE Broker → GET /api/v1/relay/events
```

Enable with `CONTROLLER_RELAY_ENABLED=true`. Clients can filter feeds via `?feeds=private_feed,chart_data`.

**Pros:** Real-time streaming to any SSE client, no JS eval, configurable feed/message filtering, multiple concurrent clients supported.
**Cons:** Requires controller to be running, relay only sees connections created after startup (page reload needed), single point of failure if controller stops.

### Option Hybrid: Direct WebSocket Client

Connect a Go WebSocket client directly to `wss://pushstream.tradingview.com/message-pipe-ws/private_feed` using the user's auth token (extracted from the chart data socket's `set_auth_token` message). This bypasses the browser entirely.

**Pros:** Lowest latency, no browser dependency, no JS eval.
**Cons:** Requires auth token management, may need periodic reconnection.
