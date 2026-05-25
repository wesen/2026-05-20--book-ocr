---
Title: Diary
Ticket: OCR-QUALITY-WORKERS-001
Status: active
Topics:
    - ocr
    - workflow
    - experiments
    - book-processing
    - implementation-guide
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: 2026-05-20--book-ocr/ttmp/2026/05/24/BOOK-OCR-HQ-001--high-quality-book-ocr-experiment-system/analysis/01-final-ocr-quality-report.md
      Note: Source for follow-up tasks
    - Path: scraper/pkg/workflows/ocrmvp/prompt.go
      Note: Context policy prompt update
    - Path: scraper/pkg/workflows/ocrquality/package.go
      Note: Main implementation for Step 1
ExternalSources: []
Summary: Diary for porting OCR QA and cleanup scripts to Go workflow workers.
LastUpdated: 2026-05-24T20:40:00-04:00
WhatFor: Use this to understand what changed while porting OCR quality scripts to Go and adding context-aware OCR inputs.
WhenToUse: Read before continuing OCR quality worker implementation or debugging the quality-pass workflow.
---


# Diary

## Goal

This diary records the work to port OCR quality Python scripts into Go workflow workers, add tasks and an implementation guide, and begin exploring surrounding-page context for OCR continuity.

## Step 1: Create the worker-port ticket and implement the first Go quality workers

I created `OCR-QUALITY-WORKERS-001` as the follow-up ticket for turning the stabilized Python experiment scripts from `BOOK-OCR-HQ-001` into proper Go workflow workers. The first implementation ports the useful stable behavior: markdown QA, deterministic list-page normalization, log import summaries, a workflow package, and a CLI `quality-pass` entry point.

I also added an exploratory `--context-window` option to `ocr-mvp run`. This lets a page OCR call include previous/next page images for continuity, while the v4 prompt explicitly instructs the model to transcribe only the target page.

### Prompt Context

**User prompt (verbatim):** "port the ocr python things to go (add  implementation guide and tasks). And then add all the recommended next steps in @2026-05-20--book-ocr/ttmp/2026/05/24/BOOK-OCR-HQ-001--high-quality-book-ocr-experiment-system/analysis/01-final-ocr-quality-report.md  as tasks as well. also potentially explore giving surrounding context to the OCR to improve continutiny.

Then optimize the whole thing to get _stellar_ results, including embedded extracted images.

Don't stop until the quality is impeccable"

**Assistant interpretation:** Create the follow-up ticket/docs/tasks, port the Python QA/log/normalization scripts into Go workflow workers, add the final report's next steps as tasks, and start improving OCR continuity and future image extraction toward higher quality.

**Inferred user intent:** The user wants the experiment-prototype tooling promoted into production-ish workflow infrastructure, then wants the OCR quality loop to continue beyond prompt-only improvements.

**Commit (code):** `eb19a4018ef5ebfbc89b730de597e686aeb5303f` — "Add OCR quality workflow workers"

### What I did

- Created ticket `OCR-QUALITY-WORKERS-001`.
- Added tasks for implementation guide, Go QA port, normalization port, workflow package, CLI, context-aware OCR, comparison run, and future embedded image extraction.
- Added the recommended next steps from `BOOK-OCR-HQ-001` final report as follow-up tasks in that ticket.
- Implemented `scraper/pkg/workflows/ocrquality`:
  - `QAResult`, `QAFinding`, `NormalizeResult`, `LogImportResult` types.
  - page-aware markdown splitting.
  - known bad term checks.
  - expected string checks.
  - adjacent duplicate line checks.
  - list-page markdown bullet/heading checks.
  - deterministic list-page dot-leader normalization.
  - NDJSON/plain log import summary with optional SQLite output.
  - workflow package with `qa-before`, `normalize-markdown`, `qa-after`, optional `import-log`, and `assemble-quality-report` steps.
- Added `ocr-mvp quality-pass` CLI command.
- Added `--context-window` to `ocr-mvp run`.
- Added context image plumbing through `PageOCRInput` and Geppetto multimodal calls.
- Added tests for QA, normalization, and context-window page selection.
- Ran the quality workflow against Experiment 007's raw markdown and confirmed it succeeded.

### Why

- The Python scripts had proven their value but were not first-class workflow workers.
- Typed QA findings and durable artifacts are needed before building operator/UI loops or targeted re-OCR.
- Surrounding-page context is a plausible next quality lever for continuity, but it must be explicit and bounded to avoid context leakage.

### What worked

