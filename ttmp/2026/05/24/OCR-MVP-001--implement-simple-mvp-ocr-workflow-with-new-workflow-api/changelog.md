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


## 2026-05-24

Completed Phase 2b profile-selection wiring tests and opt-in live OCR smoke test guard in scraper (commit 6a21bc3cbeaf420b235cec3b6ebdb36204188199).

### Related Files

- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/OCR-MVP-001--implement-simple-mvp-ocr-workflow-with-new-workflow-api/reference/01-diary.md — Diary updated with Phase 2b testing details
- /home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/workflows/ocrmvp/geppetto_ocr.go — Pinocchio selection helper used by live OCR client
- /home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/workflows/ocrmvp/geppetto_ocr_test.go — Profile-selection unit test and live smoke guard


## 2026-05-24

Implemented Phase 3 OCR MVP CLI and added operator smoke-flow documentation (commit 8a067f98c7e556ea1c9148bbc2838a0ef23a236a).

### Related Files

- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/OCR-MVP-001--implement-simple-mvp-ocr-workflow-with-new-workflow-api/design-doc/01-mvp-ocr-workflow-implementation-guide.md — Updated with CLI and operator smoke runbook
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/OCR-MVP-001--implement-simple-mvp-ocr-workflow-with-new-workflow-api/reference/01-diary.md — Diary updated with Phase 3 CLI details
- /home/manuel/workspaces/2026-05-20/book-ocr/scraper/cmd/ocr-mvp/main.go — CLI implementation


## 2026-05-24

Finalized OCR-MVP-001 implementation pass: all tasks complete, docs validated, and updated Phase 3 guide uploaded to reMarkable at /ai/2026/05/24/OCR-MVP-001.

### Related Files

- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/OCR-MVP-001--implement-simple-mvp-ocr-workflow-with-new-workflow-api/design-doc/01-mvp-ocr-workflow-implementation-guide.md — Updated implementation guide and runbook uploaded to reMarkable
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/OCR-MVP-001--implement-simple-mvp-ocr-workflow-with-new-workflow-api/reference/01-diary.md — Final diary entry with validation and upload evidence


## 2026-05-24

Added first-class OCR MVP operator subcommands for status, page listing, retry, and cancel (commit 5d0934a429bf699afb9dd88ad4ce1e90bb6648a4).

### Related Files

- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/OCR-MVP-001--implement-simple-mvp-ocr-workflow-with-new-workflow-api/design-doc/01-mvp-ocr-workflow-implementation-guide.md — Updated with concrete operator CLI runbook
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/OCR-MVP-001--implement-simple-mvp-ocr-workflow-with-new-workflow-api/reference/01-diary.md — Diary updated with operator CLI phase
- /home/manuel/workspaces/2026-05-20/book-ocr/scraper/cmd/ocr-mvp/main.go — Operator subcommands


## 2026-05-24

Uploaded refreshed OCR MVP guide with operator CLI documentation to reMarkable as 'OCR MVP 001 Workflow Guide Operator CLI.pdf'.

### Related Files

- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/OCR-MVP-001--implement-simple-mvp-ocr-workflow-with-new-workflow-api/design-doc/01-mvp-ocr-workflow-implementation-guide.md — Uploaded as part of refreshed operator CLI bundle
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/OCR-MVP-001--implement-simple-mvp-ocr-workflow-with-new-workflow-api/reference/01-diary.md — Recorded operator CLI upload evidence


## 2026-05-24

Ran live two-page provider OCR smoke test against presentation-based-uis pages; workflow succeeded and produced final markdown artifact in /tmp/ocr-mvp-live-presentation-two-pages.

### Related Files

- /home/manuel/code/wesen/claw-stuff/output/books/presentation-based-uis/pages/page_001.png — Live OCR smoke input page 1
- /home/manuel/code/wesen/claw-stuff/output/books/presentation-based-uis/pages/page_002.png — Live OCR smoke input page 2
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/OCR-MVP-001--implement-simple-mvp-ocr-workflow-with-new-workflow-api/reference/01-diary.md — Recorded live provider OCR smoke-test evidence

