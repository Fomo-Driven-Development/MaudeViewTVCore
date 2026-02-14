#!/usr/bin/env python3
"""Generate 4 choreography style JSONs from beats.json.

Each style maps beat/onset events to timed TV controller API calls.
"""

import json
import sys
from pathlib import Path

OUTPUT_DIR = Path(__file__).parent / "output"

# Symbol lists for cycling
SYMBOLS = [
    "BTCUSD", "ETHUSD", "SPY", "AAPL", "TSLA", "NVDA", "AMZN",
    "EURUSD", "XAUUSD", "SOLUSDT", "GOOGL", "MSFT", "META",
    "ADAUSD", "DOTUSD", "AVAXUSD", "LINKUSD", "MATICUSD",
]

TIMEFRAMES = ["1D", "5D", "1M", "3M", "6M", "1Y"]
RESOLUTIONS = ["1", "5", "15", "60", "1D"]

CHART_TYPES = ["candles", "bars", "line", "area", "heikin_ashi", "hollow_candles", "baseline"]

# Overlay studies that render on the main chart pane (no subplots)
OVERLAY_STUDIES = [
    "Moving Average Exponential", "Bollinger Bands", "Ichimoku Cloud",
    "Moving Average", "Volume Weighted Average Price", "Parabolic SAR",
]

# Subplot studies (use sparingly)
SUBPLOT_STUDIES = ["RSI", "MACD", "Volume"]

CHART_ID = 0
BASE = f"/api/v1/chart/{CHART_ID}"


def action(t: float, method: str, path: str, body: dict | None = None) -> dict:
    a = {"t": round(t, 3), "method": method, "path": path}
    if body:
        a["body"] = body
    return a


def gen_symbol_surf(data: dict) -> list[dict]:
    """Style 1: Symbol Surfing - change symbols on beats, timeframes every 4th beat."""
    actions = []
    beats = data["beats"]
    sym_idx = 0
    tf_idx = 0
    ct_idx = 0

    for i, beat in enumerate(beats):
        t = beat["t"]

        # Every beat → change symbol (query param)
        sym = SYMBOLS[sym_idx % len(SYMBOLS)]
        actions.append(action(t, "PUT", f"{BASE}/symbol?symbol={sym}"))
        sym_idx += 1

        # Every 4th beat → change timeframe (query params)
        if i % 4 == 0:
            preset = TIMEFRAMES[tf_idx % len(TIMEFRAMES)]
            actions.append(action(t + 0.05, "PUT", f"{BASE}/timeframe?preset={preset}&resolution=1D"))
            tf_idx += 1

        # High energy beats → change chart type (query param)
        if beat.get("energy", 0) > 0.7 and i % 3 == 0:
            ct = CHART_TYPES[ct_idx % len(CHART_TYPES)]
            actions.append(action(t + 0.02, "PUT", f"{BASE}/chart-type?type={ct}"))
            ct_idx += 1

        # Very strong beats → zoom out then snap back
        if beat.get("strength", 0) > 0.85 and i % 6 == 0:
            actions.append(action(t + 0.03, "POST", f"{BASE}/zoom", {"direction": "out"}))
            actions.append(action(t + 0.3, "POST", f"{BASE}/reset-view"))

    return actions


