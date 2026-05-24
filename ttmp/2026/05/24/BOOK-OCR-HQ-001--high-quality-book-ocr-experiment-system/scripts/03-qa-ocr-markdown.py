#!/usr/bin/env python3
"""Page-aware QA checks for OCR markdown outputs.

This script is intentionally conservative: it reports likely defects without
modifying the OCR output. It checks page-marker continuity, known bad OCR terms,
blank/image marker policy, duplicate adjacent lines, figure marker counts, and
selected book-specific expected strings.
"""
from __future__ import annotations

import argparse
import re
from dataclasses import dataclass
from pathlib import Path

PAGE_RE = re.compile(r"^<!-- page:(\d{3}) -->\s*$", re.MULTILINE)

KNOWN_BAD_TERMS = [
    "DiRed",
    "Streamer",
    "PPSBase",
    "Ciccarrelli",
    "[IMAGE:",
]

EXPECTED_STRINGS = [
    "Presentation Based User Interfaces",
    "This blank page was inserted to preserve pagination.",
    "Figure 4-1: Dired Model",
    "Figure 4-9: Sample Steamer Schematic",
    "Figure 5-1: PSBase Support of PPS Components",
    "Chapter Two",
    "The Primitive Presentation System (PPS) Model",
    "2.1 PPSCalc",
]


@dataclass
class Page:
    number: int
    text: str


def split_pages(text: str) -> list[Page]:
    matches = list(PAGE_RE.finditer(text))
    pages: list[Page] = []
    for i, match in enumerate(matches):
        start = match.end()
        end = matches[i + 1].start() if i + 1 < len(matches) else len(text)
        pages.append(Page(int(match.group(1)), text[start:end].strip("\n")))
    return pages


def line_no(text: str, needle: str) -> int | None:
    idx = text.find(needle)
    if idx < 0:
        return None
    return text[:idx].count("\n") + 1


def adjacent_duplicates(page: Page) -> list[tuple[int, str]]:
    rows = [row.strip() for row in page.text.splitlines()]
    hits: list[tuple[int, str]] = []
    prev = None
    for idx, row in enumerate(rows, start=1):
        if not row:
            prev = None
            continue
        if row == prev:
            hits.append((idx, row))
        prev = row
    return hits


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("input", type=Path)
    parser.add_argument("--out", type=Path)
    parser.add_argument("--expected-pages", type=int, default=30)
    args = parser.parse_args()

    text = args.input.read_text()
    pages = split_pages(text)
    findings: list[str] = []

    findings.append("---")
    findings.append("docType: reference")
    findings.append("status: active")
    findings.append("intent: short-term")
    findings.append("topics:")
    findings.append("  - ocr")
    findings.append("  - experiments")
    findings.append("created: 2026-05-24")
    findings.append("updated: 2026-05-24")
    findings.append("---")
    findings.append("")
    findings.append(f"# OCR Markdown QA Report")
    findings.append("")
    findings.append(f"Input: `{args.input}`")
    findings.append("")
    findings.append("## Summary")
    findings.append("")
    findings.append(f"- Page markers found: {len(pages)}")
    findings.append(f"- Expected page markers: {args.expected_pages}")
    figure_marker_count = len(re.findall(r"^\[FIGURE:", text, re.MULTILINE))
    findings.append(f"- Figure markers: {figure_marker_count}")
    findings.append("")

    page_numbers = [p.number for p in pages]
    expected = list(range(1, args.expected_pages + 1))
    if page_numbers != expected:
        findings.append("## Page continuity issues")
        findings.append("")
        findings.append(f"Observed: `{page_numbers}`")
        findings.append(f"Expected: `{expected}`")
        findings.append("")

    bad_hits: list[str] = []
    for term in KNOWN_BAD_TERMS:
        if term in text:
            bad_hits.append(f"- `{term}` at line {line_no(text, term)}")
    findings.append("## Known bad term checks")
    findings.append("")
    if bad_hits:
        findings.extend(bad_hits)
    else:
        findings.append("- PASS: no known bad terms found.")
    findings.append("")

    findings.append("## Expected string checks")
    findings.append("")
    missing = []
    for s in EXPECTED_STRINGS:
        if s in text:
            findings.append(f"- PASS: `{s}`")
        else:
            missing.append(s)
            findings.append(f"- MISSING: `{s}`")
    findings.append("")

    dup_hits: list[str] = []
    for page in pages:
        for row_no, row in adjacent_duplicates(page):
            dup_hits.append(f"- page {page.number:03d}, local line {row_no}: `{row}`")
    findings.append("## Adjacent duplicate non-empty lines")
    findings.append("")
    if dup_hits:
        findings.extend(dup_hits)
    else:
        findings.append("- PASS: no adjacent duplicate non-empty lines found.")
    findings.append("")

    list_pages = {6: "Table of Contents", 7: "Table of Contents continuation", 8: "Table of Figures", 9: "Table of Figures continuation"}
    findings.append("## List-page style checks")
    findings.append("")
    for page in pages:
        if page.number not in list_pages:
            continue
        bullet_lines = [line for line in page.text.splitlines() if re.match(r"^\s*[-*]\s+", line)]
        heading_lines = [line for line in page.text.splitlines() if re.match(r"^\s*#{1,6}\s+", line)]
        findings.append(f"### Page {page.number:03d}: {list_pages[page.number]}")
        findings.append(f"- Markdown bullet lines: {len(bullet_lines)}")
        findings.append(f"- Markdown heading lines: {len(heading_lines)}")
        if bullet_lines:
            findings.append("- Bullet samples: " + "; ".join(f"`{x}`" for x in bullet_lines[:3]))
        if heading_lines:
            findings.append("- Heading samples: " + "; ".join(f"`{x}`" for x in heading_lines[:3]))
        findings.append("")

    findings.append("## Verdict")
    findings.append("")
    if len(pages) == args.expected_pages and not bad_hits and not missing and not dup_hits:
        findings.append("PASS for automated checks. Manual visual spot-checking is still required for OCR accuracy.")
    else:
        findings.append("REVIEW REQUIRED: one or more automated checks produced findings.")

    report = "\n".join(findings) + "\n"
    if args.out:
        args.out.parent.mkdir(parents=True, exist_ok=True)
        args.out.write_text(report)
    else:
        print(report)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
