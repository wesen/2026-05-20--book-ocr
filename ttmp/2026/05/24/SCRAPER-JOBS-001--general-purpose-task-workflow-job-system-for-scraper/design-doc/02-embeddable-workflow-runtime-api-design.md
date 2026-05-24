---
Title: Embeddable Workflow Runtime API Design
Ticket: SCRAPER-JOBS-001
Status: active
Topics:
    - scraper
    - jobs
    - ocr
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: scraper/pkg/cmd/worker_runtime.go
      Note: CLI worker wiring to extract behind embeddable Runtime
    - Path: scraper/pkg/engine/model/types.go
      Note: Current workflow/op/result/artifact model to wrap as Run/Step/StepResult
    - Path: scraper/pkg/engine/runner/runner.go
      Note: Current runner interface to wrap as Executor
    - Path: scraper/pkg/engine/scheduler/scheduler.go
      Note: Current scheduler loop that Runtime.RunOnce/StartWorkers should wrap
    - Path: scraper/pkg/engine/store/store.go
      Note: Current store lifecycle contract for runs
    - Path: scraper/pkg/services/submission/service.go
      Note: Current submit path to generalize into entrypoints and StartRun
    - Path: scraper/pkg/sites/manifest/manifest.go
      Note: Current site manifest concept to generalize into workflow packages
ExternalSources: []
Summary: 'Ground-up design for an embeddable workflow runtime API that exposes scraper''s native strengths: workflow runs, dynamic DAG steps, artifacts, projections, events, and operator controls.'
LastUpdated: 2026-05-24T18:20:00-04:00
WhatFor: Use this when designing the public Go API for scraper's future general-purpose workflow runtime, especially before implementing typed executors, workflow packages, projection stores, artifact stores, or operator APIs.
WhenToUse: Before renaming scraper/site/op/runner concepts or adding an embeddable API for Go applications.
---


# Embeddable Workflow Runtime API Design

## Executive summary

The embeddable API should not be designed as a clone of a background job queue. The current `scraper` system is becoming something broader: a durable workflow runtime with dynamic DAG execution, per-step artifacts, structured results, projection databases, runtime events, and operator controls. The public Go API should expose those native strengths directly.

The recommended top-level abstraction is a `Runtime` or `Engine`, not a job queue client. Applications should be able to import the runtime, register workflow packages and executors, start runs, emit dynamic child steps, inspect artifacts/results/projections, stream runtime events, and perform operator actions such as retry, cancel, stale reset, and page-subset reprocessing.

The current code already contains most of the internal building blocks:

- workflow and step state in `pkg/engine/model/types.go`;
- executor dispatch through `pkg/engine/runner/runner.go`;
- durable store and lease lifecycle through `pkg/engine/store/store.go` and the SQLite implementation;
- scheduler loop in `pkg/engine/scheduler/scheduler.go`;
- CLI worker wiring in `pkg/cmd/worker_runtime.go`;
- site/package manifests in `pkg/sites/manifest/manifest.go`;
- submit verbs and services in `pkg/sites/submitverbs/` and `pkg/services/submission/`;
- API/UI surfaces for workflows, queues, artifacts, results, retry, cancel, metrics, and events.

The public API should be built as a facade over these pieces, while gradually renaming scraper-specific concepts:

| Current name | Future API name | Why |
|---|---|---|
| scraper | workflow runtime / engine | The system is no longer only for scraping. |
| site | workflow package / domain package | A package bundles definitions, executors, projections, migrations, and submit forms. |
| workflow | run / workflow run | A user-visible execution instance. |
| op | step / task | One durable executable node in a run. |
| runner | executor / handler | Code that executes a step kind. |
| site DB | projection store | Domain-specific query/read model. |
| submit verb | entrypoint / run starter | User-facing command/API form that starts a run. |

This document proposes an API that is workflow-native first and queue-compatible second.

## Problem statement

The current project is usable as a CLI and local API server, but it is not yet easy for another Go application to embed. A host application must currently understand internal types and wiring: `model.OpSpec`, `runner.Registry`, `scheduler.New`, store setup, site registries, metrics, runtime events, and CLI helpers.

The target developer experience should be closer to this:

