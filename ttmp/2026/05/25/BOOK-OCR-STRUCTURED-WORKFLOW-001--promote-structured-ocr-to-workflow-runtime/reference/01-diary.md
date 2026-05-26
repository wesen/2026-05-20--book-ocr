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
        min-rendered-bytes CLI flag (commit 936952d)
        embed-figures CLI flags for structured-run (commit 600dbc7)
        structured-rerun-pages operator and PDF flags (commit 52eba49)
    - Path: internal/ocrpipeline/prompts.go
      Note: Prompt rules for readable screenshot/code/table content (commit a576e96)
    - Path: internal/ocrpipeline/renderer.go
      Note: |-
        Code block rendering (commit a576e96)
        Suppress duplicate image markers for table-like figure blocks (commit 8825cd2)
    - Path: internal/ocrpipeline/structured_ocr.go
      Note: empty code/list/table validation warnings (commit 52eba49)
    - Path: internal/ocrpipeline/types.go
      Note: Structured code block type (commit a576e96)
    - Path: internal/ocrpipeline/workflow_executors.go
      Note: |-
        Discover/page/assemble/validate workflow executors (commit d4bf768)
        Short-page validation query and report generation (commit 936952d)
        Structured assembly figure embedding and artifact storage (commit 600dbc7)
        PDF rendering during structured assembly (commit 52eba49)
    - Path: internal/ocrpipeline/workflow_package.go
      Note: Structured workflow package registration and retry policy (commit d4bf768)
    - Path: internal/ocrpipeline/workflow_projection.go
      Note: Structured page projection schema and helpers (commit d4bf768)
    - Path: internal/ocrpipeline/workflow_retry_test.go
      Note: Deterministic workflow retry proof (commit 9d42c5d)
    - Path: internal/ocrpipeline/workflow_types.go
      Note: |-
        Structured workflow input/result contracts (commit d4bf768)
        Short-page validation result contracts (commit 936952d)
        Figure embedding workflow result/input fields (commit 600dbc7)
        PDF fields in structured workflow inputs/results (commit 52eba49)
    - Path: internal/ocrquality/figures.go
      Note: Avoid synthesized figure markers on rendered table pages (commit ff48bd4)
    - Path: internal/ocrquality/figures_test.go
      Note: Regression test for table pages without synthesized figures (commit ff48bd4)
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

## Step 7: Add configurable short-page completeness warnings

I added the first prose-completeness gate: a configurable `--min-rendered-bytes` validation warning for pages whose rendered Markdown is suspiciously short. This does not fail the workflow by default; it records short pages in `validation-report.json` so a reviewer can inspect likely under-transcribed pages.

This is intentionally conservative. The structured OCR output can legitimately be short for blank/front-matter pages, so the threshold is opt-in. It gives production runs a way to surface suspicious pages without hardcoding a single threshold for all books and page types.

### Prompt Context

**User prompt (verbatim):** (same as Step 6)

**Assistant interpretation:** Continue Phase 6 hardening after retry and status work.

**Inferred user intent:** The user wants production-readiness checks that catch the structured OCR's known lower prose-volume risk.

**Commit (code):** `936952db8d3fe627a1865906dc24154b40ffbffc` — "Add structured OCR short page validation"

### What I did

- Added `MinRenderedBytes` to `StructuredRunInput` and `StructuredValidateInput`.
- Added `StructuredShortPageWarning` and `ShortPages` to `StructuredValidateResult`.
- Added `--min-rendered-bytes` to `book-ocr structured-run`.
- Updated the validation executor to query `structured_pages` for successful pages with `rendered_bytes` below the threshold.
- Fixed validation report writing so the local report includes `report_path`, `report_ref_id`, and `report_uri`.
- Ran a dry-run validation smoke with a threshold of 100 bytes.

### Why

- The first-50 structured output was cleaner but shorter than freeform OCR. A production workflow needs to surface suspiciously short pages before a full-book run is accepted.

### What worked

- Full tests passed:
  - `go test ./... -count=1`
- Dry-run short-page smoke produced warnings as expected:
  - `warning_count = 3`
  - pages 1, 2, and 3 were below 100 rendered bytes in the dry-run fake output.
