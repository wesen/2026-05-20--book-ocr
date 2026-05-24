# Changelog

## 2026-05-24

- Initial workspace created


## 2026-05-24

Created OCR-MVP-001 ticket, wrote intern-oriented MVP OCR workflow implementation guide, and incorporated Geppetto OCR plus Pinocchio default profile registry requirements.

### Related Files

- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/OCR-MVP-001--implement-simple-mvp-ocr-workflow-with-new-workflow-api/design-doc/01-mvp-ocr-workflow-implementation-guide.md — Primary design and implementation guide
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/OCR-MVP-001--implement-simple-mvp-ocr-workflow-with-new-workflow-api/reference/01-diary.md — Chronological diary for ticket setup and design decisions
- /home/manuel/workspaces/2026-05-20/book-ocr/geppetto/pkg/turns/helpers_blocks.go — Multimodal image block API referenced by guide
- /home/manuel/workspaces/2026-05-20/book-ocr/pinocchio/pkg/cmds/profilebootstrap/profile_selection.go — Profile registry default behavior referenced by guide


## 2026-05-24

Validated OCR-MVP-001 docs, added missing topic vocabulary entries, and uploaded OCR MVP 001 Workflow Guide.pdf to reMarkable at /ai/2026/05/24/OCR-MVP-001.

### Related Files

- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/OCR-MVP-001--implement-simple-mvp-ocr-workflow-with-new-workflow-api/design-doc/01-mvp-ocr-workflow-implementation-guide.md — Validated and uploaded primary guide
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/OCR-MVP-001--implement-simple-mvp-ocr-workflow-with-new-workflow-api/reference/01-diary.md — Updated with validation and upload evidence
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/vocabulary.yaml — Added workflow and implementation-guide topic slugs


## 2026-05-24

Implemented Phase 1 OCR MVP workflow skeleton and fake-client integration tests in scraper (commit f827d63671369d3ea762e11e8c9bab61f0266dbf); added detailed Phase 1-3 task breakdown.

### Related Files

- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/OCR-MVP-001--implement-simple-mvp-ocr-workflow-with-new-workflow-api/reference/01-diary.md — Diary updated with Phase 1 implementation and validation details
- /home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/workflows/ocrmvp/executors.go — Discover/OCR/assemble workflow behavior
- /home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/workflows/ocrmvp/package.go — OCR MVP registration and package entrypoint
- /home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/workflows/ocrmvp/package_test.go — Integration test with fake OCR


## 2026-05-24

Implemented Phase 2a Geppetto OCR client using Pinocchio profilebootstrap default registry resolution in scraper (commit 0f3b04556260f1d07f13032b89bbca3df2a66b5f).

### Related Files

- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/OCR-MVP-001--implement-simple-mvp-ocr-workflow-with-new-workflow-api/reference/01-diary.md — Diary updated with Phase 2a implementation and validation details
- /home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/workflows/ocrmvp/geppetto_ocr.go — Live OCR client path
- /home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/workflows/ocrmvp/geppetto_ocr_test.go — Non-live tests for OCR output extraction helpers

