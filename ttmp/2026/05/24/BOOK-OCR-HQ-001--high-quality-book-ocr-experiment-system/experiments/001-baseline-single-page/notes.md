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

# Experiment 001 Notes: Baseline single-page OCR

## Summary

The first baseline attempt using the default Pinocchio registry failed because the local profile registry has a duplicate `gpt-5-nano-low` key. I preserved that failed log and loaded it into SQLite.

The second baseline run used a clean temporary profile registry copied from the local Pinocchio config and succeeded on pages 1-30 with `gpt-5-nano-low`.

## Run IDs

- Failed duplicate-profile run: `ocr-mvp-3a775e1e-e54f-4fa0-8277-175c92e07972`
- Successful clean-registry run: `ocr-mvp-593bf5b6-19c6-4c8c-b631-b48a2d1aba78`

## Key artifacts

- Final markdown: `outputs/01-final-baseline-clean.md`
- Page projection rows: `outputs/pages-clean.tsv`
- Compact timeline: `outputs/timeline-clean.tsv`
- Full noisy log: `logs/run-clean-registry.log`
- SQLite log DB: `logs/run-clean-registry.sqlite`
- Filtered summary: `logs/01-run-clean-registry-summary.md`

## Initial observations

- The successful run produced 30 page artifacts and one assembled markdown artifact.
- The raw provider log had 8687 lines, of which 8443 were trace-level SSE deltas.
- The filtered SQLite summary reduces this to 69 non-trace workflow events.
- Page 007 retried once and then succeeded; there were no final warning/error rows in the clean-registry run.

## Next review task

Review `outputs/01-final-baseline-clean.md` page by page and classify failures before changing prompts.
