#!/usr/bin/env python3
"""Regenerate traced SVG Inkforge mark / Vue component from the source PNG.

Requires Python 3 with OpenCV bindings (for example ``pip install opencv-python-headless``).
"""

from __future__ import annotations

import argparse
import pathlib

import cv2  # type: ignore[import-untyped]


def preprocess(gray):
    _, bw = cv2.threshold(gray, 250, 255, cv2.THRESH_BINARY_INV)
    kern = cv2.getStructuringElement(cv2.MORPH_ELLIPSE, (9, 9))
    bw = cv2.morphologyEx(bw, cv2.MORPH_CLOSE, kern, iterations=2)
    bw = cv2.morphologyEx(bw, cv2.MORPH_OPEN, kern, iterations=1)
    return bw


def contour_to_subpath(cnt, ox: int, oy: int, digits: int = 4) -> str:
    pts = cnt.squeeze()
    xs = pts[:, 0].astype(float) - ox
    ys = pts[:, 1].astype(float) - oy

    def fmt_val(v: float) -> str:
        return format(float(v), f".{digits}g")

    parts = [f"M{fmt_val(xs[0])},{fmt_val(ys[0])}"]
    for px, py in zip(xs[1:], ys[1:]):
        parts.append(f"L{fmt_val(px)},{fmt_val(py)}")
    parts.append("Z")
    return " ".join(parts)


def main() -> None:
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument(
        "--png",
        type=pathlib.Path,
        default=None,
        help="Raster source (default: web/console/public/branding-inkforge-source.png)",
    )
    ap.add_argument(
        "--repo-root",
        type=pathlib.Path,
        default=pathlib.Path(__file__).resolve().parents[1],
        help="Repository root containing web/console/",
    )
    args = ap.parse_args()
    repo: pathlib.Path = args.repo_root

    png = (
        args.png
        if args.png is not None
        else repo / "web" / "console" / "public" / "branding-inkforge-source.png"
    )
    out_logo = repo / "web" / "console" / "public" / "logo.svg"
    out_cmp = repo / "web" / "console" / "src" / "components" / "InkforgeLogoMark.vue"

    img = cv2.imread(str(png), cv2.IMREAD_GRAYSCALE)
    if img is None:
        raise SystemExit(f"could not read {png}")

    bw = preprocess(img)
    cnts, hier = cv2.findContours(bw, cv2.RETR_CCOMP, cv2.CHAIN_APPROX_SIMPLE)
    h0 = hier[0]
    roots = [i for i in range(len(cnts)) if h0[i][3] == -1]
    root_i = max(roots, key=lambda i: cv2.contourArea(cnts[i]))
    hole_idx = sorted(
        [j for j in range(len(cnts)) if h0[j][3] == root_i],
        key=lambda j: cv2.contourArea(cnts[j]),
        reverse=True,
    )

    ys, xs = bw.nonzero()
    pad = 8
    ox = int(xs.min()) - pad
    oy = int(ys.min()) - pad
    maxx = int(xs.max()) + pad
    maxy = int(ys.max()) + pad
    vb_w = maxx - ox + 1
    vb_h = maxy - oy + 1

    peri_outer = cv2.arcLength(cnts[root_i], True)
    outer_poly = cv2.approxPolyDP(cnts[root_i], 0.00135 * peri_outer, True)

    hole_polys = []
    for hi in hole_idx:
        peri = cv2.arcLength(cnts[hi], True)
        hole_polys.append(cv2.approxPolyDP(cnts[hi], 0.0065 * peri, True))

    combined_parts = [contour_to_subpath(outer_poly, ox, oy)]
    for hp in hole_polys:
        combined_parts.append(contour_to_subpath(hp, ox, oy))
    combined = " ".join(combined_parts)

    vb_w_r = round(vb_w, 4)
    vb_h_r = round(vb_h, 4)

    out_logo.write_text(
        f'<svg xmlns="http://www.w3.org/2000/svg" '
        f'viewBox="0 0 {vb_w_r} {vb_h_r}" aria-hidden="true">\n'
        '  <!-- Regenerated via scripts/trace_inkforge_logo.py -->\n'
        f'  <path fill="#001f4d" fill-rule="evenodd" d="{combined}" />\n'
        "</svg>\n",
        encoding="utf-8",
    )

    out_cmp.write_text(
        f"<template>\n"
        '  <svg\n'
        '    class="inkforge-logo-mark-svg"\n'
        '    xmlns="http://www.w3.org/2000/svg"\n'
        f'    viewBox="0 0 {vb_w_r} {vb_h_r}"\n'
        '    fill="currentColor"\n'
        '    aria-hidden="true"\n'
        '  >\n'
        f'    <path fill-rule="evenodd" d="{combined}" />\n'
        "  </svg>\n"
        "</template>\n\n"
        "<style scoped>\n"
        ".inkforge-logo-mark-svg {\n"
        "  display: block;\n"
        "  width: 100%;\n"
        "  height: 100%;\n"
        "}\n"
        "</style>\n",
        encoding="utf-8",
    )
    print("wrote", out_logo.relative_to(repo), "&", out_cmp.relative_to(repo))


if __name__ == "__main__":
    main()
