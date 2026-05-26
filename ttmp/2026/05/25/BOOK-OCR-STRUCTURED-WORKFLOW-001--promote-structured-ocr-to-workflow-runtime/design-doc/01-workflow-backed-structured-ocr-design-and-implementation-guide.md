---
Title: Workflow Backed Structured OCR Design and Implementation Guide
Ticket: ""
Status: active
Topics:
    - book-processing
    - ocr
    - workflow
    - geppetto
    - pinocchio
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../../scraper/pkg/workflow/context.go
      Note: Workflow StepContext artifact/projection/result APIs
    - Path: ../../../../../../../scraper/pkg/workflow/operators.go
      Note: Retry/cancel operator APIs
    - Path: ../../../../../../../scraper/pkg/workflow/runtime.go
      Note: Runtime StartRun/RunOnce/projection APIs
    - Path: cmd/book-ocr/main.go
      Note: Current structured-run direct CLI loop and runtime wiring reference
    - Path: internal/ocrmvp/discover.go
      Note: Freeform discover executor and dynamic page step emission template
    - Path: internal/ocrmvp/executors.go
      Note: Freeform page and assemble executor template
    - Path: internal/ocrmvp/package.go
      Note: Freeform OCR workflow package registration and retry policy template
    - Path: internal/ocrmvp/projection.go
      Note: Freeform page projection schema template
    - Path: internal/ocrpipeline/client.go
      Note: Structured OCR client and Geppetto profile resolution
    - Path: internal/ocrpipeline/structured_ocr.go
      Note: Reusable structured page orchestration boundary
    - Path: internal/ocrpipeline/types.go
      Note: Structured OCR contracts and parser repairs
ExternalSources: []
Summary: ""
LastUpdated: 0001-01-01T00:00:00Z
WhatFor: ""
WhenToUse: ""
---


# Workflow Backed Structured OCR Design and Implementation Guide

## Executive Summary

The Book OCR project now has a structured OCR page pipeline that can transcribe one target page image into structured JSON, render deterministic Markdown, persist Geppetto turns through the Pinocchio turn store, and assemble a first-50-page live run. The structured pipeline has proven the two most important quality properties for the Report 794 book: page 32 tables can become real Markdown tables, and neighboring prose pages do not receive false figure blocks when the primary OCR call sees only the target page image.

The remaining operational problem is that `book-ocr structured-run` is currently a direct CLI loop. It exits on the first page error, and its current resume behavior is a pragmatic artifact-level skip: if `pages/page_NNN/04-structured.json`, `05-rendered.md`, and `06-validation.json` already exist, the CLI reuses that page. This is useful, but it duplicates capabilities the `scraper/pkg/workflow` runtime already provides: durable runs, per-step leases, retry policies, artifacts, projections, status, retry, cancel, and resume.

This ticket promotes structured OCR into a workflow-backed package inside `book-ocr`. The correct end state is:

```text
book-ocr structured-run
  -> starts a durable workflow run
  -> discover pages
  -> emit one structured OCR page step per page
  -> each page step retries transient provider/network errors
  -> each page step writes structured artifacts and page projection rows
  -> assemble waits for all page steps
  -> validation summarizes warnings and acceptance gates
  -> existing status/retry/resume/cancel commands can inspect and operate on the run
```

This document is written for a new intern who needs to understand the current system, the workflow runtime, the structured OCR code, and the exact implementation path.

## Problem Statement

### Current user-visible behavior

The user asked whether `structured-run` exits on errors instead of using workflow retry. The answer is yes: the new structured OCR path is currently a sequential CLI runner. It has `--resume=true`, but that is not durable workflow retry.

Observed first-50 live structured run behavior:

- A page 6 parse-shape error stopped the run until the parser accepted string list items.
- Transient provider TLS failures stopped the run at pages 20 and 42:

```text
Post "https://api.openai.com/v1/responses": remote error: tls: bad record MAC
```

