#!/usr/bin/env python3
"""Scan Liquid DnB tracks for the highest-energy 60-second drop section.

Extracts audio from each .mp4, computes RMS energy in sliding windows,
finds the 60s window with maximum energy variance (quiet→loud transition),
and extracts that section as clip.mp3.
"""

import os
import sys
import json
import tempfile
import subprocess
from pathlib import Path

import numpy as np
import librosa

MUSIC_DIR = Path.home() / "Videos" / "Screencasts" / "MUSIC"
OUTPUT_DIR = Path(__file__).parent / "output"
CLIP_DURATION = 60  # seconds
SCAN_LIMIT = 600  # scan first 10 minutes of each track


def extract_audio_to_wav(video_path: Path, wav_path: Path, duration: int = SCAN_LIMIT):
    """Extract audio from video to WAV for librosa processing."""
    subprocess.run(
        [
            "ffmpeg", "-y", "-i", str(video_path),
            "-t", str(duration),
            "-vn", "-ac", "1", "-ar", "22050",
            "-acodec", "pcm_s16le",
            str(wav_path),
        ],
        capture_output=True,
        check=True,
    )


def find_best_window(y: np.ndarray, sr: int, window_sec: int = CLIP_DURATION) -> tuple[float, float]:
    """Find the window_sec window with highest energy variance (the drop).

    Returns (start_time, score).
    The score combines:
    - Energy variance within the window (captures quiet→loud transition)
    - Peak energy (favors loud sections)
    """
    hop = sr  # 1-second hops
    window_samples = window_sec * sr
    total_samples = len(y)

    if total_samples < window_samples:
        return 0.0, 0.0

    # Compute frame-level RMS energy (1-second frames)
    frame_length = sr
    hop_length = sr
    rms = librosa.feature.rms(y=y, frame_length=frame_length, hop_length=hop_length)[0]

    best_score = -1.0
    best_start = 0.0
    window_frames = window_sec  # 1 frame per second

    for i in range(len(rms) - window_frames + 1):
        chunk = rms[i : i + window_frames]
        variance = np.var(chunk)
        peak = np.max(chunk)
        mean = np.mean(chunk)
        # Score: variance captures the drop transition, peak/mean favor loud+dynamic sections
        score = variance * 10 + peak * 2 + mean
        if score > best_score:
            best_score = score
            best_start = float(i)  # in seconds (1 frame = 1 second)

    return best_start, best_score


def extract_clip(video_path: Path, start_sec: float, output_path: Path):
    """Extract 60s MP3 clip from the source video."""
    subprocess.run(
        [
            "ffmpeg", "-y",
            "-ss", str(start_sec),
            "-t", str(CLIP_DURATION),
            "-i", str(video_path),
            "-vn", "-acodec", "libmp3lame", "-q:a", "2",
            str(output_path),
        ],
        capture_output=True,
        check=True,
    )


def main():
    OUTPUT_DIR.mkdir(parents=True, exist_ok=True)

    videos = sorted(MUSIC_DIR.glob("*.mp4"))
    if not videos:
        print(f"No .mp4 files found in {MUSIC_DIR}", file=sys.stderr)
        sys.exit(1)

    print(f"Scanning {len(videos)} tracks for best 60s drop...\n")

    best_overall = {"score": -1, "file": None, "start": 0}

    for video in videos:
        name = video.stem
        print(f"  [{name}]")

        with tempfile.NamedTemporaryFile(suffix=".wav", delete=True) as tmp:
            try:
                extract_audio_to_wav(video, Path(tmp.name))
                y, sr = librosa.load(tmp.name, sr=22050, mono=True)
            except Exception as e:
                print(f"    SKIP: {e}")
                continue

        duration = len(y) / sr
        print(f"    Duration: {duration:.0f}s, SR: {sr}")

        start, score = find_best_window(y, sr)
        print(f"    Best window: {start:.0f}s – {start + CLIP_DURATION:.0f}s (score: {score:.4f})")

        if score > best_overall["score"]:
            best_overall = {"score": score, "file": str(video), "start": start}

    print(f"\n{'='*60}")
    print(f"WINNER: {Path(best_overall['file']).stem}")
    print(f"  Start: {best_overall['start']:.0f}s")
    print(f"  Score: {best_overall['score']:.4f}")
    print(f"{'='*60}\n")

    clip_path = OUTPUT_DIR / "clip.mp3"
    print(f"Extracting clip to {clip_path}...")
    extract_clip(Path(best_overall["file"]), best_overall["start"], clip_path)
    print(f"Done! Clip saved: {clip_path}")

    # Save metadata for downstream scripts
    meta = {
        "source_file": best_overall["file"],
        "start_sec": float(best_overall["start"]),
        "duration_sec": CLIP_DURATION,
        "score": float(best_overall["score"]),
    }
    meta_path = OUTPUT_DIR / "clip_meta.json"
    with open(meta_path, "w") as f:
        json.dump(meta, f, indent=2)
    print(f"Metadata saved: {meta_path}")


if __name__ == "__main__":
    main()
