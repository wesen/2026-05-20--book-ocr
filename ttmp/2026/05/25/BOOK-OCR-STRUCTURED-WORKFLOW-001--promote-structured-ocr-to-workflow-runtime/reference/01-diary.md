---
Title: Diary
Ticket: ""
Status: active
Topics:
    - book-processing
    - ocr
    - workflow
    - geppetto
    - pinocchio
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: cmd/book-ocr/main.go
      Note: |-
        Workflow-backed structured-run CLI wiring and resume registration (commit 325614b)
        structured-pages status command (commit 9d42c5d)
    - Path: internal/ocrpipeline/workflow_executors.go
      Note: Discover/page/assemble/validate workflow executors (commit d4bf768)
    - Path: internal/ocrpipeline/workflow_package.go
      Note: Structured workflow package registration and retry policy (commit d4bf768)
    - Path: internal/ocrpipeline/workflow_projection.go
      Note: Structured page projection schema and helpers (commit d4bf768)
    - Path: internal/ocrpipeline/workflow_retry_test.go
      Note: Deterministic workflow retry proof (commit 9d42c5d)
    - Path: internal/ocrpipeline/workflow_types.go
      Note: Structured workflow input/result contracts (commit d4bf768)
    - Path: ttmp/2026/05/25/BOOK-OCR-STRUCTURED-WORKFLOW-001--promote-structured-ocr-to-workflow-runtime/design-doc/01-workflow-backed-structured-ocr-design-and-implementation-guide.md
      Note: Primary workflow promotion design guide
    - Path: ttmp/2026/05/25/BOOK-OCR-STRUCTURED-WORKFLOW-001--promote-structured-ocr-to-workflow-runtime/experiments/001-report794-50-workflow-live-w4/01-summary.md
      Note: Workflow-backed first-50 live W4 experiment summary
    - Path: ttmp/2026/05/25/BOOK-OCR-STRUCTURED-WORKFLOW-001--promote-structured-ocr-to-workflow-runtime/reference/01-diary.md
      Note: Chronological implementation diary
    - Path: ttmp/2026/05/25/BOOK-OCR-STRUCTURED-WORKFLOW-001--promote-structured-ocr-to-workflow-runtime/tasks.md
      Note: Phase 6 hardening task list
ExternalSources: []
Summary: ""
LastUpdated: 0001-01-01T00:00:00Z
WhatFor: ""
WhenToUse: ""
---





# Diary

## Goal

This diary records the work to promote the structured OCR runner from a direct CLI loop into a workflow-runtime-backed package with durable retry, resume, projections, artifacts, and operator controls.

## Step 1: Create ticket and write workflow promotion design guide

I created a new ticket for Option B: promoting structured OCR into the existing `scraper/pkg/workflow` runtime instead of continuing to extend CLI-level artifact resume. The ticket now has a design guide that explains the current direct structured runner, the existing freeform workflow package, the workflow runtime APIs, and the planned structured workflow package.

The design guide is intentionally intern-facing. It explains the codebase boundaries, file references, current gaps, proposed step graph, retry policy, projection schema, artifact strategy, CLI behavior, implementation phases, and validation commands.

### Prompt Context

**User prompt (verbatim):** "ok, create a new ticket for option B, and create a design + implementatino guide. Create  a detailed analysis / design / implementation guide that is for a new intern, explaining all the parts of the system needed to understand what it is, with prose paragraphs and bullet point sand pseudocode and diagrams and api references and file references. It should be very clear and technical. Store in the ticket and the nupload to remarkable.


Then add detailed tasks to the ticket, and then work on them one by one, keeping a detailed diary, and committing at appropriate intervals."

**Assistant interpretation:** Create a new docmgr ticket for workflow-backed structured OCR, write a comprehensive design/implementation guide, upload it to reMarkable, add detailed tasks, then start implementing task-by-task with diary and commits.

**Inferred user intent:** The user wants to stop relying on CLI-level resume and use the workflow runtime's real retry/resume/operator capabilities for structured OCR, with enough documentation for a new engineer to pick up the work.