```go
rt, err := workflow.NewRuntime(workflow.Config{
    Store:          workflow.SQLiteStore("state/engine.db"),
    ArtifactStore:  workflow.FileArtifacts("state/artifacts"),
    ProjectionRoot: "state/projections",
})
if err != nil { panic(err) }

bookOCR := workflow.Package("book-ocr").
    Input(BookOCRInput{}).
    Projection(BookOCRProjection{}).
    Executor("pdf/render-pages", RenderPagesExecutor{}).
    Executor("ocr/vlm-page", OCRPageExecutor{}).
    Executor("markdown/postprocess", PostprocessExecutor{})

rt.RegisterPackage(bookOCR)

run, err := rt.StartRun(ctx, "book-ocr", BookOCRInput{
    BookID:    "aitr-794",
    SourcePDF: "/data/AITR-794.pdf",
    Prompt:    "universal-v2",
})

// Later, from an admin endpoint or CLI command:
err = rt.Reprocess(ctx, run.ID, workflow.SelectSteps().
    Kind("ocr/vlm-page").
    Where("page_type", "table"),
    workflow.WithPromptVersion("universal-v3"),
)
```

This API does not hide the workflow model. It makes the model explicit and ergonomic.

## Design principles

### 1. Workflow runs are first-class

The core object should be a run, not a queue item. A run has:

- stable ID;
- package/domain name;
- input;
- metadata;
- status;
- step graph;
- artifacts;
- projections;
- runtime events;
- operator actions.

This maps directly to current `WorkflowRun` plus the existing op/result/artifact tables.

### 2. Steps are durable graph nodes

A step should represent one durable executable node. It may depend on previous steps, emit child steps, write artifacts, update projections, fail retryably, or be manually retried.

The API should encourage stable step IDs and idempotency because real workflows such as OCR need selective reruns and replay.

### 3. Executors are pluggable and typed, but not the whole model

Typed Go executors are important, but they are only one extension point. The system should also support existing JS executors, HTTP/fetch executors, external-process executors, and future provider-backed executors.

### 4. Artifacts are first-class

Artifacts are a core differentiator. OCR, scraping, LLM, and ETL workflows all need inspectable intermediate files: page images, HTML, raw model output, prompts, logs, JSON, final markdown, PDFs, and debug bundles.

### 5. Projections are first-class

The current `site DB` model is powerful: engine state remains generic, while domain-specific read models live separately. The future API should expose projection stores explicitly rather than hiding them as an implementation detail.

### 6. Operator controls are part of the product

The runtime should include programmatic and HTTP-friendly operations for:

- retry step;
- retry run subset;
- cancel run;
- request cooperative cancellation;
- reset stale running steps;
- reprocess selected pages/items;
- inspect artifacts;
- stream events;
- explain why a step is blocked.

### 7. Backend flexibility should not dominate the public model

The public model is workflow/run/step/artifact/projection/event. Storage and queue backends are implementation choices. The API should not leak backend details except where the host application opts into backend-specific features.

## Current architecture evidence

### Workflow and step model

`pkg/engine/model/types.go` already contains the generic workflow primitives:

- `WorkflowRun` has ID, site, name, status, input, metadata, created/updated timestamps.
- `OpSpec` has ID, workflow ID, parent ID, site, kind, queue, dedup key, JSON input, dependencies, retry policy/state, metadata, and readiness timestamps.
- `OpResult` contains data, records, artifacts, emitted ops, emitted IDs, errors, and completion time.
- `QueuePolicy`, `RateLimitPolicy`, and `Lease` already support scheduling and concurrency policy.

The future API should rename and wrap these concepts, but it does not need a full model rewrite at first.

Proposed conceptual mapping:

```go
type Run = model.WorkflowRun      // public wrapper later
type Step = model.OpSpec          // public wrapper later
type StepResult = model.OpResult  // public wrapper later
```

Do not expose these aliases directly as the public API; expose friendlier types and convert internally.

### Executor dispatch

`pkg/engine/runner/runner.go` defines:

```go
type Runner interface {
    Kind() string
    Run(ctx context.Context, runCtx RunContext) (*model.OpResult, error)
}
```

This maps directly to `Executor`:

```go
type Executor interface {
    Kind() string
    Execute(context.Context, *StepContext) error
}
```

The new `StepContext` should hide raw JSON and `model.OpResult` assembly while still providing access to dependencies, emitted steps, artifacts, projection stores, and logs.

### Store lifecycle

`pkg/engine/store/store.go` already has a useful backend contract:

- `CreateWorkflow`;
- `Enqueue`;
- `GetOp`;
- `LeaseReadyOp`;
- `HeartbeatLease`;
- `CompleteOp`;
- `FailOp`;
- `GetResult`;
- `RefreshRunnableOps`;
- `ListQueueCandidates`;
- `GetWorkflowStats`.

This should become the internal `RuntimeStore` backend. The public API should not require ordinary users to call these methods directly.

