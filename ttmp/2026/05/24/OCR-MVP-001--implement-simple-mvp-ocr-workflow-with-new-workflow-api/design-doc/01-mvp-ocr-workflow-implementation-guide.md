---
Title: MVP OCR Workflow Implementation Guide
Ticket: OCR-MVP-001
Status: active
Topics:
    - ocr
    - workflow
    - scraper
    - implementation-guide
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: claw-stuff/ttmp/2026/05/20/AITR-794--clean-markdown-transcription-of-aitr-794-pdf/scripts/10-reprocess-universal.py
      Note: Prior OCR prompt and page processing reference
    - Path: geppetto/pkg/turns/helpers_blocks.go
      Note: Geppetto multimodal user block API for page image OCR
    - Path: pinocchio/pkg/cmds/profilebootstrap/engine_settings.go
      Note: Pinocchio helper for resolving profile-backed Geppetto engine settings
    - Path: pinocchio/pkg/cmds/profilebootstrap/profile_selection.go
      Note: Pinocchio default profile and registry selection behavior
    - Path: scraper/pkg/workflow/artifact_store.go
      Note: External artifact store used for page and book markdown artifacts
    - Path: scraper/pkg/workflow/context.go
      Note: StepContext input
    - Path: scraper/pkg/workflow/operators.go
      Note: Retry/cancel operator controls for failed OCR pages and run cancellation
    - Path: scraper/pkg/workflow/package.go
      Note: Workflow package and RunBuilder API for initial OCR step graph
    - Path: scraper/pkg/workflow/projection_store.go
      Note: |-
        SQLite projection store for OCR page status/read model
        Projection store lint fix from Phase 1 validation
    - Path: scraper/pkg/workflow/runtime.go
      Note: Runtime configuration
    - Path: scraper/pkg/workflows/ocrmvp/discover.go
      Note: Page discovery and dynamic OCR/assemble step emission
    - Path: scraper/pkg/workflows/ocrmvp/executors.go
      Note: OCR page and assemble executors using artifacts/projections
    - Path: scraper/pkg/workflows/ocrmvp/package.go
      Note: OCR MVP workflow package registration and entrypoint
    - Path: scraper/pkg/workflows/ocrmvp/package_test.go
      Note: Fake-client integration test for the workflow MVP
    - Path: scraper/pkg/workflows/ocrmvp/projection.go
      Note: OCR pages/runs projection schema and update helpers
    - Path: scraper/pkg/workflows/ocrmvp/types.go
      Note: Run/page/result/client contracts for the MVP workflow
ExternalSources: []
Summary: Design and intern-oriented implementation guide for a simple OCR workflow package built on scraper/pkg/workflow, using Geppetto for multimodal OCR and Pinocchio's profile registry defaults.
LastUpdated: 2026-05-24T20:58:00-04:00
WhatFor: Use this before implementing the OCR MVP workflow so the intern understands scraper workflow concepts, Geppetto OCR integration, profile registry resolution, artifacts, projections, and validation.
WhenToUse: Read when implementing, reviewing, or extending OCR-MVP-001.
---



# MVP OCR Workflow Implementation Guide

## Executive Summary

This ticket should implement a small but real OCR workflow package on top of the new `scraper/pkg/workflow` API. The MVP accepts a directory of pre-rendered page images, creates one durable workflow run per book, schedules one OCR step per page, stores each page transcription as an external artifact, records page status in a SQLite projection, and assembles a combined markdown artifact when all pages complete.

The important product goal is not to build the final book-OCR system in one step. The goal is to prove that the new workflow-native API can express the real AITR-style OCR workload without falling back to ad hoc SQLite queue scripts. The MVP should therefore exercise the API surfaces added in `SCRAPER-JOBS-001`: `Runtime`, `Package`, typed executors, dynamic step emission, external artifacts, projection stores, dependency result loading, and operator controls.

The important integration constraint from the user is: **use Geppetto for OCR and use Pinocchio's default profile registry path/precedence rather than hard-coding `PINOCCHIO_PROFILE` or shelling out to the `pinocchio` CLI.** In concrete terms, the OCR executor should import Geppetto inference APIs and Pinocchio's `profilebootstrap` package, resolve the selected engine profile through the same registry defaults as Pinocchio, build a Geppetto engine, and call `RunInference` with a multimodal turn containing the page image.

The recommended code location is inside the `scraper` module, because the workflow API lives there and this MVP is a validation consumer of that API:

```text
scraper/pkg/workflows/ocrmvp/
  package.go          # workflow package registration
  types.go            # public inputs/results
  executors.go        # discover, ocr-page, assemble executors
  geppetto_ocr.go     # Geppetto + Pinocchio profile-registry OCR client
  projection.go       # projection schema helpers
  prompt.go           # OCR prompt template
  package_test.go     # fake OCR client + workflow integration test
scraper/pkg/cmd/ocr_mvp.go or scraper/cmd/scraper/main.go wiring
```

A command can later wrap this package, but the package itself should be usable from embedded Go code.

## Problem Statement and Scope

The previous AITR OCR process worked, but it was script-centered:

- render PDF pages to PNGs;
- use a Python worker script to call `pinocchio code professional --images ...`;
- track page status in a custom SQLite table;
- inspect progress through a custom dashboard;
- reprocess failed/low-quality pages with separate scripts;
- assemble output markdown after the fact.

That flow is useful as a reference but it duplicates runtime concepts that now exist in `scraper/pkg/workflow`:

- page work items are workflow steps;
- worker claiming and retries are scheduler/store concerns;
- page markdown and debug payloads are artifacts;
- page status and QA summaries are projections;
- retry/cancel/reprocess are operator actions.