**Commit (code):** N/A — ticket setup and design documentation only.

### What I did

- Created ticket `BOOK-OCR-STRUCTURED-WORKFLOW-001`.
- Added design doc:
  - `design-doc/01-workflow-backed-structured-ocr-design-and-implementation-guide.md`
- Added this diary:
  - `reference/01-diary.md`
- Added initial tasks covering investigation, guide writing/upload, workflow package implementation, CLI wiring, projections, dry-run validation, and live smoke validation.
- Inspected current structured OCR CLI and page pipeline.
- Inspected old freeform OCR workflow package.
- Inspected the workflow runtime API surface in `scraper/pkg/workflow`.
- Wrote the long-form design and implementation guide.

### Why

- The structured OCR first-50 live run proved table and figure-boundary architecture, but operationally it still exits on errors.
- The project already has a workflow runtime that supports retries and operator controls, so structured OCR should use it rather than recreate retries in CLI code.

### What worked

- Ticket and docs were created successfully.
- The design guide now maps:
  - current direct structured runner,
  - reusable structured page pipeline,
  - freeform workflow template,
  - workflow runtime APIs,
  - desired structured workflow graph,
  - implementation phases.

### What didn't work

- N/A

### What I learned

- The structured page function `RunStructuredPage` is already a good workflow executor boundary. The missing layer is packaging, projections, workflow artifacts, and CLI runtime wiring.
- The old `internal/ocrmvp` workflow is the best immediate implementation template.

### What was tricky to build

- The main design tradeoff is whether to refactor shared discovery out of `internal/ocrmvp` before implementing structured workflow. The guide recommends reusing the old helper initially or moving it later only if necessary, to avoid blocking workflow promotion on cleanup.

### What warrants a second pair of eyes

- Review whether parse failures should be permanent or retryable.
- Review whether `structured-run` should keep a `--direct` fallback after workflow integration.
- Review the proposed structured projection schema before tests depend on it.

### What should be done in the future

- Upload the design guide to reMarkable.
- Implement Phase 1: `workflow_types.go` and `workflow_package.go`.

### Code review instructions

- Start with the design doc:
  - `/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/25/BOOK-OCR-STRUCTURED-WORKFLOW-001--promote-structured-ocr-to-workflow-runtime/design-doc/01-workflow-backed-structured-ocr-design-and-implementation-guide.md`
- Then inspect the implementation templates listed in the guide.

### Technical details

Ticket commands used:

```bash
docmgr ticket create-ticket \
  --ticket BOOK-OCR-STRUCTURED-WORKFLOW-001 \
  --title "Promote Structured OCR to Workflow Runtime" \
  --topics book-processing,ocr,workflow,geppetto,pinocchio

docmgr doc add \
  --ticket BOOK-OCR-STRUCTURED-WORKFLOW-001 \
  --doc-type design-doc \
  --title "Workflow Backed Structured OCR Design and Implementation Guide"

docmgr doc add \
  --ticket BOOK-OCR-STRUCTURED-WORKFLOW-001 \
  --doc-type reference \
  --title "Diary"
```

## Step 2: Upload workflow promotion guide to reMarkable

I uploaded the workflow-backed structured OCR guide and diary as a bundled PDF to reMarkable. This gives the intern-facing implementation guide an offline review copy before implementation begins.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** After writing the guide, publish it to reMarkable as requested.

**Inferred user intent:** The user wants the new ticket's design package available for reading outside the coding environment.

**Commit (code):** N/A — upload/bookkeeping only.

### What I did

- Uploaded the design guide and diary with `remarquee upload bundle`.
- Destination:
  - `/ai/2026/05/25/BOOK-OCR-STRUCTURED-WORKFLOW-001/BOOK OCR STRUCTURED WORKFLOW 001 Guide.pdf`

### Why

- The user explicitly requested upload to reMarkable after creating the design and implementation guide.

