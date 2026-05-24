---
docType: reference
status: active
intent: short-term
topics:
  - ocr
  - experiments
created: 2026-05-24
updated: 2026-05-24
---

# Experiment 007 Notes: v4 lexicon prompt, gpt-5-mini-low, pages 1-30

## Summary

Experiment 007 is the best current full-range OCR output for the first 30 pages of `presentation-based-uis`. It uses:

- profile: `gpt-5-mini-low`
- prompt: `ocr-quality-v4-report794-lexicon`
- page range: 1-30
- stdout capture: `scripts/02-run-ocr-capture-log.py`
- work dir: `/tmp/book-ocr-hq-001/007-quality-v4-mini-pages-001-030`
- run id: `ocr-mvp-4c5c9406-926a-4ecd-a6b2-e8fedba847d8`

The output is substantially better than the baseline and the earlier v2/v3 runs on the known front-matter/list-page problems. It also passed a spot-check against page images 13, 15, and 30 with the vision tool: the selected pages' major headings, captions, and diagram labels are visually consistent with the OCR output.

## Why v4 + mini was selected

The progression was:

1. Baseline `ocr-mvp-universal-v1` with `gpt-5-nano-low` produced complete output, but list pages drifted across pages and blank/title pages were not handled consistently.
2. `ocr-quality-v2` fixed several front-matter issues, but duplicated a Table of Contents chapter line and still had continuation-list drift.
3. `ocr-quality-v3-list-diplomatic` improved title-page normalization and Table of Contents structure, but `gpt-5-nano-low` still made list-page accuracy mistakes such as `Dired` page-number errors and `Steamer`/`PSBase` spelling errors.
4. `gpt-5-mini-low` with v3 improved visual accuracy on list pages, but one full front-matter run still produced `DiRed`.
5. `ocr-quality-v4-report794-lexicon` added a small Report 794 vocabulary and explicit visible-intentionally-blank-page behavior. With `gpt-5-mini-low`, pages 1-30 are now the best current output.

## Positive checks

- 30 page markers are present.
- No grep hits for the known front-matter regressions:
  - `DiRed`
  - `Streamer`
  - `PPSBase`
  - `Ciccarrelli`
  - `[BLANK PAGE]` on the visible intentionally blank page
  - `[IMAGE:` title-page markers
- Page 1 title is normalized as readable text: `Presentation Based User Interfaces`.
- Page 2 correctly transcribes the visible sentence: `This blank page was inserted to preserve pagination.`
- Pages 6-9 are plain-text list pages, not markdown bullet lists.
- Page 8 includes the visually validated entries:
  - `Figure 4-1: Dired Model ... 72`
  - `Figure 4-9: Sample Steamer Schematic ... 91`
  - `Figure 5-1: PSBase Support of PPS Components ... 101`
- Page 13 and page 15 diagrams have captions and `[FIGURE: ...]` markers.
- Page 30 starts Chapter Two and section `2.1 PPSCalc`.

## Remaining imperfections

- Dot-leader lengths are approximate. They are visually useful but not exact reproductions of the scan.
- Page 6/7 Table of Contents dot leaders are inconsistent in density because the model approximates alignment.
- Prose pages sometimes preserve source line wrapping and sometimes join lines; the output is readable but not fully normalized.
- Diagram transcription uses concise figure markers rather than a full structured diagram serialization.
- The current workflow is still single-page OCR; it does not yet use neighbor-page context for continuity or terminology reinforcement beyond the prompt lexicon.

## Quality decision

Experiment 007 is good enough to serve as the current high-quality candidate for pages 1-30. The next meaningful quality step should not be another prompt tweak; it should add a second pass or context window:

- page-level OCR first pass;
- structured continuity/cleanup pass for front matter, headings, figures, and list pages;
- optional targeted re-OCR for pages that fail consistency checks.

## Validation commands

```bash
OUT=experiments/007-quality-v4-mini-pages-001-030/outputs/01-final-quality-v4-mini-pages-001-030.md
grep -c '^<!-- page:' "$OUT"
grep -nE 'DiRed|Streamer|PPSBase|Ciccarrelli|\[BLANK PAGE\]|\[IMAGE:' "$OUT" || true
grep -c '^\[FIGURE:' "$OUT"
```

Observed:

```text
page markers: 30
known-term regression hits: 0
figure markers: 2
```