### Scheduler loop

`pkg/engine/scheduler/scheduler.go` already implements the durable poll/lease/execute loop:

```text
RefreshRunnableOps(now)
ListQueueCandidates(now)
For each queue candidate:
  apply queue policy
  LeaseReadyOp(...)
  dispatch executor by kind
  CompleteOp or FailOp
refresh workflow status
```

`Runtime.RunOnce` and `Runtime.StartWorkers` can wrap this scheduler.

### CLI worker wiring to extract

`pkg/cmd/worker_runtime.go` currently wires the worker runtime from CLI options:

- opens engine store;
- opens scraper DB;
- opens runtime event producer;
- creates metrics registry;
- registers default JS/HTTP runners;
- opens site DB provider;
- creates scheduler;
- starts worker metrics server;
- runs scheduler cycles.

This should move behind the embeddable runtime. The CLI should later become a thin wrapper over the same API used by Go applications.

## Proposed public package layout

Use a new package name that reflects the future, while keeping old internal packages stable during transition.

```text
pkg/workflow/                 # public embeddable API
  runtime.go                  # Runtime, NewRuntime, Start, Stop, RunOnce
  config.go                   # Config, StoreConfig, QueueConfig, EventConfig
  package.go                  # Package, PackageBuilder, RegisterPackage
  run.go                      # StartRun, RunHandle, RunInfo, RunSelector
  step.go                     # StepDef, StepHandle, StepSelector, Dependency
  executor.go                 # Executor, TypedExecutor, ExecutorFunc
  context.go                  # StepContext: Input, Emit, Artifact, Projection, Dep
  artifact.go                 # ArtifactStore interfaces and helpers
  projection.go               # ProjectionStore interfaces and migrations
  events.go                   # EventSink/EventStream abstractions
  operators.go                # Retry, Cancel, Reprocess, ResetStale, ExplainBlocked
  errors.go                   # Retryable/Permanent errors and classification
  sqlite.go                   # SQLite store constructor/wrapper

pkg/workflow/httpapi/          # reusable HTTP API mounting package
pkg/workflow/webembed/         # optional embedded UI later
pkg/workflow/compat/scraper/   # compatibility with current site manifests/names
```

The name `workflow` is used here as a placeholder. Alternatives: `runtime`, `flows`, `workflows`, `tasks`. Avoid naming the package after the current project if the project will be renamed.

## Core API proposal

### Runtime

The runtime owns stores, executors, packages, scheduler workers, event sinks, metrics, and operator services.

```go
type Runtime struct {
    store         RuntimeStore
    artifacts     ArtifactStore
    projections   ProjectionRegistry
    executors     ExecutorRegistry
    packages      PackageRegistry
    events        EventSink
    metrics       Metrics
    scheduler     Scheduler
}

type Config struct {
    Store         StoreConfig
    ArtifactStore ArtifactStore
    ProjectionDir string
    EventSink      EventSink
    Metrics        MetricsConfig

    Queues map[string]QueueConfig

    Worker WorkerConfig
}

func NewRuntime(cfg Config, opts ...Option) (*Runtime, error)
func (rt *Runtime) RegisterPackage(pkg *Package) error
func (rt *Runtime) RegisterExecutor(kind string, exec Executor) error
func (rt *Runtime) StartRun(ctx context.Context, packageName string, input any, opts ...RunOption) (*RunHandle, error)
func (rt *Runtime) RunOnce(ctx context.Context) (*CycleResult, error)
func (rt *Runtime) StartWorkers(ctx context.Context, opts ...WorkerOption) error
func (rt *Runtime) Stop(ctx context.Context) error
```

The initial implementation can wrap existing scheduler types internally.

### Workflow package

A package bundles definitions for a domain: entrypoints, executors, projection migrations, prompt/config files, JS scripts, default queues, and UI metadata.

```go
type Package struct {
    Name        string
    DisplayName string
    Version     string

    InputSchema any
    Entrypoints []Entrypoint
    Executors   []ExecutorRegistration
    Projection  ProjectionSpec
    Migrations  []Migration
    Queues      map[string]QueueConfig
}

func Package(name string) *PackageBuilder
```

Builder sketch:

```go
bookOCR := workflow.Package("book-ocr").
    DisplayName("Book OCR").
    Input(BookOCRInput{}).
    Projection(BookOCRProjection{}).
    Queue("ocr:io", workflow.QueueConfig{MaxWorkers: 4}).
    Queue("ocr:vlm", workflow.QueueConfig{MaxWorkers: 4}).
    Executor("pdf/render-pages", RenderPagesExecutor{}).
    Executor("ocr/vlm-page", OCRPageExecutor{}).
    Executor("markdown/postprocess", PostprocessExecutor{}).
    Entrypoint("transcribe", StartBookOCR)
```