### What worked

- Upload succeeded with:
  - `OK: uploaded BOOK OCR STRUCTURED WORKFLOW 001 Guide.pdf -> /ai/2026/05/25/BOOK-OCR-STRUCTURED-WORKFLOW-001`

### What didn't work

- N/A

### What I learned

- The ticket now has both local docmgr documentation and a reMarkable PDF handoff artifact.

### What was tricky to build

- N/A

### What warrants a second pair of eyes

- Confirm the PDF renders the Mermaid-style ASCII diagrams and long code blocks readably on device.

### What should be done in the future

- Re-upload after major implementation updates if the guide changes materially.

### Code review instructions

- Review the uploaded sources:
  - `/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/25/BOOK-OCR-STRUCTURED-WORKFLOW-001--promote-structured-ocr-to-workflow-runtime/design-doc/01-workflow-backed-structured-ocr-design-and-implementation-guide.md`
  - `/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/25/BOOK-OCR-STRUCTURED-WORKFLOW-001--promote-structured-ocr-to-workflow-runtime/reference/01-diary.md`

### Technical details

Upload command:

```bash
remarquee upload bundle \
  ttmp/2026/05/25/BOOK-OCR-STRUCTURED-WORKFLOW-001--promote-structured-ocr-to-workflow-runtime/design-doc/01-workflow-backed-structured-ocr-design-and-implementation-guide.md \
  ttmp/2026/05/25/BOOK-OCR-STRUCTURED-WORKFLOW-001--promote-structured-ocr-to-workflow-runtime/reference/01-diary.md \
  --name "BOOK OCR STRUCTURED WORKFLOW 001 Guide" \
  --remote-dir "/ai/2026/05/25/BOOK-OCR-STRUCTURED-WORKFLOW-001" \
  --toc-depth 2 \
  --non-interactive
```

## Step 3: Implement workflow package, projections, executors, and CLI wiring

I implemented the first workflow-backed structured OCR slice. Structured OCR now has a workflow package with discover, page OCR, assemble, and validate executors. The `structured-run` command now starts a durable workflow run instead of looping over pages directly.

This changes the operational model: page OCR work is now represented as workflow steps with retry policies, external artifacts, projection rows, and dependency-driven assembly/validation. The direct `structured-page` command remains available for debugging one page outside the workflow runtime.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Begin implementing the ticket tasks one by one after writing and uploading the guide.

**Inferred user intent:** The user wants Option B implemented, not just documented: structured OCR should use the workflow runtime for retries and resume.

**Commit (code):** `d4bf768a4c4829a31e8c3074b7b405f0c8f843f9` — "Add structured OCR workflow package"

**Commit (code):** `325614b8888a489ca4b56fd3107cb394b84e3457` — "Wire structured OCR command to workflow runtime"

### What I did

- Added workflow types:
  - `internal/ocrpipeline/workflow_types.go`
- Added workflow package registration:
  - `internal/ocrpipeline/workflow_package.go`
- Added projection schema/helpers:
  - `internal/ocrpipeline/workflow_projection.go`
- Added workflow executors:
  - `internal/ocrpipeline/workflow_executors.go`
- Rewired `book-ocr structured-run` to:
  - open a workflow runtime,
  - register the structured workflow package,
  - start a durable workflow run,
  - loop `RunOnce` until terminal,
  - print assemble and validation results.
- Updated `resume` so it registers the structured workflow package as well as the freeform OCR package.
- Added structured queues to `newRuntime`.

### Why

- The direct structured runner exited on transient page errors. The workflow runtime already knows how to retry page steps and persist run state, so structured OCR should use it.

### What worked

- Package tests passed:
  - `go test ./internal/ocrpipeline -count=1`
- Full tests passed:
  - `go test ./... -count=1`
- A 3-page workflow-backed dry-run succeeded:
  - `/tmp/book-ocr-structured-workflow-dry-smoke`