- `go test ./cmd/ocr-mvp ./pkg/workflows/ocrquality ./pkg/workflows/ocrmvp -count=1` passed during development.
- The full pre-commit test/lint/security suite passed before the code commit.
- `ocr-mvp quality-pass` succeeded on the Experiment 007 markdown and wrote:
  - `/tmp/ocr-quality-go-pass/out/normalized.md`
  - `/tmp/ocr-quality-go-pass/out/cleanup.diff`
- The Go normalized output is semantically equivalent to the Python cleanup but has slightly tighter page-boundary whitespace.

### What didn't work

- The first commit attempt failed lint because `file.Close()` and `db.Close()` return values were not checked:

```text
pkg/workflows/ocrquality/logimport.go:41:18: Error return value of `file.Close` is not checked (errcheck)
pkg/workflows/ocrquality/logimport.go:52:17: Error return value of `db.Close` is not checked (errcheck)
```

- The first commit attempt also failed because a local variable was named `max`, shadowing the predeclared identifier:

```text
pkg/workflows/ocrquality/normalize.go:125:2: variable max has same name as predeclared identifier (predeclared)
```

- After fixing those, gosec flagged local artifact writes as potential path traversal (`G703`). I added narrow `#nosec G703` comments explaining that the paths are explicit operator/workflow inputs for local artifact export.

### What I learned

- The Python scripts mapped cleanly into three reusable worker concepts: QA, normalization, and log import.
- A workflow-native quality pass is more useful than a CLI-only port because it stores step results and artifacts in the same runtime model as OCR.
- Context-window OCR is easy to add mechanically, but it requires careful prompt wording and targeted experiments to ensure the model does not transcribe context pages.

### What was tricky to build

- The quality workflow needed to produce both local files and workflow artifacts. Local files are convenient for direct review, while artifacts preserve runtime provenance.
- The normalizer had to stay intentionally narrow. It should improve reviewability without becoming an invisible semantic editor.
- The Geppetto multimodal context path has to preserve target-page identity. The prompt now says the first image is always the target page and any additional images are context only.

### What warrants a second pair of eyes

- Whether the `quality-pass` CLI belongs under `ocr-mvp` long-term or should become a separate command.
- Whether `--context-window` should default to zero for safety, as it does now, or whether a future profile should enable it for list/continuation pages only.
- Whether log import should later be split into a pure library and a workflow executor with projection tables.

### What should be done in the future

- Add page-level QA projection tables instead of only result JSON/artifacts.
- Add targeted re-OCR workflow steps for failed QA pages.
- Add embedded figure/image extraction and markdown references to stored image artifacts.
- Run controlled context-window OCR experiments on pages 6-9, 13-15, and 29-31 before broad use.
- Compare context-window output against v4 mini baseline with both automated QA and vision checks.

### Code review instructions

- Start with:
  - `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/workflows/ocrquality/package.go`
  - `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/workflows/ocrquality/qa.go`
  - `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/workflows/ocrquality/normalize.go`
- Then review context OCR plumbing:
  - `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/workflows/ocrmvp/types.go`
  - `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/workflows/ocrmvp/discover.go`
  - `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/workflows/ocrmvp/geppetto_ocr.go`
  - `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/workflows/ocrmvp/prompt.go`
- Validate with:
  - `cd /home/manuel/workspaces/2026-05-20/book-ocr/scraper && go test ./cmd/ocr-mvp ./pkg/workflows/ocrquality ./pkg/workflows/ocrmvp -count=1`

### Technical details

Quality pass smoke command:

```bash
cd /home/manuel/workspaces/2026-05-20/book-ocr/scraper
rm -rf /tmp/ocr-quality-go-pass

go run ./cmd/ocr-mvp quality-pass \
  --markdown /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/BOOK-OCR-HQ-001--high-quality-book-ocr-experiment-system/experiments/007-quality-v4-mini-pages-001-030/outputs/01-final-quality-v4-mini-pages-001-030.md \
  --output-dir /tmp/ocr-quality-go-pass/out \
  --work-dir /tmp/ocr-quality-go-pass/work \
  --book-id presentation-based-uis-hq-007-v4-mini-30 \
  --expected-pages 30
```

Context-window OCR pattern for future experiments:

```bash
go run ./cmd/ocr-mvp run \
  --book-id presentation-based-uis-context-window-test \
  --image-dir /home/manuel/code/wesen/claw-stuff/output/books/presentation-based-uis/pages \
  --start-page 6 \
  --end-page 9 \
  --profile gpt-5-mini-low \
  --profile-registries /tmp/book-ocr-hq-001/profiles-clean.yaml \
  --prompt-version ocr-quality-v4-report794-lexicon \
  --context-window 1 \
  --log-level warn
```
