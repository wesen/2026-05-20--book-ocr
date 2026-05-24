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

# Experiment 002 Notes: Quality v2 targeted prompt

## Summary

Experiment 002 introduced `ocr-quality-v2`, a page-type-aware prompt that explicitly handles title pages, blank pages, table-of-contents/list pages, figures, tables, footers, and consistency. It targeted pages 1-9 first because those pages expose front matter, table of contents continuation, and table of figures continuation problems.

## Run ID

- Successful run: `ocr-mvp-630348c5-4460-4fe1-b549-bc8d3781cb0f`

## Code commit

- `65b56d50bbb95778efd436dab08615686731f6b9fdbf99b40` — prompt implementation
- `c53cd92176dabeb366afdd0de56d1b9fdbf99b40` — CLI `--log-level` support

## Baseline problems identified

Manual review plus vision-tool validation found these baseline issues:

- Pages 6-7 and 8-9 changed formatting across continuation pages.
- The baseline used bullets for the table of contents and table of figures, then switched to plain lines on continuation pages.
- Page 7 inserted a markdown `## Chapter Six` heading in the middle of an ongoing table-of-contents list.
- Page 9 lost the `Table of Figures` continuation context and had weaker list styling.
- The blank-page policy emitted prose: `This blank page was inserted to preserve pagination.`
- Figure/table list pages need stable line style with labels, titles, dot leaders or spacing, and page numbers.

## Quality v2 improvements observed

- Page 2 now outputs exactly `[BLANK PAGE]`.
- Page 8 no longer uses markdown bullets for the Table of Figures.
- Page 8 preserves a consistent `Figure N-M: title page` style.
- Page 9 preserves the same non-bullet figure-entry style as page 8.
- Page 4 fixed the baseline misspelling `Ciccarrelli` to `Ciccarelli`.
- The title page no longer becomes an image marker.

## Quality v2 regressions / remaining problems

- Page 1 line-break preservation became too literal: `Presentation / Based User / Interfaces` is split across lines. For final markdown, this should probably be normalized to `Presentation Based User Interfaces` unless preserving title-page visual line breaks is desired.
- Page 6 changed the first Table of Contents page number for section 1.1 from baseline `8` to `9`; this needs visual/manual verification against the scan.
- Page 7 duplicated the chapter line:
  - `Chapter Six: Constructing Presentation Systems`
  - `Chapter Six: Constructing Presentation Systems 142`
- Page 7 still does not fully match page 6 styling; chapter headings are plain text instead of `##` headings, which may be good for ToC fidelity but is inconsistent inside the same ToC output.
- Page 9 uses figure-entry lines but lacks a repeated `Table of Figures (continued)` heading. This may be acceptable if final assembly keeps page markers, but it should be decided explicitly.
- Page 8/9 spacing approximates right alignment but is not robust markdown. Dot leaders may be more inspectable.

## Decision

`ocr-quality-v2` is an improvement over the baseline for blank pages and figure-list formatting, but it is not yet stellar. The next prompt should be more explicit about list pages:

- Use plain text list-page transcription, not markdown section headings, for all Table of Contents and Table of Figures pages.
- Preserve dot leaders when visible.
- Preserve exactly one chapter line per chapter in Table of Contents pages.
- For continuation pages, either add an explicit `[CONTINUED: Table of Contents]` / `[CONTINUED: Table of Figures]` marker or keep no heading; choose one rule and apply it consistently.
- For title pages, normalize visible title lines only if the target output is reading markdown rather than diplomatic visual transcription.

## Vision validation

The vision tool compared pages 6-9 against the baseline and confirmed the main list-page failures: bullet drift, continuation-style drift, inconsistent heading treatment, page-number/footer handling, and figure-entry punctuation/style preservation.

## Next experiment

Experiment 003 should target pages 1-9 again with `ocr-quality-v3-list-diplomatic`, then only expand to pages 1-30 if it improves the front matter without regressions.