def gen_replay(data: dict) -> list[dict]:
    """Style 2: Replay Timelapse - step through bars synced to beats."""
    actions = []
    beats = data["beats"]

    # Start: activate replay mode on crypto
    actions.append(action(0.0, "PUT", f"{BASE}/symbol?symbol=BTCUSD"))
    actions.append(action(0.3, "PUT", f"{BASE}/resolution?resolution=5"))
    actions.append(action(0.8, "POST", f"{BASE}/replay/activate/auto", None))
    actions.append(action(2.0, "POST", f"{BASE}/replay/autoplay/stop", None))

    delay_values = [1.0, 0.5, 0.25, 0.1]
    delay_idx = 0

    for i, beat in enumerate(beats):
        t = beat["t"]
        if t < 2.5:
            continue  # skip setup period

        strength = beat.get("strength", 0.5)

        # Normal beats → step 1 bar
        # Strong beats → step 3 bars
        steps = 3 if strength > 0.7 else 1
        actions.append(action(t, "POST", f"{BASE}/replay/step", {"count": steps}))

        # Every 8th beat → adjust autoplay delay based on energy
        if i % 8 == 0:
            energy = beat.get("energy", 0.5)
            if energy > 0.6:
                delay_idx = min(delay_idx + 1, len(delay_values) - 1)
            else:
                delay_idx = max(delay_idx - 1, 0)
            actions.append(action(t + 0.02, "PUT", f"{BASE}/replay/autoplay/delay",
                                  {"delay": delay_values[delay_idx]}))

        # Drop sections → zoom in, burst autoplay
        if strength > 0.9 and i % 12 == 0:
            actions.append(action(t + 0.05, "POST", f"{BASE}/zoom", {"direction": "in"}))
            actions.append(action(t + 0.1, "POST", f"{BASE}/replay/autoplay/start", None))
            actions.append(action(t + 2.0, "POST", f"{BASE}/replay/autoplay/stop", None))
            actions.append(action(t + 2.1, "POST", f"{BASE}/zoom", {"direction": "out"}))

    # End: deactivate replay
    duration = data["duration"]
    actions.append(action(duration - 1.0, "POST", f"{BASE}/replay/deactivate", None))

    return actions


def gen_chaos(data: dict) -> list[dict]:
    """Style 3: Full Chaos - everything changes all the time."""
    actions = []
    beats = data["beats"]
    onsets = data["onsets"]

    sym_idx = 0
    ct_idx = 0
    tf_idx = 0

    # Beats → cycle symbols
    for i, beat in enumerate(beats):
        t = beat["t"]
        sym = SYMBOLS[sym_idx % len(SYMBOLS)]
        actions.append(action(t, "PUT", f"{BASE}/symbol?symbol={sym}"))
        sym_idx += 1

        # Every 6th beat → shift timeframe
        if i % 6 == 0:
            preset = TIMEFRAMES[tf_idx % len(TIMEFRAMES)]
            actions.append(action(t + 0.02, "PUT", f"{BASE}/timeframe?preset={preset}&resolution=1D"))
            tf_idx += 1

        # High energy → change chart type
        if beat.get("energy", 0) > 0.65 and i % 4 == 0:
            ct = CHART_TYPES[ct_idx % len(CHART_TYPES)]
            actions.append(action(t + 0.03, "PUT", f"{BASE}/chart-type?type={ct}"))
            ct_idx += 1

        # Very strong beats → zoom out + symbol + chart type combo
        if beat.get("strength", 0) > 0.8 and i % 8 == 0:
            actions.append(action(t + 0.04, "POST", f"{BASE}/zoom", {"direction": "out"}))

    # Onsets → add a limited set of overlay studies (no subplot bloat)
    MAX_STUDIES = 8  # hard cap on total studies added
    overlay_idx = 0
    added = 0

    for i, onset in enumerate(onsets):
        if added >= MAX_STUDIES:
            break
        if i % 12 != 0:  # spread them out
            continue
        t = onset["t"]
        name = OVERLAY_STUDIES[overlay_idx % len(OVERLAY_STUDIES)]
        overlay_idx += 1
        actions.append(action(t + 0.01, "POST", f"{BASE}/indicators/add", {"query": name, "index": 0}))
        added += 1

    # Sort all actions by time
    actions.sort(key=lambda a: a["t"])
    return actions