This replaces the mental role currently played by `site.yaml`, `verbs/`, `scripts/`, and `migrations/` without requiring an immediate breaking rename.

### Entrypoint

An entrypoint starts a run and emits the initial step graph. It is the generalized version of a submit verb.

```go
type Entrypoint interface {
    Name() string
    Start(ctx context.Context, run *RunBuilder, input any) error
}

type EntrypointFunc[I any] func(context.Context, *RunBuilder, I) error
```

Example:

```go
type BookOCRInput struct {
    BookID    string `json:"book_id"`
    SourcePDF string `json:"source_pdf"`
    Prompt    string `json:"prompt"`
}

func StartBookOCR(ctx context.Context, run *workflow.RunBuilder, in BookOCRInput) error {
    run.Metadata("book_id", in.BookID)
    run.Step("render-pages", RenderPagesArgs{
        BookID: in.BookID,
        SourcePDF: in.SourcePDF,
    }, workflow.StepOpts{Queue: "ocr:io"})
    return nil
}
```

### Run builder

The run builder creates the initial durable graph.

```go
type RunBuilder struct {
    packageName string
    workflow    model.WorkflowRun
    steps       []model.OpSpec
}

func (b *RunBuilder) Name(name string)
func (b *RunBuilder) Metadata(key, value string)
func (b *RunBuilder) Step(id string, args any, opts StepOpts) StepHandle
func (b *RunBuilder) Require(handles ...StepHandle) []Dependency
```

Example with dependencies:

```go
render := run.Step("render-page-047", RenderPageArgs{Page: 47}, workflow.StepOpts{Queue: "ocr:cpu"})
ocr := run.Step("ocr-page-047", OCRPageArgs{Page: 47}, workflow.StepOpts{
    Queue: "ocr:vlm",
    DependsOn: run.Require(render),
})
run.Step("cleanup-page-047", CleanupArgs{Page: 47}, workflow.StepOpts{
    Queue: "ocr:llm",
    DependsOn: run.Require(ocr),
})
```

### Executor and step context

Executors should receive typed input and a rich context.

```go
type Executor interface {
    Kind() string
    Execute(ctx context.Context, step *StepContext) error
}

type TypedExecutor[I any] interface {
    Kind() string
    ExecuteTyped(ctx context.Context, step *StepContext, input I) error
}

type ExecutorFunc[I any] func(context.Context, *StepContext, I) error
```

`StepContext` should expose scraper's unique capabilities:

```go
type StepContext struct {
    Run      RunInfo
    Step     StepInfo
    Attempt  int
    Now      time.Time
}

func (s *StepContext) Input(out any) error
func (s *StepContext) Dependency(stepID string, out any) error
func (s *StepContext) Emit(id string, args any, opts StepOpts) (StepHandle, error)
func (s *StepContext) Result(data any) error
func (s *StepContext) Record(collection, key string, data any) error
func (s *StepContext) Artifact(name string, contentType string, body []byte, opts ...ArtifactOption) (ArtifactRef, error)
func (s *StepContext) Projection(name string) (Projection, error)
func (s *StepContext) Event(kind string, payload any)
func (s *StepContext) Log(args ...any)
func (s *StepContext) IsCancelRequested(ctx context.Context) (bool, error)
```

Example OCR executor:

```go
type OCRPageInput struct {
    BookID        string `json:"book_id"`
    Page          int    `json:"page"`
    PageImageID   string `json:"page_image_id"`
    PromptVersion string `json:"prompt_version"`
}

type OCRPageExecutor struct {
    Provider VLMProvider
}

func (OCRPageExecutor) Kind() string { return "ocr/vlm-page" }

func (e OCRPageExecutor) Execute(ctx context.Context, step *workflow.StepContext) error {
    var in OCRPageInput
    if err := step.Input(&in); err != nil { return err }

    prompt, err := step.Projection("book-ocr").GetPrompt(ctx, in.PromptVersion)
    if err != nil { return err }

    pageImage, err := step.Artifacts().Open(ctx, in.PageImageID)
    if err != nil { return workflow.Retryable(err) }

    output, err := e.Provider.Transcribe(ctx, pageImage, prompt.Text)
    if err != nil { return workflow.Retryable(err) }

    mdRef, err := step.Artifact(
        fmt.Sprintf("page_%03d.md", in.Page),
        "text/markdown",
        []byte(output.Markdown),
        workflow.ArtifactKind("ocr-markdown"),
    )
    if err != nil { return err }

    return step.Record("pages", fmt.Sprintf("%s:%03d", in.BookID, in.Page), map[string]any{
        "book_id": in.BookID,
        "page": in.Page,
        "status": "done",
        "page_type": output.PageType,
        "prompt_version": in.PromptVersion,
        "markdown_artifact_id": mdRef.ID,
    })
}
```