- Local validation report now includes:
  - `report_path`,
  - `report_ref_id`,
  - `warning_count`,
  - `short_pages`.

### What didn't work

- Initially, the local `validation-report.json` omitted `report_path` and artifact IDs because the report was marshaled before those fields were set. I fixed the order by setting `ReportPath` before the first marshal and rewriting the local report after storing the report artifact.

### What I learned

- Completeness validation needs to be configurable and page-type-aware over time. The first useful primitive is a threshold warning, not a hard failure.

### What was tricky to build

- The validation step needs data from both the assembled Markdown and the structured projection. The assembled Markdown gives page markers and adjacent captions; the projection gives per-page rendered byte counts and artifact paths.

### What warrants a second pair of eyes

- Review whether `--min-rendered-bytes` should exclude blank pages by reading each page's structured JSON page type.
- Review whether suspicious short pages should eventually fail full-book acceptance gates or remain warnings.

### What should be done in the future

- Add page-type-aware completeness checks.
- Add anchor/oracle-based checks for known risky pages.
- Add baseline comparison against the previous freeform target-only OCR output.

### Code review instructions

- Review validation types:
  - `/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/ocrpipeline/workflow_types.go`
- Review validation executor:
  - `/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/ocrpipeline/workflow_executors.go`
- Review CLI flag:
  - `/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/cmd/book-ocr/main.go`

### Technical details

Smoke command:

```bash
go run ./cmd/book-ocr structured-run \
  --book-id report-794-shortpage-dry \
  --image-dir /home/manuel/code/wesen/claw-stuff/output/books/presentation-based-uis/pages \
  --start-page 1 \
  --end-page 3 \
  --work-dir /tmp/book-ocr-structured-workflow-shortpage-dry \
  --dry-run \
  --expected-pages 3 \
  --min-rendered-bytes 100 \
  --max-workers 2 \
  --log-level warn
```

## Step 8: Connect structured figure blocks to figure image embedding

I completed the remaining Phase 6 hardening task by adding figure embedding to the workflow-backed structured assembly step. Structured runs can now pass `--embed-figures`; the assemble step writes the ordinary `assembled.md`, then uses the existing OCR quality figure extraction code to produce `embedded-figures.md`, figure crops, sidecars, debug overlays, and workflow artifacts.

This implementation deliberately reuses the existing `ocrquality.EmbedExtractedFigures` path instead of inventing a new cropper. The structured renderer already emits `[FIGURE: ...]` markers for figure blocks when no figure resolver is supplied, so the existing marker-driven extraction path can operate on structured-rendered Markdown.

### Prompt Context

**User prompt (verbatim):** "go ahead"

**Assistant interpretation:** Continue from the previous Phase 6 work and implement the remaining structured figure embedding task.

**Inferred user intent:** The user wants structured OCR output to move closer to final review quality by embedding cropped figure images, not only figure captions/placeholders.

**Commit (code):** `600dbc763264a8537796efb98dc096e467bca359` — "Embed figures in structured OCR assembly"

### What I did

- Added `EmbedFigures` and `FiguresDir` to `StructuredRunInput`.
- Added figure-embedding fields to `StructuredAssembleInput` and `StructuredAssembleResult`.
- Added CLI flags:
  - `--embed-figures`
  - `--figures-dir`
- Updated `StructuredDiscoverExecutor` to pass image dir and figure embedding options to the assemble step.
- Updated `StructuredAssembleExecutor` to:
  - write `assembled.md`,
  - optionally call `ocrquality.EmbedExtractedFigures`,
  - write `embedded-figures.md`,
  - write figure crops under `<work-dir>/figures`,
  - store embedded Markdown, figure PNGs, sidecars, and debug overlays as workflow artifacts.
- Ran a dry-run smoke over pages 12-13 with figure embedding enabled.

### Why

- Structured OCR now identifies figure blocks and captions, but final review Markdown needs image links.
- The old quality pass already has a crop/extract/embed implementation; reusing it reduces risk and keeps sidecar/debug provenance behavior consistent.

### What worked

- Full tests passed:
  - `go test ./... -count=1`