This ticket should implement an MVP that moves the page-OCR part into the workflow runtime. It should not attempt to solve every OCR production concern.

### In scope

- A workflow package named `ocr-mvp`.
- Input: a `book_id`, an image directory, optional page range, selected profile, optional profile registry overrides, output/projection roots, and prompt version.
- A `discover-pages` step that scans image files and emits page OCR steps plus a final assemble step.
- An `ocr-page` step that:
  - reads one page image;
  - runs Geppetto multimodal inference;
  - stores page markdown as an external artifact;
  - records page status in a projection table;
  - returns structured page result data.
- An `assemble-markdown` step that depends on every page OCR step, reads dependency result data, assembles ordered markdown, and stores the combined markdown artifact.
- A small CLI or command hook to start a run and run workers.
- Unit/integration tests that use a fake OCR client, not a live model provider.

### Out of scope for MVP

- PDF rendering in Go. The MVP can accept a directory of page PNG/JPEG files. Existing Python/PyMuPDF scripts remain the reference for rendering.
- Human QA UI.
- Transparent HTTP dereference of `external-artifact-ref` artifacts.
- Object storage backends.
- Complex OCR quality scoring.
- Full prompt-version experiment management.
- Multi-model voting.

These can become follow-up tickets after the API shape is proven.

## Current-State Architecture and Evidence

### Workflow runtime facade

`scraper/pkg/workflow/runtime.go` defines the embeddable runtime facade. `Config` accepts a durable `Store`, optional `ArtifactStore`, optional `ProjectionStore`, worker settings, and queue settings (`scraper/pkg/workflow/runtime.go:59-70`). `NewRuntime` opens the store, creates the runner registry, creates the scheduler, wires operator/artifact/projection stores, and installs a queue policy provider (`scraper/pkg/workflow/runtime.go:92-123`).

The runtime exposes the core operations needed by this MVP:

- `RegisterExecutor` registers typed or untyped executors (`scraper/pkg/workflow/runtime.go:163-174`).
- `RegisterPackage` registers a workflow package/entrypoint (`scraper/pkg/workflow/runtime.go:176-190`).
- `StartRun` creates a durable workflow run and initial step graph (`scraper/pkg/workflow/runtime.go:219-267`).
- `RunOnce` runs one scheduler cycle (`scraper/pkg/workflow/runtime.go:268-272`).
- `StartWorkers` runs scheduler cycles until context cancellation (`scraper/pkg/workflow/runtime.go:277-306`).
- `Projection` and `Result` let callers inspect domain projections and step results (`scraper/pkg/workflow/runtime.go:309-321`).

For the OCR MVP, this means no custom queue table is needed. The runtime already owns workflow creation, op scheduling, leasing, dependency readiness, and result persistence.

### Workflow package and step graph creation

`scraper/pkg/workflow/package.go` defines a small workflow package abstraction. A package has a name, display name, and entrypoint (`scraper/pkg/workflow/package.go:11-16`). The entrypoint receives a typed input through `EntrypointFunc[I]` and builds the initial graph through `RunBuilder` (`scraper/pkg/workflow/package.go:41-60`).

`RunBuilder.Step` appends an initial durable step with kind, queue, input, dependencies, retry policy, metadata, and site/package fields (`scraper/pkg/workflow/package.go:83-117`). `Require` turns step handles into required dependencies (`scraper/pkg/workflow/package.go:119-128`).

The OCR MVP should use this to create one initial `discover-pages` step. That step can dynamically emit the per-page OCR steps and final assembly step after it has scanned the actual image directory.

### Executor-facing step context

`scraper/pkg/workflow/context.go` defines `StepContext`, the public executor-facing view of a durable step. It wraps the lower-level runner context and accumulates result data, record writes, artifact writes, and dynamically emitted child steps (`scraper/pkg/workflow/context.go:14-28`).

Relevant methods for OCR:

- `Input(out)` decodes typed step input (`scraper/pkg/workflow/context.go:57-69`).
- `DependencyResult` and `DependencyData` load completed upstream step output (`scraper/pkg/workflow/context.go:76-103`).
- `Result(data)` stores structured step output (`scraper/pkg/workflow/context.go:105-114`).
- `Artifact(...)` stores inline artifacts in the engine DB (`scraper/pkg/workflow/context.go:149-177`).
- `Projection(name)` resolves a domain projection (`scraper/pkg/workflow/context.go:179-185`).
- `StoreArtifact(...)` writes bytes to an external artifact store and appends a JSON `external-artifact-ref` result artifact (`scraper/pkg/workflow/context.go:187-230`).
- `Emit(...)` emits child steps after the current step succeeds (`scraper/pkg/workflow/context.go:247-281`).

For OCR, use `StoreArtifact` for markdown outputs because page markdown and debug payloads can become large. Use `Projection` for queryable page status. Use `Emit` from `discover-pages` to create one `ocr-page` child per discovered image and one `assemble-markdown` child that depends on the OCR children.

### External artifact store

`scraper/pkg/workflow/artifact_store.go` defines `ArtifactStore`, `ArtifactObject`, `ArtifactRef`, and `FileArtifactStore`. `ArtifactStore.Put` stores bytes and returns a reference; `Open` reads them back (`scraper/pkg/workflow/artifact_store.go:13-38`). The file-backed implementation writes the artifact body under a local root and writes sidecar JSON metadata (`scraper/pkg/workflow/artifact_store.go:40-105`).

The MVP should configure:

