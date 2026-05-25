---
Title: Diary
Ticket: BOOK-OCR-VLM-SEPARATION-001
Status: active
Topics:
    - ocr
    - book-processing
    - experiments
    - geppetto
    - pinocchio
    - workflow
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: cmd/book-ocr/main.go
      Note: Registers vlm-separation benchmark subcommand
    - Path: internal/ocrmvp/geppetto_ocr.go
      Note: Root cause code path for multi-image OCR context
    - Path: internal/vlmseparation/command.go
      Note: Glazed benchmark command implementation
    - Path: internal/vlmseparation/runner.go
      Note: Benchmark orchestration and dry-run/live execution
    - Path: internal/vlmseparation/scenarios.go
      Note: Scenario-specific Geppetto turn/block layouts
    - Path: internal/vlmseparation/sqlite.go
      Note: Benchmark run/trial/metric SQLite persistence
    - Path: internal/vlmseparation/turns.go
      Note: Pinocchio turns DB wrapper
    - Path: ttmp/2026/05/25/BOOK-OCR-VLM-SEPARATION-001--investigate-vlm-multi-page-input-separation-for-book-ocr/design-doc/01-vlm-multi-page-separation-benchmark-design-and-implementation-guide.md
      Note: Main investigation design guide
ExternalSources: []
Summary: Diary for the VLM multi-page input separation benchmark investigation.
LastUpdated: 2026-05-25T00:00:00-04:00
WhatFor: Use this diary to understand the investigation setup before implementing the Glazed benchmark command.
WhenToUse: Read before continuing BOOK-OCR-VLM-SEPARATION-001 or changing the benchmark design.
---



# Diary

## Goal

Design and implement a benchmark tool that investigates whether VLMs can keep target-page OCR separate from neighboring page images under different prompt and Geppetto turn/block layouts.

## Step 1: Create the ticket and write the benchmark design before implementation

I created `BOOK-OCR-VLM-SEPARATION-001` as a focused investigation ticket. This ticket comes before changing the production OCR pipeline because the user wants to know whether the full-book context bleed was mainly a prompting problem, a block-layout problem, or a more fundamental limitation of passing multiple page images to a VLM.

I wrote an intern-facing implementation guide that explains the observed full-book regression, the current OCR/Geppetto/Pinocchio architecture, the benchmark scenarios, the SQLite/file/turns-db persistence plan, and the Glazed command shape. Per the user's instruction, I did this before writing the tool.

### Prompt Context

**User prompt (verbatim):** "before we continue, create a tool that allows us to invsstigate how good our vlm is at separating multiple pages passed as input. Maybe it's a prompting issue, or maybe images blocks need to be separated by text blocks (we can put multiple user blocks in sequence). I think we should investigate that first. 

Create a ticket for this investigation into prompting, and write a program that allows us to benchmark that under different scenarios, logging results properly both in files and in sqlite, including also a turns db (that way we already get some stuff done in that regard). Use galzed command for this. 

Create  a detailed analysis / design / implementation guide that is for a new intern, explaining all the parts of the system needed to understand what it is, with prose paragraphs and bullet point sand pseudocode and diagrams and api references and file references. It should be very clear and technical. Store in the ticket and the nupload to remarkable.

Keep a diary as you worok"

**Assistant interpretation:** Create the investigation ticket and documentation package first, then implement a Glazed benchmark command that tests multi-page VLM input separation across prompt/block-layout scenarios and logs to files, SQLite, and a Pinocchio turns DB.

**Inferred user intent:** The user wants evidence before committing to a production OCR redesign: benchmark whether prompt/block separation can make neighboring page images safe, while also beginning turns-db observability work.

**Commit (code):** N/A — documentation/ticket setup only.

### What I did

- Created docmgr ticket `BOOK-OCR-VLM-SEPARATION-001`.
- Added tasks for:
  - guide creation,
  - reMarkable upload,
  - Glazed benchmark command implementation,
  - file/SQLite/turns-db persistence,
  - dry-run smoke validation.
- Wrote:
  - `design-doc/01-vlm-multi-page-separation-benchmark-design-and-implementation-guide.md`
