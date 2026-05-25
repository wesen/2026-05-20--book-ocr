# Changelog

## 2026-05-24

- Initial workspace created


## 2026-05-24

Created high-quality OCR experiment ticket, intern-oriented design guide, diary, and experiment folder layout; wrote and pushed Obsidian project report.

### Related Files

- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/BOOK-OCR-HQ-001--high-quality-book-ocr-experiment-system/design-doc/01-high-quality-book-ocr-experiment-system.md — Initial design and implementation guide
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/BOOK-OCR-HQ-001--high-quality-book-ocr-experiment-system/experiments/001-baseline-single-page/manifest.yaml — Baseline experiment manifest
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/BOOK-OCR-HQ-001--high-quality-book-ocr-experiment-system/reference/01-experiment-diary.md — Initial experiment diary


## 2026-05-24

Uploaded initial BOOK-OCR-HQ-001 design guide and diary bundle to reMarkable at /ai/2026/05/24/BOOK-OCR-HQ-001.

### Related Files

- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/BOOK-OCR-HQ-001--high-quality-book-ocr-experiment-system/design-doc/01-high-quality-book-ocr-experiment-system.md — Uploaded in initial design bundle
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/BOOK-OCR-HQ-001--high-quality-book-ocr-experiment-system/reference/01-experiment-diary.md — Recorded upload and doctor fixes


## 2026-05-24

Ran baseline pages 1-30 with clean Pinocchio registry, added SQLite log filtering for noisy SSE traces, and captured final baseline artifacts.

### Related Files

- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/BOOK-OCR-HQ-001--high-quality-book-ocr-experiment-system/experiments/001-baseline-single-page/outputs/01-final-baseline-clean.md — Baseline output
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/BOOK-OCR-HQ-001--high-quality-book-ocr-experiment-system/reference/01-experiment-diary.md — Recorded baseline/log-filter step
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/BOOK-OCR-HQ-001--high-quality-book-ocr-experiment-system/scripts/01-filter-ndjson-log-to-sqlite.py — Log filtering script


## 2026-05-24

Assessed baseline OCR quality, validated list-page failures with the vision tool, added ocr-quality-v2 prompt and CLI logging controls, ran targeted pages 1-9 experiment, and captured outputs.

### Related Files

- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/BOOK-OCR-HQ-001--high-quality-book-ocr-experiment-system/experiments/002-quality-v2-targeted/notes.md — Experiment evidence
- /home/manuel/workspaces/2026-05-20/book-ocr/scraper/cmd/ocr-mvp/main.go — Logging and prompt CLI controls
- /home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/workflows/ocrmvp/prompt.go — Quality prompt


## 2026-05-24

Iterated OCR prompts through v3/v4, compared nano versus mini on hard list pages, ran the best-current v4 mini OCR for pages 1-30, and recorded QA notes with vision spot-checks.

### Related Files

- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/BOOK-OCR-HQ-001--high-quality-book-ocr-experiment-system/experiments/007-quality-v4-mini-pages-001-030/notes.md — QA notes
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/BOOK-OCR-HQ-001--high-quality-book-ocr-experiment-system/experiments/007-quality-v4-mini-pages-001-030/outputs/01-final-quality-v4-mini-pages-001-030.md — Best output
- /home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/workflows/ocrmvp/prompt.go — Prompt iterations


## 2026-05-24

Added deterministic OCR markdown QA and list-page continuity cleanup, produced normalized 30-page review artifact, preserved cleanup diff, and recorded Experiment 008.

### Related Files

- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/BOOK-OCR-HQ-001--high-quality-book-ocr-experiment-system/experiments/008-deterministic-continuity-cleanup/outputs/02-final-quality-v4-mini-pages-001-030-normalized.md — Normalized output
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/BOOK-OCR-HQ-001--high-quality-book-ocr-experiment-system/scripts/03-qa-ocr-markdown.py — QA script
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/BOOK-OCR-HQ-001--high-quality-book-ocr-experiment-system/scripts/04-normalize-ocr-markdown.py — Cleanup script


## 2026-05-24

Wrote final OCR quality report selecting Experiment 008 normalized markdown as the review artifact and Experiment 007 raw markdown as provenance.

### Related Files

- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/BOOK-OCR-HQ-001--high-quality-book-ocr-experiment-system/analysis/01-final-ocr-quality-report.md — Final report
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/BOOK-OCR-HQ-001--high-quality-book-ocr-experiment-system/experiments/008-deterministic-continuity-cleanup/outputs/02-final-quality-v4-mini-pages-001-030-normalized.md — Selected output


## 2026-05-24

Ticket closed


## 2026-05-24

Added final report recommended next steps as follow-up tasks and started implementation in OCR-QUALITY-WORKERS-001.

### Related Files

- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/BOOK-OCR-HQ-001--high-quality-book-ocr-experiment-system/tasks.md — Follow-up tasks
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/OCR-QUALITY-WORKERS-001--port-ocr-qa-and-cleanup-scripts-to-go-workflow-workers/design-doc/01-ocr-quality-workers-implementation-guide.md — Follow-up implementation guide