```go
workflow.NewFileArtifactStore(filepath.Join(workDir, "artifacts"))
```

Then each page OCR step should call:

```go
ref, err := step.StoreArtifact(
    fmt.Sprintf("page_%03d.md", input.PageNumber),
    "text/markdown",
    []byte(markdown),
    workflow.ArtifactKind("ocr-markdown"),
    workflow.ArtifactMetadata(map[string]string{
        "book_id": input.BookID,
        "page": strconv.Itoa(input.PageNumber),
        "prompt_version": input.PromptVersion,
    }),
)
```

### Projection store

`scraper/pkg/workflow/projection_store.go` defines `ProjectionStore` and `Projection`. A projection is intentionally separate from engine scheduling state and exposes SQL-shaped `Exec` and `Query` methods (`scraper/pkg/workflow/projection_store.go:15-26`). `SQLiteProjectionStore` stores one SQLite DB per projection name under a local directory (`scraper/pkg/workflow/projection_store.go:28-68`).

For OCR, use one projection named `ocr-mvp` with tables such as:

```sql
CREATE TABLE IF NOT EXISTS pages (
  book_id TEXT NOT NULL,
  page_num INTEGER NOT NULL,
  image_path TEXT NOT NULL,
  status TEXT NOT NULL,
  ocr_step_id TEXT,
  markdown_artifact_id TEXT,
  markdown_artifact_uri TEXT,
  char_count INTEGER,
  error_code TEXT,
  error_message TEXT,
  prompt_version TEXT,
  updated_at TEXT NOT NULL,
  PRIMARY KEY(book_id, page_num)
);
```

This projection lets an operator answer: Which pages are pending? Which failed? Which have artifacts? Which prompt version produced each page?

### Operator controls

`scraper/pkg/workflow/operators.go` exposes retry and cancellation through the runtime. `RetryStep` moves a failed step back to ready, and `CancelRun` cancels pending/ready/running work (`scraper/pkg/workflow/operators.go:11-36`). The MVP does not need a UI, but the implementation should demonstrate how to retry a failed page step from Go or a CLI command.

### Geppetto multimodal inference

Geppetto represents a model request as a `turns.Turn` containing blocks. `turns.NewUserMultimodalBlock` creates a user block with text plus optional image maps; the documented image map shape supports `media_type`, `url` or `content`, and optional `detail` (`geppetto/pkg/turns/helpers_blocks.go:24-39`). Geppetto docs also explain that OpenAI Responses consumes image-bearing turns as mixed `input_text` and `input_image` content, so the page OCR executor should not manually build provider-specific JSON.

The OCR executor should therefore do this:

```go
turn := &turns.Turn{}
turns.AppendBlock(turn, turns.NewSystemTextBlock(systemPrompt))
turns.AppendBlock(turn, turns.NewUserMultimodalBlock(userPrompt, []map[string]any{{
    "media_type": "image/png",
    "content": imageBytes,
    "detail": "high",
}}))
updated, err := engine.RunInference(ctx, turn)
```

Then extract the last `turns.BlockKindLLMText` block and read `turns.PayloadKeyText`.

### Pinocchio profile registry defaults

The MVP must use Pinocchio's profile resolution rules. Do **not** hard-code `PINOCCHIO_PROFILE=gpt-5-nano-low`, do **not** shell out to `pinocchio code professional`, and do **not** manually parse `~/.config/pinocchio/profiles.yaml`.

Use `github.com/go-go-golems/pinocchio/pkg/cmds/profilebootstrap`:

- `profilebootstrap.BootstrapConfig()` is configured with `AppName: "pinocchio"`, `EnvPrefix: "PINOCCHIO"`, Geppetto sections, and Pinocchio's config plan (`pinocchio/pkg/cmds/profilebootstrap/profile_selection.go:41-57`).
- `profilebootstrap.NewCLISelectionValues(...)` creates parsed values for explicit profile/profile-registry overrides while preserving default behavior (`pinocchio/pkg/cmds/profilebootstrap/profile_selection.go:64-65`).
- `ResolveCLIProfileRuntime` reads the resolved Pinocchio config document, extracts `profile.active` and `profile.registries`, applies explicit overrides, imports external registries, composes inline profiles, and selects the default registry/profile (`pinocchio/pkg/cmds/profilebootstrap/profile_selection.go:72-151`).
- `ResolveCLIEngineSettings` resolves hidden base inference settings, profile runtime, profile registry, selected engine profile, and merged final inference settings (`pinocchio/pkg/cmds/profilebootstrap/engine_settings.go:50-122`).
- `NewEngineFromResolvedCLIEngineSettings` builds a Geppetto engine from those final settings (`pinocchio/pkg/cmds/profilebootstrap/engine_settings.go:146-166`).

Underneath, Geppetto's generic bootstrap falls back to `${XDG_CONFIG_HOME:-~/.config}/pinocchio/profiles.yaml` when no explicit registries are supplied and that file exists (`geppetto/pkg/cli/bootstrap/profile_registry_defaults.go:9-27`). This matches Pinocchio's documented registry-source precedence:

1. `--profile-registries`
2. `PINOCCHIO_PROFILE_REGISTRIES`
3. `profile.registries` from Pinocchio config
4. `${XDG_CONFIG_HOME:-~/.config}/pinocchio/profiles.yaml`

## Proposed MVP Architecture

### Component diagram

