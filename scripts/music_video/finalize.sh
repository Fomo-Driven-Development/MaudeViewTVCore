#!/usr/bin/env bash
# Mux raw screen recordings with the audio clip to produce final music videos.
#
# Usage: bash finalize.sh [style]
#   No args = finalize all styles
#   With arg = finalize just that style (symbol_surf, replay, chaos, drawing)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
OUTPUT_DIR="$SCRIPT_DIR/output"
FINAL_DIR="$OUTPUT_DIR/final"
CLIP="$OUTPUT_DIR/clip.mp3"

mkdir -p "$FINAL_DIR"

if [ ! -f "$CLIP" ]; then
    echo "ERROR: clip.mp3 not found at $CLIP"
    exit 1
fi

STYLES=("symbol_surf" "replay" "chaos" "drawing")

if [ $# -gt 0 ]; then
    STYLES=("$1")
fi

for style in "${STYLES[@]}"; do
    raw="$OUTPUT_DIR/raw_${style}.mkv"
    final="$FINAL_DIR/music_video_${style}.mp4"

    if [ ! -f "$raw" ]; then
        echo "SKIP: $raw not found"
        continue
    fi

    echo "Muxing: $style"
    echo "  Video: $raw"
    echo "  Audio: $CLIP"
    echo "  Output: $final"

    ffmpeg -y \
        -i "$raw" \
        -i "$CLIP" \
        -c:v copy \
        -c:a aac -b:a 192k \
        -shortest \
        -movflags +faststart \
        "$final"

    echo "  Done: $final"
    echo ""
done

echo "All finalizations complete!"
echo "Output files:"
ls -lh "$FINAL_DIR"/*.mp4 2>/dev/null || echo "  (none found)"