### Dynamic child step emission

Dynamic emission is one of the current system's most important capabilities. The API should make it obvious.

```go
func (e RenderPagesExecutor) Execute(ctx context.Context, step *workflow.StepContext) error {
    var in RenderPagesInput
    _ = step.Input(&in)

    pages, err := renderPDF(in.SourcePDF)
    if err != nil { return err }

    cleanupSteps := []workflow.StepHandle{}

    for _, page := range pages {
        imageRef, _ := step.Artifact(page.FileName, "image/png", page.Bytes)

        ocr, _ := step.Emit(
            fmt.Sprintf("ocr-page-%03d", page.Number),
            OCRPageInput{BookID: in.BookID, Page: page.Number, PageImageID: imageRef.ID, PromptVersion: in.PromptVersion},
            workflow.StepOpts{Queue: "ocr:vlm"},
        )

        cleanup, _ := step.Emit(
            fmt.Sprintf("cleanup-page-%03d", page.Number),
            CleanupPageInput{BookID: in.BookID, Page: page.Number},
            workflow.StepOpts{Queue: "ocr:postprocess", DependsOn: workflow.Require(ocr)},
        )
        cleanupSteps = append(cleanupSteps, cleanup)
    }

    _, _ = step.Emit("assemble-book", AssembleBookInput{BookID: in.BookID}, workflow.StepOpts{
        Queue: "ocr:io",
        DependsOn: workflow.Require(cleanupSteps...),
    })

    return nil
}
```

### Artifact store

Current SQLite BLOB artifacts are useful, but OCR proves that artifact storage must be pluggable.

```go
type ArtifactStore interface {
    Put(ctx context.Context, artifact ArtifactWrite) (ArtifactRef, error)
    Open(ctx context.Context, id string) (io.ReadCloser, ArtifactMeta, error)
    Delete(ctx context.Context, id string) error
}
```

Initial backends:

- `SQLiteArtifacts`: current behavior, good for small demos.
- `FileArtifacts`: local filesystem, good for OCR page images and final files.
- `ObjectArtifacts`: future S3-compatible storage.

The engine DB should store artifact metadata and references even when bytes live outside SQLite.

### Projection store

The generalized version of the current site DB should be a projection store. Projection stores are domain-specific query models, not scheduler state.

```go
type Projection interface {
    Exec(ctx context.Context, query string, args ...any) error
    Query(ctx context.Context, query string, args ...any) (Rows, error)
}

type ProjectionSpec struct {
    Name       string
    Migrations []Migration
}
```

For OCR:

```sql
books(book_id, source_uri, status, page_count, created_at, updated_at)
pages(book_id, page_number, status, page_type, prompt_version, attempt, markdown_artifact_id, error_code)
prompts(prompt_version, prompt_hash, prompt_text, created_at)
exports(book_id, kind, artifact_id, created_at)
```

### Runtime events

Runtime events should be embedded as an explicit API surface.

```go
type EventSink interface {
    Publish(ctx context.Context, event Event) error
}

type Event struct {
    RunID string
    StepID string
    Package string
    Kind string
    Severity string
    Message string
    Payload any
    OccurredAt time.Time
}
```

Applications should be able to choose:

- no events;
- in-memory events;
- Redis stream events;
- custom event sink;
- OpenTelemetry bridge later.

### Operator API

The runtime should expose operations that map to admin UI and API actions.

```go
func (rt *Runtime) RetryStep(ctx context.Context, runID, stepID string, opts ...RetryOption) error
func (rt *Runtime) RetryFailed(ctx context.Context, runID string, selector StepSelector) error
func (rt *Runtime) CancelRun(ctx context.Context, runID string, opts ...CancelOption) error
func (rt *Runtime) RequestCancelStep(ctx context.Context, runID, stepID string) error
func (rt *Runtime) ResetStale(ctx context.Context, runID string, olderThan time.Duration) (int, error)
func (rt *Runtime) Reprocess(ctx context.Context, runID string, selector StepSelector, opts ...ReprocessOption) error
func (rt *Runtime) ExplainBlocked(ctx context.Context, runID, stepID string) (*BlockedExplanation, error)
```