- Rerunning the same command with `--resume=true` reused existing page artifacts and continued.

That was sufficient for Phase 5 validation, but it is not sufficient for full-book production OCR.

### What must change

The structured OCR runner must stop being a one-process page loop. It should become a workflow package that uses the same runtime capabilities as the older freeform OCR workflow.

The workflow-backed implementation must provide:

- durable run state in the workflow engine DB,
- page-level retry policies,
- retryable vs permanent error classification,
- status inspection,
- manual page retry,
- run resume after process exit,
- assembled Markdown as a step result/artifact,
- structured page projection rows,
- validation summary projection/report rows,
- compatibility with existing `book-ocr status`, `retry`, `resume`, and `cancel` commands or clearly named structured equivalents.

## Scope

### In scope

- Add a new workflow package for structured OCR, probably under:

```text
internal/ocrpipeline/package.go
internal/ocrpipeline/workflow_types.go
internal/ocrpipeline/workflow_executors.go
internal/ocrpipeline/workflow_projection.go
```

- Reuse current page-level structured OCR functions:

```text
internal/ocrpipeline/RunStructuredPage
internal/ocrpipeline/StructuredOCRClient
internal/ocrpipeline/DryRunStructuredOCRClient
internal/ocrpipeline/GeppettoStructuredOCRClient
```

- Wire `book-ocr structured-run` to the workflow runtime.
- Add queue configuration for structured OCR.
- Add projection tables for structured page status and validation warnings.
- Keep the old direct `structured-page` command for debugging.
- Preserve direct artifact files under `<work-dir>/pages/page_NNN/` for local review.
- Store workflow artifacts for operator discovery and stable result references.

### Out of scope for this ticket

- Full 202-page production rerun.
- Figure crop/image embedding into final structured Markdown.
- VLM-backed figure QA.
- Text-only normalization pass.
- Major workflow runtime changes inside `scraper` unless a missing API blocks clean integration.

Those belong to the broader Phase 6 hardening work after workflow promotion.

## Current-State Architecture

### Repositories and boundaries

The relevant repositories are:

```text
/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr
/home/manuel/workspaces/2026-05-20/book-ocr/scraper
```

The Book OCR repository imports the workflow runtime from scraper. The scraper repository must remain a generic workflow runtime and must not import OCR code.

### Current structured OCR CLI

The current `structured-run` command lives in:

```text
cmd/book-ocr/main.go
```

Evidence:

```text
cmd/book-ocr/main.go:279 defines runStructuredRun.
cmd/book-ocr/main.go:280-293 parses flags such as --book-id, --image-dir, --start-page, --end-page, --work-dir, --profile, --dry-run, and --resume.
cmd/book-ocr/main.go:318-351 loops over discovered pages and calls ocrpipeline.RunStructuredPage directly.
cmd/book-ocr/main.go:374-397 implements existingStructuredPageResult, the artifact-level resume helper.
```

The direct loop currently does this:

```text
DiscoverPageImages
for each page:
  maybe reuse page artifacts
  else RunStructuredPage
  read rendered.md
  append to assembled string
write assembled.md
write validation-report.json
```

That is easy to understand but not durable. If a provider call fails, the command exits. The next invocation can skip completed artifacts, but the workflow engine has no idea which page failed, why it failed, or whether it should retry automatically.

### Current structured page pipeline

The structured page pipeline lives in:

```text
internal/ocrpipeline/client.go
internal/ocrpipeline/prompts.go
internal/ocrpipeline/structured_ocr.go
internal/ocrpipeline/types.go
internal/ocrpipeline/renderer.go
internal/ocrpipeline/session.go
```

Key contracts:

```text
internal/ocrpipeline/client.go:14-25 defines StructuredOCRInput.
internal/ocrpipeline/client.go:27-32 defines StructuredOCRResult.
internal/ocrpipeline/client.go:34-35 defines StructuredOCRClient.
internal/ocrpipeline/client.go:38-44 defines GeppettoStructuredOCRClient.
internal/ocrpipeline/client.go:44-83 resolves a Pinocchio profile, runs Geppetto inference, and returns raw response plus turns.
internal/ocrpipeline/structured_ocr.go:209 defines RunStructuredPage.
internal/ocrpipeline/structured_ocr.go:216 reads the target page image.
internal/ocrpipeline/structured_ocr.go:223 opens the OCR turn store.
internal/ocrpipeline/structured_ocr.go:241-262 persists input/final turns and writes turn/raw artifacts.
internal/ocrpipeline/structured_ocr.go:267 parses the raw structured response.
internal/ocrpipeline/structured_ocr.go:278 renders deterministic Markdown.
internal/ocrpipeline/structured_ocr.go:279 computes validation.
```

The page-level function is reusable. The workflow package should not duplicate OCR logic. It should call `RunStructuredPage` inside a workflow executor and then publish the result into workflow artifacts/projections.

### Current older freeform workflow package

The older freeform OCR workflow is the implementation template. It lives in:

```text
internal/ocrmvp/
```

Important files:

```text
internal/ocrmvp/types.go
internal/ocrmvp/package.go
internal/ocrmvp/discover.go
internal/ocrmvp/executors.go
internal/ocrmvp/projection.go
```

Evidence:

```text
internal/ocrmvp/types.go:7-14 defines package name, step kinds, and queues.
internal/ocrmvp/types.go:17-28 defines RunInput.
internal/ocrmvp/types.go:49-59 defines PageOCRInput.
internal/ocrmvp/types.go:61-67 defines PageOCRResult.
internal/ocrmvp/package.go:18-39 registers the package and executors.
internal/ocrmvp/package.go:95-103 defines a retry policy with 3 attempts and exponential backoff.
internal/ocrmvp/discover.go:54-110 discovers pages, seeds projection rows, emits page OCR steps, and emits assemble-markdown with dependencies.
internal/ocrmvp/executors.go:13-63 executes one OCR page step and records page state.
internal/ocrmvp/executors.go:66-103 assembles dependency page results into final Markdown.
internal/ocrmvp/projection.go:10-28 creates the pages projection table.
```

The structured workflow should use the same shape, but with structured-specific inputs/results and artifacts.

### Workflow runtime capabilities

The workflow runtime lives in scraper:

```text
/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/workflow
```

Relevant runtime APIs:

```text
pkg/workflow/package.go:12-16 defines Package.
pkg/workflow/package.go:22-38 defines NewPackage and builder methods.
pkg/workflow/executor.go:70-73 defines NewTypedExecutor.
pkg/workflow/context.go:18-29 defines StepContext.
pkg/workflow/context.go:89-97 defines StepContext.Result.
pkg/workflow/context.go:150-178 defines StepContext.Artifact.
pkg/workflow/context.go:180-183 defines StepContext.Projection.
pkg/workflow/context.go:190-230 defines StepContext.StoreArtifact for external artifacts.
pkg/workflow/runtime.go:196-249 defines Runtime.StartRun.
pkg/workflow/runtime.go:251-257 defines Runtime.RunOnce.
pkg/workflow/runtime.go:313-319 defines Runtime.Projection.
pkg/workflow/operators.go:17-24 defines RetryStep.
pkg/workflow/operators.go:28-35 defines CancelRun.
```

The runtime already provides what `structured-run` needs. The implementation task is integration, not invention.

## Gap Analysis

### Gap 1: Direct loop instead of durable step graph

Current direct loop:

```go
for _, page := range pages {
    result, err := ocrpipeline.RunStructuredPage(...)
    if err != nil {
        return err
    }
}
```

Desired workflow graph:

```text
start run
  discover-structured-pages
    emits structured-page-001
    emits structured-page-002
    ...
    emits structured-page-050
    emits assemble-structured-markdown depending on all page steps
    emits validate-structured-run depending on assemble
```