- Wrote this diary.
- Reviewed relevant APIs and source references:
  - Glazed command authoring skill,
  - current `cmd/book-ocr/main.go`,
  - current `internal/ocrmvp/geppetto_ocr.go`,
  - Geppetto `turns` helpers,
  - Pinocchio `chatstore` and CLI persistence patterns.

### Why

- The current full-book artifact showed adjacent-page bleed, but we do not yet know whether a different prompt or turn/block layout can prevent it.
- The benchmark will let us compare target-only, single-block multi-image, multi-block labeled, context-first, and text-context scenarios.
- File logs, SQLite rows, and turns-db snapshots are all needed because prompt/layout debugging requires exact replayable evidence.

### What worked

- The current code already has the dependencies needed: Geppetto, Pinocchio, Glazed, SQLite.
- Pinocchio already provides the durable turns DB (`chatstore.SQLiteTurnStore`), so no new turn database needs to be invented.
- The design can be implemented as a Glazed command without rewriting the entire existing manual `book-ocr` CLI.

### What didn't work

- N/A. Implementation has not started yet by design.

### What I learned

- The benchmark should explicitly test Geppetto block layout, not just prompt wording.
- The most important scenarios are `target-only`, `single-block-target-first`, `multi-block-labeled`, `context-first-negative-control`, and `target-plus-text-context`.
- The turn IDs must encode page and scenario to make the resulting turns DB useful.

### What was tricky to build

- The guide had to separate benchmark concerns from production OCR concerns. This tool is not the final OCR pipeline; it is a measurement harness that informs the next pipeline.
- The tool must be useful in dry-run mode so tests can run without live provider calls, while still preserving the same file/SQLite/turn persistence paths used in live runs.

### What warrants a second pair of eyes

- The exact page oracles for page 12/13 and other adjacent figure pages should be reviewed manually before live scoring.
- The Glazed command integration approach should be checked because `cmd/book-ocr/main.go` is currently a manual flag-based CLI, not a fully Glazed root.
- The choice of response schema should be reviewed to make sure it is easy to score and robust against model formatting mistakes.

### What should be done in the future

- Upload the guide to reMarkable.
- Commit the ticket documents.
- Implement `internal/vlmseparation` with dry-run tests first.
- Add the Glazed `vlm-separation benchmark` command.
- Run a tiny live benchmark only after dry-run/file/SQLite/turn persistence passes.

### Code review instructions

- Start with the guide:
  - `/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/25/BOOK-OCR-VLM-SEPARATION-001--investigate-vlm-multi-page-input-separation-for-book-ocr/design-doc/01-vlm-multi-page-separation-benchmark-design-and-implementation-guide.md`
- Verify that the scenarios actually distinguish single-block vs multi-block image layouts.
- Verify that the implementation plan starts with dry-run persistence tests before live model calls.

### Technical details

Preferred command shape:

```bash
go run ./cmd/book-ocr vlm-separation benchmark \
  --book-id report-794 \
  --image-dir /home/manuel/code/wesen/claw-stuff/output/books/presentation-based-uis/pages \
  --target-pages 12,13 \
  --scenarios target-only,single-block-target-first,multi-block-labeled \
  --out-dir /tmp/book-ocr-vlm-separation-dry \
  --sqlite /tmp/book-ocr-vlm-separation-dry/results.sqlite \
  --turns-db /tmp/book-ocr-vlm-separation-dry/turns.db \
  --dry-run=true \
  --output json
```

Persistence targets:

```text
files:    <out-dir>/manifest.json and <out-dir>/trials/trial-*/...
sqlite:   <out-dir>/results.sqlite
turns db: <out-dir>/turns.db
```

## Step 2: Upload the investigation guide to reMarkable

I uploaded the benchmark design guide and diary as a bundled PDF to reMarkable before beginning implementation. This keeps the design reviewable as a standalone reading artifact and satisfies the requested order: ticket, diary, guide, upload, then tool work.

The upload succeeded and the ticket now records the reMarkable location.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Finish the documentation package and make it available on reMarkable before writing code.

**Inferred user intent:** The user wants a reviewed implementation plan before the benchmark tool changes the codebase.

**Commit (code):** N/A — upload/bookkeeping only.

