#!/usr/bin/env python3
"""Execute a choreography style against the TV controller.

OBS recording is handled externally (via MCP). This script only does:
  1. Prep the chart
  2. Play audio
  3. Fire timed API calls

Usage:
    python3 perform.py --style symbol_surf
    python3 perform.py --style replay
    python3 perform.py --style chaos
    python3 perform.py --style drawing
"""

import argparse
import asyncio
import json
import random
import subprocess
import sys
import time
from pathlib import Path

import aiohttp

OUTPUT_DIR = Path(__file__).parent / "output"
CONTROLLER_BASE = "http://127.0.0.1:8188"
BUFFER_SEC = 3  # extra seconds before/after for clean cuts

ALL_STYLES = ["symbol_surf", "replay", "chaos", "drawing"]


async def discover_chart_id() -> str:
    """Query the controller to find the first available chart ID."""
    async with aiohttp.ClientSession() as s:
        async with s.get(f"{CONTROLLER_BASE}/api/v1/charts") as r:
            data = await r.json()
            charts = data.get("charts", [])
            if not charts:
                raise RuntimeError("No charts found — is TradingView open?")
            return charts[0]["chart_id"]


async def fetch_chart_bounds(session: aiohttp.ClientSession, chart_id: str) -> dict:
    """Get visible time range and estimate price bounds from chart title."""
    # Time range from visible-range endpoint
    async with session.get(f"{CONTROLLER_BASE}/api/v1/chart/{chart_id}/visible-range") as r:
        vr = await r.json()

    # Parse current price from chart title (format: "BTCUSD 68,896 ▲ +0.12% ...")
    price = None
    async with session.get(f"{CONTROLLER_BASE}/api/v1/charts") as r:
        data = await r.json()
        for chart in data.get("charts", []):
            if chart.get("chart_id") == chart_id:
                import re
                match = re.search(r'[\d,]+\.?\d*', chart.get("title", "").split("▲")[0].split("▼")[0])
                if match:
                    price = float(match.group().replace(",", ""))
                break

    if price is None or price == 0:
        price = 50000  # fallback

    return {
        "time_from": vr["from"],
        "time_to": vr["to"],
        "price": price,
    }


def resolve_drawing_placeholders(actions: list[dict], bounds: dict):
    """Replace __RAND_TIME/PRICE__ placeholders with random coordinates in the visible range."""
    time_from = bounds["time_from"]
    time_to = bounds["time_to"]
    price = bounds["price"]
    time_span = time_to - time_from

    for act in actions:
        body = act.get("body")
        if body is None:
            continue
        raw = json.dumps(body)
        if "__RAND_" not in raw:
            continue
        # Each action gets fresh random values — scatter drawings across the visible chart
        t1 = time_from + random.random() * time_span
        t2 = time_from + (0.3 + random.random() * 0.5) * time_span
        p1 = price * (0.95 + random.random() * 0.10)  # ±5% of current price
        p2 = price * (0.95 + random.random() * 0.10)
        raw = raw.replace('"__RAND_TIME__"', str(int(t1)))
        raw = raw.replace('"__RAND_TIME2__"', str(int(t2)))
        raw = raw.replace('"__RAND_PRICE__"', f"{p1:.2f}")
        raw = raw.replace('"__RAND_PRICE2__"', f"{p2:.2f}")
        act["body"] = json.loads(raw)


async def prep_chart(session: aiohttp.ClientSession, style: str, chart_id: str):
    """Reset chart to clean state before performing."""
    print(f"  Prepping chart for {style}...")

    async with session.post(f"{CONTROLLER_BASE}/api/v1/chart/{chart_id}/reset-view") as r:
        pass
    async with session.delete(f"{CONTROLLER_BASE}/api/v1/chart/{chart_id}/drawings") as r:
        pass
    async with session.put(f"{CONTROLLER_BASE}/api/v1/chart/{chart_id}/symbol?symbol=BTCUSD") as r:
        pass
    async with session.put(f"{CONTROLLER_BASE}/api/v1/chart/{chart_id}/chart-type?type=candles") as r:
        pass

    await asyncio.sleep(1)
    print("  Chart prepped.")