- The workflow wrote:
  - `engine.db`,
  - `turns.db`,
  - per-page artifacts,
  - `assembled.md`,
  - `validation-report.json`,
  - projection DB `projections/book_ocr_structured.db`.
- Projection query showed structured page success rows:
  - `succeeded|3`

### What didn't work

- I initially queried the projection file with the wrong extension:
  - `/tmp/book-ocr-structured-workflow-dry-smoke/projections/book_ocr_structured.sqlite`
- The actual projection file created by `workflow.NewSQLiteProjectionStore` is:
  - `/tmp/book-ocr-structured-workflow-dry-smoke/projections/book_ocr_structured.db`

### What I learned

- `RunStructuredPage` is a clean executor boundary. The workflow layer can call it without duplicating Geppetto, rendering, parsing, or turn-store behavior.
- The workflow package can keep the local artifact layout while also storing workflow artifacts for operator tools.

### What was tricky to build

- The discover executor emits both the page steps and downstream aggregate steps. The validate step depends on assemble, while assemble depends on all page OCR steps.
- The command previously had `--resume=true` artifact skip behavior. Workflow-backed resume is now provided by the engine DB and `book-ocr resume`, not by skipping files in a direct loop.

### What warrants a second pair of eyes

- Review whether parse errors should stay permanent or become retryable with a low retry count.
- Review whether structured workflow should continue using `ocrmvp.DiscoverPageImages`, or whether discovery should be moved into a shared package.
- Review projection table fields before relying on them in operator UI/tools.

### What should be done in the future

- Run pages 1-50 in workflow-backed dry-run mode and inspect DB/projections/artifacts.
- Run a limited live workflow smoke and verify automatic retry or manual resume.

### Code review instructions

- Start with:
  - `/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/ocrpipeline/workflow_executors.go`
- Then review:
  - `/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/ocrpipeline/workflow_package.go`
  - `/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/ocrpipeline/workflow_projection.go`
  - `/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/cmd/book-ocr/main.go`
- Validate with:
  - `go test ./... -count=1`

### Technical details

Smoke command:

```bash
go run ./cmd/book-ocr structured-run \
  --book-id report-794-structured-workflow-dry \
  --image-dir /home/manuel/code/wesen/claw-stuff/output/books/presentation-based-uis/pages \
  --start-page 1 \
  --end-page 3 \
  --work-dir /tmp/book-ocr-structured-workflow-dry-smoke \
  --dry-run \
  --expected-pages 3 \
  --max-workers 2 \
  --log-level warn
```

Projection query:

```bash
sqlite3 /tmp/book-ocr-structured-workflow-dry-smoke/projections/book_ocr_structured.db \
  'select status,count(*) from structured_pages group by status;'
```

## Step 4: Validate workflow-backed structured OCR over 50 dry-run pages

I ran the new workflow-backed `structured-run` command over the first 50 Report 794 pages in dry-run mode. This validates the durable workflow graph at the same page count used by the earlier direct structured runner, but now with workflow engine state, page steps, projections, external workflow artifacts, turn persistence, assembly, and validation.

The run succeeded. The workflow engine recorded 53 succeeded ops: one discover step, fifty page OCR steps, one assemble step, and one validate step. The structured projection recorded all 50 pages as succeeded.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Continue working through the ticket tasks after implementing workflow support.

**Inferred user intent:** The user wants proof that the workflow-backed structured OCR path works before live provider validation.

**Commit (code):** N/A — validation/documentation only.

### What I did

- Ran `book-ocr structured-run` with:
  - pages 1-50,
  - `--dry-run`,
  - `--expected-pages 50`,
  - `--max-workers 4`.
- Inspected assembled Markdown page markers.
- Inspected workflow engine op status counts.
- Inspected structured projection status counts.
- Inspected turn DB counts.
- Inspected artifact count.

### Why

- The workflow implementation should be validated at realistic page count before spending live provider calls.
- This confirms that dynamic page step emission, dependency assembly, validation, projections, and artifacts work together.

### What worked

