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
<body style="height: 100vh; margin: 0;">
  <elements-api
    apiDescriptionUrl="/openapi.json"
    router="hash"
    layout="sidebar"
    tryItCredentialsPolicy="same-origin"
    darkMode
  />
</body>
</html>`