OCR-specific reprocessing should be built on generic step selection:

```go
err := rt.Reprocess(ctx, run.ID,
    workflow.SelectSteps().Kind("ocr/vlm-page").Where("page_type", "table"),
    workflow.WithInputPatch(map[string]any{"prompt_version": "universal-v3"}),
)
```

### Step selectors

Selectors are critical for operator workflows.

```go
type StepSelector struct {
    Kinds []string
    Statuses []StepStatus
    Metadata map[string]string
    ProjectionFilter ProjectionFilter
}

workflow.SelectSteps().
    Kind("ocr/vlm-page").
    Status(workflow.Failed).
    WhereMetadata("prompt_version", "universal-v2")
```

For AITR-style OCR:

```go
workflow.SelectSteps().Kind("ocr/vlm-page").WhereProjection("pages.page_type = ?", "blank")
workflow.SelectSteps().Kind("ocr/vlm-page").WhereProjection("pages.page_number IN (?, ?, ?)", 13, 47, 182)
workflow.SelectSteps().Kind("ocr/vlm-page").Status(workflow.Failed)
```

## Backend model

### Runtime store

The internal store should be split conceptually into:

```text
Workflow state store
  runs, steps, dependencies, leases, retry state, results, artifact metadata

Queue backend
  ready-step discovery, leasing, queue concurrency, rate limiting, scheduled readiness

Projection stores
  domain-specific query models

Artifact store
  artifact bytes and metadata references
```

The current SQLite implementation combines most of workflow state and queue backend. That is fine for v1. The public API should allow future separation.

### Backend interface sketch

```go
type RuntimeStore interface {
    Runs() RunStore
    Steps() StepStore
    Results() ResultStore
    Scheduler() SchedulerStore
}

type SchedulerStore interface {
    RefreshRunnableSteps(ctx context.Context, now time.Time) (int, error)
    ListQueueCandidates(ctx context.Context, now time.Time) ([]QueueCandidate, error)
    LeaseReadyStep(ctx context.Context, req LeaseRequest) (*Step, *Lease, error)
    HeartbeatLease(ctx context.Context, stepID string, lease Lease, extendBy time.Duration) error
    CompleteStep(ctx context.Context, stepID string, result StepResult) error
    FailStep(ctx context.Context, stepID string, failure StepFailure) error
}
```

The first implementation can adapt the existing `store.Store` contract.

### Queue backends

Do not expose a queue-first API, but allow different queue implementations:

- SQLite queue backend: current local-first backend.
- Postgres queue backend: future native implementation.
- External queue backend: future bridge for teams that need a dedicated queue system.

The public API should remain the same across these backends.

## API examples

### Example 1: Simple embedded one-step workflow

```go
type ResizeImageInput struct {
    ImagePath string `json:"image_path"`
    Width     int    `json:"width"`
}

type ResizeExecutor struct{}
func (ResizeExecutor) Kind() string { return "image/resize" }
func (ResizeExecutor) Execute(ctx context.Context, step *workflow.StepContext) error {
    var in ResizeImageInput
    if err := step.Input(&in); err != nil { return err }
    out, err := resize(in.ImagePath, in.Width)
    if err != nil { return workflow.Retryable(err) }
    _, err = step.Artifact("resized.png", "image/png", out)
    return err
}

rt, _ := workflow.NewRuntime(workflow.Config{Store: workflow.SQLiteStore("state/engine.db")})
rt.RegisterExecutor("image/resize", ResizeExecutor{})
run, _ := rt.StartRun(ctx, workflow.AdHocPackage("images"), ResizeImageInput{ImagePath: "in.png", Width: 800})
_ = run
```

### Example 2: Book OCR package

```go
bookOCR := workflow.Package("book-ocr").
    Queue("ocr:io", workflow.QueueConfig{MaxWorkers: 2}).
    Queue("ocr:vlm", workflow.QueueConfig{MaxWorkers: 4}).
    Projection(BookOCRProjection()).
    Entrypoint("transcribe", workflow.EntrypointFunc[BookOCRInput](StartBookOCR)).
    Executor("pdf/render-pages", RenderPagesExecutor{}).
    Executor("ocr/vlm-page", OCRPageExecutor{Provider: pinocchio}).
    Executor("markdown/postprocess", PostprocessExecutor{}).
    Build()

rt.RegisterPackage(bookOCR)
run, err := rt.StartRun(ctx, "book-ocr", BookOCRInput{
    BookID: "aitr-794",
    SourcePDF: "~/Downloads/AITR-794.pdf",
    PromptVersion: "universal-v2",
})
```