- Run succeeded at:
  - `/tmp/book-ocr-structured-workflow-dry-50`
- Page markers:
  - `50`
- Structured projection:
  - `succeeded|50`
- Engine ops:
  - `succeeded|53`
- Turn DB:
  - `50` turns
  - `input|100`
  - `final|150`
- Workflow artifact files:
  - `152`
- Assemble result:
  - `/tmp/book-ocr-structured-workflow-dry-50/assembled.md`
- Validation result:
  - `/tmp/book-ocr-structured-workflow-dry-50/validation-report.json`

### What didn't work

- N/A

### What I learned

- The workflow-backed path preserves the same local page artifact layout while also writing workflow artifacts and projections.
- The engine op count matches the expected graph shape: discover + 50 page steps + assemble + validate.

### What was tricky to build

- Nothing new was built in this step, but this validation confirms the earlier dependency graph was wired correctly. If assemble had run before page dependencies, page count and dependency reads would have failed.

### What warrants a second pair of eyes

- Review artifact count expectations. There are per-page structured/rendered/validation artifacts plus assemble and validation artifacts, which accounts for the high count.

### What should be done in the future

- Run limited live workflow smoke with `gpt-5-mini-low` and verify automatic retry/resume behavior.

### Code review instructions

- Review output directory:
  - `/tmp/book-ocr-structured-workflow-dry-50`
- Review projection:
  - `/tmp/book-ocr-structured-workflow-dry-50/projections/book_ocr_structured.db`
- Review engine DB:
  - `/tmp/book-ocr-structured-workflow-dry-50/engine.db`

### Technical details

Command:

```bash
go run ./cmd/book-ocr structured-run \
  --book-id report-794-structured-workflow-dry-50 \
  --image-dir /home/manuel/code/wesen/claw-stuff/output/books/presentation-based-uis/pages \
  --start-page 1 \
  --end-page 50 \
  --work-dir /tmp/book-ocr-structured-workflow-dry-50 \
  --dry-run \
  --expected-pages 50 \
  --max-workers 4 \
  --log-level warn
```

Validation commands:

```bash
rg -c '<!-- page:' /tmp/book-ocr-structured-workflow-dry-50/assembled.md
sqlite3 /tmp/book-ocr-structured-workflow-dry-50/projections/book_ocr_structured.db \
  'select status,count(*) from structured_pages group by status;'
sqlite3 /tmp/book-ocr-structured-workflow-dry-50/engine.db \
  'select status,count(*) from ops group by status;'
sqlite3 /tmp/book-ocr-structured-workflow-dry-50/turns.db \
  'select count(*) from turns;'
```

## Step 5: Run workflow-backed first-50 live OCR with four workers

I ran the workflow-backed structured OCR pipeline on pages 1-50 with `--max-workers 4` from the start. This was the first real parallel live run of the structured workflow path.

The run succeeded without provider failures. It produced 50 structured page outputs, 50 projection success rows, 53 succeeded workflow ops, an assembled Markdown file, a validation report, and a shared turns DB. No separate `--max-vision-workers` flag was needed for this experiment; the existing workflow `--max-workers` setting was enough to exercise page-level parallelism.

### Prompt Context

**User prompt (verbatim):** "ok, let's try 4 from the start, just for fun."

**Assistant interpretation:** Run the workflow-backed structured OCR live path with four concurrent workers immediately, rather than adding a separate max-vision-workers flag first.

**Inferred user intent:** The user wants practical evidence about whether the workflow-backed structured OCR can parallelize live OCR calls at worker concurrency 4.

**Commit (code):** N/A — live validation/documentation only.

### What I did

- Ran `book-ocr structured-run` live over pages 1-50 with:
  - `--max-workers 4`,
  - `--dry-run=false`,
  - `--profile gpt-5-mini-low`,
  - `--expected-pages 50`.
- Measured assembled Markdown, projection, engine DB, and turns DB.
- Wrote experiment summary:
  - `/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/25/BOOK-OCR-STRUCTURED-WORKFLOW-001--promote-structured-ocr-to-workflow-runtime/experiments/001-report794-50-workflow-live-w4/01-summary.md`