### Gap 2: Artifact-level resume is not retry policy

Current resume is:

```text
if page_NNN/04-structured.json and 05-rendered.md and 06-validation.json exist:
    skip page
```

This does not classify failures or schedule automatic retry. It also cannot answer operator questions such as:

```text
Which pages failed?
Which error code occurred?
How many attempts were made?
Can I retry only page 042?
```

Workflow retry solves that at the step level.

### Gap 3: No structured page projection

The old freeform workflow has a `pages` projection table. The structured runner only has files. A structured projection should include status and artifact paths for each page.

Minimum projection table:

```sql
CREATE TABLE IF NOT EXISTS structured_pages (
  book_id TEXT NOT NULL,
  page_num INTEGER NOT NULL,
  image_path TEXT NOT NULL,
  status TEXT NOT NULL,
  step_id TEXT,
  page_dir TEXT,
  structured_json_path TEXT,
  rendered_markdown_path TEXT,
  validation_json_path TEXT,
  raw_response_path TEXT,
  rendered_artifact_id TEXT,
  structured_artifact_id TEXT,
  validation_artifact_id TEXT,
  warning_count INTEGER NOT NULL DEFAULT 0,
  table_count INTEGER NOT NULL DEFAULT 0,
  figure_count INTEGER NOT NULL DEFAULT 0,
  error_code TEXT,
  error_message TEXT,
  updated_at TEXT NOT NULL,
  PRIMARY KEY(book_id, page_num)
);
```

### Gap 4: No workflow-backed validation step

Current validation is per page and stored in per-page JSON. The assembled run should have an aggregate validation step that can fail or warn based on acceptance gates.

Initial gates:

- expected page count,
- no parse failures,
- no page-level warnings unless allowed,
- no adjacent duplicate figure captions,
- selected risky pages match expected figure-boundary behavior,
- input turn snapshots contain exactly one target image for sampled pages.

## Proposed Architecture

### Package layout

Add the structured workflow implementation to the existing `ocrpipeline` package:

```text
internal/ocrpipeline/workflow_types.go
internal/ocrpipeline/workflow_package.go
internal/ocrpipeline/workflow_discover.go
internal/ocrpipeline/workflow_executors.go
internal/ocrpipeline/workflow_projection.go
internal/ocrpipeline/workflow_validation.go
internal/ocrpipeline/workflow_test.go
```

Why use `internal/ocrpipeline` instead of a new package?

- The package already owns structured OCR contracts and orchestration.
- The workflow executors can reuse unexported helpers if needed.
- The CLI can register one package without cross-package glue.

If files grow too large, a later cleanup can split to `internal/structuredworkflow`, but that is not necessary for the first implementation.

### Workflow package constants

Proposed constants:

```go
const (
    StructuredPackageName = "book-ocr/structured"

    KindStructuredDiscover = "book-ocr/structured/discover-pages"
    KindStructuredPage     = "book-ocr/structured/ocr-page"
    KindStructuredAssemble = "book-ocr/structured/assemble-markdown"
    KindStructuredValidate = "book-ocr/structured/validate-run"

    QueueStructuredControl  = "structured-control"
    QueueStructuredVision   = "structured-vision"
    QueueStructuredAssemble = "structured-assemble"
    QueueStructuredValidate = "structured-validate"

    StructuredProjectionName = "book_ocr_structured"
)
```

### Workflow input and result types

Proposed run input:

```go
type StructuredRunInput struct {
    BookID            string   `json:"book_id"`
    ImageDir          string   `json:"image_dir"`
    PageGlob          string   `json:"page_glob,omitempty"`
    StartPage         int      `json:"start_page,omitempty"`
    EndPage           int      `json:"end_page,omitempty"`
    WorkDir           string   `json:"work_dir"`
    RunID             string   `json:"run_id,omitempty"`
    Profile           string   `json:"profile,omitempty"`
    ProfileRegistries []string `json:"profile_registries,omitempty"`
    DryRun            bool     `json:"dry_run,omitempty"`
    ExpectedPages     int      `json:"expected_pages,omitempty"`
}
```

