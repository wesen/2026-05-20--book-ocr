#!/usr/bin/env python3
"""Deterministic second-pass cleanup for OCR markdown.

The cleanup is deliberately narrow and auditable. It preserves page markers and
content, but normalizes Table of Contents / Table of Figures dot-leader rows to a
consistent readable `label ... page` style. It does not rewrite prose or call an
LLM; use the diff/report as an inspectable continuity pass.
"""
from __future__ import annotations

import argparse
import difflib
import re
from dataclasses import dataclass
from pathlib import Path

PAGE_RE = re.compile(r"^<!-- page:(\d{3}) -->\s*$", re.MULTILINE)
LEADER_RE = re.compile(r"^(?P<label>\S.*?)(?:\s*\.\s*){3,}\s*(?P<page>\d{1,3})\s*$")
SPACED_PAGE_RE = re.compile(r"^(?P<label>\S.*?\S)\s{2,}(?P<page>\d{1,3})\s*$")

LIST_PAGES = {6, 7, 8, 9}


@dataclass
class Segment:
    page: int | None
    header: str
    body: str


def split_segments(text: str) -> tuple[str, list[Segment]]:
    matches = list(PAGE_RE.finditer(text))
    preamble = text[: matches[0].start()] if matches else text
    segments: list[Segment] = []
    for i, match in enumerate(matches):
        end = matches[i + 1].start() if i + 1 < len(matches) else len(text)
        segments.append(Segment(int(match.group(1)), match.group(0), text[match.end() : end]))
    return preamble, segments


def normalize_list_line(line: str) -> str:
    stripped = line.rstrip()
    if not stripped:
        return stripped
    for pattern in (LEADER_RE, SPACED_PAGE_RE):
        match = pattern.match(stripped)
        if not match:
            continue
        label = re.sub(r"\s+", " ", match.group("label")).strip()
        page = match.group("page")
        # Do not normalize chapter/table headings without an actual label body.
        if len(label) < 3:
            return stripped
        return f"{label} ... {page}"
    return stripped


def normalize_page(page: int, body: str) -> str:
    body = body.strip("\n")
    if page not in LIST_PAGES:
        return body + "\n\n"
    lines = body.splitlines()
    out = [normalize_list_line(line) for line in lines]
    # Collapse only excessive blank runs introduced by page assembly, not content.
    collapsed: list[str] = []
    blank_count = 0
    for line in out:
        if line.strip() == "":
            blank_count += 1
            if blank_count <= 2:
                collapsed.append("")
        else:
            blank_count = 0
            collapsed.append(line)
    return "\n".join(collapsed).strip("\n") + "\n\n"


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("input", type=Path)
    parser.add_argument("output", type=Path)
    parser.add_argument("--diff", type=Path)
    args = parser.parse_args()

    old = args.input.read_text()
    preamble, segments = split_segments(old)
    parts = [preamble.rstrip(), "\n"]
    for seg in segments:
        assert seg.page is not None
        parts.append(f"{seg.header}\n\n")
        parts.append(normalize_page(seg.page, seg.body))
    new = "".join(parts)
    if not new.endswith("\n"):
        new += "\n"

    args.output.parent.mkdir(parents=True, exist_ok=True)
    args.output.write_text(new)

    if args.diff:
        diff = difflib.unified_diff(
            old.splitlines(keepends=True),
            new.splitlines(keepends=True),
            fromfile=str(args.input),
            tofile=str(args.output),
        )
        args.diff.parent.mkdir(parents=True, exist_ok=True)
        args.diff.write_text("".join(diff))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