### Why

- The user asked whether we could parallelize OCR. The structured workflow graph should allow page steps to run concurrently after discovery. A live W4 run is the simplest proof.

### What worked

- Run succeeded:
  - run ID `book-ocr/structured-f32e3de4-eb1f-49cd-b000-f4364241238b`
  - work dir `/tmp/book-ocr-structured-workflow-live-50-w4`
- Final artifacts:
  - `/tmp/book-ocr-structured-workflow-live-50-w4/assembled.md`
  - `/tmp/book-ocr-structured-workflow-live-50-w4/validation-report.json`
  - `/tmp/book-ocr-structured-workflow-live-50-w4/engine.db`
  - `/tmp/book-ocr-structured-workflow-live-50-w4/projections/book_ocr_structured.db`
  - `/tmp/book-ocr-structured-workflow-live-50-w4/turns.db`
- Counts:
  - page markers: `50`
  - assembled bytes: `78,976`
  - table Markdown lines: `74`
  - structured table blocks: `10`
  - structured figure blocks: `17`
  - figure blocks with captions: `17`
  - validation warnings: `0`
  - projection rows: `succeeded|50`
  - engine ops: `succeeded|53`
  - turn rows: `50`
  - turn phases: `input|200`, `final|200`

### What didn't work

- No provider failures occurred, so this run did not demonstrate automatic retry in action. It demonstrated successful parallel execution, not retry recovery.

### What I learned

- Worker concurrency 4 is viable for this 50-page live structured workflow run under current provider/profile conditions.
- The workflow-backed graph created the expected number of ops: discover + 50 page steps + assemble + validate.
- A deterministic retry test is still needed because relying on provider TLS failures is not a reliable validation strategy.

### What was tricky to build

- Nothing new was built in this step. The main operational nuance is interpreting the absence of failures: it is good news for throughput, but it does not prove retry behavior beyond the configured workflow retry policy.

### What warrants a second pair of eyes

- Review the assembled output for prose completeness, because previous structured runs were still shorter than freeform OCR.
- Review provider usage/cost before trying higher concurrency or full-book W4.

### What should be done in the future

- Add a fake retrying structured OCR client test that fails once with a retryable error and succeeds on the second attempt.
- Add structured page-status command for the `structured_pages` projection.
- Continue Phase 6 hardening.

### Code review instructions

- Review experiment summary:
  - `/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/25/BOOK-OCR-STRUCTURED-WORKFLOW-001--promote-structured-ocr-to-workflow-runtime/experiments/001-report794-50-workflow-live-w4/01-summary.md`
- Inspect live output:
  - `/tmp/book-ocr-structured-workflow-live-50-w4/assembled.md`

### Technical details

Command:

```bash
go run ./cmd/book-ocr structured-run \
  --book-id report-794-structured-workflow-live-50-w4 \
  --image-dir /home/manuel/code/wesen/claw-stuff/output/books/presentation-based-uis/pages \
  --start-page 1 \
  --end-page 50 \
  --work-dir /tmp/book-ocr-structured-workflow-live-50-w4 \
  --profile gpt-5-mini-low \
  --profile-registries /tmp/book-ocr-hq-001/profiles-clean.yaml \
  --dry-run=false \
  --expected-pages 50 \
  --max-workers 4 \
  --log-level warn
```

## Step 6: Add deterministic retry test and structured page status command

I started Phase 6 hardening by proving workflow retry deterministically and adding a projection-backed page status command. The retry test uses a fake structured OCR client that fails once with a retryable provider-style error and succeeds on the second attempt. This closes the gap left by the live W4 run, which succeeded without exercising retry behavior.

I also added `book-ocr structured-pages`, which queries the `structured_pages` projection and prints page status, warning counts, table/figure counts, rendered bytes, errors, and rendered Markdown paths. This gives operators a structured equivalent of the older `pages` command.