Proposed page input:

```go
type StructuredPageInput struct {
    BookID            string   `json:"book_id"`
    RunID             string   `json:"run_id"`
    PageNumber        int      `json:"page_number"`
    ImagePath         string   `json:"image_path"`
    WorkDir           string   `json:"work_dir"`
    TurnsDB           string   `json:"turns_db,omitempty"`
    Profile           string   `json:"profile,omitempty"`
    ProfileRegistries []string `json:"profile_registries,omitempty"`
    DryRun            bool     `json:"dry_run,omitempty"`
}
```

Proposed page result:

```go
type StructuredPageWorkflowResult struct {
    BookID             string `json:"book_id"`
    PageNumber         int    `json:"page_number"`
    PageDir            string `json:"page_dir"`
    RawResponsePath    string `json:"raw_response_path"`
    StructuredJSONPath string `json:"structured_json_path"`
    RenderedMDPath     string `json:"rendered_markdown_path"`
    ValidationJSONPath string `json:"validation_json_path"`
    RenderedArtifactID string `json:"rendered_artifact_id,omitempty"`
    WarningCount       int    `json:"warning_count"`
    TableCount         int    `json:"table_count"`
    FigureCount        int    `json:"figure_count"`
}
```

### Workflow graph

Diagram:

```text
StructuredRunInput
       |
       v
+-----------------------------+
| discover-structured-pages   |
| - validate input            |
| - discover page PNGs        |
| - seed projection rows      |
| - emit page steps           |
+-----------------------------+
       |
       +----------+-----------+-------------------+
                  |                               |
                  v                               v
        +-------------------+           +-------------------+
        | structured-page-1 |    ...    | structured-page-N |
        | - RunStructured   |           | - RunStructured   |
        | - write artifacts |           | - write artifacts |
        | - update project. |           | - update project. |
        +-------------------+           +-------------------+
                  |                               |
                  +---------------+---------------+
                                  v
                    +-----------------------------+
                    | assemble-structured-markdown|
                    | - load dependency results   |
                    | - read rendered.md          |
                    | - write assembled.md        |
                    +-----------------------------+
                                  |
                                  v
                    +-----------------------------+
                    | validate-structured-run     |
                    | - page count                |
                    | - warning summary           |
                    | - adjacent captions         |
                    | - risky page checks         |
                    +-----------------------------+
```

### Error classification

The page executor should classify errors carefully.

Permanent errors:

- missing image file,
- invalid run input,
- invalid page number,
- impossible work dir path,
- parse/schema errors after raw response is saved if no retry is expected to help.

Retryable errors:

- provider network errors,
- TLS failures,
- rate limits,
- transient profile/engine failures,
- projection/artifact store availability issues.

Initial implementation can use helper functions:

```go
func classifyStructuredPageError(err error) error {
    msg := err.Error()
    switch {
    case strings.Contains(msg, "tls: bad record MAC"):
        return workflow.Retryable("structured_provider_tls", err)
    case strings.Contains(msg, "429") || strings.Contains(msg, "rate limit"):
        return workflow.Retryable("structured_provider_rate_limit", err)
    case strings.Contains(msg, "run structured OCR inference"):
        return workflow.Retryable("structured_provider_failed", err)
    case strings.Contains(msg, "parse structured OCR JSON"):
        return workflow.Permanent("structured_parse_failed", err)
    case strings.Contains(msg, "read target image"):
        return workflow.Permanent("structured_image_read_failed", err)
    default:
        return workflow.Retryable("structured_page_failed", err)
    }
}
```

This is intentionally conservative. If a parse failure is caused by model variability, retrying could help, but it may also waste calls. The first implementation should treat parse failures as permanent after saving raw artifacts. Later work can add `--retry-parse-failures` or provider-native JSON schema.