### Example 3: Mount reusable HTTP API into an app

```go
mux := http.NewServeMux()
workflowhttp.Mount(mux, rt, workflowhttp.Options{
    Prefix: "/internal/workflows",
    EnableArtifacts: true,
    EnableEvents: true,
    EnableOperators: true,
})

http.ListenAndServe(":8080", mux)
```

This lets host applications embed the runtime and still get the admin/operator surfaces.

## What this changes relative to the previous framing

### Drop queue-first naming

Do not center the API on `Client.InsertJob`. A convenience method for one-step runs is fine, but the main API should be `StartRun`, `RegisterPackage`, `RegisterExecutor`, and `Reprocess`.

### Drop copied job-queue vocabulary where it conflicts

Avoid making `JobArgs`, `WorkerDefaults`, or queue-style `AddWorker` the dominant concepts. Prefer:

- `RunInput`;
- `StepInput`;
- `Executor`;
- `Runtime`;
- `WorkflowPackage`;
- `StepContext`.

Typed inputs are still useful, but they serve workflow steps rather than defining the whole system.

### Promote projection and artifact APIs

The previous framing underemphasized projection stores. For this system, projection and artifact APIs are central because they turn execution into inspectable products.

### Promote operator controls

The previous framing mentioned retry/cancel but did not make operator actions central enough. A generalized scraper runtime should be designed for human-in-the-loop production work.

## Implementation plan

### Phase 1: Introduce future-facing names without breaking internals

Add a new public package that wraps old internals.

Files to add:

```text
pkg/workflow/runtime.go
pkg/workflow/executor.go
pkg/workflow/context.go
pkg/workflow/package.go
pkg/workflow/run.go
pkg/workflow/step.go
pkg/workflow/errors.go
```

Implementation goals:

- `Executor` adapts to `runner.Runner`.
- `StepContext` wraps `runner.RunContext` and accumulates `model.OpResult` writes.
- `Runtime.RegisterExecutor` registers into an internal `runner.Registry`.

Do not rename internal packages yet.

### Phase 2: Runtime construction over SQLite

Add:

```go
func NewRuntime(cfg Config, opts ...Option) (*Runtime, error)
func SQLiteStore(path string) StoreConfig
```

Internally use:

- `sqlitestore.Open`;
- `scheduler.New`;
- existing metrics/runtime event wrappers if configured.

### Phase 3: StartRun and RunBuilder

Add package registration and run creation.

- `Runtime.RegisterPackage` stores package definitions.
- `Runtime.StartRun` invokes the entrypoint and creates workflow + initial steps.
- Convert `RunBuilder.Step` calls to `model.OpSpec`.

Tests:

- start a one-step run;
- start a three-step DAG;
- verify dependencies block child until parent succeeds.

### Phase 4: Worker lifecycle

Add:

```go
func (rt *Runtime) RunOnce(ctx context.Context) (*CycleResult, error)
func (rt *Runtime) StartWorkers(ctx context.Context, opts ...WorkerOption) error
func (rt *Runtime) Stop(ctx context.Context) error
```

Wrap the existing scheduler loop, but review concurrency semantics before documenting strong parallelism guarantees.

### Phase 5: ArtifactStore abstraction

Introduce file-backed artifacts before OCR porting.

- Keep SQLite BLOB compatibility.
- Add file artifact store.
- Store refs in the engine DB.
- Update API/artifact read paths to support external artifact bytes.

### Phase 6: ProjectionStore API

Generalize the site DB pattern.

- Keep current site DB provider as a compatibility implementation.
- Add `ProjectionSpec` and migrations.
- Add `StepContext.Projection(name)`.

### Phase 7: Operator service

Expose reusable operator methods:

- retry step;
- retry selector;
- cancel run;
- reset stale;
- reprocess selector;
- explain blocked.

Then mount them in the existing HTTP API.

### Phase 8: Compatibility and rename path

Keep current CLI behavior working:

- `site` commands continue to work as compatibility aliases.
- docs introduce `workflow package` as the preferred term.
- internal package renames happen only after the public API stabilizes.

## Testing strategy

### Unit tests

- `Executor` adapter decodes typed input.
- `StepContext.Artifact` creates result artifacts.
- `StepContext.Emit` creates child steps with correct parent/dependency metadata.
- `RunBuilder` produces stable step IDs and dependencies.
- Error classification maps retryable/permanent errors to existing `OpError`.

### Integration tests