```text
Operator / CLI
    |
    | StartRun(book_id, image_dir, profile, registries, prompt_version)
    v
scraper/pkg/workflow.Runtime
    |
    | package: ocr-mvp
    v
+------------------+
| discover-pages   |
| - scan image dir |
| - seed projection|
| - emit OCR steps |
| - emit assemble  |
+------------------+
      | emits N page steps                    | emits final step with deps
      v                                       v
+------------------+                 +-------------------+
| ocr-page N       |  ... depends -> | assemble-markdown |
| - read image     |                 | - read dep data   |
| - Geppetto OCR   |                 | - order pages     |
| - StoreArtifact  |                 | - StoreArtifact   |
| - update pages   |                 | - update run row  |
+------------------+                 +-------------------+
      |                                      |
      v                                      v
FileArtifactStore                     FileArtifactStore
(page_NNN.md)                         book.md
      |
      v
SQLiteProjectionStore("ocr-mvp")
(pages, runs, artifacts/prompt metadata)
```

### Runtime wiring diagram

```text
NewRuntime(Config{
  Store:           SQLiteStore("work/engine.db"),
  ArtifactStore:   NewFileArtifactStore("work/artifacts"),
  ProjectionStore: NewSQLiteProjectionStore("work/projections"),
  MaxWorkers:      4,
  Queues: {
    "ocr":      {MaxWorkers: 4},
    "assemble": {MaxWorkers: 1},
  },
})

RegisterPackage(ocrmvp.Package())
RegisterExecutor(discoverPagesExecutor)
RegisterExecutor(ocrPageExecutor)       // uses Geppetto OCR client
RegisterExecutor(assembleMarkdownExecutor)
StartRun("ocr-mvp", RunInput{...})
StartWorkers(ctx)
```

### Data model

#### Run input

```go
type RunInput struct {
    BookID            string   `json:"book_id"`
    ImageDir          string   `json:"image_dir"`
    PageGlob          string   `json:"page_glob,omitempty"` // default page_*.png
    StartPage         int      `json:"start_page,omitempty"`
    EndPage           int      `json:"end_page,omitempty"`
    Profile           string   `json:"profile,omitempty"`
    ProfileRegistries []string `json:"profile_registries,omitempty"`
    PromptVersion     string   `json:"prompt_version,omitempty"`
    DryRun            bool     `json:"dry_run,omitempty"`
}
```

Rules:

- `BookID` is required and should be safe for filenames.
- `ImageDir` is required and should be absolute or resolved before run start.
- If `Profile` is empty, Pinocchio defaults choose the active/default profile.
- If `ProfileRegistries` is empty, Pinocchio defaults choose registries from env/config/XDG.
- `DryRun` should skip live inference and return deterministic fake text; useful for local smoke tests.

#### Discover result

```go
type DiscoverResult struct {
    BookID     string     `json:"book_id"`
    PageCount  int        `json:"page_count"`
    OCRStepIDs []string   `json:"ocr_step_ids"`
    Pages      []PageSpec `json:"pages"`
}
```

#### Page OCR input

```go
type PageOCRInput struct {
    BookID            string   `json:"book_id"`
    PageNumber        int      `json:"page_number"`
    ImagePath         string   `json:"image_path"`
    Profile           string   `json:"profile,omitempty"`
    ProfileRegistries []string `json:"profile_registries,omitempty"`
    PromptVersion     string   `json:"prompt_version"`
    DryRun            bool     `json:"dry_run,omitempty"`
}
```

#### Page OCR result

```go
type PageOCRResult struct {
    BookID          string `json:"book_id"`
    PageNumber      int    `json:"page_number"`
    Markdown        string `json:"markdown,omitempty"` // optional; can omit if large
    MarkdownRefID   string `json:"markdown_ref_id"`
    MarkdownRefURI  string `json:"markdown_ref_uri"`
    CharCount       int    `json:"char_count"`
    PromptVersion   string `json:"prompt_version"`
    Profile         string `json:"profile,omitempty"`
}
```

The MVP can include `Markdown` for small outputs, but the safer long-term approach is to rely on `MarkdownRefID`/`MarkdownRefURI` and the artifact store. The assemble step can read artifact refs from dependency results or open files through `FileArtifactStore`.

#### Assemble result

```go
type AssembleResult struct {
    BookID        string `json:"book_id"`
    PageCount     int    `json:"page_count"`
    MarkdownRefID string `json:"markdown_ref_id"`
    MarkdownURI   string `json:"markdown_uri"`
    CharCount     int    `json:"char_count"`
}
```

## Geppetto OCR Client Design

### Interface for testability

The live OCR client should be behind an interface so workflow tests can run without hitting a model provider.

```go
type OCRClient interface {
    OCRPage(ctx context.Context, input PageOCRInput, imageBytes []byte) (OCRTextResult, error)
}

type OCRTextResult struct {
    Text            string
    ProfileSlug     string
    RegistrySlug    string
    ConfigFiles     []string
    PromptVersion   string
    ProviderMetadata map[string]string
}
```

### Live implementation using Geppetto and Pinocchio defaults

Pseudocode:

