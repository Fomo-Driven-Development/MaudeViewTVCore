#!/usr/bin/env python3
"""Draw red coverage boxes on TradingView chart screenshot — API coverage map."""

from PIL import Image, ImageDraw, ImageFont
import os

IMG_PATH = os.path.join(os.path.dirname(__file__), "..", "docs", "chart.png")
OUT_PATH = os.path.join(os.path.dirname(__file__), "..", "docs", "chart_coverage.png")

# Image is 2303x1449

# Each region: (x1, y1, x2, y2, label, label_pos)
# label_pos: "inside" (default), "above", "below", "right", or (x, y) tuple
REGIONS = [
    # === Top toolbar buttons ===
    # Symbol selector (BTCUSD.P dropdown)
    (38, 56, 162, 84, "Symbol", "above"),
    # Resolution buttons (1m 5m 15m 1h 4h 6h 12h D 3D W 2W M)
    (195, 56, 620, 84, "Resolution", "above"),
    # Chart type icon (candle/bar selector)
    (640, 56, 680, 84, "Chart Type", "above"),
    # Compare / overlay (the ⊕ icon area)
    # Indicators button
    (795, 56, 905, 84, "Indicators", "above"),
    # Alert button
    (1088, 56, 1140, 84, "Alert", "above"),
    # Replay button
    (1152, 56, 1243, 84, "Replay", "above"),

    # Layout selector ("Untitled" dropdown, top-right area)
    (1720, 56, 1830, 84, "Layout", "above"),

    # === Left sidebar — drawing tools ===
    (0, 90, 42, 480, "Drawing Tools", "right"),

    # === Main chart area ===
    (43, 90, 2155, 1345, "Chart Viewport (zoom / scroll / visible-range / go-to-date / reset-view / snapshot)", "inside"),

    # OHLC data bar + Buy/Sell
    (43, 88, 460, 130, "OHLC + Trade", "inside"),

    # === Price axis (right) ===
    (2155, 90, 2270, 1345, "Price Axis", "inside"),

    # === Time axis (bottom of chart) ===
    (43, 1345, 2155, 1380, "Time Axis", "below"),

    # === Timeframe presets bar (very bottom) ===
    (43, 1400, 365, 1435, "Timeframe Presets", "below"),

    # === Right sidebar icons (watchlist, pine editor, alerts panel, etc.) ===
    (2270, 90, 2303, 1345, "Right Sidebar\n(Watchlist / Pine / Alerts)", "left"),
]

# Colors
BOX_OUTLINE = (255, 50, 50, 220)
BOX_FILL    = (255, 40, 40, 30)
LABEL_BG    = (200, 20, 20, 230)
LABEL_FG    = (255, 255, 255, 255)
LINE_W      = 3


def find_font(size):
    for fp in [
        "/usr/share/fonts/dejavu-sans-fonts/DejaVuSans-Bold.ttf",
        "/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf",
        "/usr/share/fonts/liberation-sans/LiberationSans-Bold.ttf",
        "/usr/share/fonts/google-noto/NotoSans-Bold.ttf",
    ]:
        if os.path.exists(fp):
            return ImageFont.truetype(fp, size)
    return ImageFont.load_default()


def text_size(font, text):
    """Return (width, height) for a single line of text."""
    bbox = font.getbbox(text)
    return bbox[2] - bbox[0], bbox[3] - bbox[1]


def draw_label(draw, font, label, x1, y1, x2, y2, pos):
    lines = label.split("\n")
    line_sizes = [text_size(font, l) for l in lines]
    max_w = max(w for w, h in line_sizes)
    line_h = max(h for w, h in line_sizes)
    pad = 5
    gap = 3
    total_h = line_h * len(lines) + gap * (len(lines) - 1)
    block_w = max_w + pad * 2
    block_h = total_h + pad * 2

    # Determine label anchor point
    if pos == "inside":
        lx = x1 + 8
        ly = y1 + 8
    elif pos == "above":
        lx = x1
        ly = y1 - block_h - 6
    elif pos == "below":
        lx = x1
        ly = y2 + 6
    elif pos == "right":
        lx = x2 + 8
        ly = y1
    elif pos == "left":
        lx = x1 - block_w - 8
        ly = y1
    else:
        lx, ly = pos

    # Draw background
    draw.rectangle([lx, ly, lx + block_w, ly + block_h], fill=LABEL_BG)
    # Draw text
    for i, line in enumerate(lines):
        tx = lx + pad
        ty = ly + pad + i * (line_h + gap)
        draw.text((tx, ty), line, fill=LABEL_FG, font=font)


def main():
    img = Image.open(IMG_PATH).convert("RGBA")
    overlay = Image.new("RGBA", img.size, (0, 0, 0, 0))
    draw = ImageDraw.Draw(overlay)

    font = find_font(20)

    for entry in REGIONS:
        x1, y1, x2, y2, label, pos = entry

        # Box outline
        for i in range(LINE_W):
            draw.rectangle([x1 - i, y1 - i, x2 + i, y2 + i], outline=BOX_OUTLINE)

        # Semi-transparent fill (composite per-box to avoid stacking opacity)
        fill_layer = Image.new("RGBA", img.size, (0, 0, 0, 0))
        ImageDraw.Draw(fill_layer).rectangle([x1, y1, x2, y2], fill=BOX_FILL)
        overlay = Image.alpha_composite(overlay, fill_layer)
        draw = ImageDraw.Draw(overlay)

        # Label
        draw_label(draw, font, label, x1, y1, x2, y2, pos)

    result = Image.alpha_composite(img, overlay)
    result = result.convert("RGB")
    result.save(OUT_PATH, quality=95)
    print(f"Saved: {OUT_PATH}")


if __name__ == "__main__":
    main()
