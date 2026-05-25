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

# OCR Markdown QA Report

Input: `/tmp/ocr-quality-go-figures-v2/out/embedded-figures.md`

## Summary

- Page markers found: 30
- Expected page markers: 30
- Figure markers: 0

## Known bad term checks

- PASS: no known bad terms found.

## Expected string checks

- PASS: `2.1 PPSCalc`
- PASS: `Chapter Two`
- PASS: `Figure 4-1: Dired Model`
- PASS: `Figure 4-9: Sample Steamer Schematic`
- PASS: `Figure 5-1: PSBase Support of PPS Components`
- PASS: `Presentation Based User Interfaces`
- PASS: `The Primitive Presentation System (PPS) Model`
- PASS: `This blank page was inserted to preserve pagination.`

## Adjacent duplicate non-empty lines

- PASS: no adjacent duplicate non-empty lines found.

## List-page style checks

### Page 006: Table of Contents
- Markdown bullet lines: 0
- Markdown heading lines: 0

### Page 007: Table of Contents continuation
- Markdown bullet lines: 0
- Markdown heading lines: 0

### Page 008: Table of Figures
- Markdown bullet lines: 0
- Markdown heading lines: 0

### Page 009: Table of Figures continuation
- Markdown bullet lines: 0
- Markdown heading lines: 0

## Verdict

PASS for automated checks. Manual visual spot-checking is still required for OCR accuracy.