```go
type GeppettoOCRClient struct{}

func (c *GeppettoOCRClient) OCRPage(ctx context.Context, input PageOCRInput, imageBytes []byte) (OCRTextResult, error) {
    parsed, err := profilebootstrap.NewCLISelectionValues(profilebootstrap.CLISelectionInput{
        Profile:           input.Profile,
        ProfileRegistries: input.ProfileRegistries,
    })
    if err != nil { return OCRTextResult{}, err }

    resolved, err := profilebootstrap.ResolveCLIEngineSettings(ctx, parsed)
    if err != nil { return OCRTextResult{}, err }
    if resolved.Close != nil { defer resolved.Close() }

    eng, err := profilebootstrap.NewEngineFromResolvedCLIEngineSettings(resolved)
    if err != nil { return OCRTextResult{}, err }

    prompt := RenderPagePrompt(input.PageNumber, input.PromptVersion)
    turn := &turns.Turn{}
    turns.AppendBlock(turn, turns.NewSystemTextBlock(OCRSystemPrompt))
    turns.AppendBlock(turn, turns.NewUserMultimodalBlock(prompt, []map[string]any{{
        "media_type": mediaTypeFromPath(input.ImagePath),
        "content":    imageBytes,
        "detail":     "high",
    }}))

    updated, err := eng.RunInference(ctx, turn)
    if err != nil { return OCRTextResult{}, err }

    text, err := lastLLMText(updated)
    if err != nil { return OCRTextResult{}, err }

    out := OCRTextResult{
        Text:          strings.TrimSpace(text),
        PromptVersion: input.PromptVersion,
        ConfigFiles:   resolved.ConfigFiles,
    }
    if resolved.ResolvedEngineProfile != nil {
        out.ProfileSlug = resolved.ResolvedEngineProfile.ProfileSlug.String()
        out.RegistrySlug = resolved.ResolvedEngineProfile.RegistrySlug.String()
    }
    return out, nil
}
```

Important implementation notes:

- The executor should construct the OCR client once when registering executors, not on every function call if that becomes expensive. The MVP may resolve/build per page for simplicity, but it should be documented as a performance tradeoff.
- Do not use `os.Getenv("PINOCCHIO_PROFILE")` directly. Let `profilebootstrap` handle precedence.
- Do not shell out to `pinocchio`. Shelling out loses structured errors, profile resolution metadata, and testability.
- Store selected profile/registry/config metadata in the page projection so later QA can explain which model produced a transcription.

### Extracting the model output

Pseudocode:

```go
func lastLLMText(t *turns.Turn) (string, error) {
    if t == nil { return "", errors.New("nil turn") }
    blocks := turns.FindLastBlocksByKind(*t, turns.BlockKindLLMText)
    if len(blocks) == 0 { return "", errors.New("no LLM text block") }
    last := blocks[len(blocks)-1]
    text, _ := last.Payload[turns.PayloadKeyText].(string)
    if strings.TrimSpace(text) == "" { return "", errors.New("empty OCR response") }
    return text, nil
}
```

## Workflow Package Design

### Package registration

```go
func Package(cfg Config) *workflow.Package {
    return workflow.NewPackage("ocr-mvp").
        DisplayName("MVP OCR Workflow").
        Entrypoint(workflow.EntrypointFunc[RunInput](func(ctx context.Context, run *workflow.RunBuilder, input RunInput) error {
            run.Metadata("book_id", input.BookID)
            run.Metadata("image_dir", input.ImageDir)
            run.Metadata("prompt_version", defaultPromptVersion(input.PromptVersion))
            _, err := run.Step("discover-pages", input, workflow.StepOpts{
                Kind:  "ocr-mvp/discover-pages",
                Queue: "ocr-control",
                Retry: noRetry(),
            })
            return err
        })).
        Build()
}
```

### Discover pages executor

Responsibilities:

1. Validate `BookID` and `ImageDir`.
2. Resolve image files by `PageGlob` and optional page range.
3. Ensure projection schema exists.
4. Insert or reset page rows to `pending`.
5. Emit one `ocr-page` step per page.
6. Emit one `assemble-markdown` step depending on all `ocr-page` steps.
7. Return a `DiscoverResult` with emitted step IDs and page specs.

Pseudocode:

```go
func DiscoverPagesExecutor() workflow.Executor {
    return workflow.NewTypedExecutor("ocr-mvp/discover-pages", func(ctx context.Context, step *workflow.StepContext, input RunInput) error {
        input = normalizeRunInput(input)
        if err := validateRunInput(input); err != nil {
            return workflow.Permanent("ocr_input_invalid", err)
        }

        projection, err := step.Projection("ocr-mvp")
        if err != nil { return workflow.Permanent("ocr_projection_unavailable", err) }
        if err := EnsureProjectionSchema(ctx, projection); err != nil {
            return workflow.Retryable("ocr_projection_schema_failed", err)
        }

        pages, err := DiscoverPageImages(input)
        if err != nil { return workflow.Permanent("ocr_discover_pages_failed", err) }

        ocrHandles := make([]workflow.StepHandle, 0, len(pages))
        for _, page := range pages {
            stepID := model.OpID(fmt.Sprintf("ocr-page-%03d", page.PageNumber))
            _, _ = projection.Exec(ctx, `INSERT OR REPLACE INTO pages ... status='pending' ...`)
            emittedID, err := step.Emit(string(stepID), PageOCRInput{...}, workflow.StepOpts{
                Kind:  "ocr-mvp/ocr-page",
                Queue: "ocr",
                Retry: defaultOCRRetryPolicy(),
                Metadata: map[string]string{
                    "book_id": input.BookID,
                    "page": strconv.Itoa(page.PageNumber),
                    "prompt_version": input.PromptVersion,
                },
            })
            if err != nil { return err }
            ocrHandles = append(ocrHandles, workflow.StepHandle{ID: emittedID})
        }

        _, err = step.Emit("assemble-markdown", AssembleInput{BookID: input.BookID}, workflow.StepOpts{
            Kind:      "ocr-mvp/assemble-markdown",
            Queue:     "ocr-assemble",
            DependsOn: workflow.Require(ocrHandles...),
        })
        if err != nil { return err }

        return step.Result(DiscoverResult{BookID: input.BookID, PageCount: len(pages), ...})
    })
}
```