- Dry-run figure embedding smoke succeeded:
  - `/tmp/book-ocr-structured-embed-dry/embedded-figures.md`
- The embedded Markdown contains a figure image link:
  - `![Diagram showing a user interface connected to an application data base.](figures/page_013_figure_01.png)`
- The figures directory contains:
  - `page_013_figure_01.png`
  - `page_013_figure_01.json`
  - `page_013_figure_01.debug.png`
- Workflow artifacts were written for embedded Markdown and figure-related files.

### What didn't work

- An earlier ad hoc attempt to inspect the assemble result with an incorrectly quoted SQLite command failed. The workflow run itself succeeded; I switched to direct file checks for the smoke validation.

### What I learned

- The structured renderer and legacy figure embedding path already meet at a useful compatibility boundary: `[FIGURE: ...]` markers.
- A future improvement can replace marker synthesis with a direct `FigureResolver` path, but the marker-based bridge is enough to produce reviewable embedded Markdown now.

### What was tricky to build

- The assembly result needs to report both raw assembled Markdown and embedded Markdown. `assembled.md` remains the renderer-native output, while `embedded-figures.md` is the review artifact with image links.
- Workflow artifacts should include not only the embedded Markdown and figure PNG but also sidecar and debug overlay files, because those files explain crop provenance.

### What warrants a second pair of eyes

- Review whether `structured_embed_figures_failed` should be permanent or retryable. Missing source images are permanent, but filesystem/transient decode issues may be retryable.
- Review whether structured figure embedding should eventually bypass marker-based extraction and use a structured `FigureResolver` that maps `page_number + block_id` to crop paths.

### What should be done in the future

- Run a first-50 workflow-backed live structured OCR with `--embed-figures` and inspect the resulting `embedded-figures.md`.
- Add a dedicated test around structured assemble figure embedding, ideally with a small generated page image.
- Consider adding figure count and embedded path to `structured-pages` or a run-level status command.

### Code review instructions

- Review type additions:
  - `/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/ocrpipeline/workflow_types.go`
- Review assembly changes:
  - `/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/ocrpipeline/workflow_executors.go`
- Review CLI flags:
  - `/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/cmd/book-ocr/main.go`

### Technical details

Smoke command:

```bash
go run ./cmd/book-ocr structured-run \
  --book-id report-794-embed-dry \
  --image-dir /home/manuel/code/wesen/claw-stuff/output/books/presentation-based-uis/pages \
  --start-page 12 \
  --end-page 13 \
  --work-dir /tmp/book-ocr-structured-embed-dry \
  --dry-run \
  --expected-pages 2 \
  --embed-figures \
  --max-workers 2 \
  --log-level warn
```

## Step 9: Tighten figure-versus-readable-text classification

After reviewing the full-book PDF, the user found pages where readable screenshot/text/table/code-like content was treated as figures instead of being transcribed. I tightened the structured OCR contract so screenshots, boxed displays, command/code listings, quoted text windows, and visible tabular content are not automatically treated as image-only figures.

The change adds a first-class `code` block type and renderer support for fenced text blocks. It also updates the prompt to require readable text/code/table content inside screenshots to be emitted as paragraph/list/code/table blocks, with figure blocks reserved for visual content that cannot be faithfully represented as text.

### Prompt Context

**User prompt (verbatim):** "page 20, 30, 31, 83, 84, 86, 96, haven't been properly recognized as text (or text with code + tables), and instead kept as figures. This points to a prompt issue to keep things as figures when code is recognized or a table is recognized."

**Assistant interpretation:** The full-book structured run overuses figure blocks for regions that contain readable text, code/listings, or tables; adjust the structured OCR schema/prompt so the model decomposes readable content instead of hiding it behind figure images.

**Inferred user intent:** Improve the next OCR pass so final Markdown contains searchable/editable text for screenshots and boxed listings where possible, not only cropped figure images.

**Commit (code):** `a576e968f61f8ba35dace5633d0bc9564a54baf2` — "Tighten structured OCR text figure classification"

### What I did

