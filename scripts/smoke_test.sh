#!/usr/bin/env bash
# smoke_test.sh — comprehensive endpoint smoke tests for tv_controller
# Usage: bash scripts/smoke_test.sh [BASE_URL]
set -uo pipefail

BASE="${1:-http://127.0.0.1:8188}"
PASS=0; FAIL=0; SKIP=0
ERRORS=""
START_TIME=$(date +%s)

# Colors
RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[0;33m'; CYAN='\033[0;36m'; NC='\033[0m'

ok()   { PASS=$((PASS+1)); printf "${GREEN}  PASS${NC} %s\n" "$1" >&2; }
fail() { FAIL=$((FAIL+1)); ERRORS="${ERRORS}\n  - $1: $2"; printf "${RED}  FAIL${NC} %s: %s\n" "$1" "$2" >&2; }
skip() { SKIP=$((SKIP+1)); printf "${YELLOW}  SKIP${NC} %s: %s\n" "$1" "$2" >&2; }

# Helper: GET request, check HTTP 200 and valid JSON
get() {
  local label="$1" path="$2"
  local resp code
  resp=$(curl -s -w "\n%{http_code}" "$BASE$path" 2>/dev/null) || { fail "$label" "curl error"; return 1; }
  code=$(echo "$resp" | tail -1)
  resp=$(echo "$resp" | sed '$d')
  if [ "$code" != "200" ]; then fail "$label" "HTTP $code"; return 1; fi
  echo "$resp" | jq . >/dev/null 2>&1 || { fail "$label" "invalid JSON"; return 1; }
  ok "$label"
  echo "$resp"
}

# Helper: request with method/body, check HTTP 200
req() {
  local label="$1" method="$2" path="$3" body="${4:-}"
  local resp code args=(-s -w "\n%{http_code}" -X "$method")
  if [ -n "$body" ]; then
    args+=(-H "Content-Type: application/json" -d "$body")
  fi
  resp=$(curl "${args[@]}" "$BASE$path" 2>/dev/null) || { fail "$label" "curl error"; return 1; }
  code=$(echo "$resp" | tail -1)
  resp=$(echo "$resp" | sed '$d')
  if [ "$code" != "200" ]; then fail "$label" "HTTP $code"; return 1; fi
  echo "$resp" | jq . >/dev/null 2>&1 || { fail "$label" "invalid JSON"; return 1; }
  ok "$label"
  echo "$resp"
}

# Helper: request that's expected to return a specific status
req_status() {
  local label="$1" method="$2" path="$3" expected="${4:-200}" body="${5:-}"
  local resp code args=(-s -w "\n%{http_code}" -X "$method")
  if [ -n "$body" ]; then
    args+=(-H "Content-Type: application/json" -d "$body")
  fi
  resp=$(curl "${args[@]}" "$BASE$path" 2>/dev/null) || { fail "$label" "curl error"; return 1; }
  code=$(echo "$resp" | tail -1)
  if [ "$code" != "$expected" ]; then fail "$label" "HTTP $code (expected $expected)"; return 1; fi
  ok "$label"
}

###############################################################################
printf "${CYAN}=== TV Agent Smoke Test ===${NC}\n"
printf "Target: %s\n\n" "$BASE"

###############################################################################
printf "${CYAN}--- Phase 0: Prerequisites ---${NC}\n"

HEALTH=$(get "GET /health" "/health") || { echo "Controller not running. Exiting." >&2; exit 1; }

CHARTS_JSON=$(get "GET /api/v1/charts" "/api/v1/charts") || { echo "Cannot list charts. Exiting." >&2; exit 1; }
CHART_ID=$(echo "$CHARTS_JSON" | jq -r '.charts[0].chart_id // empty')
if [ -z "$CHART_ID" ]; then
  echo "No chart tabs found. Open TradingView in the browser first." >&2
  exit 1
fi
printf "  Using chart_id: %s\n\n" "$CHART_ID" >&2

###############################################################################
printf "${CYAN}--- Phase 1: Read-Only GETs ---${NC}\n"

get "GET /api/v1/charts/active" "/api/v1/charts/active" >/dev/null
get "GET /api/v1/health/deep" "/api/v1/health/deep" >/dev/null
get "GET chart symbol" "/api/v1/chart/$CHART_ID/symbol" >/dev/null
get "GET symbol info" "/api/v1/chart/$CHART_ID/symbol/info" >/dev/null
get "GET resolution" "/api/v1/chart/$CHART_ID/resolution" >/dev/null
get "GET studies" "/api/v1/chart/$CHART_ID/studies" >/dev/null
get "GET watchlists" "/api/v1/watchlists" >/dev/null
get "GET active watchlist" "/api/v1/watchlists/active" >/dev/null
get "GET visible range" "/api/v1/chart/$CHART_ID/visible-range" >/dev/null
get "GET chart-api probe" "/api/v1/chart/$CHART_ID/chart-api/probe" >/dev/null
get "GET chart-api deep" "/api/v1/chart/$CHART_ID/chart-api/probe/deep" >/dev/null
get "GET replay probe" "/api/v1/chart/$CHART_ID/replay/probe" >/dev/null
get "GET replay probe deep" "/api/v1/chart/$CHART_ID/replay/probe/deep" >/dev/null
get "GET replay status" "/api/v1/chart/$CHART_ID/replay/status" >/dev/null
get "GET replay scan" "/api/v1/chart/$CHART_ID/replay/scan" >/dev/null
get "GET strategy probe" "/api/v1/chart/$CHART_ID/strategy/probe" >/dev/null
get "GET strategy list" "/api/v1/chart/$CHART_ID/strategy/list" >/dev/null
get "GET strategy active" "/api/v1/chart/$CHART_ID/strategy/active" >/dev/null
get "GET strategy report" "/api/v1/chart/$CHART_ID/strategy/report" >/dev/null
get "GET strategy date-range" "/api/v1/chart/$CHART_ID/strategy/date-range" >/dev/null
get "GET alerts scan" "/api/v1/chart/$CHART_ID/alerts/scan" >/dev/null
get "GET alerts probe" "/api/v1/chart/$CHART_ID/alerts/probe" >/dev/null
get "GET alerts probe deep" "/api/v1/chart/$CHART_ID/alerts/probe/deep" >/dev/null
get "GET alerts list" "/api/v1/alerts" >/dev/null
get "GET alert fires" "/api/v1/alerts/fires" >/dev/null
get "GET drawings" "/api/v1/chart/$CHART_ID/drawings" >/dev/null
get "GET drawing toggles" "/api/v1/chart/$CHART_ID/drawings/toggles" >/dev/null
get "GET drawing tool" "/api/v1/chart/$CHART_ID/drawings/tool" >/dev/null
get "GET drawings state" "/api/v1/chart/$CHART_ID/drawings/state" >/dev/null
get "GET snapshots" "/api/v1/snapshots" >/dev/null
get "GET pine status" "/api/v1/pine/status" >/dev/null
get "GET pine console" "/api/v1/pine/console" >/dev/null
get "GET layouts" "/api/v1/layouts" >/dev/null
get "GET layout status" "/api/v1/layout/status" >/dev/null

printf "\n"

###############################################################################
printf "${CYAN}--- Phase 2: Idempotent Toggles ---${NC}\n"

req "POST fullscreen on" "POST" "/api/v1/layout/fullscreen" >/dev/null
sleep 0.5
req "POST fullscreen off" "POST" "/api/v1/layout/fullscreen" >/dev/null
sleep 0.5

req "POST maximize on" "POST" "/api/v1/chart/maximize" >/dev/null
sleep 0.5
req "POST maximize off" "POST" "/api/v1/chart/maximize" >/dev/null
sleep 0.5

printf "\n"

###############################################################################
printf "${CYAN}--- Phase 3: Symbol/Resolution Round-Trip ---${NC}\n"

ORIG_SYM=$(curl -s "$BASE/api/v1/chart/$CHART_ID/symbol" | jq -r '.current_symbol // empty')
if [ -n "$ORIG_SYM" ]; then
  req "PUT symbol AAPL" "PUT" "/api/v1/chart/$CHART_ID/symbol?symbol=NASDAQ:AAPL" >/dev/null
  sleep 1
  NEW_SYM=$(curl -s "$BASE/api/v1/chart/$CHART_ID/symbol" | jq -r '.current_symbol // empty')
  if echo "$NEW_SYM" | grep -qi "AAPL"; then ok "verify symbol=AAPL"; else fail "verify symbol=AAPL" "got $NEW_SYM"; fi
  ENCODED_SYM=$(printf '%s' "$ORIG_SYM" | jq -sRr @uri)
  req "PUT symbol restore" "PUT" "/api/v1/chart/$CHART_ID/symbol?symbol=$ENCODED_SYM" >/dev/null
  sleep 1
else
  skip "symbol round-trip" "could not read original symbol"
fi

ORIG_RES=$(curl -s "$BASE/api/v1/chart/$CHART_ID/resolution" | jq -r '.current_resolution // empty')
if [ -n "$ORIG_RES" ]; then
  req "PUT resolution 60" "PUT" "/api/v1/chart/$CHART_ID/resolution?resolution=60" >/dev/null
  sleep 1
  NEW_RES=$(curl -s "$BASE/api/v1/chart/$CHART_ID/resolution" | jq -r '.current_resolution // empty')
  if [ "$NEW_RES" = "60" ]; then ok "verify resolution=60"; else fail "verify resolution=60" "got $NEW_RES"; fi
  req "PUT resolution restore" "PUT" "/api/v1/chart/$CHART_ID/resolution?resolution=$ORIG_RES" >/dev/null
  sleep 1
else
  skip "resolution round-trip" "could not read original resolution"
fi

printf "\n"

###############################################################################
printf "${CYAN}--- Phase 4: Study CRUD ---${NC}\n"

ADD_RESP=$(req "POST add RSI study" "POST" "/api/v1/chart/$CHART_ID/studies" '{"name":"Relative Strength Index"}')
STUDY_ID=$(echo "$ADD_RESP" | jq -r '.study.id // empty')
if [ -n "$STUDY_ID" ]; then
  get "GET study detail" "/api/v1/chart/$CHART_ID/studies/$STUDY_ID" >/dev/null
  req "PATCH modify study" "PATCH" "/api/v1/chart/$CHART_ID/studies/$STUDY_ID" '{"inputs":{"length":21}}' >/dev/null
  req_status "DELETE study" "DELETE" "/api/v1/chart/$CHART_ID/studies/$STUDY_ID" "204"
else
  skip "study CRUD" "could not add study"
fi

printf "\n"

###############################################################################
printf "${CYAN}--- Phase 5: Drawing CRUD ---${NC}\n"

# Get a recent bar timestamp for drawing placement
VR=$(curl -s "$BASE/api/v1/chart/$CHART_ID/visible-range")
BAR_TIME=$(echo "$VR" | jq -r '.from // empty')
BAR_PRICE=100

if [ -n "$BAR_TIME" ]; then
  DRAW_RESP=$(req "POST create drawing" "POST" "/api/v1/chart/$CHART_ID/drawings" \
    "{\"point\":{\"time\":$BAR_TIME,\"price\":$BAR_PRICE},\"options\":{\"shape\":\"horizontal_ray\"}}")
  SHAPE_ID=$(echo "$DRAW_RESP" | jq -r '.id // empty')
  if [ -n "$SHAPE_ID" ]; then
    get "GET drawing detail" "/api/v1/chart/$CHART_ID/drawings/$SHAPE_ID" >/dev/null
    req "POST clone drawing" "POST" "/api/v1/chart/$CHART_ID/drawings/$SHAPE_ID/clone" >/dev/null
    req "POST z-order" "POST" "/api/v1/chart/$CHART_ID/drawings/$SHAPE_ID/z-order" '{"action":"bring_to_front"}' >/dev/null
    req_status "DELETE drawing" "DELETE" "/api/v1/chart/$CHART_ID/drawings/$SHAPE_ID" "204"
    # Clean up clone
    req_status "DELETE all drawings" "DELETE" "/api/v1/chart/$CHART_ID/drawings" "204"
  else
    skip "drawing ops" "could not create drawing"
  fi
else
  skip "drawing CRUD" "could not get visible range"
fi

printf "\n"

###############################################################################
printf "${CYAN}--- Phase 6: Watchlist CRUD ---${NC}\n"

WL_RESP=$(req "POST create watchlist" "POST" "/api/v1/watchlists" '{"name":"__smoke_test__"}')
WL_ID=$(echo "$WL_RESP" | jq -r '.id // empty')
if [ -n "$WL_ID" ]; then
  req "PATCH rename watchlist" "PATCH" "/api/v1/watchlist/$WL_ID" '{"name":"__smoke_renamed__"}' >/dev/null
  req "POST add symbols" "POST" "/api/v1/watchlist/$WL_ID/symbols" '{"symbols":["NASDAQ:AAPL","NASDAQ:MSFT"]}' >/dev/null
  get "GET watchlist detail" "/api/v1/watchlist/$WL_ID" >/dev/null
  req "DELETE symbols" "DELETE" "/api/v1/watchlist/$WL_ID/symbols" '{"symbols":["NASDAQ:MSFT"]}' >/dev/null
  req_status "DELETE watchlist" "DELETE" "/api/v1/watchlist/$WL_ID" "204"
else
  skip "watchlist CRUD" "could not create watchlist"
fi

printf "\n"

###############################################################################
printf "${CYAN}--- Phase 7: Layout Operations ---${NC}\n"

req "POST save layout" "POST" "/api/v1/layout/save" >/dev/null
req "POST rename layout" "POST" "/api/v1/layout/rename" '{"name":"__smoke_layout__"}' >/dev/null
req "POST grid 2h" "POST" "/api/v1/layout/grid" '{"template":"2h"}' >/dev/null
sleep 1
req "POST grid s (restore)" "POST" "/api/v1/layout/grid" '{"template":"s"}' >/dev/null
sleep 1
req "POST rename layout restore" "POST" "/api/v1/layout/rename" '{"name":"Untitled"}' >/dev/null

printf "\n"

###############################################################################
printf "${CYAN}--- Phase 8: Snapshot CRUD ---${NC}\n"

SNAP_RESP=$(req "POST take snapshot" "POST" "/api/v1/chart/$CHART_ID/snapshot" '{"format":"png","quality":"standard"}')
SNAP_ID=$(echo "$SNAP_RESP" | jq -r '.snapshot.id // empty')
if [ -n "$SNAP_ID" ]; then
  get "GET snapshot list" "/api/v1/snapshots" >/dev/null
  get "GET snapshot meta" "/api/v1/snapshots/$SNAP_ID/metadata" >/dev/null
  req "DELETE snapshot" "DELETE" "/api/v1/snapshots/$SNAP_ID" >/dev/null
else
  skip "snapshot CRUD" "could not take snapshot"
fi

printf "\n"

###############################################################################
printf "${CYAN}--- Phase 9: Pine Editor (non-destructive) ---${NC}\n"

# Count studies before pine ops
ORIG_STUDY_COUNT=$(curl -s "$BASE/api/v1/chart/$CHART_ID/studies" | jq '.studies | length')

PINE_STATUS=$(curl -s "$BASE/api/v1/pine/status" | jq -r '.is_visible // "false"')
if [ "$PINE_STATUS" = "false" ]; then
  req "POST pine toggle open" "POST" "/api/v1/pine/toggle" >/dev/null
  sleep 1
fi

PINE_VIS=$(curl -s "$BASE/api/v1/pine/status" | jq -r '.is_visible // "false"')
if [ "$PINE_VIS" = "true" ]; then
  # Create new indicator — doesn't touch existing script
  req "POST pine new-indicator" "POST" "/api/v1/pine/new-indicator" >/dev/null
  sleep 1

  # Write smoke test source
  req "PUT pine source" "PUT" "/api/v1/pine/source" '{"source":"// smoke_test_indicator\nplot(close, title=\"Smoke\")"}' >/dev/null
  sleep 0.5

  # Add to chart (Ctrl+Enter) — skip explicit save because new indicators
  # trigger a "Save Script As" naming dialog that blocks Ctrl+S.
  # Ctrl+Enter handles "Save and add to chart" in one step.
  req "POST pine add-to-chart" "POST" "/api/v1/pine/add-to-chart" >/dev/null
  sleep 2

  # Verify study count increased
  NEW_STUDY_COUNT=$(curl -s "$BASE/api/v1/chart/$CHART_ID/studies" | jq '.studies | length')
  if [ "$NEW_STUDY_COUNT" -gt "$ORIG_STUDY_COUNT" ] 2>/dev/null; then
    ok "verify study added by pine"
  else
    fail "verify study added by pine" "count $ORIG_STUDY_COUNT -> $NEW_STUDY_COUNT"
  fi

  # Read source back
  get "GET pine source" "/api/v1/pine/source" >/dev/null

  # Close editor
  req "POST pine toggle close" "POST" "/api/v1/pine/toggle" >/dev/null
  sleep 0.5

  # Clean up: remove the added study (last one in list)
  SMOKE_STUDY_ID=$(curl -s "$BASE/api/v1/chart/$CHART_ID/studies" | jq -r '.studies[-1].id // empty')
  if [ -n "$SMOKE_STUDY_ID" ]; then
    req_status "DELETE smoke study cleanup" "DELETE" "/api/v1/chart/$CHART_ID/studies/$SMOKE_STUDY_ID" "204"
  fi
else
  skip "pine editor ops" "editor not visible after toggle"
fi

printf "\n"

###############################################################################
printf "${CYAN}--- Phase 10: Chart Navigation ---${NC}\n"

req "POST grid 2h" "POST" "/api/v1/layout/grid" '{"template":"2h"}' >/dev/null
sleep 1
req "POST next chart" "POST" "/api/v1/chart/next" >/dev/null
sleep 0.5
req "POST prev chart" "POST" "/api/v1/chart/prev" >/dev/null
sleep 0.5
req "POST activate chart 0" "POST" "/api/v1/chart/activate" '{"index":0}' >/dev/null
sleep 0.5
req "POST grid s (restore)" "POST" "/api/v1/layout/grid" '{"template":"s"}' >/dev/null
sleep 1

printf "\n"

###############################################################################
printf "${CYAN}--- Phase 11: Keyboard Shortcuts ---${NC}\n"

req "POST dismiss dialog" "POST" "/api/v1/layout/dismiss-dialog" >/dev/null
req "POST scroll +5" "POST" "/api/v1/chart/$CHART_ID/scroll" '{"bars":5}' >/dev/null
req "POST scroll -5" "POST" "/api/v1/chart/$CHART_ID/scroll" '{"bars":-5}' >/dev/null
req "POST zoom in" "POST" "/api/v1/chart/$CHART_ID/zoom" '{"direction":"in"}' >/dev/null
req "POST zoom out" "POST" "/api/v1/chart/$CHART_ID/zoom" '{"direction":"out"}' >/dev/null
req "POST scroll to realtime" "POST" "/api/v1/chart/$CHART_ID/scroll/to-realtime" >/dev/null
req "POST reset scales" "POST" "/api/v1/chart/$CHART_ID/reset-scales" >/dev/null

printf "\n"

###############################################################################
printf "${CYAN}--- Phase 12: Pine Keyboard Shortcuts ---${NC}\n"

# Ensure editor is open
PINE12=$(curl -s "$BASE/api/v1/pine/status" | jq -r '.is_visible // "false"')
if [ "$PINE12" = "false" ]; then
  req "POST pine open (phase12)" "POST" "/api/v1/pine/toggle" >/dev/null
  sleep 1
fi

PINE12_VIS=$(curl -s "$BASE/api/v1/pine/status" | jq -r '.is_visible // "false"')
if [ "$PINE12_VIS" = "true" ]; then
  req "POST pine undo" "POST" "/api/v1/pine/undo" >/dev/null
  req "POST pine redo" "POST" "/api/v1/pine/redo" >/dev/null
  req "POST pine toggle-console" "POST" "/api/v1/pine/toggle-console" >/dev/null
  sleep 0.3
  req "POST pine toggle-console off" "POST" "/api/v1/pine/toggle-console" >/dev/null
  req "POST pine go-to-line" "POST" "/api/v1/pine/go-to-line" '{"line":1}' >/dev/null
  req "POST pine insert-line" "POST" "/api/v1/pine/insert-line" >/dev/null
  req "POST pine toggle-comment" "POST" "/api/v1/pine/toggle-comment" >/dev/null
  req "POST pine delete-line" "POST" "/api/v1/pine/delete-line" '{"count":1}' >/dev/null

  # Close editor
  req "POST pine close (phase12)" "POST" "/api/v1/pine/toggle" >/dev/null
  sleep 0.5
else
  skip "pine keyboard shortcuts" "editor not visible"
fi

printf "\n"

###############################################################################
printf "${CYAN}--- Phase 13: Drawing Toggles & Advanced Ops ---${NC}\n"

# Read visible range for drawing placement
VR13=$(curl -s "$BASE/api/v1/chart/$CHART_ID/visible-range")
BAR_TIME13=$(echo "$VR13" | jq -r '.from // empty')
BAR_PRICE13=100

if [ -n "$BAR_TIME13" ]; then
  # Toggle hide on/off
  req "PUT hide drawings on" "PUT" "/api/v1/chart/$CHART_ID/drawings/toggles/hide" '{"value":true}' >/dev/null
  req "PUT hide drawings off" "PUT" "/api/v1/chart/$CHART_ID/drawings/toggles/hide" '{"value":false}' >/dev/null

  # Toggle lock on/off
  req "PUT lock drawings on" "PUT" "/api/v1/chart/$CHART_ID/drawings/toggles/lock" '{"value":true}' >/dev/null
  req "PUT lock drawings off" "PUT" "/api/v1/chart/$CHART_ID/drawings/toggles/lock" '{"value":false}' >/dev/null

  # Toggle magnet on/off
  req "PUT magnet on" "PUT" "/api/v1/chart/$CHART_ID/drawings/toggles/magnet" '{"enabled":true,"mode":1}' >/dev/null
  req "PUT magnet off" "PUT" "/api/v1/chart/$CHART_ID/drawings/toggles/magnet" '{"enabled":false}' >/dev/null

  # Set drawing tool
  req "PUT set drawing tool" "PUT" "/api/v1/chart/$CHART_ID/drawings/tool" '{"tool":"LineToolTrendLine"}' >/dev/null
  TOOL_RESP=$(get "GET drawing tool verify" "/api/v1/chart/$CHART_ID/drawings/tool")

  # Create multipoint drawing (trend line needs 2 points)
  BAR_TIME13_END=$(echo "$VR13" | jq -r '.to // empty')
  if [ -n "$BAR_TIME13_END" ]; then
    MP_RESP=$(req "POST multipoint drawing" "POST" "/api/v1/chart/$CHART_ID/drawings/multipoint" \
      "{\"points\":[{\"time\":$BAR_TIME13,\"price\":$BAR_PRICE13},{\"time\":$BAR_TIME13_END,\"price\":$(echo "$BAR_PRICE13 + 10" | bc)}],\"options\":{\"shape\":\"trend_line\"}}")
    MP_ID=$(echo "$MP_RESP" | jq -r '.id // empty')
    if [ -n "$MP_ID" ]; then
      req "PUT drawing visibility off" "PUT" "/api/v1/chart/$CHART_ID/drawings/$MP_ID/visibility" '{"visible":false}' >/dev/null
      req "PUT drawing visibility on" "PUT" "/api/v1/chart/$CHART_ID/drawings/$MP_ID/visibility" '{"visible":true}' >/dev/null
      req_status "DELETE multipoint drawing" "DELETE" "/api/v1/chart/$CHART_ID/drawings/$MP_ID" "204"
    else
      skip "multipoint drawing ops" "could not create multipoint drawing"
    fi
  fi

  # State export/import round-trip
  STATE_JSON=$(get "GET drawings state" "/api/v1/chart/$CHART_ID/drawings/state")
  STATE_DATA=$(echo "$STATE_JSON" | jq -c '.state // empty')
  if [ -n "$STATE_DATA" ] && [ "$STATE_DATA" != "null" ]; then
    req "PUT import drawings state" "PUT" "/api/v1/chart/$CHART_ID/drawings/state" "{\"state\":$STATE_DATA}" >/dev/null
    ok "drawings state round-trip"
  else
    skip "drawings state round-trip" "no state data to import"
  fi
else
  skip "drawing toggles" "could not get visible range"
fi

printf "\n"

###############################################################################
printf "${CYAN}--- Phase 14: Replay Lifecycle ---${NC}\n"

# Activate replay at first available date
REPLAY_ACT=$(req "POST replay activate/auto" "POST" "/api/v1/chart/$CHART_ID/replay/activate/auto")
REPLAY_ACT_OK=$?
sleep 1

if [ $REPLAY_ACT_OK -eq 0 ]; then
  # Check status
  REPLAY_ST=$(get "GET replay status (active)" "/api/v1/chart/$CHART_ID/replay/status")
  sleep 0.5

  # Step forward twice
  req "POST replay step 1" "POST" "/api/v1/chart/$CHART_ID/replay/step" >/dev/null
  sleep 0.3
  req "POST replay step 2" "POST" "/api/v1/chart/$CHART_ID/replay/step" >/dev/null
  sleep 0.3

  # Change autoplay delay
  req "PUT autoplay delay 500" "PUT" "/api/v1/chart/$CHART_ID/replay/autoplay/delay" '{"delay":500}' >/dev/null

  # Start autoplay
  req "POST autoplay start" "POST" "/api/v1/chart/$CHART_ID/replay/autoplay/start" >/dev/null
  sleep 1

  # Stop autoplay
  req "POST autoplay stop" "POST" "/api/v1/chart/$CHART_ID/replay/autoplay/stop" >/dev/null
  sleep 0.5

  # Reset replay
  req "POST replay reset" "POST" "/api/v1/chart/$CHART_ID/replay/reset" >/dev/null
  sleep 0.5

  # Deactivate replay
  req "POST replay deactivate" "POST" "/api/v1/chart/$CHART_ID/replay/deactivate" >/dev/null
  sleep 1

  # Verify replay not active
  get "GET replay status (deactivated)" "/api/v1/chart/$CHART_ID/replay/status" >/dev/null
else
  skip "replay lifecycle" "could not activate replay"
fi

# Extra settle time after replay deactivation before chart navigation
sleep 2

printf "\n"

###############################################################################
printf "${CYAN}--- Phase 15: Navigation Extras ---${NC}\n"

# Go-to-date: use the from timestamp of current visible range
VR15=$(curl -s "$BASE/api/v1/chart/$CHART_ID/visible-range")
VR15_FROM=$(echo "$VR15" | jq -r '.from // empty')
VR15_TO=$(echo "$VR15" | jq -r '.to // empty')

if [ -n "$VR15_FROM" ]; then
  req "POST go-to-date" "POST" "/api/v1/chart/$CHART_ID/go-to-date" "{\"timestamp\":$VR15_FROM}" >/dev/null
  sleep 0.5
fi

# Set visible range (read current, write it back — idempotent)
if [ -n "$VR15_FROM" ] && [ -n "$VR15_TO" ]; then
  req "PUT set visible-range" "PUT" "/api/v1/chart/$CHART_ID/visible-range" "{\"from\":$VR15_FROM,\"to\":$VR15_TO}" >/dev/null
  sleep 0.5
fi

# Resolve symbol
get "GET resolve-symbol AAPL" "/api/v1/chart/$CHART_ID/chart-api/resolve-symbol?symbol=AAPL" >/dev/null

# Timezone round-trip
TZ_RESP=$(get "GET chart-api probe (tz)" "/api/v1/chart/$CHART_ID/chart-api/probe")
ORIG_TZ=$(echo "$TZ_RESP" | jq -r '.timezone // empty')
req "PUT timezone NY" "PUT" "/api/v1/chart/$CHART_ID/chart-api/timezone" '{"timezone":"America/New_York"}' >/dev/null
sleep 0.5
if [ -n "$ORIG_TZ" ]; then
  req "PUT timezone restore" "PUT" "/api/v1/chart/$CHART_ID/chart-api/timezone" "{\"timezone\":\"$ORIG_TZ\"}" >/dev/null
fi

# Scroll back to realtime after navigation
req "POST scroll to realtime (nav)" "POST" "/api/v1/chart/$CHART_ID/scroll/to-realtime" >/dev/null

printf "\n"

###############################################################################
printf "${CYAN}--- Phase 16: Watchlist Set Active ---${NC}\n"

# Read current active watchlist
ORIG_ACTIVE_WL=$(curl -s "$BASE/api/v1/watchlists/active" | jq -r '.id // empty')

# Create temp watchlist
WL16_RESP=$(req "POST create temp watchlist" "POST" "/api/v1/watchlists" '{"name":"__smoke_active_test__"}')
WL16_ID=$(echo "$WL16_RESP" | jq -r '.id // empty')

if [ -n "$WL16_ID" ]; then
  # Set active to temp
  req "PUT set active watchlist" "PUT" "/api/v1/watchlists/active" "{\"id\":\"$WL16_ID\"}" >/dev/null
  sleep 0.5

  # Verify active changed
  NEW_ACTIVE=$(curl -s "$BASE/api/v1/watchlists/active" | jq -r '.id // empty')
  if [ "$NEW_ACTIVE" = "$WL16_ID" ]; then
    ok "verify active watchlist changed"
  else
    fail "verify active watchlist changed" "expected $WL16_ID got $NEW_ACTIVE"
  fi

  # Restore original active
  if [ -n "$ORIG_ACTIVE_WL" ]; then
    req "PUT restore active watchlist" "PUT" "/api/v1/watchlists/active" "{\"id\":\"$ORIG_ACTIVE_WL\"}" >/dev/null
  fi

  # Delete temp watchlist
  req_status "DELETE temp watchlist" "DELETE" "/api/v1/watchlist/$WL16_ID" "204"
else
  skip "watchlist set active" "could not create temp watchlist"
fi

printf "\n"

###############################################################################
printf "${CYAN}--- Phase 17: Alert Read & Fires ---${NC}\n"

# List alerts
ALERTS_JSON=$(get "GET alerts list (phase17)" "/api/v1/alerts")
FIRST_ALERT_ID=$(echo "$ALERTS_JSON" | jq -r '.alerts[0].alert_id // empty')

if [ -n "$FIRST_ALERT_ID" ]; then
  get "GET single alert" "/api/v1/alerts/$FIRST_ALERT_ID" >/dev/null
else
  skip "GET single alert" "no alerts exist"
fi

# Delete all fires (safe — clears fired notifications)
req "DELETE all alert fires" "DELETE" "/api/v1/alerts/fires/all" >/dev/null

printf "\n"

###############################################################################
printf "${CYAN}--- Phase 18: Layout Clone ---${NC}\n"

ORIG_LAYOUT_URL=$(curl -s "$BASE/api/v1/layout/status" | jq -r '.layout_id // empty')
ORIG_LAYOUT_NUM_ID=$(curl -s "$BASE/api/v1/layouts" | jq -r "[.layouts[] | select(.url==\"$ORIG_LAYOUT_URL\")][0].id // empty")
req "POST clone layout" "POST" "/api/v1/layout/clone" '{"name":"__smoke_clone__"}' >/dev/null

# Poll up to 8 seconds for the clone to appear in the list
CLONE_NUM_ID=""
for i in 1 2 3 4; do
  sleep 2
  CLONE_NUM_ID=$(curl -s "$BASE/api/v1/layouts" | jq -r '[.layouts[] | select(.name=="__smoke_clone__")][0].id // empty')
  [ -n "$CLONE_NUM_ID" ] && break
done

if [ -n "$CLONE_NUM_ID" ]; then
  ok "verify layout clone created"

  # Clean up: switch back to original layout, then delete the clone
  if [ -n "$ORIG_LAYOUT_NUM_ID" ]; then
    req "POST switch to original layout" "POST" "/api/v1/layout/switch" "{\"id\":$ORIG_LAYOUT_NUM_ID}" >/dev/null
    sleep 2
    req_status "DELETE clone layout" "DELETE" "/api/v1/layout/$CLONE_NUM_ID" "204"
  fi
else
  skip "verify layout clone created" "clone not found in layout list after 8s"
fi

printf "\n"

###############################################################################
printf "${CYAN}--- Phase 19: Page Reload (final) ---${NC}\n"

req "POST page reload" "POST" "/api/v1/page/reload" '{"mode":"normal"}' >/dev/null
sleep 5

# Verify controller still responds after reload
get "GET health (post-reload)" "/health" >/dev/null
sleep 2

# Verify charts reconnected
CHARTS_POST=$(get "GET charts (post-reload)" "/api/v1/charts")
CHART_COUNT_POST=$(echo "$CHARTS_POST" | jq '.charts | length')
if [ "$CHART_COUNT_POST" -gt 0 ] 2>/dev/null; then
  ok "verify charts reconnected after reload"
else
  fail "verify charts reconnected after reload" "got $CHART_COUNT_POST charts"
fi

printf "\n"

###############################################################################
END_TIME=$(date +%s)
ELAPSED=$((END_TIME - START_TIME))
TOTAL=$((PASS + FAIL + SKIP))

printf "${CYAN}=== Summary ===${NC}\n"
printf "  Total: %d  |  ${GREEN}Pass: %d${NC}  |  ${RED}Fail: %d${NC}  |  ${YELLOW}Skip: %d${NC}\n" "$TOTAL" "$PASS" "$FAIL" "$SKIP"
printf "  Duration: %ds\n" "$ELAPSED"

if [ "$FAIL" -gt 0 ]; then
  printf "\n${RED}Failures:${NC}%b\n" "$ERRORS"
  exit 1
fi
