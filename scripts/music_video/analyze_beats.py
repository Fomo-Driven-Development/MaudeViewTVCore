#!/usr/bin/env python3
"""Analyze clip.mp3 for beats, onsets, energy, and sections.

Outputs beats.json with everything needed to generate choreographies.
"""

import json
import sys
from pathlib import Path

import numpy as np
import librosa


class NumpyEncoder(json.JSONEncoder):
    def default(self, obj):
        if isinstance(obj, (np.floating, np.float32, np.float64)):
            return float(obj)
        if isinstance(obj, (np.integer, np.int32, np.int64)):
            return int(obj)
        if isinstance(obj, np.ndarray):
            return obj.tolist()
        return super().default(obj)

OUTPUT_DIR = Path(__file__).parent / "output"


def detect_sections(energy_curve: list[list[float]], beats: list[dict]) -> list[dict]:
    """Segment the clip into sections based on energy transitions.

    Simple heuristic: split into 4 roughly equal segments and classify by
    relative energy level.
    """
    if not energy_curve:
        return []

    times = np.array([e[0] for e in energy_curve])
    values = np.array([e[1] for e in energy_curve])
    duration = times[-1]

    # Split into ~8s segments and classify
    seg_duration = 8.0
    sections = []
    t = 0.0
    while t < duration:
        end = min(t + seg_duration, duration)
        mask = (times >= t) & (times < end)
        if mask.any():
            seg_energy = float(np.mean(values[mask]))
        else:
            seg_energy = 0.0
        sections.append({"start": round(t, 2), "end": round(end, 2), "energy": seg_energy})
        t = end

    # Classify sections by energy relative to global mean
    global_mean = float(np.mean(values))
    global_std = float(np.std(values))

    for sec in sections:
        e = sec["energy"]
        if e > global_mean + global_std:
            sec["type"] = "drop"
        elif e > global_mean:
            sec["type"] = "high"
        elif e > global_mean - global_std * 0.5:
            sec["type"] = "mid"
        else:
            sec["type"] = "buildup"
        del sec["energy"]

    # Merge adjacent sections of same type
    merged = [sections[0]]
    for sec in sections[1:]:
        if sec["type"] == merged[-1]["type"]:
            merged[-1]["end"] = sec["end"]
        else:
            merged.append(sec)

    return merged


def main():
    clip_path = OUTPUT_DIR / "clip.mp3"
    if not clip_path.exists():
        print(f"clip.mp3 not found at {clip_path}. Run find_drop.py first.", file=sys.stderr)
        sys.exit(1)

    print(f"Loading {clip_path}...")
    y, sr = librosa.load(str(clip_path), sr=22050, mono=True)
    duration = len(y) / sr
    print(f"  Duration: {duration:.2f}s, SR: {sr}")

    # Beat tracking
    print("Detecting beats...")
    tempo, beat_frames = librosa.beat.beat_track(y=y, sr=sr, units="frames")
    beat_times = librosa.frames_to_time(beat_frames, sr=sr)

    # Handle tempo - it might be an array
    if hasattr(tempo, '__len__'):
        bpm = float(tempo[0])
    else:
        bpm = float(tempo)
    print(f"  BPM: {bpm:.1f}, Beats: {len(beat_times)}")

    # Onset detection
    print("Detecting onsets...")
    onset_frames = librosa.onset.onset_detect(y=y, sr=sr, units="frames")
    onset_times = librosa.frames_to_time(onset_frames, sr=sr)

    # Onset strength envelope (for beat strength)
    onset_env = librosa.onset.onset_strength(y=y, sr=sr)
    onset_env_times = librosa.times_like(onset_env, sr=sr)

    # RMS energy envelope
    print("Computing energy...")
    rms = librosa.feature.rms(y=y, frame_length=2048, hop_length=512)[0]
    rms_times = librosa.times_like(rms, sr=sr, hop_length=512)

    # Normalize RMS to 0-1
    rms_max = rms.max() if rms.max() > 0 else 1.0
    rms_norm = rms / rms_max

    # Spectral centroid (brightness)
    print("Computing spectral centroid...")
    centroid = librosa.feature.spectral_centroid(y=y, sr=sr, hop_length=512)[0]
    centroid_times = librosa.times_like(centroid, sr=sr, hop_length=512)
    centroid_max = centroid.max() if centroid.max() > 0 else 1.0
    centroid_norm = centroid / centroid_max

    # Build beat list with strength
    beats = []
    for bt in beat_times:
        # Find nearest onset envelope value for strength
        idx = np.argmin(np.abs(onset_env_times - bt))
        strength = float(onset_env[idx])
        # Find nearest RMS for energy at this beat
        rms_idx = np.argmin(np.abs(rms_times - bt))
        energy = float(rms_norm[rms_idx])
        beats.append({
            "t": round(float(bt), 3),
            "strength": round(strength / (onset_env.max() if onset_env.max() > 0 else 1), 3),
            "energy": round(energy, 3),
        })

    # Build onset list
    onsets = []
    for ot in onset_times:
        idx = np.argmin(np.abs(onset_env_times - ot))
        strength = float(onset_env[idx])
        onsets.append({
            "t": round(float(ot), 3),
            "strength": round(strength / (onset_env.max() if onset_env.max() > 0 else 1), 3),
        })

    # Downsample energy curve to ~0.1s resolution for JSON
    step = max(1, len(rms_norm) // 600)
    energy_curve = [
        [round(float(rms_times[i]), 2), round(float(rms_norm[i]), 3)]
        for i in range(0, len(rms_norm), step)
    ]

    # Detect sections
    sections = detect_sections(energy_curve, beats)

    result = {
        "bpm": round(bpm, 1),
        "duration": round(duration, 2),
        "beat_count": len(beats),
        "onset_count": len(onsets),
        "beats": beats,
        "onsets": onsets,
        "energy_curve": energy_curve,
        "sections": sections,
    }

    out_path = OUTPUT_DIR / "beats.json"
    with open(out_path, "w") as f:
        json.dump(result, f, indent=2, cls=NumpyEncoder)

    print(f"\nResults saved to {out_path}")
    print(f"  BPM: {bpm:.1f}")
    print(f"  Beats: {len(beats)}")
    print(f"  Onsets: {len(onsets)}")
    print(f"  Sections: {len(sections)}")
    for sec in sections:
        print(f"    {sec['start']:.0f}s–{sec['end']:.0f}s: {sec['type']}")


if __name__ == "__main__":
    main()