- Start run with one step, run scheduler, assert succeeded.
- Step emits child step, child runs after parent completes.
- Required dependency failure cancels child.
- Retryable error schedules retry.
- File artifact store writes bytes and API can read them back.
- Projection store migration creates tables and executor writes projection rows.
- Operator reprocess selector creates new steps or attempts as intended.

### OCR-shaped smoke test

Create a tiny fixture workflow:

```text
render-pages -> ocr-page-001 -> assemble-book
```

Assertions:

- run reaches succeeded;
- page projection row exists;
- OCR markdown artifact exists;
- final markdown artifact exists;
- runtime events include step started/succeeded;
- retry/reprocess can rerun `ocr-page-001` with a new prompt version.

## Intern implementation guide

### Read these files first

1. `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/engine/model/types.go`
2. `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/engine/runner/runner.go`
3. `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/engine/store/store.go`
4. `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/engine/scheduler/scheduler.go`
5. `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/engine/store/sqlite/lease_store.go`
6. `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/engine/store/sqlite/result_store.go`
7. `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/cmd/worker_runtime.go`
8. `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/sites/manifest/manifest.go`
9. `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/sites/submitverbs/host.go`
10. `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/api/server/routes_engine.go`

### First PR

Implement only executor wrapping:

- `pkg/workflow.Executor`;
- `pkg/workflow.ExecutorFunc[I]`;
- `pkg/workflow.StepContext` with `Input` and `Result` only;
- adapter to `runner.Runner`;
- tests.

Do not touch scheduler or store yet.

### Second PR

Implement `Runtime.RegisterExecutor` and an in-memory test harness that can execute one adapted runner with a fake `runner.RunContext`.

### Third PR

Implement SQLite-backed `Runtime.StartRun` for one-step runs.

### Fourth PR

Implement `RunOnce` over the existing scheduler.

### Fifth PR

Add `StepContext.Emit`, `Artifact`, and `Record`.

This sequencing avoids a large, risky rewrite.

## Design decisions

### Decision 1: Public API starts with `Runtime`, not `Client`

A client suggests a remote service or queue. Runtime better matches embedded execution, registered packages, workers, stores, and operator APIs.

### Decision 2: Public API starts with workflow packages, not sites

The term `site` is now too narrow. Keep it as a compatibility layer, but introduce `Package` or `WorkflowPackage` in the new API.

### Decision 3: Public API starts with steps, not ops

`op` is an internal engine word. `Step` is clearer for users, docs, and UI.

### Decision 4: Executors should be typed but context-rich

Typed inputs help Go developers, but the context must expose artifacts/projections/events/dynamic emission. Otherwise the API loses what makes this system useful.

### Decision 5: Operator APIs are not optional polish

The concrete OCR workflow needed monitoring, reset, retry, and reprocessing. These should be part of the runtime design, not later UI-only additions.

## Open questions

1. Should the package be called `workflow`, `runtime`, `flows`, or something else?
2. Should the project rename happen before or after adding the new public package?
3. Should `Package` be code-first, manifest-first, or both equally?
4. Should `ProjectionStore` be SQLite-only in v1, or generic from the start?
5. How should external artifact references be represented in the current artifacts table?
6. Should step reprocessing create new step IDs, new attempts under the same step ID, or both depending on policy?
7. How should the API express stable IDs for dynamic fan-out without making simple workflows verbose?

## Final recommendation

Rewrite the second design direction around a workflow-native embeddable runtime. The API should fit this system's model rather than imitating a queue library. The central concepts should be `Runtime`, `WorkflowPackage`, `Run`, `Step`, `Executor`, `ArtifactStore`, `ProjectionStore`, `EventSink`, and `Operator` actions.

The first implementation should be conservative: add a `pkg/workflow` facade over the existing runner/scheduler/store internals. Do not rename the internal code yet. Once the public model proves itself with the book OCR workflow, migrate internal names and CLI terminology gradually.

## References

### Scraper files

- `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/engine/model/types.go`
- `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/engine/runner/runner.go`
- `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/engine/store/store.go`
- `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/engine/scheduler/scheduler.go`
- `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/engine/store/sqlite/lease_store.go`
- `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/engine/store/sqlite/result_store.go`
- `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/cmd/worker_runtime.go`
- `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/cmd/runtime_helpers.go`
- `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/sites/manifest/manifest.go`
- `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/sites/submitverbs/host.go`
- `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/services/submission/service.go`
- `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/api/server/routes_engine.go`

### Related ticket docs

- `design-doc/01-scraper-task-workflow-job-system-design-and-book-ocr-mapping.md`
- `reference/01-diary.md`
