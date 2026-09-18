#!/usr/bin/env python3
"""Generate the AI Interview Pro / 智聘 AI app icon."""

from __future__ import annotations

import math
from pathlib import Path

from PIL import Image, ImageDraw, ImageFilter

ROOT = Path(__file__).resolve().parent
MASTER = 2048  # supersample, then downsample


def lerp(a: float, b: float, t: float) -> float:
    return a + (b - a) * t


def mix(c1, c2, t: float):
    return tuple(int(lerp(c1[i], c2[i], t)) for i in range(len(c1)))


def rounded_mask(size: int, radius: int) -> Image.Image:
    mask = Image.new("L", (size, size), 0)
    ImageDraw.Draw(mask).rounded_rectangle(
        (0, 0, size - 1, size - 1), radius=radius, fill=255
    )
    return mask


def vertical_gradient(size: int, stops: list[tuple[float, tuple[int, int, int]]]) -> Image.Image:
    img = Image.new("RGB", (size, size))
    px = img.load()
    stops = sorted(stops, key=lambda s: s[0])
    for y in range(size):
        t = y / max(size - 1, 1)
        for i in range(len(stops) - 1):
            t0, c0 = stops[i]
            t1, c1 = stops[i + 1]
            if t0 <= t <= t1 or i == len(stops) - 2:
                local = 0 if t1 == t0 else (t - t0) / (t1 - t0)
                local = max(0.0, min(1.0, local))
                color = mix(c0, c1, local)
                break
        for x in range(size):
            # slight diagonal bias so it isn't a flat wash
            diag = (x / max(size - 1, 1) - 0.5) * 0.08
            shade = mix(color, (255, 255, 255) if diag > 0 else (15, 23, 42), abs(diag))
            px[x, y] = shade
    return img


def radial_glow(canvas: Image.Image, cx: float, cy: float, radius: float, color, strength: float) -> None:
    overlay = Image.new("RGBA", canvas.size, (0, 0, 0, 0))
    draw = ImageDraw.Draw(overlay)
    steps = 96
    for i in range(steps, 0, -1):
        t = i / steps
        alpha = int(255 * strength * (1.0 - t) ** 1.7)
        r = radius * t
        draw.ellipse((cx - r, cy - r, cx + r, cy + r), fill=(*color[:3], alpha))
    canvas.alpha_composite(overlay)


def star(cx: float, cy: float, outer: float, inner: float, points: int = 4, rotation: float = 0.0):
    coords = []
    for i in range(points * 2):
        r = outer if i % 2 == 0 else inner
        a = rotation + -math.pi / 2 + i * math.pi / points
        coords.append((cx + r * math.cos(a), cy + r * math.sin(a)))
    return coords


def speech_bubble(draw: ImageDraw.ImageDraw, body, radius: int, fill) -> None:
    x0, y0, x1, y1 = body
    draw.rounded_rectangle(body, radius=radius, fill=fill)
    w = x1 - x0
    h = y1 - y0
    # Wide chat-bubble tail (no pin-dot).
    attach_l = (x0 + w * 0.16, y1 - h * 0.02)
    attach_r = (x0 + w * 0.42, y1 - h * 0.08)
    tip = (x0 + w * 0.10, y1 + h * 0.30)
    mid = (x0 + w * 0.22, y1 + h * 0.08)
    draw.polygon([attach_l, attach_r, tip, mid], fill=fill)


