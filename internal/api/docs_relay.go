package api

const relayDocsHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>WebSocket Relay — TV Agent</title>
  <style>
    *, *::before, *::after { box-sizing: border-box; }

    body {
      margin: 0;
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", sans-serif;
      font-size: 14px;
      line-height: 1.65;
      background: #0d1117;
      color: #c9d1d9;
      display: flex;
      flex-direction: column;
      min-height: 100vh;
    }

    a { color: #58a6ff; text-decoration: none; }
    a:hover { text-decoration: underline; }

    /* ── top nav ── */
    nav {
      background: #161b22;
      border-bottom: 1px solid #30363d;
      padding: 0 24px;
      height: 48px;
      display: flex;
      align-items: center;
      gap: 24px;
      flex-shrink: 0;
    }
    nav .brand {
      font-weight: 600;
      font-size: 15px;
      color: #e6edf3;
    }
    nav .sep { color: #484f58; }
    nav .current { color: #e6edf3; font-weight: 500; }
    nav .back { font-size: 13px; }

    /* ── layout ── */
    .layout {
      display: flex;
      flex: 1;
      max-width: 1100px;
      width: 100%;
      margin: 0 auto;
      padding: 0 16px;
    }

    /* ── sidebar ── */
    aside {
      width: 220px;
      flex-shrink: 0;
      padding: 32px 16px 32px 0;
      position: sticky;
      top: 0;
      height: calc(100vh - 48px);
      overflow-y: auto;
    }
    aside h4 {
      margin: 0 0 8px;
      font-size: 11px;
      font-weight: 600;
      text-transform: uppercase;
      letter-spacing: .08em;
      color: #8b949e;
    }
    aside ul {
      list-style: none;
      margin: 0 0 24px;
      padding: 0;
    }
    aside ul li a {
      display: block;
      padding: 4px 8px;
      border-radius: 4px;
      font-size: 13px;
      color: #8b949e;
    }
    aside ul li a:hover {
      background: #21262d;
      color: #c9d1d9;
      text-decoration: none;
    }

    /* ── main content ── */
    main {
      flex: 1;
      padding: 32px 0 64px 32px;
      border-left: 1px solid #21262d;
      min-width: 0;
    }

    h1 {
      margin: 0 0 8px;
      font-size: 28px;
      font-weight: 600;
      color: #e6edf3;
    }
    .subtitle {
      color: #8b949e;
      margin: 0 0 36px;
      font-size: 15px;
    }

    h2 {
      margin: 40px 0 12px;
      font-size: 18px;
      font-weight: 600;
      color: #e6edf3;
      padding-bottom: 8px;
      border-bottom: 1px solid #21262d;
    }
    h3 {
      margin: 28px 0 10px;
      font-size: 15px;
      font-weight: 600;
      color: #e6edf3;
    }

    p { margin: 0 0 12px; }

    /* ── method + path badge ── */
    .endpoint {
      display: inline-flex;
      align-items: center;
      gap: 10px;
      background: #161b22;
      border: 1px solid #30363d;
      border-radius: 6px;
      padding: 10px 16px;
      margin-bottom: 20px;
      font-family: "SFMono-Regular", Consolas, "Liberation Mono", Menlo, monospace;
      font-size: 14px;
    }
    .method {
      background: #1f6feb;
      color: #fff;
      font-weight: 700;
      font-size: 11px;
      padding: 2px 7px;
      border-radius: 4px;
      letter-spacing: .04em;
    }
    .path { color: #e6edf3; }

    /* ── tables ── */
    table {
      width: 100%;
      border-collapse: collapse;
      margin-bottom: 20px;
      font-size: 13px;
    }
    th {
      text-align: left;
      padding: 8px 12px;
      background: #161b22;
      color: #8b949e;
      font-weight: 600;
      border-bottom: 1px solid #30363d;
    }
    td {
      padding: 8px 12px;
      border-bottom: 1px solid #21262d;
      vertical-align: top;
    }
    tr:last-child td { border-bottom: none; }
    code {
      font-family: "SFMono-Regular", Consolas, "Liberation Mono", Menlo, monospace;
      font-size: 12px;
      background: #161b22;
      border: 1px solid #30363d;
      border-radius: 3px;
      padding: 1px 5px;
      color: #e6edf3;
    }

    /* ── code blocks ── */
    pre {
      background: #161b22;
      border: 1px solid #30363d;
      border-radius: 6px;
      padding: 16px;
      overflow-x: auto;
      margin: 0 0 20px;
    }
    pre code {
      background: none;
      border: none;
      padding: 0;
      font-size: 13px;
      line-height: 1.6;
      color: #c9d1d9;
    }

    /* ── callout ── */
    .callout {
      background: #161b22;
      border-left: 3px solid #1f6feb;
      border-radius: 0 6px 6px 0;
      padding: 12px 16px;
      margin-bottom: 20px;
      font-size: 13px;
    }
    .callout.warning { border-color: #d29922; }
    .callout strong { color: #e6edf3; }

    /* ── feed cards ── */
    .feed-card {
      background: #161b22;
      border: 1px solid #30363d;
      border-radius: 8px;
      padding: 16px 20px;
      margin-bottom: 14px;
    }
    .feed-card h3 { margin: 0 0 10px; font-size: 14px; }
    .feed-card code { font-size: 13px; }
    .feed-meta {
      display: flex;
      flex-wrap: wrap;
      gap: 8px;
      margin-bottom: 10px;
      font-size: 12px;
    }
    .feed-meta span { color: #8b949e; }
    .tag {
      background: #21262d;
      border: 1px solid #30363d;
      border-radius: 3px;
      padding: 1px 6px;
      font-family: "SFMono-Regular", Consolas, "Liberation Mono", Menlo, monospace;
      font-size: 11px;
      color: #8b949e;
    }

    /* ── SSE format visualization ── */
    .sse-block {
      background: #161b22;
      border: 1px solid #30363d;
      border-radius: 6px;
      padding: 16px;
      margin-bottom: 20px;
      font-family: "SFMono-Regular", Consolas, "Liberation Mono", Menlo, monospace;
      font-size: 13px;
      line-height: 1.8;
    }
    .sse-key { color: #79c0ff; }
    .sse-value { color: #a5d6ff; }
    .sse-comment { color: #484f58; }
  </style>
</head>
<body>

<nav>
  <span class="brand">TV Agent</span>
  <span class="sep">/</span>
  <span class="current">WebSocket Relay</span>
  <a class="back" href="/docs">← REST API Docs</a>
</nav>

<div class="layout">

  <aside>
    <h4>On this page</h4>
    <ul>
      <li><a href="#overview">Overview</a></li>
      <li><a href="#enable">Enabling the Relay</a></li>
      <li><a href="#endpoint">Endpoint</a></li>
      <li><a href="#feeds">Available Feeds</a></li>
      <li><a href="#sse-format">SSE Event Format</a></li>
      <li><a href="#examples">Examples</a></li>
      <li><a href="#config">Relay Config File</a></li>
      <li><a href="#notes">Notes</a></li>
    </ul>
  </aside>

  <main>
    <h1>WebSocket Relay</h1>
    <p class="subtitle">Stream TradingView WebSocket frames to HTTP clients via Server-Sent Events.</p>

    <!-- OVERVIEW -->
    <h2 id="overview">Overview</h2>
    <p>
      The relay bridges TradingView's internal WebSocket connections to any HTTP client.
      It listens to the browser via Chrome DevTools Protocol (CDP), intercepts WebSocket
      frames matching configured URL patterns, and re-publishes them as an SSE stream.
    </p>
    <p>
      This lets you subscribe to real-time events — alert firings, chart data updates,
      public feed messages — without needing to authenticate directly with TradingView's
      WebSocket servers.
    </p>
    <div class="callout warning">
      <strong>Relay is opt-in.</strong> It must be enabled via environment variable before
      the <code>tv_controller</code> binary starts. See <a href="#enable">Enabling the Relay</a>.
    </div>

    <!-- ENABLE -->
    <h2 id="enable">Enabling the Relay</h2>
    <p>Set these in your <code>.env</code> file before starting <code>tv_controller</code>:</p>
    <pre><code># Required — relay is disabled by default
CONTROLLER_RELAY_ENABLED=true

# Optional — path to feed config (default shown)
CONTROLLER_RELAY_CONFIG=./config/relay.yaml</code></pre>
    <p>Then (re)start the controller:</p>
    <pre><code>just run-tv-controller</code></pre>

    <!-- ENDPOINT -->
    <h2 id="endpoint">Endpoint</h2>
    <div class="endpoint">
      <span class="method">GET</span>
      <span class="path">/api/v1/relay/events</span>
    </div>

    <h3>Query Parameters</h3>
    <table>
      <thead>
        <tr><th>Name</th><th>Type</th><th>Required</th><th>Description</th></tr>
      </thead>
      <tbody>
        <tr>
          <td><code>feeds</code></td>
          <td>string</td>
          <td>No</td>
          <td>
            Comma-separated list of feed names to receive. Omit to receive events from
            all feeds. Example: <code>?feeds=private_feed,chart_data</code>
          </td>
        </tr>
      </tbody>
    </table>

    <h3>Response Headers</h3>
    <table>
      <thead>
        <tr><th>Header</th><th>Value</th></tr>
      </thead>
      <tbody>
        <tr><td><code>Content-Type</code></td><td><code>text/event-stream</code></td></tr>
        <tr><td><code>Cache-Control</code></td><td><code>no-cache</code></td></tr>
        <tr><td><code>Connection</code></td><td><code>keep-alive</code></td></tr>
        <tr><td><code>X-Accel-Buffering</code></td><td><code>no</code> (disables nginx buffering)</td></tr>
      </tbody>
    </table>

    <!-- FEEDS -->
    <h2 id="feeds">Available Feeds</h2>
    <p>
      Feeds are defined in <code>config/relay.yaml</code>. The defaults shipped with
      tv_agent are:
    </p>

    <div class="feed-card">
      <h3><code>private_feed</code></h3>
      <div class="feed-meta">
        <span>URL pattern:</span> <code>private_feed</code>
        &nbsp;·&nbsp;
        <span>Message types:</span>
        <span class="tag">alert_fired</span>
        <span class="tag">alerts_created</span>
        <span class="tag">alerts_updated</span>
        <span class="tag">fires_updated</span>
      </div>
      <p>
        TradingView's authenticated private WebSocket. Carries real-time alert events
        for the logged-in account.
      </p>
    </div>

    <div class="feed-card">
      <h3><code>public</code></h3>
      <div class="feed-meta">
        <span>URL pattern:</span> <code>public</code>
        &nbsp;·&nbsp;
        <span>Message types:</span> <em>all (no filter)</em>
      </div>
      <p>
        TradingView's public WebSocket feed. All frames are forwarded without
        message-type filtering.
      </p>
    </div>

    <div class="feed-card">
      <h3><code>chart_data</code></h3>
      <div class="feed-meta">
        <span>URL pattern:</span> <code>socket.io/websocket</code>
        &nbsp;·&nbsp;
        <span>Message types:</span>
        <span class="tag">du</span>
        <span class="tag">qsd</span>
      </div>
      <p>
        TradingView's chart data socket (socket.io). Filtered to data-update
        (<code>du</code>) and quote-series-data (<code>qsd</code>) messages.
      </p>
    </div>

    <!-- SSE FORMAT -->
    <h2 id="sse-format">SSE Event Format</h2>
    <p>Each event follows the standard SSE format. The <code>event</code> field is the feed name.</p>
    <div class="sse-block">
      <span class="sse-key">event:</span> <span class="sse-value">private_feed</span><br>
      <span class="sse-key">data:</span> <span class="sse-value">{"m":"alert_fired","alert_id":3999574105,...}</span><br>
      <br>
      <span class="sse-key">event:</span> <span class="sse-value">chart_data</span><br>
      <span class="sse-key">data:</span> <span class="sse-value">{"m":"du","p":["cs_...","sds_...",[...]]}</span><br>
      <br>
    </div>
    <p>
      All payloads are JSON. Messages from TradingView's socket.io transport include
      an <code>"m"</code> field identifying the message type, and a <code>"p"</code>
      array with parameters.
    </p>

    <!-- EXAMPLES -->
    <h2 id="examples">Examples</h2>

    <h3>Browser — EventSource</h3>
    <pre><code>// Subscribe to all feeds
const sse = new EventSource('http://127.0.0.1:8188/api/v1/relay/events');

sse.addEventListener('private_feed', (e) => {
  const msg = JSON.parse(e.data);
  if (msg.m === 'alert_fired') {
    console.log('Alert fired:', msg.alert_id);
  }
});

sse.addEventListener('chart_data', (e) => {
  const msg = JSON.parse(e.data);
  console.log(msg.m, msg.p);
});

sse.onerror = (e) => console.error('SSE error', e);</code></pre>

    <h3>Browser — Filter to one feed</h3>
    <pre><code>const sse = new EventSource(
  'http://127.0.0.1:8188/api/v1/relay/events?feeds=private_feed'
);</code></pre>

    <h3>curl</h3>
    <pre><code>curl -N http://127.0.0.1:8188/api/v1/relay/events
curl -N 'http://127.0.0.1:8188/api/v1/relay/events?feeds=private_feed,chart_data'</code></pre>

    <h3>Python — sseclient</h3>
    <pre><code>import json, sseclient, requests

resp = requests.get(
    'http://127.0.0.1:8188/api/v1/relay/events',
    params={'feeds': 'private_feed'},
    stream=True,
)
for event in sseclient.SSEClient(resp).events():
    msg = json.loads(event.data)
    print(event.event, msg.get('m'), msg)</code></pre>

    <!-- CONFIG -->
    <h2 id="config">Relay Config File</h2>
    <p>
      Feeds are defined in <code>config/relay.yaml</code> (path overridable via
      <code>CONTROLLER_RELAY_CONFIG</code>). Each feed entry specifies:
    </p>
    <table>
      <thead>
        <tr><th>Field</th><th>Required</th><th>Description</th></tr>
      </thead>
      <tbody>
        <tr>
          <td><code>name</code></td>
          <td>Yes</td>
          <td>Feed identifier used as the SSE <code>event:</code> name and the <code>?feeds=</code> filter value.</td>
        </tr>
        <tr>
          <td><code>url_pattern</code></td>
          <td>Yes</td>
          <td>Substring matched against the WebSocket URL. The first matching feed wins.</td>
        </tr>
        <tr>
          <td><code>message_types</code></td>
          <td>No</td>
          <td>List of <code>"m"</code> field values to accept. Omit to forward all frames.</td>
        </tr>
      </tbody>
    </table>

    <pre><code># config/relay.yaml
feeds:
  - name: private_feed
    url_pattern: "private_feed"
    message_types: ["alert_fired", "alerts_created", "alerts_updated", "fires_updated"]

  - name: public
    url_pattern: "public"

  - name: chart_data
    url_pattern: "socket.io/websocket"
    message_types: ["du", "qsd"]</code></pre>

    <!-- NOTES -->
    <h2 id="notes">Notes</h2>
    <ul>
      <li>
        <strong>Buffer &amp; back-pressure:</strong> each subscriber has a 256-event in-memory buffer.
        Slow clients will have events silently dropped — the broker is non-blocking.
      </li>
      <li>
        <strong>Tab attachment:</strong> the relay attaches to browser tabs matching
        <code>CONTROLLER_TAB_URL_FILTER</code> (default <code>tradingview.com</code>).
        Open a TradingView chart tab before starting the controller.
      </li>
      <li>
        <strong>Reconnection:</strong> the browser's built-in <code>EventSource</code>
        automatically reconnects on disconnect. For other clients, implement reconnect
        with exponential backoff.
      </li>
      <li>
        <strong>Authentication:</strong> the relay endpoint has no authentication. Bind
        the controller to <code>127.0.0.1</code> (the default) to prevent external access.
      </li>
    </ul>

  </main>
</div>

</body>
</html>`