### OCR page executor

Responsibilities:

1. Mark page row `running`.
2. Read page image bytes.
3. Call the Geppetto OCR client.
4. Store page markdown through `StoreArtifact`.
5. Mark projection row `done` with artifact/profile/prompt metadata.
6. Return `PageOCRResult`.
7. Classify transient provider failures as retryable.

Pseudocode:

```go
func OCRPageExecutor(client OCRClient) workflow.Executor {
    return workflow.NewTypedExecutor("ocr-mvp/ocr-page", func(ctx context.Context, step *workflow.StepContext, input PageOCRInput) error {
        projection, err := step.Projection("ocr-mvp")
        if err != nil { return workflow.Retryable("ocr_projection_unavailable", err) }

        _, _ = projection.Exec(ctx, `UPDATE pages SET status='running', updated_at=? WHERE book_id=? AND page_num=?`, now(), input.BookID, input.PageNumber)

        imageBytes, err := os.ReadFile(input.ImagePath)
        if err != nil {
            markProjectionError(ctx, projection, input, "image_read_failed", err)
            return workflow.Permanent("ocr_image_read_failed", err)
        }

        result, err := client.OCRPage(ctx, input, imageBytes)
        if err != nil {
            markProjectionError(ctx, projection, input, "geppetto_ocr_failed", err)
            return workflow.Retryable("ocr_geppetto_failed", err)
        }

        ref, err := step.StoreArtifact(
            fmt.Sprintf("page_%03d.md", input.PageNumber),
            "text/markdown",
            []byte(result.Text),
            workflow.ArtifactKind("ocr-markdown"),
            workflow.ArtifactMetadata(map[string]string{"book_id": input.BookID, "page": strconv.Itoa(input.PageNumber)}),
        )
        if err != nil {
            markProjectionError(ctx, projection, input, "artifact_store_failed", err)
            return workflow.Retryable("ocr_artifact_store_failed", err)
        }

        _, err = projection.Exec(ctx, `UPDATE pages SET status='done', markdown_artifact_id=?, markdown_artifact_uri=?, char_count=?, prompt_version=?, updated_at=? WHERE book_id=? AND page_num=?`, ...)
        if err != nil { return workflow.Retryable("ocr_projection_update_failed", err) }

        return step.Result(PageOCRResult{BookID: input.BookID, PageNumber: input.PageNumber, MarkdownRefID: ref.ID, MarkdownRefURI: ref.URI, CharCount: len(result.Text), PromptVersion: input.PromptVersion})
    })
}
```

### Assemble markdown executor

Responsibilities:

1. Load all dependency `PageOCRResult` objects.
2. Sort by `PageNumber`.
3. Read each page markdown artifact.
4. Concatenate pages with stable page comments.
5. Store `book.md` through `StoreArtifact`.
6. Update projection run summary.

Pseudocode:

```go
func AssembleMarkdownExecutor(store workflow.ArtifactStore) workflow.Executor {
    return workflow.NewTypedExecutor("ocr-mvp/assemble-markdown", func(ctx context.Context, step *workflow.StepContext, input AssembleInput) error {
        deps := step.Step().DependsOn
        pages := []PageOCRResult{}
        for _, dep := range deps {
            var page PageOCRResult
            if err := step.DependencyData(dep.OpID, &page); err != nil {
                return workflow.Retryable("ocr_dependency_load_failed", err)
            }
            pages = append(pages, page)
        }
        sort.Slice(pages, func(i, j int) bool { return pages[i].PageNumber < pages[j].PageNumber })

        var out strings.Builder
        for _, page := range pages {
            out.WriteString(fmt.Sprintf("\n\n<!-- page:%03d -->\n\n", page.PageNumber))
            body, err := readArtifactByURI(page.MarkdownRefURI)
            if err != nil { return workflow.Retryable("ocr_artifact_read_failed", err) }
            out.Write(body)
        }

        ref, err := step.StoreArtifact(input.BookID+".md", "text/markdown", []byte(out.String()), workflow.ArtifactKind("ocr-book-markdown"))
        if err != nil { return workflow.Retryable("ocr_assemble_store_failed", err) }
        return step.Result(AssembleResult{BookID: input.BookID, PageCount: len(pages), MarkdownRefID: ref.ID, MarkdownURI: ref.URI, CharCount: out.Len()})
    })
}
```

Implementation note: `StepContext` does not currently expose `ArtifactStore.Open`. For MVP, either parse `file://` artifact URIs directly when using `FileArtifactStore`, or pass the store into the assemble executor constructor. Passing the store is cleaner and keeps the executor testable.

## Prompt Design

Use a single universal prompt based on the AITR v2 script, but store it in Go with a stable `PromptVersion`.

MVP prompt requirements:

- Output only markdown.
- Preserve headings, paragraphs, footnotes, citations, math, tables, and code.
- Emit `[IMAGE: single-line description]` for figures/diagrams/screenshots.
- Do not include standalone page numbers.
- Do not invent content.
- Do not duplicate text.

Example:

```text
You are transcribing a scanned book/report page into clean markdown.

Rules:
1. Output only markdown. No commentary.
2. Preserve headings, paragraphs, footnotes, citations, math, code, and tables.
3. If the page is blank, output an empty string.
4. If an image/figure/diagram appears, insert exactly one single-line marker:
   [IMAGE: concise description of what the figure shows]
5. Do not include standalone page numbers.
6. Do not duplicate text.
7. Do not add text that is not visible on the page.

Page number: {{.PageNumber}}
Book ID: {{.BookID}}
```

