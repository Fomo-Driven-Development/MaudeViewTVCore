

  Category: Alerts
  Endpoints: get/create/update/delete alerts, fires, clone, restart, stop, probes
  Count: 14
  ────────────────────────────────────────
  Category: Strategy
  Endpoints: list, active, date-range, probe, report, goto, set active, set input
  Count: 8
  ────────────────────────────────────────
  Category: Layout (COVERED in layout_test.go)
  Covered: list, status, clone, save, rename, switch, delete, grid, fullscreen, dismiss-dialog (+ validation tests)
  Deferred: preview (2 page reloads), batch-delete happy-path (needs 2+ clones)
  Count: 2 remaining
  ────────────────────────────────────────
  Category: Snapshots
  Endpoints: list, metadata, delete
  Count: 3
  ────────────────────────────────────────
  Category: Watchlist extras
  Endpoints: flag, active get/set
  Count: 3
  ────────────────────────────────────────
  Category: Chart extras
  Endpoints: symbol info, chart-api probes, action, zoom, scroll, snapshot, visible-range set
  Covered in layout_test.go: panes, activate, next, prev, maximize
  Count: ~10
  ────────────────────────────────────────
  Category: Drawing extras
  Endpoints: list/get/delete drawings, clone, z-order, tool/state/toggles get/set
  Count: ~12
  ────────────────────────────────────────
  Category: Pine extras
  Endpoints: command-palette, new-tab, open-script
  Count: 3
  ────────────────────────────────────────
  Category: System
  Endpoints: health, deep health, page reload, browser screenshot
  Count: 4

  The biggest untested areas are Alerts (14), Chart extras (~10), and Strategy (8).
