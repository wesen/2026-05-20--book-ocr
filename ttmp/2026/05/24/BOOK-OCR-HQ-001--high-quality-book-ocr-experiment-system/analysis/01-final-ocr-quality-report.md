---
Title: Final OCR Quality Report
Ticket: BOOK-OCR-HQ-001
Status: active
Topics:
    - ocr
    - workflow
    - experiments
    - book-processing
    - implementation-guide
DocType: analysis
Intent: long-term
Owners: []
RelatedFiles:
    - Path: 2026-05-20--book-ocr/ttmp/2026/05/24/BOOK-OCR-HQ-001--high-quality-book-ocr-experiment-system/experiments/007-quality-v4-mini-pages-001-030/outputs/01-final-quality-v4-mini-pages-001-030.md
      Note: Raw best OCR model output used as provenance
    - Path: 2026-05-20--book-ocr/ttmp/2026/05/24/BOOK-OCR-HQ-001--high-quality-book-ocr-experiment-system/experiments/008-deterministic-continuity-cleanup/outputs/02-final-quality-v4-mini-pages-001-030-normalized.md
      Note: Selected normalized review artifact
    - Path: 2026-05-20--book-ocr/ttmp/2026/05/24/BOOK-OCR-HQ-001--high-quality-book-ocr-experiment-system/experiments/008-deterministic-continuity-cleanup/outputs/03-cleanup-diff.patch
      Note: Diff from raw model output to normalized review output
    - Path: 2026-05-20--book-ocr/ttmp/2026/05/24/BOOK-OCR-HQ-001--high-quality-book-ocr-experiment-system/experiments/008-deterministic-continuity-cleanup/outputs/04-qa-after-cleanup.md
      Note: Final automated QA report
ExternalSources: []
Summary: Final report for high-quality OCR experiments on the first 30 pages of Presentation Based User Interfaces.
LastUpdated: 2026-05-24T19:34:51.725983608-04:00
WhatFor: Use this to understand the final selected OCR output, the experiment path that led to it, validation evidence, and remaining limitations.
WhenToUse: Read this when reviewing BOOK-OCR-HQ-001, selecting an OCR artifact for downstream use, or planning the next OCR quality iteration.
---


# Final OCR Quality Report

## Executive Summary

`BOOK-OCR-HQ-001` produced a high-quality, inspectable OCR candidate for the first 30 pages of MIT Technical Report 794, *Presentation Based User Interfaces* by Eugene C. Ciccarelli IV.

The best raw model output is Experiment 007:

```text
experiments/007-quality-v4-mini-pages-001-030/outputs/01-final-quality-v4-mini-pages-001-030.md
```

The best review artifact is Experiment 008:

```text
experiments/008-deterministic-continuity-cleanup/outputs/02-final-quality-v4-mini-pages-001-030-normalized.md
```

Experiment 008 does not rerun OCR. It applies a deterministic, auditable cleanup to Experiment 007, normalizing list-page dot leaders and producing before/after QA reports plus a diff.

## Recommendation

Use the Experiment 008 normalized markdown for human review, downstream markdown reading, and report sharing.

Use the Experiment 007 raw markdown when exact model output provenance matters.

Do not discard either artifact. The raw output and the normalized output serve different review purposes.

## Selected Configuration

The selected OCR configuration is:

```yaml
profile: gpt-5-mini-low
prompt_version: ocr-quality-v4-report794-lexicon
page_range: 1-30
source_pages: /home/manuel/code/wesen/claw-stuff/output/books/presentation-based-uis/pages
workflow: ocr-mvp
stdout_capture: scripts/02-run-ocr-capture-log.py
```

The selected run was:

```text
run id: ocr-mvp-4c5c9406-926a-4ecd-a6b2-e8fedba847d8
work dir: /tmp/book-ocr-hq-001/007-quality-v4-mini-pages-001-030
```

The run completed successfully with 30 `done` page projection rows.

## Why This Configuration Won

The project started with a baseline single-page OCR pass, then iterated through prompt and model choices based on concrete observed failures.

### Baseline: Experiment 001

The baseline used the original `ocr-mvp-universal-v1` prompt with `gpt-5-nano-low`. It completed, but exposed quality problems:

- front matter was inconsistent;
- list pages drifted between bullets, headings, and plain text;
- Table of Contents and Table of Figures continuation pages changed style;
- footer/page-number handling required explicit policy;
- blank/title-page behavior needed stronger page-type rules.

### Quality v2: Experiment 002

`ocr-quality-v2` added explicit page-type rules. It improved blank pages, title pages, and figure-list formatting, but still left problems:

- some title-page output became too visually literal;
- Table of Contents continuation still had duplicate/formatting issues;
- page/list style was not yet consistently diplomatic.

### Quality v3: Experiments 003-005

`ocr-quality-v3-list-diplomatic` focused on contents and figure lists. It made list pages plain-text, reduced markdown heading drift, and improved title-page normalization.

The remaining problems were no longer mostly prompt-policy failures. They were visual-recognition failures on small list entries:

