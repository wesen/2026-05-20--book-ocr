# Changelog

## 2026-05-25

- Initial workspace created


## 2026-05-25

Created the workflow-backed structured OCR ticket, added detailed tasks, investigated current structured/freeform workflow code, and wrote the intern-facing design and implementation guide.

### Related Files

- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/cmd/book-ocr/main.go — Current CLI behavior
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/ocrmvp/package.go — Workflow package template
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/25/BOOK-OCR-STRUCTURED-WORKFLOW-001--promote-structured-ocr-to-workflow-runtime/design-doc/01-workflow-backed-structured-ocr-design-and-implementation-guide.md — Design guide
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/25/BOOK-OCR-STRUCTURED-WORKFLOW-001--promote-structured-ocr-to-workflow-runtime/reference/01-diary.md — Diary


## 2026-05-25

Uploaded the workflow-backed structured OCR guide bundle to reMarkable at /ai/2026/05/25/BOOK-OCR-STRUCTURED-WORKFLOW-001.

### Related Files

- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/25/BOOK-OCR-STRUCTURED-WORKFLOW-001--promote-structured-ocr-to-workflow-runtime/design-doc/01-workflow-backed-structured-ocr-design-and-implementation-guide.md — Uploaded design guide source
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/25/BOOK-OCR-STRUCTURED-WORKFLOW-001--promote-structured-ocr-to-workflow-runtime/reference/01-diary.md — Upload diary


## 2026-05-25

Implemented structured OCR workflow package, projection schema, workflow executors, queue wiring, and workflow-backed structured-run CLI (commits d4bf768, 325614b).

### Related Files

- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/cmd/book-ocr/main.go — CLI wiring
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/ocrpipeline/workflow_executors.go — Workflow executors
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/ocrpipeline/workflow_projection.go — Projection schema
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/25/BOOK-OCR-STRUCTURED-WORKFLOW-001--promote-structured-ocr-to-workflow-runtime/reference/01-diary.md — Implementation diary


## 2026-05-25

Validated workflow-backed structured OCR over pages 1-50 in dry-run mode; engine, projections, turns, artifacts, assemble, and validation all succeeded.

### Related Files

- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/25/BOOK-OCR-STRUCTURED-WORKFLOW-001--promote-structured-ocr-to-workflow-runtime/reference/01-diary.md — Dry-run validation diary


## 2026-05-25

Ran workflow-backed structured OCR live over pages 1-50 with --max-workers 4; all 50 page steps succeeded, validation warnings were zero, and the W4 experiment summary was recorded.

### Related Files

- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/25/BOOK-OCR-STRUCTURED-WORKFLOW-001--promote-structured-ocr-to-workflow-runtime/experiments/001-report794-50-workflow-live-w4/01-summary.md — Live W4 summary
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/25/BOOK-OCR-STRUCTURED-WORKFLOW-001--promote-structured-ocr-to-workflow-runtime/reference/01-diary.md — Live W4 diary


## 2026-05-25

Phase 6: added deterministic structured workflow retry test and projection-backed structured-pages status command (commit 9d42c5d).

### Related Files

- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/cmd/book-ocr/main.go — Structured pages command
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/ocrpipeline/workflow_retry_test.go — Retry test
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/25/BOOK-OCR-STRUCTURED-WORKFLOW-001--promote-structured-ocr-to-workflow-runtime/reference/01-diary.md — Phase 6 diary


## 2026-05-25

Phase 6: added opt-in short-page completeness warnings via --min-rendered-bytes and included short pages in validation-report.json (commit 936952d).

### Related Files

- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/cmd/book-ocr/main.go — CLI flag
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/ocrpipeline/workflow_executors.go — Short-page validation
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/25/BOOK-OCR-STRUCTURED-WORKFLOW-001--promote-structured-ocr-to-workflow-runtime/reference/01-diary.md — Completeness validation diary


## 2026-05-25

Phase 6: connected structured figure blocks to figure image embedding during assembly, producing embedded-figures.md plus crop/sidecar/debug artifacts (commit 600dbc7).

### Related Files

- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/cmd/book-ocr/main.go — CLI flags
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/ocrpipeline/workflow_executors.go — Figure embedding implementation
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/25/BOOK-OCR-STRUCTURED-WORKFLOW-001--promote-structured-ocr-to-workflow-runtime/reference/01-diary.md — Figure embedding diary


## 2026-05-26

Tightened structured OCR figure-versus-readable-text classification after PDF review; added code blocks and prompt rules for screenshot text/tables/listings (commit a576e96).

### Related Files

- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/ocrpipeline/prompts.go — Classification contract
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/25/BOOK-OCR-STRUCTURED-WORKFLOW-001--promote-structured-ocr-to-workflow-runtime/reference/01-diary.md — Review findings and follow-ups


## 2026-05-26

Added targeted structured page rerun operator and workflow PDF rendering; patched pages 20,30,31,83,84,86,96 in the previous full-book run and opened regenerated PDF (commit 52eba49).

### Related Files

- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/cmd/book-ocr/main.go — Targeted operator
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/ocrpipeline/workflow_executors.go — PDF assembly
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/25/BOOK-OCR-STRUCTURED-WORKFLOW-001--promote-structured-ocr-to-workflow-runtime/reference/01-diary.md — Step 10 diary


## 2026-05-26

Fixed duplicate image embedding for spreadsheet/table figures; reprocessed pages 34 and 48 and regenerated the PDF without PPSCalc spreadsheet image refs (commits 8825cd2, ff48bd4).

### Related Files

- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/ocrpipeline/renderer.go — Table-figure rendering
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/ocrquality/figures.go — Figure synthesis guard
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/25/BOOK-OCR-STRUCTURED-WORKFLOW-001--promote-structured-ocr-to-workflow-runtime/reference/01-diary.md — Step 11 diary


## 2026-05-26

Rendered structured code as Common Lisp, fixed targeted rerun downstream ordering, reprocessed code-heavy pages, and regenerated the review PDF (commit c74f34e).

### Related Files

- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/cmd/book-ocr/main.go — Targeted rerun dependency fix
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/ocrpipeline/renderer.go — Code fence rendering
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/25/BOOK-OCR-STRUCTURED-WORKFLOW-001--promote-structured-ocr-to-workflow-runtime/reference/01-diary.md — Step 12 diary

