# Changelog

## 2026-05-25

- Initial workspace created


## 2026-05-25

Created VLM separation investigation ticket and wrote the intern-facing benchmark design guide before implementing the tool.

### Related Files

- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/25/BOOK-OCR-VLM-SEPARATION-001--investigate-vlm-multi-page-input-separation-for-book-ocr/design-doc/01-vlm-multi-page-separation-benchmark-design-and-implementation-guide.md — Benchmark design and implementation guide
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/25/BOOK-OCR-VLM-SEPARATION-001--investigate-vlm-multi-page-input-separation-for-book-ocr/reference/01-diary.md — Investigation diary


## 2026-05-25

Uploaded the VLM separation benchmark guide and diary to reMarkable at /ai/2026/05/25/BOOK-OCR-VLM-SEPARATION-001.

### Related Files

- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/25/BOOK-OCR-VLM-SEPARATION-001--investigate-vlm-multi-page-input-separation-for-book-ocr/design-doc/01-vlm-multi-page-separation-benchmark-design-and-implementation-guide.md — Uploaded guide source
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/25/BOOK-OCR-VLM-SEPARATION-001--investigate-vlm-multi-page-input-separation-for-book-ocr/reference/01-diary.md — Uploaded diary source


## 2026-05-25

Implemented the Glazed VLM separation benchmark command with dry-run trials, file outputs, benchmark SQLite tables, and Pinocchio turns DB snapshots (commit 07db987).

### Related Files

- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/vlmseparation/command.go — Glazed command
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/vlmseparation/runner.go — Benchmark runner
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/vlmseparation/sqlite.go — Results DB
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/vlmseparation/turns.go — Turns DB wrapper


## 2026-05-25

Completed dry-run smoke validation: go test ./... passes and vlm-separation benchmark writes files, results.sqlite, and turns.db under /tmp/book-ocr-vlm-separation-dry.

### Related Files

- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/vlmseparation/command.go — Smoke-tested command
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/vlmseparation/runner_test.go — Dry-run persistence regression test


## 2026-05-25

Ran the first live VLM separation benchmark on pages 12/13 with gpt-5-mini-low, inspected results, and fixed nested Glazed logging initialization to suppress provider trace deltas (commit 3a13734).

### Related Files

- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/vlmseparation/command.go — Logging initialization fix
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/25/BOOK-OCR-VLM-SEPARATION-001--investigate-vlm-multi-page-input-separation-for-book-ocr/reference/01-diary.md — Live benchmark diary


## 2026-05-25

Added sanitize-backed response repair so live benchmark scoring separates schema drift from target/context bleed (commit 3e994dd).

### Related Files

- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/vlmseparation/scoring.go — Response repair
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/vlmseparation/sqlite.go — Metric schema migration
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/25/BOOK-OCR-VLM-SEPARATION-001--investigate-vlm-multi-page-input-separation-for-book-ocr/reference/01-diary.md — Diary update


## 2026-05-25

Updated and pushed the Obsidian VLM benchmark report with the repaired-parser interpretation (vault commit 0ff992f).

### Related Files

- /home/manuel/code/wesen/go-go-golems/go-go-parc/Projects/2026/05/25/ARTICLE - VLM Separation Benchmark for Book OCR - Prompt Block Layouts and Turn Persistence.md — Updated benchmark interpretation
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/25/BOOK-OCR-VLM-SEPARATION-001--investigate-vlm-multi-page-input-separation-for-book-ocr/reference/01-diary.md — Report update diary


## 2026-05-25

Added a first-class VLM benchmark rescore command and used it to rewrite the first live run metrics from saved responses (commit b606549).

### Related Files

- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/vlmseparation/command.go — CLI wiring
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/vlmseparation/rescore.go — Rescore implementation
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/25/BOOK-OCR-VLM-SEPARATION-001--investigate-vlm-multi-page-input-separation-for-book-ocr/reference/01-diary.md — Diary update


## 2026-05-25

Updated and pushed the Obsidian report with the new saved-run rescore workflow (vault commit 80f638f).

### Related Files

- /home/manuel/code/wesen/go-go-golems/go-go-parc/Projects/2026/05/25/ARTICLE - VLM Separation Benchmark for Book OCR - Prompt Block Layouts and Turn Persistence.md — Rescore workflow article update
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/25/BOOK-OCR-VLM-SEPARATION-001--investigate-vlm-multi-page-input-separation-for-book-ocr/reference/01-diary.md — Report update diary


## 2026-05-25

Ran the broad risky-page live benchmark, retried transient provider failures, hardened scoring, and pushed a rewritten deep-dive Obsidian article (vault commit 802b1b5).

### Related Files

- /home/manuel/code/wesen/go-go-golems/go-go-parc/Projects/2026/05/25/ARTICLE - VLM Separation Benchmark for Book OCR - Prompt Block Layouts and Turn Persistence.md — Deep-dive article
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/vlmseparation/oracle.go — Risky-page oracles
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/vlmseparation/scoring.go — Scoring hardening
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/25/BOOK-OCR-VLM-SEPARATION-001--investigate-vlm-multi-page-input-separation-for-book-ocr/reference/01-diary.md — Broad benchmark diary


## 2026-05-25

Added a formal VLM benchmark report command, generated broad-run report.md/report.json, and pushed the updated Obsidian article (code commit 99446a4, vault commit 1954655).

### Related Files

- /home/manuel/code/wesen/go-go-golems/go-go-parc/Projects/2026/05/25/ARTICLE - VLM Separation Benchmark for Book OCR - Prompt Block Layouts and Turn Persistence.md — Updated project report
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/vlmseparation/report.go — Report implementation
- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/25/BOOK-OCR-VLM-SEPARATION-001--investigate-vlm-multi-page-input-separation-for-book-ocr/reference/01-diary.md — Report diary

