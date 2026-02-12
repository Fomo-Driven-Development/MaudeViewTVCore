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
  get "GET snapshot meta" "/api/v1/snapshots/$SNAP_ID" >/dev/null
  req "DELETE snapshot" "DELETE" "/api/v1/snapshots/$SNAP_ID" >/dev/null
else
  skip "snapshot CRUD" "could not take snapshot"
fi

printf "\n"

###############################################################################
printf "${CYAN}--- Phase 9: Pine Editor ---${NC}\n"

PINE_STATUS=$(curl -s "$BASE/api/v1/pine/status" | jq -r '.is_visible // "false"')
if [ "$PINE_STATUS" = "false" ]; then
  req "POST pine toggle open" "POST" "/api/v1/pine/toggle" >/dev/null
  sleep 1
fi

PINE_VIS=$(curl -s "$BASE/api/v1/pine/status" | jq -r '.is_visible // "false"')
if [ "$PINE_VIS" = "true" ]; then
  get "GET pine source" "/api/v1/pine/source" >/dev/null
  req "PUT pine source" "PUT" "/api/v1/pine/source" '{"source":"// smoke test\nplot(close)"}' >/dev/null
  sleep 0.5
  req "POST pine save" "POST" "/api/v1/pine/save" >/dev/null
  sleep 0.5
  # Close editor
  req "POST pine toggle close" "POST" "/api/v1/pine/toggle" >/dev/null
  sleep 0.5
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
