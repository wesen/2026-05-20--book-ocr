---
docType: reference
status: active
intent: short-term
topics:
  - ocr
  - experiments
  - book-processing
created: 2026-05-24
updated: 2026-05-24
---

# Experiment 002: Figure-aware marker recovery

This experiment fixes the missing graphical-page signal discovered after Experiment 001. Figure 1-2 and Figure 1-3 are full-page diagrams, but the previous OCR output had transcribed them as plain diagram text without `[FIGURE: ...]` markers. Because the embedding worker only extracted pages with figure markers, those two figures were not embedded.

## Change

Two changes were made:

1. Add prompt version `ocr-quality-v5-figure-aware`, which explicitly instructs OCR to emit `[FIGURE: ...]` markers for full-page diagrams, flowcharts, architecture diagrams, and other graphical pages even when diagram labels are also transcribed.
2. Add a post-processing fallback in the figure worker that detects caption-only diagram pages such as `Figure 1-2: The Representation Shift Model`, synthesizes a missing figure marker, and extracts the source-page crop without requiring a model rerun.

## Command

```bash
go run ./cmd/ocr-mvp quality-pass \
  --markdown /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/BOOK-OCR-HQ-001--high-quality-book-ocr-experiment-system/experiments/007-quality-v4-mini-pages-001-030/outputs/01-final-quality-v4-mini-pages-001-030.md \
  --output-dir /tmp/ocr-quality-go-figures-v2/out \
  --work-dir /tmp/ocr-quality-go-figures-v2/work \
  --book-id presentation-based-uis-hq-007-v4-mini-30 \
  --expected-pages 30 \
  --image-dir /home/manuel/code/wesen/claw-stuff/output/books/presentation-based-uis/pages \
  --embed-figures
```

## Result

The embedded markdown now links four figures:

- `figures/page_013_figure_01.png` — Figure 1-1
- `figures/page_015_figure_01.png` — Figure 1-2
- `figures/page_017_figure_01.png` — Figure 1-3
- `figures/page_021_figure_01.png` — Figure 1-4

The new Markdown links are visible in `outputs/02-embedded-figures.md` around the figure pages.

## Visual validation

The vision check for the newly recovered Figure 1-2 and Figure 1-3 crops reported that both crops include the title and complete diagram/labels, and do not include page numbers or footer elements. Figure 1-3 has noticeable background speckling; a future enhancement pass should add optional denoising/contrast improvement while preserving raw crops.

## Remaining work

- Add structured figure QA counts to the quality report.
- Add sidecar crop metadata and debug overlays.
- Add optional denoise/contrast-enhanced image output.
- Run the `ocr-quality-v5-figure-aware` prompt on targeted live pages to confirm future OCR emits markers directly rather than relying on fallback recovery.