def render_master() -> Image.Image:
    s = MASTER
    canvas = Image.new("RGBA", (s, s), (0, 0, 0, 0))

    indigo = (79, 70, 229)
    indigo_deep = (67, 56, 202)
    teal = (13, 148, 136)
    teal_bright = (20, 184, 166)
    cyan = (8, 145, 178)

    plate = vertical_gradient(
        s,
        [
            (0.0, (67, 56, 202)),
            (0.42, (79, 70, 229)),
            (1.0, (15, 118, 110)),
        ],
    ).convert("RGBA")
    radial_glow(plate, s * 0.28, s * 0.22, s * 0.55, (165, 180, 252), 0.50)
    radial_glow(plate, s * 0.78, s * 0.82, s * 0.50, (45, 212, 191), 0.40)
    radial_glow(plate, s * 0.85, s * 0.18, s * 0.28, (199, 210, 254), 0.22)
    plate.putalpha(rounded_mask(s, int(s * 0.223)))
    canvas.alpha_composite(plate)

    # Inner glass rim
    rim = Image.new("RGBA", (s, s), (0, 0, 0, 0))
    rim_draw = ImageDraw.Draw(rim)
    inset = int(s * 0.027)
    rim_draw.rounded_rectangle(
        (inset, inset, s - inset - 1, s - inset - 1),
        radius=int(s * 0.201),
        outline=(255, 255, 255, 46),
        width=max(4, s // 170),
    )
    canvas.alpha_composite(rim)

    # --- centered conversation mark ---
    body = (int(s * 0.20), int(s * 0.23), int(s * 0.80), int(s * 0.61))
    radius = int(s * 0.11)

    # Soft drop shadow
    shadow = Image.new("RGBA", (s, s), (0, 0, 0, 0))
    shadow_draw = ImageDraw.Draw(shadow)
    offset_body = tuple(v + 18 for v in body)
    speech_bubble(shadow_draw, offset_body, radius, (15, 23, 42, 90))
    canvas.alpha_composite(shadow.filter(ImageFilter.GaussianBlur(s // 40)))

    bubble = Image.new("RGBA", (s, s), (0, 0, 0, 0))
    bdraw = ImageDraw.Draw(bubble)
    speech_bubble(bdraw, body, radius, (255, 255, 255, 244))

    # Inner top highlight on the bubble
    hx0, hy0, hx1, hy1 = body
    highlight = Image.new("RGBA", (s, s), (0, 0, 0, 0))
    hdraw = ImageDraw.Draw(highlight)
    hdraw.rounded_rectangle(
        (hx0 + 18, hy0 + 14, hx1 - 18, hy0 + int((hy1 - hy0) * 0.42)),
        radius=int(radius * 0.7),
        fill=(255, 255, 255, 40),
    )
    bubble.alpha_composite(highlight)

    # Voice / growth / AI equalizer bars
    bar_w = int(s * 0.055)
    gap = int(s * 0.042)
    heights = [int(s * 0.11), int(s * 0.175), int(s * 0.135)]
    colors = [indigo, teal_bright, indigo_deep]
    total_w = 3 * bar_w + 2 * gap
    start_x = (hx0 + hx1 - total_w) // 2
    base_y = hy0 + int((hy1 - hy0) * 0.78)
    for i, (h, color) in enumerate(zip(heights, colors)):
        x = start_x + i * (bar_w + gap)
        bdraw.rounded_rectangle(
            (x, base_y - h, x + bar_w, base_y),
            radius=bar_w // 2,
            fill=(*color, 240),
        )

    canvas.alpha_composite(bubble)

    # AI spark badge, docked on the top-right of the bubble
    bx = int(s * 0.74)
    by = int(s * 0.30)
    badge_r = int(s * 0.092)
    badge = Image.new("RGBA", (s, s), (0, 0, 0, 0))
    d = ImageDraw.Draw(badge)
    # shadow
    d.ellipse(
        (bx - badge_r + 8, by - badge_r + 14, bx + badge_r + 8, by + badge_r + 14),
        fill=(15, 23, 42, 50),
    )
    badge = badge.filter(ImageFilter.GaussianBlur(s // 90))
    d = ImageDraw.Draw(badge)
    d.ellipse((bx - badge_r, by - badge_r, bx + badge_r, by + badge_r), fill=(255, 255, 255, 255))
    inner_r = int(badge_r * 0.78)
    d.ellipse(
        (bx - inner_r, by - inner_r, bx + inner_r, by + inner_r),
        fill=(*teal, 255),
    )
    d.polygon(star(bx, by, badge_r * 0.48, badge_r * 0.20, 4), fill=(255, 255, 255, 255))
    d.ellipse(
        (bx - badge_r * 0.12, by - badge_r * 0.12, bx + badge_r * 0.12, by + badge_r * 0.12),
        fill=(*cyan, 255),
    )
    canvas.alpha_composite(badge)

    # Specular arc (iOS-like glass)
    spec = Image.new("RGBA", (s, s), (0, 0, 0, 0))
    sdraw = ImageDraw.Draw(spec)
    sdraw.arc(
        (int(s * 0.08), int(s * 0.05), int(s * 0.62), int(s * 0.42)),
        start=200,
        end=312,
        fill=(255, 255, 255, 52),
        width=max(10, s // 110),
    )
    canvas.alpha_composite(spec.filter(ImageFilter.GaussianBlur(2)))

    return canvas


def save_png(master: Image.Image, path: Path, size: int) -> None:
    out = master.resize((size, size), Image.Resampling.LANCZOS)
    out.save(path, "PNG", optimize=True)
    print(f"wrote {path} ({size}x{size})")


def write_svg(path: Path) -> None:
    svg = """<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1024 1024" role="img" aria-label="智聘 AI / AI Interview Pro">
  <defs>
    <linearGradient id="bg" x1="160" y1="0" x2="880" y2="1024" gradientUnits="userSpaceOnUse">
      <stop offset="0%" stop-color="#4338CA"/>
      <stop offset="42%" stop-color="#4F46E5"/>
      <stop offset="100%" stop-color="#0F766E"/>
    </linearGradient>
    <radialGradient id="glowA" cx="28%" cy="22%" r="55%">
      <stop offset="0%" stop-color="#A5B4FC" stop-opacity=".5"/>
      <stop offset="100%" stop-color="#A5B4FC" stop-opacity="0"/>
    </radialGradient>
    <radialGradient id="glowB" cx="78%" cy="82%" r="50%">
      <stop offset="0%" stop-color="#2DD4BF" stop-opacity=".4"/>
      <stop offset="100%" stop-color="#2DD4BF" stop-opacity="0"/>
    </radialGradient>
    <filter id="shadow" x="-25%" y="-10%" width="150%" height="150%">
      <feGaussianBlur in="SourceAlpha" stdDeviation="16"/>
      <feOffset dy="16" result="off"/>
      <feColorMatrix in="off" values="0 0 0 0 0.06  0 0 0 0 0.09  0 0 0 0 0.16  0 0 0 0.28 0"/>
    </filter>
  </defs>

  <rect width="1024" height="1024" rx="228" fill="url(#bg)"/>
  <rect width="1024" height="1024" rx="228" fill="url(#glowA)"/>
  <rect width="1024" height="1024" rx="228" fill="url(#glowB)"/>
  <rect x="28" y="28" width="968" height="968" rx="206" fill="none" stroke="#fff" stroke-opacity=".18" stroke-width="6"/>

  <g filter="url(#shadow)">
    <path fill="#F8FAFF" d="
      M 205 236
      h 614
      a 112 112 0 0 1 112 112
      v 164
      a 112 112 0 0 1 -112 112
      H 430
      L 268 738
      L 310 624
      H 205
      a 112 112 0 0 1 -112 -112
      v -164
      a 112 112 0 0 1 112 -112
      z"/>
  </g>

  <rect x="400" y="392" width="56" height="112" rx="28" fill="#4F46E5"/>
  <rect x="484" y="326" width="56" height="178" rx="28" fill="#14B8A6"/>
  <rect x="568" y="364" width="56" height="140" rx="28" fill="#4338CA"/>

  <circle cx="758" cy="308" r="94" fill="#fff"/>
  <circle cx="758" cy="308" r="72" fill="#0D9488"/>
  <path fill="#fff" d="M758 248 L772 292 L816 308 L772 324 L758 368 L744 324 L700 308 L744 292 Z"/>
  <circle cx="758" cy="308" r="11" fill="#0891B2"/>
</svg>
"""
    path.write_text(svg, encoding="utf-8")
    print(f"wrote {path}")


def main() -> None:
    ROOT.mkdir(parents=True, exist_ok=True)
    master = render_master()
    save_png(master, ROOT / "logo-1024.png", 1024)
    save_png(master, ROOT / "logo.png", 512)
    save_png(master, ROOT / "logo-256.png", 256)
    write_svg(ROOT / "logo.svg")


if __name__ == "__main__":
    main()