### What I did

- Uploaded a bundle containing:
  - `design-doc/01-vlm-multi-page-separation-benchmark-design-and-implementation-guide.md`
  - `reference/01-diary.md`
- Destination:
  - `/ai/2026/05/25/BOOK-OCR-VLM-SEPARATION-001/BOOK OCR VLM SEPARATION 001 Guide.pdf`
- Checked task 2.
- Updated the changelog.

### Why

- The user explicitly requested upload to reMarkable.
- Uploading before implementation gives the design a stable review point.

### What worked

- `remarquee upload bundle` succeeded on the first attempt.

### What didn't work

- N/A.

### What I learned

- The guide is acceptable to the current Markdown-to-PDF upload path.

### What was tricky to build

- N/A for upload.

### What warrants a second pair of eyes

- Review the PDF for code block readability and table formatting on reMarkable.

### What should be done in the future

- Commit the ticket documentation.
- Begin the Glazed benchmark implementation with dry-run persistence tests.

### Code review instructions

- Validate the ticket with:
  - `docmgr doctor --ticket BOOK-OCR-VLM-SEPARATION-001 --stale-after 30`

### Technical details

Upload command shape:

```bash
remarquee upload bundle DESIGN.md DIARY.md \
  --name "BOOK OCR VLM SEPARATION 001 Guide" \
  --remote-dir "/ai/2026/05/25/BOOK-OCR-VLM-SEPARATION-001" \
  --toc-depth 2 \
  --non-interactive
```

## Step 3: Implement the dry-run VLM separation benchmark command

I implemented the first version of the benchmark tool as a Glazed command under `book-ocr vlm-separation benchmark`. The command can run dry-run scenario/page trials, emit Glazed rows, write a file tree with manifests and per-trial artifacts, write benchmark metrics to SQLite, and persist exact input/final Geppetto turns into a Pinocchio-compatible turns DB.

This implementation deliberately starts with deterministic dry-run behavior. Live provider benchmarking is now possible through the same command shape by passing `--dry-run=false`, but the normal test path does not call external models. This gives us the persistence, scoring, and block-layout machinery before spending provider tokens.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** After creating the ticket/guide/diary/upload, continue by implementing the benchmark tool with Glazed, files, SQLite, and turns DB persistence.

**Inferred user intent:** The user wants a practical measurement harness that can test whether prompt/block-layout choices make multi-page VLM input safe enough for OCR context.

**Commit (code):** `07db987bd8a6ce7f8908b469a1ff13482b1e3229` — "Add VLM separation benchmark command"

### What I did

- Added `internal/vlmseparation` with:
  - `types.go` for run/trial/scenario/metric/result contracts,
  - `oracle.go` for page presets and conservative page oracles,
  - `scoring.go` for JSON parsing and bleed scoring,
  - `scenarios.go` for Geppetto turn construction under different image/text block layouts,
  - `files.go` for manifest/trial/summary file outputs,
  - `sqlite.go` for benchmark run/trial/metric tables,
  - `turns.go` for a wrapper around Pinocchio `chatstore.SQLiteTurnStore`,
  - `runner.go` for benchmark orchestration and fake/live execution path,
  - `command.go` for the Glazed command.
- Added tests:
  - `scoring_test.go`,
  - `scenarios_test.go`,
  - `runner_test.go`.
- Registered the command in `cmd/book-ocr/main.go` under:
  - `vlm-separation benchmark`.
- Ran:
  - `go test ./internal/vlmseparation -count=1`
  - `go test ./... -count=1`
- Ran a dry-run command against Report 794 page images:
  - pages 12 and 13,
  - scenarios `target-only`, `single-block-target-first`, `multi-block-labeled`,
  - output directory `/tmp/book-ocr-vlm-separation-dry`.

### Why

- We need evidence before deciding whether neighboring page images are always unsafe or whether better block layout can make them safe.
- The benchmark needs exact turn snapshots because block layout is the variable under test.
- The benchmark also exercises the planned turns-db plumbing for future structured OCR work.

### What worked