async def execute_actions(session: aiohttp.ClientSession, actions: list[dict], start_time: float):
    """Execute timed actions against the controller API."""
    for i, act in enumerate(actions):
        target_time = act["t"]
        now = time.monotonic() - start_time

        delay = target_time - now
        if delay > 0:
            await asyncio.sleep(delay)

        method = act["method"]
        path = act["path"]
        body = act.get("body")
        url = f"{CONTROLLER_BASE}{path}"

        try:
            if method == "GET":
                async with session.get(url) as r:
                    pass
            elif method == "PUT":
                if body:
                    async with session.put(url, json=body) as r:
                        pass
                else:
                    async with session.put(url) as r:
                        pass
            elif method == "POST":
                if body:
                    async with session.post(url, json=body) as r:
                        pass
                else:
                    async with session.post(url) as r:
                        pass
            elif method == "DELETE":
                if body:
                    async with session.delete(url, json=body) as r:
                        pass
                else:
                    async with session.delete(url) as r:
                        pass
        except Exception:
            pass  # Don't let a single failed call break timing

        if (i + 1) % 20 == 0:
            elapsed = time.monotonic() - start_time
            print(f"    [{elapsed:.1f}s] {i+1}/{len(actions)} actions fired")


async def perform_style(style: str):
    """Perform a single style: prep chart, play audio, execute choreography."""
    style_path = OUTPUT_DIR / f"style_{style}.json"
    clip_path = OUTPUT_DIR / "clip.mp3"

    if not style_path.exists():
        print(f"  Style JSON not found: {style_path}", file=sys.stderr)
        return False

    if not clip_path.exists():
        print(f"  clip.mp3 not found: {clip_path}", file=sys.stderr)
        return False

    with open(style_path) as f:
        choreo = json.load(f)

    chart_id = await discover_chart_id()
    print(f"  Discovered chart ID: {chart_id}")

    actions = choreo["actions"]
    for a in actions:
        a["path"] = a["path"].replace("/chart/0/", f"/chart/{chart_id}/")

    duration = choreo["duration"]

    print(f"\n{'='*60}")
    print(f"PERFORMING: {style} ({len(actions)} actions, {duration:.0f}s)")
    print(f"{'='*60}")

    async with aiohttp.ClientSession(timeout=aiohttp.ClientTimeout(total=10)) as session:
        try:
            async with session.get(f"{CONTROLLER_BASE}/health") as r:
                if r.status != 200:
                    print(f"  Controller not healthy: {r.status}", file=sys.stderr)
                    return False
        except Exception:
            print("  Controller not reachable", file=sys.stderr)
            return False

        await prep_chart(session, style, chart_id)

        # For drawing style: resolve coordinate placeholders with real chart data
        if style == "drawing":
            bounds = await fetch_chart_bounds(session, chart_id)
            print(f"  Chart bounds: time={bounds['time_from']}-{bounds['time_to']}, price≈{bounds['price']:.0f}")
            resolve_drawing_placeholders(actions, bounds)

        # Signal ready — OBS recording should already be started externally
        print("  READY — starting audio + choreography now")

        # Start audio playback
        player = subprocess.Popen(
            ["ffplay", "-nodisp", "-autoexit", str(clip_path)],
            stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL,
        )

        # Execute choreography
        print(f"  Executing {len(actions)} actions...")
        start_time = time.monotonic()
        await execute_actions(session, actions, start_time)

        # Wait for remaining duration
        elapsed = time.monotonic() - start_time
        remaining = duration - elapsed + BUFFER_SEC
        if remaining > 0:
            print(f"  Waiting {remaining:.1f}s for clip to finish...")
            await asyncio.sleep(remaining)

        # Stop audio player
        player.terminate()
        try:
            player.wait(timeout=5)
        except subprocess.TimeoutExpired:
            player.kill()

    print(f"  Performance complete — stop OBS recording externally")
    return True


async def main():
    parser = argparse.ArgumentParser(description="Perform beat-synced TradingView choreography")
    parser.add_argument("--style", required=True, choices=ALL_STYLES + ["all"],
                        help="Which choreography style to perform")
    args = parser.parse_args()

    styles = ALL_STYLES if args.style == "all" else [args.style]

    for style in styles:
        success = await perform_style(style)
        if not success:
            print(f"  FAILED: {style}", file=sys.stderr)
            if len(styles) > 1:
                continue
            sys.exit(1)
        print(f"  DONE: {style}")

        if len(styles) > 1:
            print("\n  Pausing 5s before next style...")
            await asyncio.sleep(5)

    print(f"\nAll performances complete!")


if __name__ == "__main__":
    asyncio.run(main())
