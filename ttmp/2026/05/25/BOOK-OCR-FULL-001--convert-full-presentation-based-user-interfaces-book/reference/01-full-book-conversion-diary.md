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
      Note: |-
        External OCR CLI used for full conversion
        Added resume command used to finish full-book run
    - Path: internal/ocrmvp/package.go
      Note: Page OCR workflow package for full conversion
    - Path: internal/ocrquality/package.go
      Note: Quality pass workflow for full-book artifacts
    - Path: ttmp/2026/05/25/BOOK-OCR-FULL-001--convert-full-presentation-based-user-interfaces-book/experiments/001-full-book-v5-mini/outputs/01-full-book-raw.md
      Note: Complete 202-page raw OCR artifact
    - Path: ttmp/2026/05/25/BOOK-OCR-FULL-001--convert-full-presentation-based-user-interfaces-book/experiments/001-full-book-v5-mini/outputs/quality-pass/embedded-figures.md
      Note: Complete normalized OCR artifact with embedded figure links
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

## Step 2: Run the 202-page live OCR conversion and quality pass

I ran the full 202-page OCR workflow through the external `book-ocr` CLI using `gpt-5-mini-low`, the `ocr-quality-v5-figure-aware` prompt, and a one-page context window. The first long-running pass produced OCR artifacts for 201 pages and failed on page 129 after provider retries; I retried page 129, added a small `resume` operator command so an existing workflow can continue processing queued work, and then resumed the run.

After page 129 succeeded, the run still had a canceled `assemble-markdown` step because the workflow had already transitioned to failed when the page failure occurred. I manually reset that local SQLite op from `canceled` to `failed`, used the existing retry command to make it ready, and resumed once more. The assembler then produced a complete 202-page raw markdown artifact. I copied the final raw markdown and logs into the ticket experiment directory, then ran the quality pass with expected page count 202 and figure embedding enabled.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Execute the full conversion now that extraction is complete.

**Inferred user intent:** Produce complete full-book OCR artifacts, not just a partial experiment.

**Commit (code):** `593744414123b398ccbe681fa5c880f4b818849b` — "Add OCR run resume command"

### What I did

- Ran full OCR over pages 1–202 with:
  - model/profile: `gpt-5-mini-low`
  - prompt: `ocr-quality-v5-figure-aware`
  - context window: `1`
  - max workers: `2`
- Initial run ID:
  - `ocr-mvp-2456c4ae-6728-4511-aecc-3de87eb335e2`
- Work directory:
  - `/tmp/book-ocr-full-report794-v5-mini/work`
- Retried failed page step:
  - `ocr-page-129`
- Added `book-ocr resume --work-dir DIR --run-id RUN_ID` so queued/retried work in an existing run can be processed without starting a new run.
- Resumed the run and completed page 129.
- Reset the canceled assembler op locally, retried it, and resumed once more.
- Copied final raw artifact to:
  - `experiments/001-full-book-v5-mini/outputs/01-full-book-raw.md`
- Ran quality pass with figure embedding into:
  - `experiments/001-full-book-v5-mini/outputs/quality-pass/`
- Exported QA and figure-embedding result JSON from the quality workflow DB.

### Why

- The full-book artifact should live in the book OCR repository ticket, not only in `/tmp`.
- A failed long run needs a durable resume path. Retrying a step is not sufficient if no workers are running afterward.
- The quality pass is needed to normalize deterministic OCR artifacts, embed figure crops, write discovery/profile patch files, and verify page count.

### What worked

- Final raw markdown was assembled successfully:
  - 202 page markers
  - 409,099 bytes
- Quality pass succeeded:
  - normalized markdown: 404,606 bytes
  - embedded markdown: 406,481 bytes
  - page markers after QA: 202 / 202
  - embedded figure links: 75
  - figure PNGs: 75
  - figure JSON sidecars: 75
- Final primary artifacts:
  - `experiments/001-full-book-v5-mini/outputs/01-full-book-raw.md`
  - `experiments/001-full-book-v5-mini/outputs/quality-pass/normalized.md`
  - `experiments/001-full-book-v5-mini/outputs/quality-pass/embedded-figures.md`
  - `experiments/001-full-book-v5-mini/outputs/quality-pass/book.discovery.yaml`
  - `experiments/001-full-book-v5-mini/outputs/quality-pass/book.profile.patch.yaml`
  - `experiments/001-full-book-v5-mini/outputs/quality-pass/run-log.sqlite`

### What didn't work

