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
    - Path: internal/ocrmvp/geppetto_ocr.go
      Note: Root cause code path for multi-image OCR context
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