Do not put provider names or model names into the prompt. The profile controls the provider/model.

## Implementation Phases

### Phase 1: Package skeleton and fake OCR path

Files:

- `scraper/pkg/workflows/ocrmvp/types.go`
- `scraper/pkg/workflows/ocrmvp/prompt.go`
- `scraper/pkg/workflows/ocrmvp/package.go`
- `scraper/pkg/workflows/ocrmvp/projection.go`
- `scraper/pkg/workflows/ocrmvp/executors.go`
- `scraper/pkg/workflows/ocrmvp/package_test.go`

Build:

- Define inputs/results.
- Define package registration.
- Implement `discover-pages`, `ocr-page`, and `assemble-markdown` with a fake OCR client.
- Use a temp image directory and fake text in tests.
- Assert projection rows and artifact refs.

Validation:

```bash
cd /home/manuel/workspaces/2026-05-20/book-ocr/scraper
go test ./pkg/workflows/ocrmvp ./pkg/workflow -count=1
```

### Phase 2: Geppetto OCR client

Files:

- `scraper/pkg/workflows/ocrmvp/geppetto_ocr.go`
- `scraper/go.mod`
- maybe `scraper/go.work` remains unchanged because the root workspace already includes `geppetto` and `pinocchio`.

Build:

- Add imports for Geppetto turns/inference and Pinocchio profile bootstrap.
- Add a `GeppettoOCRClient` implementing `OCRClient`.
- Resolve engine through `profilebootstrap.ResolveCLIEngineSettings`.
- Use `turns.NewUserMultimodalBlock` with image bytes.
- Extract the last LLM text block.
- Store selected profile/registry/config metadata.

Validation:

- Unit test with fake engine/client only.
- Optional live test behind an explicit environment flag such as `OCR_MVP_LIVE=1`.

Do not make normal tests require API keys or a live model provider.

### Phase 3: CLI or command wiring

Options:

1. Add a `scraper ocr-mvp run` command under `scraper/pkg/cmd`.
2. Add a small example binary under `scraper/cmd/ocr-mvp`.
3. Add a separate application command in `2026-05-20--book-ocr` that imports `scraper/pkg/workflows/ocrmvp`.

Recommended first implementation: option 2, a small example binary, because it keeps the MVP isolated.

CLI shape:

```bash
go run ./cmd/ocr-mvp \
  --book-id aitr-794 \
  --image-dir ../output/books/presentation-based-uis/pages \
  --work-dir /tmp/ocr-mvp \
  --profile gpt-5-nano-low \
  --max-workers 4
```

Flags:

- `--book-id`
- `--image-dir`
- `--page-glob` default `page_*.png`
- `--start-page`
- `--end-page`
- `--work-dir`
- `--profile`
- `--profile-registries` repeatable or comma-separated
- `--max-workers` default `4`
- `--dry-run`

Important profile behavior:

- `--profile` maps to `profilebootstrap.CLISelectionInput.Profile`.
- `--profile-registries` maps to `profilebootstrap.CLISelectionInput.ProfileRegistries`.
- If both are omitted, Pinocchio defaults apply.

### Phase 4: Operator smoke flows

Add a short README or test helper documenting:

```go
_ = rt.RetryStep(ctx, runID, "ocr-page-047")
_ = rt.CancelRun(ctx, runID)
```

For MVP this can be code-level documentation; a full UI can come later.

## Testing Strategy

### Unit tests

- `DiscoverPageImages` handles page ordering and range filtering.
- `RenderPagePrompt` includes prompt version and page number.
- `EnsureProjectionSchema` is idempotent.
- `lastLLMText` extracts the final LLM text block and errors on empty output.
- `GeppettoOCRClient` profile wiring can be tested with a fake engine factory if the implementation introduces an injectable factory.

### Workflow integration tests with fake OCR

Use `workflow.NewRuntime` with temp stores:

```go
rt, err := workflow.NewRuntime(ctx, workflow.Config{
    Store:           workflow.SQLiteStore(filepath.Join(tmp, "engine.db")),
    ArtifactStore:   workflow.NewFileArtifactStore(filepath.Join(tmp, "artifacts")),
    ProjectionStore: workflow.NewSQLiteProjectionStore(filepath.Join(tmp, "projections")),
    MaxWorkers:      4,
    Queues: map[model.QueueKey]workflow.QueueConfig{
        "ocr-control":  {MaxWorkers: 1},
        "ocr":          {MaxWorkers: 4},
        "ocr-assemble": {MaxWorkers: 1},
    },
})
```

Then:

1. Create three fake page image files.
2. Register `ocr-mvp` package with fake OCR client.
3. Start a run.
4. Call `StartWorkers(ctx, workflow.WithWorkerMaxCycles(10))` or loop `RunOnce`.
5. Assert:
   - workflow succeeded;
   - three page projection rows are `done`;
   - page artifacts exist;
   - final combined markdown artifact exists;
   - results contain ordered pages.

### Live OCR smoke test

Make live provider tests opt-in:

```bash
OCR_MVP_LIVE=1 go test ./pkg/workflows/ocrmvp -run TestLiveGeppettoOCR -count=1
```

The test should skip unless `OCR_MVP_LIVE=1`. It should process one tiny fixture image and require a configured Pinocchio profile registry.

### Manual validation