- `Dired` page-number drift;
- `Steamer` misread as `Streamer`;
- `PSBase` misread as `PPSBase`;
- occasional `DiRed` casing.

Switching from `gpt-5-nano-low` to `gpt-5-mini-low` improved these list-heavy pages.

### Quality v4: Experiments 006-007

`ocr-quality-v4-report794-lexicon` added a small book-specific lexicon:

- `Dired`, not `DiRed`;
- `Steamer`, not `Streamer`;
- `PSBase`, not `PPSBase`;
- `PPS` and `PPSCalc` remain valid;
- `Zmacs` and `Xerox Star` are expected terms;
- the title and author names are known.

With `gpt-5-mini-low`, this produced the best current 30-page result.

## Automated QA Results

Experiment 008 adds repeatable checks with:

```text
scripts/03-qa-ocr-markdown.py
```

Both pre-cleanup and post-cleanup QA passed.

Observed post-cleanup checks:

```text
Page markers found: 30
Expected page markers: 30
Figure markers: 2
Known bad term checks: pass
Expected string checks: pass
Adjacent duplicate non-empty lines: pass
List pages 006-009: no markdown bullets, no markdown headings
```

Known bad terms checked:

```text
DiRed
Streamer
PPSBase
Ciccarrelli
[IMAGE:
```

Expected strings checked:

```text
Presentation Based User Interfaces
This blank page was inserted to preserve pagination.
Figure 4-1: Dired Model
Figure 4-9: Sample Steamer Schematic
Figure 5-1: PSBase Support of PPS Components
Chapter Two
The Primitive Presentation System (PPS) Model
2.1 PPSCalc
```

## Vision Validation

The vision tool was used to validate selected hard pages and spot-check the final output.

### Pages 6-9

The vision tool confirmed:

- page 002 visibly contains `This blank page was inserted to preserve pagination.`;
- page 006 has `Chapter One: Introduction and Overview` on page 8;
- page 006 has `1.1 The Primitive Presentation System Model` on page 9;
- page 008 has `Figure 4-1: Dired Model` on page 72;
- page 008 has `Figure 4-9: Sample Steamer Schematic` on page 91;
- page 008 has `Figure 5-1: PSBase Support of PPS Components` on page 101.

### Pages 13, 15, and 30

The vision tool also spot-checked later selected pages:

- page 013: `Figure 1-1: A Rudimentary User Interface` and major diagram labels were present;
- page 015: `Figure 1-2: The Representation Shift Model` and major diagram labels were present;
- page 030: the page visibly starts `Chapter Two`, `The Primitive Presentation System (PPS) Model`, and section `2.1 PPSCalc`.

## Cleanup Pass

Experiment 008 uses:

```text
scripts/04-normalize-ocr-markdown.py
```

The cleanup normalizes list-page dot leaders on pages 006-009 into a consistent readable form:

```text
Figure 4-1: Dired Model ... 72
```

It does not call an LLM. It does not rewrite prose. It preserves the raw Experiment 007 output and writes a diff:

```text
experiments/008-deterministic-continuity-cleanup/outputs/03-cleanup-diff.patch
```

This is important: the cleanup is a review convenience, not hidden OCR correction.

## Remaining Limitations

The result is high quality, but not a proof-perfect transcription.

Known limitations:

- Dot leaders are normalized, not exact visual reproductions.
- Figure diagrams use concise `[FIGURE: ...]` markers, not structured diagram encodings.
- The OCR workflow is still single-page first pass; it does not use neighbor-page context during OCR.
- Prose line wrapping is readable but not globally normalized into publication-grade paragraphs.
- QA is heuristic. It catches known failure modes but cannot prove every word is correct.

## Review Instructions

Start with the final review artifact:

```text
experiments/008-deterministic-continuity-cleanup/outputs/02-final-quality-v4-mini-pages-001-030-normalized.md
```

Then inspect the diff from raw model output:

```text
experiments/008-deterministic-continuity-cleanup/outputs/03-cleanup-diff.patch
```

Then inspect the QA reports:

```text
experiments/008-deterministic-continuity-cleanup/outputs/01-qa-before-cleanup.md
experiments/008-deterministic-continuity-cleanup/outputs/04-qa-after-cleanup.md
```

For provenance, inspect raw output:

```text
experiments/007-quality-v4-mini-pages-001-030/outputs/01-final-quality-v4-mini-pages-001-030.md
```

## Future Work

The next quality jump should come from adding a true continuity workflow rather than more single-page prompt edits.

Recommended next steps:

1. Add a structured second-pass cleanup workflow using OCR output plus page metadata.
2. Add automated page-level QA projections to the workflow runtime.
3. Add targeted re-OCR for pages that fail QA.
4. Add figure-specific structured extraction for diagram pages.
5. Expand the first-30-page approach to the full book once the second-pass loop is stable.

## Final Status

`BOOK-OCR-HQ-001` is complete for the requested first-30-page high-quality OCR iteration.

The selected artifact is:

```text
experiments/008-deterministic-continuity-cleanup/outputs/02-final-quality-v4-mini-pages-001-030-normalized.md
```