- Added `BlockCode = "code"` to the structured OCR block types.
- Added deterministic renderer support for fenced `text` code blocks.
- Updated renderer tests to cover table plus code rendering.
- Updated the structured OCR prompt to clarify:
  - screenshots are not automatically figures,
  - readable screenshot text should be transcribed,
  - code/listings must use the `text` field and must not be empty,
  - tables/grids must become table blocks,
  - figure blocks should only cover remaining visual content.
- Ran targeted live page reruns for pages 20, 30, 31, 83, 84, 86, and 96 under `/tmp/book-ocr-structured-prompt-fix-targeted`.
- Ran `go test ./... -count=1`.

### Why

- The embedded PDF made it clear that figure extraction alone is not enough; if a screenshot contains readable text, the Markdown must include that text.
- Searchable text/code/table output is the reason for structured OCR. Figure crops are supplementary provenance/review artifacts.

### What worked

- Tests passed.
- The prompt/schema now has a place to put listing-like content (`code`).
- A targeted rerun changed page 83 from a pure figure to a heading plus code block shape, proving the model noticed the content class change.

### What didn't work

- The first targeted rerun for page 83 emitted an empty code block, so the prompt still needs stronger enforcement and perhaps validation against empty non-blank code/table/list blocks.
- Pages 86 and 96 remained figure blocks in the targeted rerun; visual inspection suggests they are largely screenshot/diagram pages, but they should be reviewed page-by-page before deciding whether to force text extraction.

### What I learned

- The problematic class is not simply "figure versus not figure". Some pages need both: a figure/caption for visual context and separate text/code/table blocks for readable content inside the screenshot.
- Empty structured blocks are a validation gap. The parser/validator should warn when a non-blank block type lacks required content.

### What was tricky to build

- The existing schema did not distinguish paragraph text from line-preserving listings. Adding a code block type is a small schema change, but it affects prompt contract, parser tolerance, renderer behavior, and tests.
- The model may comply with block type selection but still omit the content. That requires validation, not only prompt text.

### What warrants a second pair of eyes

- Review whether `code` should be rendered as plain fenced `text` or whether some pages need language-specific fences.
- Review pages 83, 86, and 96 visually against the new targeted artifacts before deciding whether further prompt changes should force screenshot text extraction.

### What should be done in the future

- Add validation warnings for empty `code`, empty `table`, empty `list`, and figure-only pages that contain OCR-detectable text density.
- Rerun the affected pages after the prompt update and patch the full-book artifact or run a full reprocess.
- Consider an explicit `screenshot_text` block or a figure-with-transcription structure if mixed screenshot pages remain hard for the model.

### Code review instructions

- Start with:
  - `/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/ocrpipeline/prompts.go`
  - `/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/ocrpipeline/types.go`
  - `/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/ocrpipeline/renderer.go`
- Validate with:
  - `go test ./... -count=1`

### Technical details

Targeted rerun directory:

```text
/tmp/book-ocr-structured-prompt-fix-targeted
```

## Step 10: Add targeted structured page rerun operator and workflow PDF rendering

I added a targeted operator for the exact workflow repair path we needed after manual PDF review. Instead of rerunning the whole 202-page book, the new command requeues selected successful page OCR steps, requeues downstream assembly and validation, optionally enables PDF rendering in the assembly step, and resumes the workflow with the current prompt/schema code.

I also added PDF rendering to structured workflow assembly itself. When `--render-pdf` is enabled, the assembly step renders the assembled Markdown (or embedded-figures Markdown when figures are enabled) through `pandoc`/`xelatex`, writes `book.pdf`, and stores it as a workflow artifact.

### Prompt Context

**User prompt (verbatim):** "go ahead. also show me the pdf when done, so i can o manual validation."

**Assistant interpretation:** Continue by adding validation/rerun support, regenerate the affected pages, produce a PDF, and open it for manual validation.

**Inferred user intent:** The user wants a tight review loop: fix only the bad pages in the nearly-good full-book artifact and immediately inspect the resulting PDF.

**Commit (code):** `52eba495c26ebd1fd8e8c50cca4028867ccd49be` — "Render structured OCR PDF in workflow"

### Prompt Context

**User prompt (verbatim):** "in fact can you add pdf rendering to the workflow?