### Prompt Context

**User prompt (verbatim):** "ok, continue, keep a detailed diary, commit at appropriate intervals"

**Assistant interpretation:** Continue Phase 6 hardening work incrementally, recording what changes and validating each slice.

**Inferred user intent:** The user wants production hardening to proceed beyond the successful W4 run, with the same documentation and commit discipline as previous phases.

**Commit (code):** `9d42c5d75bd968b298055e2ba5d280f30eab8fd6` — "Add structured workflow retry test and page status"

### What I did

- Added `internal/ocrpipeline/workflow_retry_test.go`.
- Implemented a fake `failOnceStructuredClient` that returns:
  - `run structured OCR inference: transient test failure`
  on the first call, then delegates to `DryRunStructuredOCRClient`.
- Added an integration-style workflow runtime test that:
  - starts a structured workflow run,
  - observes workflow retry behavior,
  - verifies the workflow succeeds,
  - verifies the client was called at least twice,
  - verifies the structured projection row is `succeeded`.
- Added `book-ocr structured-pages` command.
- Verified it against the W4 live run projection:
  - `/tmp/book-ocr-structured-workflow-live-50-w4/projections/book_ocr_structured.db`
- Added Phase 6 hardening tasks for retry proof, page status, prose completeness, and figure embedding.

### Why

- The W4 live run proved parallel execution but did not prove retry because no provider failure occurred.
- A deterministic retry test is better than waiting for a random TLS failure.
- Operators need page-level status for structured OCR runs, especially as we prepare for larger live runs.

### What worked

- Targeted retry test passed:
  - `go test ./internal/ocrpipeline -run TestStructuredWorkflowRetriesTransientPageFailure -count=1 -v`
- Full tests passed:
  - `go test ./... -count=1`
- Test logs showed the expected sequence:
  - page step attempt 1 failed and was scheduled for retry,
  - page step attempt 2 succeeded,
  - assemble and validate succeeded,
  - workflow succeeded.
- `structured-pages` printed rows from the W4 run, for example:
  - `page_num=1 status=succeeded warning_count=0 rendered_markdown_path=/tmp/book-ocr-structured-workflow-live-50-w4/pages/page_001/05-rendered.md`

### What didn't work

- N/A

### What I learned

- The workflow retry policy is working for structured page OCR when the executor returns a retryable workflow error.
- A page-status command is a useful complement to workflow-level status because it exposes OCR-domain fields: warnings, table count, figure count, rendered bytes, and rendered Markdown path.

### What was tricky to build

- The retry test has to wait through the workflow retry backoff. The test loops `RunOnce` until the workflow reaches a terminal state instead of assuming retry occurs in the next immediate scheduler cycle.
- The projection file suffix is `.db`, matching `workflow.NewSQLiteProjectionStore`, not `.sqlite`.

### What warrants a second pair of eyes

- Review whether the retry test should reduce retry backoff through a test-only policy hook to avoid waiting around one second.
- Review whether `structured-pages` should sort or format columns in a more human-friendly table instead of reusing the generic key/value row printer.

### What should be done in the future

- Add suspicious-short-page/prose completeness validation.
- Connect structured figure blocks to figure image embedding.
- Consider adding `--output json/table` for `structured-pages` later.

### Code review instructions

- Review retry test:
  - `/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/ocrpipeline/workflow_retry_test.go`
- Review status command:
  - `/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/cmd/book-ocr/main.go`
- Validate with:
  - `go test ./... -count=1`
  - `go run ./cmd/book-ocr structured-pages --work-dir /tmp/book-ocr-structured-workflow-live-50-w4 --book-id report-794-structured-workflow-live-50-w4 --limit 5`

### Technical details

Structured page status command:

```bash
go run ./cmd/book-ocr structured-pages \
  --work-dir /tmp/book-ocr-structured-workflow-live-50-w4 \
  --book-id report-794-structured-workflow-live-50-w4 \
  --limit 5
```