def gen_drawing(data: dict) -> list[dict]:
    """Style 4: Drawing Art - create drawings synced to beats.

    Uses "__RAND__" as coordinate placeholders — perform.py replaces these
    with actual chart coordinates at runtime via the visible-range API.
    """
    actions = []
    beats = data["beats"]

    # Valid snake_case tool names from GET /drawings/shapes
    tools = ["trend_line", "horizontal_line", "rectangle", "fib_retracement", "arrow_up", "text"]
    single_shapes = ["horizontal_line", "arrow_up", "arrow_down", "flag", "text", "price_label"]
    multi_shapes = ["trend_line", "rectangle", "fib_retracement", "regression_trend", "ray", "info_line"]
    tool_idx = 0
    single_idx = 0
    multi_idx = 0
    drawing_count = 0

    # Start with a clean slate
    actions.append(action(0.0, "PUT", f"{BASE}/symbol?symbol=BTCUSD"))
    actions.append(action(0.2, "PUT", f"{BASE}/timeframe?preset=1M&resolution=15"))
    actions.append(action(0.5, "DELETE", f"{BASE}/drawings"))

    for i, beat in enumerate(beats):
        t = beat["t"]
        if t < 1.0:
            continue

        strength = beat.get("strength", 0.5)
        energy = beat.get("energy", 0.5)

        # Every beat → create a single-point drawing (cycle through shapes)
        if i % 2 == 0:
            shape = single_shapes[single_idx % len(single_shapes)]
            single_idx += 1
            actions.append(action(t, "POST", f"{BASE}/drawings",
                                  {"point": {"time": "__RAND_TIME__", "price": "__RAND_PRICE__"},
                                   "options": {"shape": shape}}))
            drawing_count += 1

        # Every 3rd beat → create multi-point drawing (cycle through shapes)
        if i % 3 == 0:
            shape = multi_shapes[multi_idx % len(multi_shapes)]
            multi_idx += 1
            actions.append(action(t + 0.02, "POST", f"{BASE}/drawings/multipoint",
                                  {"points": [{"time": "__RAND_TIME__", "price": "__RAND_PRICE__"},
                                              {"time": "__RAND_TIME2__", "price": "__RAND_PRICE2__"}],
                                   "options": {"shape": shape}}))
            drawing_count += 1

        # Strong beats → change drawing tool
        if strength > 0.7 and i % 4 == 0:
            tool = tools[tool_idx % len(tools)]
            actions.append(action(t + 0.03, "PUT", f"{BASE}/drawings/tool", {"tool": tool}))
            tool_idx += 1

        # Drop moments → clear all and start fresh
        if energy > 0.85 and strength > 0.8 and drawing_count > 10:
            actions.append(action(t + 0.04, "DELETE", f"{BASE}/drawings"))
            drawing_count = 0

        # Flash hide/show on every 5th beat for visual effect
        if i % 5 == 0:
            actions.append(action(t + 0.05, "PUT", f"{BASE}/drawings/toggles/hide", {"value": True}))
            actions.append(action(t + 0.3, "PUT", f"{BASE}/drawings/toggles/hide", {"value": False}))

    actions.sort(key=lambda a: a["t"])
    return actions


def main():
    beats_path = OUTPUT_DIR / "beats.json"
    if not beats_path.exists():
        print(f"beats.json not found at {beats_path}. Run analyze_beats.py first.", file=sys.stderr)
        sys.exit(1)

    with open(beats_path) as f:
        data = json.load(f)

    print(f"Loaded beats.json: BPM={data['bpm']}, beats={data['beat_count']}, onsets={data['onset_count']}")

    styles = {
        "symbol_surf": gen_symbol_surf,
        "replay": gen_replay,
        "chaos": gen_chaos,
        "drawing": gen_drawing,
    }

    for name, generator in styles.items():
        actions = generator(data)
        output = {
            "style": name,
            "bpm": data["bpm"],
            "duration": data["duration"],
            "action_count": len(actions),
            "actions": actions,
        }
        out_path = OUTPUT_DIR / f"style_{name}.json"
        with open(out_path, "w") as f:
            json.dump(output, f, indent=2)
        print(f"  {name}: {len(actions)} actions → {out_path}")


if __name__ == "__main__":
    main()