don't forget to keep a diary as you work (using the well known recommended format, with things like "What worked / what didn't work""

**Assistant interpretation:** Add PDF generation as part of the workflow rather than only as an ad hoc pandoc command, and keep the diary updated in the established format.

**Inferred user intent:** Make PDF rendering reproducible and durable inside workflow artifacts while preserving a clear implementation record.

### Prompt Context

**User prompt (verbatim):** "hold on, we only want to rerun the broken pages, in the previous run, is that something the workflow would allow us to do?"

**Assistant interpretation:** Stop the full rerun and determine whether the existing workflow can selectively reprocess only selected succeeded pages.

**Inferred user intent:** Avoid unnecessary provider calls and keep the mostly-good full-book run as the base artifact.

### Prompt Context

**User prompt (verbatim):** "add the targeted operator, then rerun those pages in our previous almost perfect pdf"

**Assistant interpretation:** Implement the missing targeted rerun operator and apply it to the previous full-book run for the manually identified pages.

**Inferred user intent:** Patch the previous full-book artifact in place by reprocessing only pages 20, 30, 31, 83, 84, 86, and 96 with the improved prompt/schema, then regenerate Markdown and PDF.

### What I did

- Added structured workflow PDF fields:
  - `RenderPDF`
  - `PDFPath`
  - `PandocPath`
- Added assemble result PDF fields:
  - `PDFPath`
  - `PDFRefID`
  - `PDFURI`
- Added `--render-pdf`, `--pdf-path`, and `--pandoc-path` to `structured-run`.
- Added `renderStructuredPDF`, which invokes `pandoc --pdf-engine=xelatex` with the same font/margin settings used in the earlier ad hoc PDF render.
- Added `book-ocr structured-rerun-pages` with flags:
  - `--work-dir`
  - `--run-id`
  - `--pages`
  - `--render-pdf`
  - `--pdf-path`
  - `--pandoc-path`
  - `--max-workers`
- The operator requeues selected `structured-page-NNN` ops, requeues `assemble-structured-markdown` and `validate-structured-run`, marks the workflow running again, optionally patches the assemble op input to render PDF, and resumes workers.
- Added empty block validation for empty code/list/table blocks in `ValidateStructuredPage`.
- Stopped the accidental full-book rerun path and used the targeted operator on the previous full-book run.
- Backed up the previous engine DB before targeted requeue:
  - `/tmp/book-ocr-structured-workflow-full-live-w4-figures/engine.db.before-targeted-rerun-pages.bak`
- Reprocessed pages:
  - `20,30,31,83,84,86,96`
- Opened the regenerated PDF in Okular.

### Why

- Manual review found only a small set of bad pages. Rerunning all 202 pages would waste provider calls and potentially introduce new regressions.
- PDF rendering belongs in the workflow because the PDF is part of the review artifact set and should be produced from the same assembly inputs as Markdown.

### What worked

- Full tests passed:
  - `go test ./... -count=1`
- The targeted rerun operator successfully reprocessed the selected pages and reassembled the full book.
- The previous full-book run ended in `succeeded` again.
- The regenerated PDF exists and opened in Okular:
  - `/tmp/book-ocr-structured-workflow-full-live-w4-figures/book.pdf`
- Page 83 improved substantially: it now has a heading plus a populated `code` block with the visible Zmacs scroll-bar text.
- Page 96 changed away from a pure figure to heading/paragraph text.

### What didn't work

- I initially started a fresh full-book rerun after adding PDF rendering, but the user correctly stopped that path because only selected pages needed reprocessing.
- The current targeted operator directly updates workflow SQLite state. It works for this local workflow, but it should eventually become a first-class runtime/operator API rather than CLI-local SQL.
- Pages 31 and 86 still classify as figure blocks after targeted rerun; based on visual review, page 31 is a model diagram and page 86 is mostly a screenshot, but they remain review targets in the PDF.

### What I learned

- The workflow graph can support targeted repair as long as selected page ops and downstream assembly/validation are reset consistently.
- The operator needs to patch existing assemble input when a feature such as PDF rendering is added after the original run was created.
- Manual review through PDF remains essential: structured block counts alone cannot decide whether screenshot text is sufficiently transcribed.

