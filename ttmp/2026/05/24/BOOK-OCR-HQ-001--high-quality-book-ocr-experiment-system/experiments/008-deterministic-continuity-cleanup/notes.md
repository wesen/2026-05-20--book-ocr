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

# Experiment 008 Notes: Deterministic continuity cleanup and QA

## Summary

Experiment 008 adds an auditable second pass over Experiment 007 instead of running another OCR prompt. The pass has two scripts:

- `scripts/03-qa-ocr-markdown.py`: page-aware QA report.
- `scripts/04-normalize-ocr-markdown.py`: narrow list-page dot-leader normalization.

The cleanup is intentionally deterministic and conservative. It does not call a model and does not rewrite prose. Its purpose is to make the current best OCR easier to review while preserving provenance and producing a diff.

## Inputs

- Source experiment: `007-quality-v4-mini-pages-001-030`
- Source output: `experiments/007-quality-v4-mini-pages-001-030/outputs/01-final-quality-v4-mini-pages-001-030.md`

## Outputs

- QA before cleanup: `outputs/01-qa-before-cleanup.md`
- Normalized markdown: `outputs/02-final-quality-v4-mini-pages-001-030-normalized.md`
- Cleanup diff: `outputs/03-cleanup-diff.patch`
- QA after cleanup: `outputs/04-qa-after-cleanup.md`

## QA result

Both pre-cleanup and post-cleanup automated checks pass:

- Page markers found: 30
- Expected page markers: 30
- Figure markers: 2
- Known bad term checks: pass
- Expected string checks: pass
- Adjacent duplicate line checks: pass
- List pages 006-009 have:
  - markdown bullet lines: 0
  - markdown heading lines: 0

The expected strings checked include:

- `Presentation Based User Interfaces`
- `This blank page was inserted to preserve pagination.`
- `Figure 4-1: Dired Model`
- `Figure 4-9: Sample Steamer Schematic`
- `Figure 5-1: PSBase Support of PPS Components`
- `Chapter Two`
- `The Primitive Presentation System (PPS) Model`
- `2.1 PPSCalc`

## Cleanup behavior

The cleanup normalizes list-page lines with long or irregular dot leaders into a consistent readable form:

```text
Figure 4-1: Dired Model .................................................. 72
```

becomes:

```text
Figure 4-1: Dired Model ... 72
```

The same policy is applied to pages 006-009. Non-list pages are preserved except for page-boundary whitespace normalization.

## Decision

The normalized output is the best current review artifact when the goal is readable markdown with stable list-page style. The raw Experiment 007 output remains the best provenance artifact when the goal is to inspect the exact model output.

## Remaining limitations

- QA checks are heuristic and do not prove full OCR correctness.
- The cleanup normalizes dot leaders rather than reconstructing exact scan alignment.
- Figures still use concise `[FIGURE: ...]` markers rather than structured diagram extraction.
- Prose continuity and line wrapping still need a future semantic cleanup pass if publication-grade markdown is required.