### Retry policy

Use a retry policy similar to the freeform OCR workflow:

```go
func structuredOCRRetryPolicy() model.RetryPolicy {
    return model.RetryPolicy{
        MaxAttempts:    3,
        BackoffKind:    model.BackoffKindExponential,
        InitialBackoff: time.Second,
        MaxBackoff:     30 * time.Second,
        Multiplier:     2,
    }
}
```

Freeform evidence:

```text
internal/ocrmvp/package.go:95-103 defines the existing OCR retry policy.
```

### Projection updates

The projection functions should mirror old `ocrmvp/projection.go` but include structured-specific fields.

Pseudocode:

```go
func EnsureStructuredProjectionSchema(ctx context.Context, projection workflow.Projection) error {
    projection.Exec(ctx, `CREATE TABLE IF NOT EXISTS structured_pages (...)`)
    projection.Exec(ctx, `CREATE TABLE IF NOT EXISTS structured_warnings (...)`)
    projection.Exec(ctx, `CREATE INDEX IF NOT EXISTS ...`)
}

func markStructuredPagePending(ctx, projection, input, stepID) error
func markStructuredPageRunning(ctx, projection, input) error
func markStructuredPageSucceeded(ctx, projection, input, result) error
func markStructuredPageError(ctx, projection, input, code, err) error
```

Projection statuses:

```text
pending
running
succeeded
failed
```

### Artifacts

The page executor should keep both local files and workflow artifacts.

Local files are useful for debugging:

```text
<work-dir>/pages/page_NNN/01-turn-input.yaml
<work-dir>/pages/page_NNN/02-turn-final.yaml
<work-dir>/pages/page_NNN/03-raw-response.json
<work-dir>/pages/page_NNN/04-structured.json
<work-dir>/pages/page_NNN/05-rendered.md
<work-dir>/pages/page_NNN/06-validation.json
```

Workflow artifacts are useful for operator tooling:

```go
structuredID, _ := step.StoreArtifact(
    fmt.Sprintf("page_%03d_structured.json", input.PageNumber),
    "application/json",
    structuredJSON,
    workflow.ArtifactKind("structured-ocr-json"),
)
renderedID, _ := step.StoreArtifact(
    fmt.Sprintf("page_%03d_rendered.md", input.PageNumber),
    "text/markdown",
    renderedMarkdown,
    workflow.ArtifactKind("structured-rendered-markdown"),
)
validationID, _ := step.StoreArtifact(
    fmt.Sprintf("page_%03d_validation.json", input.PageNumber),
    "application/json",
    validationJSON,
    workflow.ArtifactKind("structured-validation"),
)
```

### CLI behavior

`structured-run` should become workflow-backed by default.

Recommended flags:

```text
--book-id
--image-dir
--page-glob
--start-page
--end-page
--work-dir
--profile
--profile-registries
--dry-run
--max-workers
--poll-interval
--run-id
--log-level
--direct=false     # optional escape hatch for old direct loop during transition
```

The command should print:

```text
started structured run <run-id> in <work-dir>
status=running processed=... succeeded=... failed=... retried=...
...
assembled result: { ... }
```

Operator commands can be reused if they open the same runtime and register the structured package/executors:

```bash
book-ocr status --work-dir DIR --run-id RUN_ID
book-ocr retry --work-dir DIR --run-id RUN_ID --step-id structured-page-042
book-ocr resume --work-dir DIR --run-id RUN_ID
book-ocr cancel --work-dir DIR --run-id RUN_ID
```

The current `resume` command registers only the freeform OCR package. It should register both freeform and structured packages so it can resume either kind of run.

## Implementation Plan

### Phase 1: Add workflow types and package registration

Files:

```text
internal/ocrpipeline/workflow_types.go
internal/ocrpipeline/workflow_package.go
```

Tasks:

- Define structured package constants.
- Define `StructuredRunInput`, `StructuredPageInput`, `StructuredDiscoverResult`, `StructuredAssembleInput`, `StructuredAssembleResult`, and `StructuredValidateResult`.
- Add `StructuredWorkflowConfig` with client and projection name.
- Add `StructuredPackage()` and `RegisterStructuredWorkflow(rt, cfg)`.

Acceptance:

```bash
go test ./internal/ocrpipeline -count=1
```

### Phase 2: Add structured projection schema

Files:

```text
internal/ocrpipeline/workflow_projection.go
internal/ocrpipeline/workflow_projection_test.go
```

Tasks:

- Create `structured_pages` table.
- Create `structured_warnings` table.
- Add helpers for pending/running/succeeded/failed.
- Add tests using `workflow.NewSQLiteProjectionStore`.

Acceptance:

```bash
go test ./internal/ocrpipeline -run StructuredProjection -count=1
```

### Phase 3: Add workflow executors

Files:

```text
internal/ocrpipeline/workflow_discover.go
internal/ocrpipeline/workflow_executors.go
```

Tasks:

- Discover pages with `ocrmvp.DiscoverPageImages` or move shared discovery into a neutral package later.
- Emit one page step per page.
- Page executor calls `RunStructuredPage`.
- Page executor stores workflow artifacts and updates projection.
- Assemble executor reads dependency data and writes `assembled.md`.
- Validate executor emits aggregate report.

Acceptance:

```bash
go test ./internal/ocrpipeline -count=1
```

### Phase 4: Wire CLI to workflow runtime

File:

```text
cmd/book-ocr/main.go
```

Tasks:

- Add structured queues to `newRuntime`.
- Register structured package/executors in `structured-run`.
- Update `resume` to register both freeform OCR and structured OCR packages.
- Optionally update `status`, `retry`, and `cancel` only if needed. They operate at runtime level and likely do not need package-specific registration except for `resume` execution.

Acceptance:

```bash
go run ./cmd/book-ocr structured-run --dry-run ...
book-ocr status --work-dir DIR --run-id RUN_ID
book-ocr resume --work-dir DIR --run-id RUN_ID
```

### Phase 5: Dry-run validation

Run:

```bash
go run ./cmd/book-ocr structured-run \
  --book-id report-794-structured-workflow-dry \
  --image-dir /home/manuel/code/wesen/claw-stuff/output/books/presentation-based-uis/pages \
  --start-page 1 \
  --end-page 50 \
  --work-dir /tmp/book-ocr-structured-workflow-dry \
  --dry-run \
  --max-workers 4 \
  --log-level warn
```

Check:

```bash
sqlite3 /tmp/book-ocr-structured-workflow-dry/engine.db '.tables'
sqlite3 /tmp/book-ocr-structured-workflow-dry/projections/book_ocr_structured.sqlite \
  'select status, count(*) from structured_pages group by status;'
rg -c '<!-- page:' /tmp/book-ocr-structured-workflow-dry/assembled.md
```

### Phase 6: Limited live retry validation

Run a limited set likely to succeed quickly:

```bash
go run ./cmd/book-ocr structured-run \
  --book-id report-794-structured-workflow-live-smoke \
  --image-dir /home/manuel/code/wesen/claw-stuff/output/books/presentation-based-uis/pages \
  --start-page 31 \
  --end-page 32 \
  --work-dir /tmp/book-ocr-structured-workflow-live-smoke \
  --profile gpt-5-mini-low \
  --profile-registries /tmp/book-ocr-hq-001/profiles-clean.yaml \
  --dry-run=false \
  --max-workers 1 \
  --log-level warn
```

If a transient provider error occurs, confirm automatic retry. If no transient error occurs, manually induce a retryable fake error in a test rather than depending on provider instability.

## Testing Strategy

### Unit tests