### What was tricky to build

- Reprocessing succeeded pages is different from retrying failed pages. The existing `retry` operator intentionally only handles failed ops, so the new operator had to reset succeeded page ops to `ready` and make the workflow `running` again.
- Downstream assemble/validate ops also need to be reset, otherwise the page outputs change but the final Markdown/PDF stays stale.
- Existing runs may have old assemble input JSON without `render_pdf`; the operator patches that input only when requested.

### What warrants a second pair of eyes

- Review the direct SQL state transition in `structured-rerun-pages`; it should be promoted into the workflow runtime as a supported reset/reprocess operation.
- Review whether downstream artifacts should be garbage-collected when reassembling, because artifact IDs are replaced for deterministic IDs but stale extra artifacts may remain.
- Review the regenerated PDF pages 31 and 86 to decide whether they need stronger screenshot-text extraction or are acceptable figure-only pages.

### What should be done in the future

- Add a runtime-level `ResetStep`/`ReprocessStep` operator instead of direct CLI SQL.
- Add a `--reassemble-only` mode for cases where page artifacts are edited manually.
- Add validation warnings for figure-only pages with suspicious OCR text density if an external OCR/text detector is available.

### Code review instructions

- Start with:
  - `/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/cmd/book-ocr/main.go`
  - `/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/ocrpipeline/workflow_types.go`
  - `/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/ocrpipeline/workflow_executors.go`
  - `/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/ocrpipeline/structured_ocr.go`
- Validate with:
  - `go test ./... -count=1`
- Manual validation PDF:
  - `/tmp/book-ocr-structured-workflow-full-live-w4-figures/book.pdf`

### Technical details

Targeted operator command:

```bash
book-ocr structured-rerun-pages \
  --work-dir /tmp/book-ocr-structured-workflow-full-live-w4-figures \
  --run-id book-ocr/structured-499f1718-bfb6-4135-a52f-56d35001d0bd \
  --pages 20,30,31,83,84,86,96 \
  --render-pdf \
  --max-workers 2 \
  --log-level warn
```

Regenerated artifacts:

```text
/tmp/book-ocr-structured-workflow-full-live-w4-figures/assembled.md
/tmp/book-ocr-structured-workflow-full-live-w4-figures/embedded-figures.md
/tmp/book-ocr-structured-workflow-full-live-w4-figures/book.pdf
/tmp/book-ocr-structured-workflow-full-live-w4-figures/validation-report.json
```

## Step 11: Remove duplicate figure images for table-like spreadsheet figures

Manual PDF review showed that source pages containing PPSCalc spreadsheet tables were still producing image-only figure floats in the PDF even though the OCR had already recognized the table contents. The issue was not only the structured OCR model output: page 48 had proper table blocks, but the figure embedding pass synthesized image markers from figure captions and then cropped the source page again.

I fixed both sides of the boundary. The structured renderer now suppresses figure image markers when a table-like figure is immediately followed by a structured table transcription, and the legacy figure embedding pass no longer synthesizes missing figure markers on pages that already contain Markdown tables.

### Prompt Context

**User prompt (verbatim):** "i swear, in the pdf it's still the figure. figure 10 and figure 11 on pages 30 and 31 should be text with tables."

**Assistant interpretation:** The PDF still contains image floats for spreadsheet figures that have already been recognized as text/table content, so the final rendering/embedding path is duplicating table figures as images.

**Inferred user intent:** The review PDF should show searchable tables for spreadsheet-like figures and should not include duplicate cropped page images for those same figures.

**Commit (code):** `8825cd2c2dce21f7567c4e3df7756902e223aca2` — "Suppress duplicate images for table figures"

**Commit (code):** `ff48bd4` — "Avoid synthesizing figures for rendered tables"

### Prompt Context

**User prompt (verbatim):** "/tmp/pi-clipboard-6571b388-4245-41f3-bce9-2c2a136fa90e.png Here's page 31 and here's page 32 which is well scanned /tmp/pi-clipboard-ff876c4d-c52a-458c-9c08-0024dda54d53.png (you can look with read tool)"

