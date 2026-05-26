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
      Note: Workflow-backed structured-run CLI wiring and resume registration (commit 325614b)
    - Path: internal/ocrpipeline/workflow_executors.go
      Note: Discover/page/assemble/validate workflow executors (commit d4bf768)
    - Path: internal/ocrpipeline/workflow_package.go
      Note: Structured workflow package registration and retry policy (commit d4bf768)
    - Path: internal/ocrpipeline/workflow_projection.go
      Note: Structured page projection schema and helpers (commit d4bf768)
    - Path: internal/ocrpipeline/workflow_types.go
      Note: Structured workflow input/result contracts (commit d4bf768)
    - Path: ttmp/2026/05/25/BOOK-OCR-STRUCTURED-WORKFLOW-001--promote-structured-ocr-to-workflow-runtime/design-doc/01-workflow-backed-structured-ocr-design-and-implementation-guide.md
      Note: Primary workflow promotion design guide
    - Path: ttmp/2026/05/25/BOOK-OCR-STRUCTURED-WORKFLOW-001--promote-structured-ocr-to-workflow-runtime/reference/01-diary.md
      Note: Chronological implementation diary
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
