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

# Experiment 001: Go quality pass with embedded figures

This experiment runs the new Go `ocr-quality` workflow workers against the raw `BOOK-OCR-HQ-001` Experiment 007 OCR output.

## Command

```bash
go run ./cmd/ocr-mvp quality-pass \
  --markdown /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/BOOK-OCR-HQ-001--high-quality-book-ocr-experiment-system/experiments/007-quality-v4-mini-pages-001-030/outputs/01-final-quality-v4-mini-pages-001-030.md \
  --output-dir /tmp/ocr-quality-go-figures/out \
  --work-dir /tmp/ocr-quality-go-figures/work \
  --book-id presentation-based-uis-hq-007-v4-mini-30 \
  --expected-pages 30 \
  --image-dir /home/manuel/code/wesen/claw-stuff/output/books/presentation-based-uis/pages \
  --embed-figures
```

## Outputs

- `outputs/01-normalized.md`: deterministic dot-leader normalized markdown.
- `outputs/02-embedded-figures.md`: normalized markdown with figure markers replaced by embedded image links.
- `outputs/03-cleanup.diff`: diff from raw OCR to normalized output.
- `outputs/figures/page_013_figure_01.png`: extracted Figure 1-1 image.
- `outputs/figures/page_021_figure_01.png`: extracted Figure 1-4 image.
- `outputs/04-qa-before.md`: Go QA report before cleanup.
- `outputs/05-qa-after.md`: Go QA report after embedded figure output.
- `outputs/06-quality-report.md`: workflow-level report artifact.

## Visual validation

The figure crop algorithm was iterated with vision feedback. The final crop pass removes page numbers/footers and preserves the full diagrams for both extracted figures. Remaining possible improvements are minor whitespace tightening and contrast enhancement for Figure 1-1.

## Decision

This is the current best review artifact with embedded images. The crop algorithm is an acceptable first pass for the two figure pages in the first 30 pages, but should become more structured before scaling to the full book.
