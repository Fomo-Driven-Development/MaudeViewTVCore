package api

const docsHTML = `<!doctype html>
<html lang="en" data-theme="dark">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>TV Agent Controller API Docs</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
  <style>
    html, body { margin: 0; background: #0b1220; color: #e5e7eb; }
    #swagger-ui { max-width: 1200px; margin: 0 auto; }
    .swagger-ui .topbar { background: #111827; border-bottom: 1px solid #1f2937; }
    .swagger-ui .info .title, .swagger-ui, .swagger-ui .opblock-tag, .swagger-ui .opblock-summary-path, .swagger-ui .response-col_status { color: #e5e7eb; }
    .swagger-ui .opblock .opblock-summary, .swagger-ui .opblock, .swagger-ui .scheme-container, .swagger-ui .model-box, .swagger-ui section.models { background: #0f172a; border-color: #1f2937; }
    .swagger-ui .btn, .swagger-ui input, .swagger-ui select, .swagger-ui textarea { background: #111827; color: #e5e7eb; border-color: #374151; }
    .swagger-ui .opblock-description-wrapper p, .swagger-ui .response-col_description__inner p, .swagger-ui table thead tr td, .swagger-ui table thead tr th { color: #cbd5e1; }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.onload = function() {
      window.ui = SwaggerUIBundle({
        url: '/openapi.json',
        dom_id: '#swagger-ui',
        deepLinking: true,
        displayRequestDuration: true,
        tryItOutEnabled: true,
      });
    };
  </script>
</body>
</html>`