- Page 129 failed during the initial live run after three provider attempts:
  - log event: `error_code="ocr_geppetto_failed"`, `op_id="ocr-page-129"`, `attempt=3`
- The workflow then marked the dependent `assemble-markdown` op as `canceled`. The existing retry command only accepts failed ops, not canceled ops.
- I used a one-off SQLite repair for this local run:
  - changed `assemble-markdown` from `canceled` to `failed`
  - changed workflow status back to `failed`
  - then used the normal retry command and resume command.

### What I learned

- The workflow runtime needs an operator path for resuming/reprocessing canceled downstream ops after a failed dependency is later repaired.
- Full-book OCR with two workers completed most of the book successfully; the only live OCR failure observed was page 129.
- The figure-aware prompt emits many more figure markers across the whole book than the first-30-page experiments, and the existing crop sidecar system handled 75 figures.

### What was tricky to build

- The original CLI had `retry` but no worker-driving `resume`. Once a failed step was retried, there was no command to process that ready op inside the existing run. The fix was to add a `resume` subcommand that opens the existing runtime, registers the OCR workflow package with a live Geppetto client, and loops `RunOnce` until the target workflow reaches a terminal state.
- The canceled assembler exposed a runtime semantic edge case: when a workflow fails, dependent pending work can be canceled. Repairing the failed page afterward does not automatically revive canceled downstream work. The local SQL reset was safe for this experiment because all page OCR ops had already succeeded, but this should become a first-class operator control before relying on it generally.

### What warrants a second pair of eyes

- Spot-check page 129, since it was the failed/retried page.
- Spot-check the 75 figure crops, especially pages with multiple figures such as 31, 32, 33, 48, and 60.
- Review whether `resume` should also support quality workflows or whether it should be generalized at the workflow runtime/operator layer.
- Decide whether retrying canceled downstream ops should be supported by `RetryStep` or a separate `requeue`/`repair-run` command.

### What should be done in the future

- Add a generalized resume/requeue operator command to the runtime package rather than only the book OCR CLI.
- Add automated QA for missing/canceled downstream assembly after page retries.
- Add richer figure QA over all 75 crops.
- Consider uploading the final embedded full-book artifact to reMarkable after a quick visual/text spot-check.

### Code review instructions

- Review the new resume command:
  - `/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/cmd/book-ocr/main.go`
- Review final artifacts:
  - `/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/25/BOOK-OCR-FULL-001--convert-full-presentation-based-user-interfaces-book/experiments/001-full-book-v5-mini/outputs/01-full-book-raw.md`
  - `/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/25/BOOK-OCR-FULL-001--convert-full-presentation-based-user-interfaces-book/experiments/001-full-book-v5-mini/outputs/quality-pass/embedded-figures.md`
- Validate with:
  - `cd /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr && go test ./... -count=1`

### Technical details

Full run command:

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

Retry/resume commands:

```bash
go run ./cmd/book-ocr retry \
  --work-dir /tmp/book-ocr-full-report794-v5-mini/work \
  --run-id ocr-mvp-2456c4ae-6728-4511-aecc-3de87eb335e2 \
  --step-id ocr-page-129

go run ./cmd/book-ocr resume \
  --work-dir /tmp/book-ocr-full-report794-v5-mini/work \
  --run-id ocr-mvp-2456c4ae-6728-4511-aecc-3de87eb335e2 \
  --max-workers 1 \
  --log-level warn
```

Assembler repair used for this local experiment:

```bash
sqlite3 /tmp/book-ocr-full-report794-v5-mini/work/engine.db \
  "update ops set status='failed', updated_at=strftime('%Y-%m-%dT%H:%M:%fZ','now') where workflow_id='ocr-mvp-2456c4ae-6728-4511-aecc-3de87eb335e2' and id='assemble-markdown' and status='canceled'; update workflows set status='failed', updated_at=strftime('%Y-%m-%dT%H:%M:%fZ','now') where id='ocr-mvp-2456c4ae-6728-4511-aecc-3de87eb335e2';"
```

Quality pass command:

```bash
go run ./cmd/book-ocr quality-pass \
  --markdown experiments/001-full-book-v5-mini/outputs/01-full-book-raw.md \
  --output-dir experiments/001-full-book-v5-mini/outputs/quality-pass \
  --work-dir /tmp/book-ocr-full-report794-v5-mini/quality-work \
  --book-id report-794-full-v5-mini \
  --expected-pages 202 \
  --log experiments/001-full-book-v5-mini/logs/01-run.log \
  --image-dir /home/manuel/code/wesen/claw-stuff/output/books/presentation-based-uis/pages \
  --embed-figures
```