- The Glazed command emits one row per benchmark trial.
- Dry-run wrote:
  - `/tmp/book-ocr-vlm-separation-dry/manifest.json`,
  - `/tmp/book-ocr-vlm-separation-dry/scenarios.json`,
  - `/tmp/book-ocr-vlm-separation-dry/summary.json`,
  - `/tmp/book-ocr-vlm-separation-dry/results.sqlite`,
  - `/tmp/book-ocr-vlm-separation-dry/turns.db`,
  - per-trial `turn-input.yaml`, `turn-final.yaml`, `response.txt`, `response.json`, `metrics.json`, and `trial.json`.
- SQLite dry-run counts:
  - `benchmark_trials`: 6 rows,
  - turns DB `turns`: 6 rows,
  - turns DB `blocks`: 28 rows,
  - turns DB `turn_block_membership`: 50 rows.
- `go test ./... -count=1` passes.

### What didn't work

- No live provider benchmark has been run yet. The implementation is validated in dry-run mode only.
- The page oracles are intentionally conservative and need manual review before relying on live scores.
- The benchmark results DB and turns DB are separate files for now. That avoids coupling benchmark analytics tables to Pinocchio's canonical turn-store schema, but it means users inspect two SQLite files.

### What I learned

- The existing Geppetto turn model can represent the scenarios we need:
  - one multimodal block with all images,
  - multiple user blocks separated by text,
  - target image plus text-only context.
- Pinocchio's turn store works as-is for benchmark input/final turn snapshots.
- Glazed can be integrated incrementally into the existing manual `book-ocr` CLI without rewriting all current subcommands.

### What was tricky to build

- The current `book-ocr` CLI is not a fully Glazed root; it is mostly manual `flag` parsing. I avoided a risky root rewrite by adding only a `vlm-separation` subcommand whose `benchmark` child is built from a Glazed command.
- The same turn is saved with different phases (`input`, `final`) in the Pinocchio store. The canonical `turns` table has one row per `(conv_id, session_id, turn_id)`, while phase-specific block membership is recorded in `turn_block_membership`.
- The file logger and SQLite logger need to cooperate: file paths are written after trial artifacts are emitted, then the final `TrialResult` is inserted into SQLite.

### What warrants a second pair of eyes

- Review the scenario block layouts in `internal/vlmseparation/scenarios.go` and confirm they match the intended experiment.
- Review whether `single-block-labeled-images` is meaningfully different from `single-block-target-first` for the active providers, given that provider adapters may ignore custom image metadata.
- Review the page oracles in `internal/vlmseparation/oracle.go`, especially page 12 forbidden captions.
- Decide whether benchmark tables should optionally live in the same SQLite file as the turns DB.

### What should be done in the future

- Add an `inspect` command for reading benchmark result SQLite and turns DB summaries.
- Add more curated oracles for known duplicate-caption pages.
- Run a small opt-in live benchmark with `gpt-5-mini-low` on pages 12/13 and two scenarios.
- Compare live outputs before changing the production OCR context strategy.

### Code review instructions

- Start at:
  - `/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/vlmseparation/command.go`
- Then review:
  - `scenarios.go` for block layout construction,
  - `runner.go` for persistence flow,
  - `turns.go` for Pinocchio store reuse,
  - `sqlite.go` for benchmark tables,
  - tests in `internal/vlmseparation/*_test.go`.
- Validate with:
  - `go test ./internal/vlmseparation -count=1`
  - `go test ./... -count=1`
  - dry-run command from the technical details below.

### Technical details

Dry-run command used:

```bash
go run ./cmd/book-ocr vlm-separation benchmark \
  --book-id report-794 \
  --image-dir /home/manuel/code/wesen/claw-stuff/output/books/presentation-based-uis/pages \
  --target-pages 12,13 \
  --scenarios target-only,single-block-target-first,multi-block-labeled \
  --out-dir /tmp/book-ocr-vlm-separation-dry \
  --dry-run \
  --output json
```

SQLite checks used:

```bash
sqlite3 /tmp/book-ocr-vlm-separation-dry/results.sqlite \
  "select count(*) from benchmark_trials;"

sqlite3 /tmp/book-ocr-vlm-separation-dry/turns.db \
  "select count(*) from turns; select count(*) from blocks; select count(*) from turn_block_membership;"
```