- Package registration rejects nil runtime.
- Projection schema is created.
- Pending/running/succeeded/failed helpers update rows correctly.
- Discover executor emits expected number of page steps in dry-run fixtures.
- Page executor with fake client writes local artifacts and workflow artifacts.
- Assemble executor sorts page results numerically.
- Validate executor detects adjacent duplicate captions with fixtures.

### Integration tests

Use dry-run client and temporary image files.

Expected flow:

```go
rt := workflow.NewRuntime(...)
ocrpipeline.RegisterStructuredWorkflow(rt, cfg{Client: DryRunStructuredOCRClient{}})
handle := rt.StartRun(ctx, StructuredPackageName, input)
for !terminal { rt.RunOnce(ctx) }
result := rt.Result(ctx, handle.ID, "validate-structured-run")
```

### Live tests

Live tests must remain opt-in. They should not run in normal `go test ./...`.

Manual command:

```bash
go run ./cmd/book-ocr structured-run ... --dry-run=false --profile gpt-5-mini-low
```

## Risks and Mitigations

### Risk: Workflow package duplicates old OCR workflow code

Mitigation: start with a small structured-specific package, then extract shared page discovery helpers only if duplication becomes painful. Do not prematurely generalize.

### Risk: Parse failures are permanent but might succeed on retry

Mitigation: preserve raw response artifacts. Start with permanent parse failures to avoid runaway provider spend. Add configurable parse retry later.

### Risk: Local artifacts and workflow artifacts drift

Mitigation: local files remain the human-debug artifact; workflow artifacts point to key data for operator tools. Page projection rows should record both local paths and workflow artifact IDs.

### Risk: `resume` needs package registration

Mitigation: update `resumeRun` to register structured executors as well as freeform executors. A resumed structured workflow cannot run if its executor kinds are not registered.

### Risk: Existing direct structured-run behavior changes

Mitigation: keep `structured-page` direct. If needed, keep `structured-run --direct` temporarily, but prefer workflow-backed behavior by default.

## Open Questions

- Should parse failures be permanent or retryable for live structured OCR?
- Should `structured-run` keep a `--direct` fallback after workflow integration?
- Should structured OCR reuse the old `pages` projection table or use a new `structured_pages` table? This guide recommends a new table.
- Should full local page artifacts be written by the workflow page executor, or should workflow artifacts become canonical? This guide recommends both for now.
- Should page discovery move out of `internal/ocrmvp` to avoid structured OCR importing old freeform package helpers?

## File Reference Index

### Book OCR CLI

```text
/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/cmd/book-ocr/main.go
```

Key areas:

- `runStructuredRun`: current direct structured loop.
- `existingStructuredPageResult`: artifact-level resume helper.
- `newRuntime`: queue configuration.
- `resumeRun`: currently registers freeform OCR package.

### Structured OCR

```text
/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/ocrpipeline/client.go
/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/ocrpipeline/structured_ocr.go
/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/ocrpipeline/types.go
/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/ocrpipeline/renderer.go
/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/ocrpipeline/session.go
```

### Freeform OCR workflow template

```text
/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/ocrmvp/package.go
/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/ocrmvp/discover.go
/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/ocrmvp/executors.go
/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/ocrmvp/projection.go
/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/ocrmvp/types.go
```

### Workflow runtime

```text
/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/workflow/package.go
/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/workflow/executor.go
/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/workflow/context.go
/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/workflow/runtime.go
/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/workflow/operators.go
```

## Intern Quick Start

Run these commands before coding:

```bash
cd /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr
go test ./... -count=1
docmgr task list --ticket BOOK-OCR-STRUCTURED-WORKFLOW-001
```

Then inspect these files in order:

```text
internal/ocrmvp/package.go
internal/ocrmvp/discover.go
internal/ocrmvp/executors.go
internal/ocrpipeline/structured_ocr.go
cmd/book-ocr/main.go
```

First implementation target:

```text
Add workflow_types.go and workflow_package.go, then make `go test ./internal/ocrpipeline -count=1` pass.
```
