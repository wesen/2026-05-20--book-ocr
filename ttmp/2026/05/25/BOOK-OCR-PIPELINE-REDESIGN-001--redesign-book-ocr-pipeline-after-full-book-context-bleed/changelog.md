# Changelog

## 2026-05-25

- Initial workspace created


## 2026-05-25

Created pipeline redesign ticket and wrote intern-facing guide covering target-page-only structured OCR, Geppetto turn sessions, Pinocchio turn persistence, deterministic rendering, figure QA, and validation gates.

### Related Files

- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/25/BOOK-OCR-PIPELINE-REDESIGN-001--redesign-book-ocr-pipeline-after-full-book-context-bleed/design-doc/01-structured-book-ocr-pipeline-redesign-and-implementation-guide.md — Design and implementation guide
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/25/BOOK-OCR-PIPELINE-REDESIGN-001--redesign-book-ocr-pipeline-after-full-book-context-bleed/reference/01-diary.md — Work diary


## 2026-05-25

Uploaded the structured Book OCR pipeline redesign guide and diary to reMarkable at /ai/2026/05/25/BOOK-OCR-PIPELINE-REDESIGN-001.

### Related Files

- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/25/BOOK-OCR-PIPELINE-REDESIGN-001--redesign-book-ocr-pipeline-after-full-book-context-bleed/design-doc/01-structured-book-ocr-pipeline-redesign-and-implementation-guide.md — Uploaded guide source
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/25/BOOK-OCR-PIPELINE-REDESIGN-001--redesign-book-ocr-pipeline-after-full-book-context-bleed/reference/01-diary.md — Uploaded diary source


## 2026-05-25

Updated the structured OCR redesign with VLM separation benchmark evidence, diagnostic/production boundaries, and validation-first implementation guidance.

### Related Files

- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/25/BOOK-OCR-PIPELINE-REDESIGN-001--redesign-book-ocr-pipeline-after-full-book-context-bleed/design-doc/01-structured-book-ocr-pipeline-redesign-and-implementation-guide.md — Updated redesign guide
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/25/BOOK-OCR-PIPELINE-REDESIGN-001--redesign-book-ocr-pipeline-after-full-book-context-bleed/reference/01-diary.md — Redesign diary


## 2026-05-25

Implemented the first deterministic structured OCR contracts: renderer, turn-store wrapper, and validation helpers (commit 5011269).

### Related Files

- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/ocrpipeline/renderer.go — Renderer
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/ocrpipeline/session.go — Turn persistence
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/ocrpipeline/types.go — Structured contracts
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/ocrvalidation/anchors.go — Anchor validation
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/25/BOOK-OCR-PIPELINE-REDESIGN-001--redesign-book-ocr-pipeline-after-full-book-context-bleed/reference/01-diary.md — Implementation diary


## 2026-05-25

Updated and pushed the Obsidian project report with refined benchmark oracle results and structured OCR contract progress (vault commit 54c97ee).

### Related Files

- /home/manuel/code/wesen/go-go-golems/go-go-parc/Projects/2026/05/25/ARTICLE - VLM Separation Benchmark for Book OCR - Prompt Block Layouts and Turn Persistence.md — Updated vault report
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/25/BOOK-OCR-PIPELINE-REDESIGN-001--redesign-book-ocr-pipeline-after-full-book-context-bleed/reference/01-diary.md — Vault update diary


## 2026-05-25

Ran a live first-50 Report 794 target-page-only OCR rerun, quality pass, and md-view review; recorded the experiment summary (run ocr-mvp-d8701e29-d511-4d6a-9860-a44b75be1b20, summary commit 742bfde).

### Related Files

- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/25/BOOK-OCR-PIPELINE-REDESIGN-001--redesign-book-ocr-pipeline-after-full-book-context-bleed/experiments/001-report794-50-target-only-v5/01-summary.md — Experiment summary
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/25/BOOK-OCR-PIPELINE-REDESIGN-001--redesign-book-ocr-pipeline-after-full-book-context-bleed/reference/01-diary.md — Rerun diary


## 2026-05-25

Added restart-ready structured OCR implementation phases and docmgr tasks for tomorrow's work.

### Related Files

- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/25/BOOK-OCR-PIPELINE-REDESIGN-001--redesign-book-ocr-pipeline-after-full-book-context-bleed/design-doc/01-structured-book-ocr-pipeline-redesign-and-implementation-guide.md — Fresh-start implementation plan
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/25/BOOK-OCR-PIPELINE-REDESIGN-001--redesign-book-ocr-pipeline-after-full-book-context-bleed/reference/01-diary.md — Planning diary
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/25/BOOK-OCR-PIPELINE-REDESIGN-001--redesign-book-ocr-pipeline-after-full-book-context-bleed/tasks.md — Phase task list


## 2026-05-25

Refreshed the structured OCR design guide upload on reMarkable after adding restart-ready implementation phases.

### Related Files

- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/25/BOOK-OCR-PIPELINE-REDESIGN-001--redesign-book-ocr-pipeline-after-full-book-context-bleed/design-doc/01-structured-book-ocr-pipeline-redesign-and-implementation-guide.md — Uploaded source guide
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/25/BOOK-OCR-PIPELINE-REDESIGN-001--redesign-book-ocr-pipeline-after-full-book-context-bleed/reference/01-diary.md — Upload diary


## 2026-05-25

Phase 1: implemented structured-page dry-run with target-page-only turn construction, fake page 32 table fixture, artifacts, validation output, and turn input/final persistence (commit cab6b6f).

### Related Files

- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/cmd/book-ocr/main.go — CLI command
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/ocrpipeline/client.go — One-image turn invariant
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/ocrpipeline/prompts.go — Prompt contract
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/ocrpipeline/structured_ocr.go — Dry-run orchestration
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/25/BOOK-OCR-PIPELINE-REDESIGN-001--redesign-book-ocr-pipeline-after-full-book-context-bleed/reference/01-diary.md — Phase 1 diary