```bash
# dry run, no provider calls
cd /home/manuel/workspaces/2026-05-20/book-ocr/scraper
go run ./cmd/ocr-mvp \
  --book-id smoke \
  --image-dir /path/to/pages \
  --work-dir /tmp/ocr-mvp-smoke \
  --dry-run \
  --max-workers 2

# live run, default Pinocchio registry/profile resolution
go run ./cmd/ocr-mvp \
  --book-id aitr-794 \
  --image-dir /home/manuel/workspaces/2026-05-20/book-ocr/output/books/presentation-based-uis/pages \
  --work-dir /tmp/ocr-mvp-aitr \
  --profile gpt-5-nano-low \
  --max-workers 4
```

## Risks, Tradeoffs, and Open Questions

### Risk: importing Pinocchio into scraper may be heavy

Using `pinocchio/pkg/cmds/profilebootstrap` is the correct way to inherit Pinocchio defaults, but it means `scraper` imports `pinocchio`. If that dependency is too heavy, an alternative is to create a tiny shared package in Pinocchio or Geppetto that exposes only profile bootstrap. Do not reimplement profile resolution locally unless that refactor happens.

### Risk: building one engine per page may be expensive

The simplest implementation resolves the profile and builds an engine in every `ocr-page` executor call. That is correct but may be inefficient. A later optimization can cache an engine/client per worker process keyed by `(profile, registry sources, prompt version)`. Do not cache until the MVP works and tests pass.

### Risk: external artifact refs are not yet first-class downloads

`StoreArtifact` writes large bytes externally but persists a JSON reference artifact in the engine DB. Existing artifact APIs see the reference JSON, not the external markdown directly. That is acceptable for the MVP, but the follow-up should teach artifact readers to dereference `external-artifact-ref`.

### Risk: assemble step needs to read artifacts

`StepContext` can write external artifacts but does not expose the store for reading. Pass the artifact store into the assemble executor constructor, or read `file://` URIs in the MVP. Passing the store is cleaner.

### Risk: retry does not automatically clear old projection errors

If a page fails, then succeeds after retry, the OCR executor must explicitly clear `error_code` and `error_message` in the projection row when marking the page `done`.

### Open question: should page markdown be duplicated in result data?

For small pages, duplicating markdown in `PageOCRResult.Markdown` simplifies assembly. For large pages, it increases SQLite result size. Recommended MVP compromise: store `Markdown` only in dry-run/test mode or keep it optional; always store artifact refs.

### Open question: should PDF rendering become a step?

Eventually yes, but not in this MVP. Keeping rendering out of scope lets this ticket validate workflow runtime behavior first.

## Implementation Checklist for the Intern

1. Read this document fully.
2. Read these files:
   - `scraper/pkg/workflow/runtime.go`
   - `scraper/pkg/workflow/package.go`
   - `scraper/pkg/workflow/context.go`
   - `scraper/pkg/workflow/artifact_store.go`
   - `scraper/pkg/workflow/projection_store.go`
   - `scraper/pkg/workflow/operators.go`
   - `pinocchio/pkg/cmds/profilebootstrap/profile_selection.go`
   - `pinocchio/pkg/cmds/profilebootstrap/engine_settings.go`
   - `geppetto/pkg/turns/helpers_blocks.go`
3. Implement Phase 1 with fake OCR and tests.
4. Run `go test ./pkg/workflows/ocrmvp ./pkg/workflow -count=1`.
5. Implement Geppetto OCR client with Pinocchio profile bootstrap.
6. Run full scraper tests: `go test ./... -count=1`.
7. Add CLI/example command.
8. Run dry-run CLI smoke test.
9. Optionally run one live page with `--profile` and configured registries.
10. Update diary/changelog/tasks.

## References

### Scraper workflow API

- `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/workflow/runtime.go`
- `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/workflow/package.go`
- `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/workflow/context.go`
- `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/workflow/executor.go`
- `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/workflow/errors.go`
- `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/workflow/artifact_store.go`
- `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/workflow/projection_store.go`
- `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/workflow/operators.go`
- `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/workflow/runtime_test.go`

### Geppetto and Pinocchio integration

- `/home/manuel/workspaces/2026-05-20/book-ocr/geppetto/pkg/turns/helpers_blocks.go`
- `/home/manuel/workspaces/2026-05-20/book-ocr/geppetto/pkg/doc/topics/06-inference-engines.md`
- `/home/manuel/workspaces/2026-05-20/book-ocr/geppetto/pkg/doc/topics/08-turns.md`
- `/home/manuel/workspaces/2026-05-20/book-ocr/pinocchio/pkg/cmds/profilebootstrap/profile_selection.go`
- `/home/manuel/workspaces/2026-05-20/book-ocr/pinocchio/pkg/cmds/profilebootstrap/engine_settings.go`
- `/home/manuel/workspaces/2026-05-20/book-ocr/pinocchio/README.md`

### Prior OCR script reference

- `/home/manuel/workspaces/2026-05-20/book-ocr/claw-stuff/ttmp/2026/05/20/AITR-794--clean-markdown-transcription-of-aitr-794-pdf/scripts/01-extract-pages.py`
- `/home/manuel/workspaces/2026-05-20/book-ocr/claw-stuff/ttmp/2026/05/20/AITR-794--clean-markdown-transcription-of-aitr-794-pdf/scripts/03-ocr-batch.py`
- `/home/manuel/workspaces/2026-05-20/book-ocr/claw-stuff/ttmp/2026/05/20/AITR-794--clean-markdown-transcription-of-aitr-794-pdf/scripts/10-reprocess-universal.py`
- `/home/manuel/workspaces/2026-05-20/book-ocr/claw-stuff/ttmp/2026/05/20/AITR-794--clean-markdown-transcription-of-aitr-794-pdf/reference/01-diary.md`