**Assistant interpretation:** Compare the user's screenshots with the regenerated PDF and source markdown to verify that the bad pages correspond to table-like spreadsheet figures still being embedded as images.

**Inferred user intent:** Validate visually that the specific PDF pages now render as text/tables rather than as large scanned/cropped images.

### What I did

- Read the two clipboard screenshots with the image-capable `read` tool.
- Confirmed the screenshot labeled page 31 corresponded to source page 48 in the generated Markdown/PDF, not source page 31.
- Found remaining spreadsheet image refs in `embedded-figures.md`:
  - page 34 Figure 2-7
  - page 48 Figure 2-12
  - page 48 Figure 2-13
- Added renderer logic to suppress image markers for table-like figure blocks when the next block is a structured table.
- Updated the structured prompt so primarily table/grid/spreadsheet figures should be emitted as caption text plus table blocks, not as figure blocks.
- Added a renderer unit test for suppressing a PPSCalc figure image when a table follows.
- Fixed `ocrquality.synthesizeMissingFigureMarkers` so it does not synthesize `[FIGURE: ...]` markers on pages that already contain Markdown tables.
- Added a unit test preventing marker synthesis for rendered table pages.
- Reprocessed source pages 34 and 48 with `structured-rerun-pages`, reassembled Markdown, re-rendered PDF, and reopened it in Okular.

### Why

- The final review PDF should be optimized for readable/searchable OCR text and tables.
- Figure crops are useful for diagrams, but not when they duplicate a table that is already represented as Markdown.

### What worked

- Tests passed:
  - `go test ./... -count=1`
- After the second fix, `embedded-figures.md` no longer contains spreadsheet/PPSCalc image refs:
  - `rg '^!\\[.*(Spreadsheet|spreadsheet|columns A B C|PPSCalc)' embedded-figures.md` returns no matches.
- Page 48 now renders as prose plus Markdown tables and captions, without spreadsheet image links.
- The regenerated PDF was opened in Okular:
  - `/tmp/book-ocr-structured-workflow-full-live-w4-figures/book.pdf`

### What didn't work

- The first renderer-only fix did not remove the duplicate images from the final `embedded-figures.md` because the legacy figure embedding pass synthesized new `[FIGURE: ...]` markers from captions on table-heavy pages.
- The PDF physical page numbers shifted after the duplicate image floats were removed, so validating by Okular page number alone is unreliable.

### What I learned

- There are two independent sources of image links: explicit renderer `[FIGURE: ...]` markers and synthesized figure markers in `ocrquality`.
- A page can have correct structured JSON and correct per-page rendered Markdown, then still become wrong in `embedded-figures.md` if the embedding pass synthesizes markers too aggressively.

### What was tricky to build

- The PDF issue appeared as a model/OCR issue, but the actual remaining failure was post-processing: synthesis treated captions plus table text as a diagram-like page.
- The right fix is conservative: if Markdown tables are present, do not synthesize missing full-page figure markers for that page.

### What warrants a second pair of eyes

- Review whether the `shouldSuppressFigureImageForFollowingTable` cues are too narrow or too broad.
- Review whether `containsMarkdownTable` should become a more general page-classifier helper used by other figure-synthesis heuristics.

### What should be done in the future

- Add a report section listing image links whose alt text contains table/spreadsheet/grid terms; those are review suspects.
- Consider visible source page markers in the PDF to make manual validation less confusing.

### Code review instructions

- Review:
  - `/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/ocrpipeline/renderer.go`
  - `/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/ocrquality/figures.go`
- Validate with:
  - `go test ./... -count=1`
  - `rg '^!\\[.*(Spreadsheet|spreadsheet|columns A B C|PPSCalc)' /tmp/book-ocr-structured-workflow-full-live-w4-figures/embedded-figures.md`

### Technical details

Regenerated artifact:

```text
/tmp/book-ocr-structured-workflow-full-live-w4-figures/book.pdf
```

Relevant backup before rerun:

```text
/tmp/book-ocr-structured-workflow-full-live-w4-figures/engine.db.before-rerun-table-synthesis-fix.bak
```
