---
Title: Full Book Conversion Diary
Ticket: BOOK-OCR-FULL-001
Status: active
Topics:
    - ocr
    - book-processing
    - experiments
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: cmd/book-ocr/main.go
      Note: External OCR CLI used for full conversion
    - Path: internal/ocrmvp/package.go
      Note: Page OCR workflow package for full conversion
    - Path: internal/ocrquality/package.go
      Note: Quality pass workflow for full-book artifacts
ExternalSources: []
Summary: Diary for the full 202-page OCR conversion of Presentation Based User Interfaces.
LastUpdated: 2026-05-25T00:00:00-04:00
WhatFor: Use this to retrace commands, outputs, failures, and quality decisions for the full-book conversion.
WhenToUse: Read before resuming, validating, or publishing the full-book OCR artifacts.
---


# Diary

## Goal

Convert all 202 rendered page images of *Presentation Based User Interfaces* into OCR markdown using the externalized `book-ocr` workflow application, then run the quality pass and preserve final artifacts.

## Step 1: Prepare full-book conversion ticket and validate prerequisites

I created `BOOK-OCR-FULL-001` to track the full-book conversion separately from the extraction/migration work. The external `book-ocr` application is now the active OCR command, while `scraper` only provides workflow/runtime infrastructure.

Before starting the live conversion, I checked that the rendered page directory contains 202 PNG pages, that the temporary clean Pinocchio profile registry still exists, and that the external `book-ocr` Go module tests pass.

### Prompt Context

**User prompt (verbatim):** "then when you are done, convert the whole book"

**Assistant interpretation:** The OCR extraction work is done; proceed to convert the entire 202-page book rather than only the first 30 pages.

**Inferred user intent:** The user wants the pipeline to produce a complete book OCR artifact using the improved, externalized OCR workflow.

**Commit (code):** N/A — setup/tracking only.

### What I did

- Created docmgr ticket `BOOK-OCR-FULL-001`.
- Added tasks for full OCR run, quality pass, and artifact preservation.
- Verified source pages:
  - `/home/manuel/code/wesen/claw-stuff/output/books/presentation-based-uis/pages`
  - 202 `page_*.png` files.
- Verified profile registry workaround exists:
  - `/tmp/book-ocr-hq-001/profiles-clean.yaml`
- Ran:
  - `cd /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr && go test ./... -count=1`

### Why

- Full-book conversion is a long-running live provider workflow. It needs its own ticket, diary, and artifact plan.
- The clean profile registry is still needed because the default local Pinocchio registry has a known duplicate `gpt-5-nano-low` key.

### What worked

- 202 source page images are available.
- `/tmp/book-ocr-hq-001/profiles-clean.yaml` exists.
- External `book-ocr` tests pass.

### What didn't work

- N/A in setup.

### What I learned

- The full conversion should use the external command:
  - `go run ./cmd/book-ocr run ...`
- The old `scraper/cmd/ocr-mvp` no longer exists after the extraction.

### What was tricky to build

- The full run should preserve provenance separately from the earlier first-30-page experiments, because it has a larger scope and may expose new quality classes.

### What warrants a second pair of eyes

- Whether to use `--context-window 1` for the whole book. It improves continuity but increases multimodal context and can risk context leakage.

### What should be done in the future

- Run the live 202-page OCR workflow.
- Run quality pass with figure embedding.
- Preserve final markdown, embedded markdown, discovery, patch, QA reports, and logs.

### Code review instructions

- Start with the external CLI:
  - `/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/cmd/book-ocr/main.go`
- Then inspect the resulting work directory and artifacts.

### Technical details

Expected live run shape:

```bash
go run ./cmd/book-ocr run \
  --book-id report-794-full-v5-mini \
  --image-dir /home/manuel/code/wesen/claw-stuff/output/books/presentation-based-uis/pages \
  --work-dir /tmp/book-ocr-full-report794-v5-mini/work \
  --start-page 1 \
  --end-page 202 \
  --profile gpt-5-mini-low \
  --profile-registries /tmp/book-ocr-hq-001/profiles-clean.yaml \
  --prompt-version ocr-quality-v5-figure-aware \
  --context-window 1 \
  --max-workers 2 \
  --log-level warn
```
