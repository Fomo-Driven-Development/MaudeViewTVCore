package api

const docsHTML = `<!doctype html>
<html lang="en" data-theme="dark">
<head>
  <meta charset="utf-8" />
  <meta name="referrer" content="same-origin" />
  <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no" />
  <title>TV Agent Controller API</title>
  <link href="https://unpkg.com/@stoplight/elements@9.0.0/styles.min.css" rel="stylesheet" />
  <script src="https://unpkg.com/@stoplight/elements@9.0.0/web-components.min.js" crossorigin="anonymous"></script>
</head>
<body style="height: 100vh; margin: 0; position: relative;">
  <a href="/docs/relay" style="
    position: fixed;
    top: 12px;
    right: 16px;
    z-index: 9999;
    background: #161b22;
    border: 1px solid #30363d;
    border-radius: 6px;
    color: #58a6ff;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
    font-size: 12px;
    font-weight: 500;
    padding: 5px 12px;
    text-decoration: none;
  ">WebSocket Relay Docs →</a>
  <elements-api
    apiDescriptionUrl="/openapi.json"
    router="hash"
    layout="sidebar"
    tryItCredentialsPolicy="same-origin"
    darkMode
  />
</body>
</html>`
