---
Title: Scraper Task Workflow Job System Design and Book OCR Mapping
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
    - Path: claw-stuff/ttmp/2026/05/20/AITR-794--clean-markdown-transcription-of-aitr-794-pdf/reference/01-diary.md
      Note: Concrete OCR implementation diary that refined the book OCR mapping
    - Path: claw-stuff/ttmp/2026/05/20/AITR-794--clean-markdown-transcription-of-aitr-794-pdf/scripts/03-ocr-batch.py
      Note: SQLite page queue
    - Path: claw-stuff/ttmp/2026/05/20/AITR-794--clean-markdown-transcription-of-aitr-794-pdf/scripts/04-ocr-dashboard.py
      Note: Concrete operator progress dashboard model
    - Path: claw-stuff/ttmp/2026/05/20/AITR-794--clean-markdown-transcription-of-aitr-794-pdf/scripts/10-reprocess-universal.py
      Note: Universal prompt and prompt-version/page-type OCR pass
    - Path: claw-stuff/ttmp/2026/05/20/AITR-794--clean-markdown-transcription-of-aitr-794-pdf/scripts/11-postprocess.py
      Note: Final post-processing and assembly pass
    - Path: scraper/pkg/engine/model/types.go
      Note: Core generic workflow/op/lease/retry/queue/artifact model
    - Path: scraper/pkg/engine/scheduler/scheduler.go
      Note: Current durable poll/lease/execute scheduler and event source
    - Path: scraper/pkg/engine/store/sqlite/lease_store.go
      Note: SQLite queue policy and leasing backend behavior
    - Path: scraper/pkg/engine/store/sqlite/result_store.go
      Note: Completion
    - Path: scraper/pkg/sites/submitverbs/host.go
      Note: Submission path that creates workflows and initial ops
    - Path: scraper/sites/nereval/scripts/extract_list.js
      Note: Existing fan-out workflow pattern used as OCR mapping evidence
    - Path: scraper/web/src/api/workflowApi.ts
      Note: Existing frontend workflow/admin API surface for retry/cancel/artifacts
ExternalSources:
    - sources/river/01-river-getting-started.md
    - sources/river/02-river-scheduled-jobs.md
    - sources/river/03-river-periodic-jobs.md
    - sources/river/04-river-unique-jobs.md
    - sources/river/05-river-job-retries.md
    - sources/river/06-river-multiple-queues.md
Summary: Analysis and implementation guide for evolving scraper into a general task/workflow/job system and mapping a book OCR pipeline onto it.
LastUpdated: 2026-05-24T17:10:00-04:00
WhatFor: Use this when planning a generic workflow/job runtime from the scraper engine or onboarding an intern to implement a book OCR workflow on top of it.
WhenToUse: Before renaming abstractions, adding new backends such as River, or building production/admin surfaces for non-scraping workloads.
---



# Scraper Task Workflow Job System Design and Book OCR Mapping

## Executive summary

The current `scraper` repository is already much closer to a general-purpose durable workflow system than its name suggests. Its core primitives are generic: a `WorkflowRun` contains durable `OpSpec` records, each op has dependencies, retry state, a queue key, an execution kind, leases, results, artifacts, and emitted child ops. The scraping-specific parts mostly live in naming (`site`, `scripts`, `verbs`) and in the default runners (`js`, `http/fetch`) rather than in the scheduler or store.

The recommended path is therefore **not** to replace the scheduler immediately. The recommended path is:

1. Treat the current SQLite scheduler as the first backend for a generic `workflow/job` runtime.
2. Introduce clearer names and interfaces around the existing concepts without breaking the working scraper behavior.
3. Build a book OCR workflow as the first non-scraping workload, using custom runners and/or JS scripts to prove the general model.
4. Add production/admin features that book OCR exposes sharply: long-running task heartbeats, cooperative cancellation, stage-level progress, large artifact storage, operator runbooks, and worker capability routing.
5. Add a backend abstraction that can later target River/Postgres if production needs distributed scheduling, stronger multi-process behavior, transactional enqueueing with Postgres application data, or River UI/operations features.

River is a plausible future backend for the **job queue** portion, especially if the system moves to Postgres and multi-node workers. River provides Postgres-backed workers, queues, retries, scheduled jobs, unique jobs, and periodic jobs. However, River is not a full workflow DAG engine by itself. The scraper engine already owns workflow graphs, dependencies, artifacts, and per-workflow admin operations. A River integration should be designed as a backend adapter for leasing/executing ready ops, not as a wholesale rewrite of the workflow model.

After studying the concrete AITR-794 OCR ticket, the main update is that book OCR should be modeled as an **iterative page-processing campaign**, not only as a one-shot `split -> OCR all pages -> aggregate` DAG. The real workflow used multiple prompt versions, selective reprocessing passes for image/table/code pages, a SQLite page-status queue, tmux workers, a live page-grid dashboard, filesystem artifacts, and post-processing/build scripts. That reinforces the recommendation to keep the current scheduler for now, but it adds stronger requirements for prompt/version provenance, page attempts, subset reprocessing, stale-job reset, and operator dashboards.

## Problem statement and scope

The user request is to create a docmgr ticket and design package for evolving `@scraper/` into a more general-purpose task/workflow/job system. The concrete proving example is a book OCR workflow like `@2026-05-20--book-ocr/`. The design must explain the current system deeply enough for a new intern, map a book OCR workload onto it, and cover production/admin concerns: monitoring, restarting, signaling, user-visible progress, and debugging.

This document focuses on design and implementation guidance. It does not change runtime code yet.

### Observed scope boundaries

- `scraper/` contains the mature workflow engine, runtime events, metrics, API, frontend, site manifest system, default JS/HTTP runners, and production-ish monitoring scaffolding.
- `2026-05-20--book-ocr/` is currently a Go template stub rather than a finished OCR app. Its `cmd/XXX/main.go` contains only an empty `main` function, its `go.mod` still declares `github.com/go-go-golems/XXX`, and its README is template ASCII art.
- The concrete OCR implementation evidence lives in `/home/manuel/workspaces/2026-05-20/book-ocr/claw-stuff/ttmp/2026/05/20/AITR-794--clean-markdown-transcription-of-aitr-794-pdf/`. It is Python/script-based rather than Go-based: PyMuPDF page rendering, Pinocchio VLM OCR, SQLite page tracking, tmux workers, an ANSI dashboard, prompt reprocessing passes, figure extraction, and final markdown assembly. The OCR mapping below should therefore be read as a design to port/generalize that proven workflow into the scraper runtime.

## Terms for the intern

The current code uses scraper-oriented language. The generic system can preserve behavior while exposing clearer generic names.

| Current term | Generic term | Meaning |
|---|---|---|
| Site | Workflow package / domain | A bundle of manifests, scripts, migrations, submit verbs, and optional modules. For OCR this might be `book-ocr`. |
| Submit verb | Workflow entrypoint / job submitter | Code that validates user input and creates initial durable work. |
| Workflow | Workflow / job run | One user-visible run with ID, status, input, metadata, and many tasks. |
| Op | Task / step / node | One durable executable unit in the workflow graph. |
| Runner kind | Executor / task handler | The implementation that knows how to execute an op, such as JS, HTTP fetch, OCR, PDF split, or LLM cleanup. |
| Queue key | Work queue / capability lane | A routing and throttling key such as `site:nereval:http`, `ocr:cpu`, or `ocr:gpu`. |
| Site DB | Projection DB / domain DB | Per-domain read model for outputs that users query. |
| Engine DB | Runtime DB | Durable workflow state: workflows, ops, leases, results, artifacts, queue limits. |
| Runtime event | Operator/user event | Structured event stream for UI and logs. |

## Current-state architecture, with evidence

### The engine model is already generic

The central model in `pkg/engine/model/types.go` defines workflow and op states, retry policy, queue policy, leases, op specs, records, artifacts, errors, and results. The fields are not inherently scraping-specific except for `SiteName`.

Key evidence:

- Workflow statuses are generic: pending, running, succeeded, failed, canceled (`pkg/engine/model/types.go:27-33` for op status and `WorkflowStatus` nearby).
- `WorkflowRun` has ID, site, name, status, input, metadata, and timestamps (`pkg/engine/model/types.go:44-53`).
- `OpSpec` has workflow ID, parent ID, kind, queue, dedup key, input, dependencies, retry, metadata, and readiness timestamps (`pkg/engine/model/types.go:124-135`).
- Queue policy already supports max in-flight and token-bucket rate limiting (`pkg/engine/model/types.go:78-115`).
- A lease has worker ID, token, acquisition time, and expiry (`pkg/engine/model/types.go:117-122`).

A generic task runtime can keep these semantics almost unchanged. The first refactor should be naming/API layering, not data model replacement.

### The scheduler is a durable poll/lease/execute loop

The scheduler configuration is simple: max workers, poll interval, and default lease duration. It validates those values and emits structured scheduler events such as workflow created, op leased, op succeeded, op retried, op failed, queue rate limited, and idle (`pkg/engine/scheduler/scheduler.go:18-75`).

The high-level scheduler cycle is:

```text
RunOnce(now)
  -> RefreshRunnableOps(now)
  -> ListQueueCandidates(now)
  -> for each queue candidate, up to MaxWorkers:
       -> apply queue policy
       -> LeaseReadyOp(workerID, queue, site, policy, lease duration)
       -> execute leased op using runner registry
       -> CompleteOp or FailOp
  -> refresh workflow status for touched workflows
```

This is visible in `RunOnce`: it refreshes runnable ops, lists queue candidates, emits idle if no work is ready, and iterates queue candidates up to `MaxWorkers` (`pkg/engine/scheduler/scheduler.go:197-226`). After a lease is acquired it calls `executeLeasedOp` (`pkg/engine/scheduler/scheduler.go:240-288` in the file; line snippets collected during investigation show the lease call and event emission).

`executeLeasedOp` loads the workflow, finds a runner by op kind, resolves the site DB, calls the runner with workflow/op/lease/dependency context, and persists success or failure (`pkg/engine/scheduler/scheduler.go:303-380`). This is the exact seam where new OCR runners would plug in.

### The SQLite store is the current scheduling backend

The engine DB schema stores workflows and ops in `001_engine_core.sql`, and dependencies, leases, queue limiter state, results, and artifacts in `002_engine_runtime.sql`.

```text
workflows
ops
op_dependencies
leases
queue_limit_state
results
artifacts
```

The lease path is transactionally implemented in SQLite. `LeaseReadyOp` begins a transaction, normalizes queue policy, checks active leases, refills/consumes token-bucket state, selects one ready op by site and queue, writes a lease, and marks the op running (`pkg/engine/store/sqlite/lease_store.go:14-162`).

Completion is also transactional. `CompleteOp` persists the result row, persists artifacts, inserts emitted child ops, deletes the lease by token, and marks the original op succeeded (`pkg/engine/store/sqlite/result_store.go:17-95`). This is important for OCR because page OCR tasks will emit artifacts such as text, hOCR, per-page logs, and page images.

### The current runtime split is submit-time vs execution-time JS

The existing docs explain the key conceptual split: submit verbs create initial durable work, while execution scripts run later through the worker. The doc states that submit verbs are not workers and do not crawl for minutes; their job is to emit the initial durable graph (`pkg/doc/topics/scraper-runtime-model.md:25-36`). Execution scripts can read input, inspect workflow/op metadata, read dependency results, emit child ops, write records/artifacts, and use site/scraper DB modules (`pkg/doc/topics/scraper-runtime-model.md:42-57`).

This is a good model for OCR:

- The submit verb should not OCR the book inline.
- The submit verb should create the workflow and first tasks: ingest, split, classify pages, or load a manifest.
- Workers should process page-level and aggregation tasks durably.

### The submit host already creates workflows through a normal service path

`Host.Submit` opens the engine DB, opens the scraper DB, migrates the site DB, creates a workflow, executes the submit verb, creates a scheduler, and calls `CreateWorkflow` with the workflow and initial ops (`pkg/sites/submitverbs/host.go:63-147`).

The submit JS context exposes values, sections, command metadata, workflow metadata, logging, `ctx.emit`, `ctx.setTargetOpID`, `ctx.setWorkflowName`, and `ctx.setWorkflowMetadata` (`pkg/sites/submitverbs/runtime.go:172-238`). This is enough to expose an OCR CLI/API submitter such as:

```bash
scraper --sites-manifest-dir ./sites site book-ocr run ingest \
  --book-id poetics-001 \
  --source /data/uploads/poetics.pdf \
  --language eng \
  --ocr-engine tesseract
```

### Site manifests are already package manifests

`pkg/sites/manifest/manifest.go` defines a site manifest with name, database file, scripts root, verbs root, SQL/JS migrations, help root, modules, and queue policies (`pkg/sites/manifest/manifest.go:5-28`). This can be generalized into a `workflowPackage.yaml` or `domain.yaml` later, but the fields already match an extensible package model.

For OCR, a manifest might look like:

```yaml
name: book-ocr
databaseFileName: book_ocr.db
scriptsRoot: scripts
verbsRoot: verbs
sqlMigrationsRoot: migrations
queuePolicies:
  - queue: ocr:io
    maxInFlight: 4
  - queue: ocr:cpu
    maxInFlight: 2
  - queue: ocr:gpu
    maxInFlight: 1
  - queue: ocr:llm
    maxInFlight: 1
    rateLimit:
      kind: token_bucket
      ratePerSecond: 0.2
      burst: 2
```

### The NEREVAL site is the best existing pattern for fan-out workflows

NEREVAL shows dynamic graph growth and dependency wiring. Its `extract_list.js` writes normalized records into a site DB, emits detail fetch/extract tasks for each property, and emits the next list page fetch/extract pair if pagination continues (`sites/nereval/scripts/extract_list.js:65-165`).

The OCR equivalent is page fan-out:

```text
split_pdf
  -> emits render_page_001, render_page_002, ...
render_page_001
  -> emits ocr_page_001
ocr_page_001
  -> emits cleanup_page_001
all cleanup_page_N
  -> aggregate_book
```

The current engine supports this style because completed ops may insert emitted child ops inside `CompleteOp` (`pkg/engine/store/sqlite/result_store.go:72-75`).

### Admin, API, and UI surfaces already exist

The API exposes workflow, op, artifact, result, queue, cancel, retry, and metrics routes (`pkg/api/server/routes_engine.go:11-35`). The engine view mutation service can retry a failed op by setting it back to ready, and cancel a workflow by marking pending/ready/running ops canceled and deleting leases (`pkg/services/engineview/workflow_mutation_service.go:21-78`).

The frontend RTK Query API consumes these surfaces: list workflows, workflow detail, workflow ops, op result, workflow artifacts, workflow results, op artifacts, retry op, and cancel workflow (`web/src/api/workflowApi.ts:27-149`). The workflow table shows ID, site, name, status, progress, and created time.

For OCR, this means the existing UI can already show:

- one row per book OCR workflow;
- op progress as completed/total;
- failed page OCR ops;
- artifacts such as logs, page images, OCR text, and final export files;
- retry/cancel buttons.

The missing OCR-specific UX is stage-level progress and user-friendly book/page labels.

### Metrics and runtime events are production-oriented already

`pkg/metrics/metrics.go` defines counters/histograms/gauges for HTTP requests, submissions, scheduler cycles, ops leased/completed/failed/retried, queue rate limiting, queue wait duration, op duration, HTTP runner requests, and worker liveness (`pkg/metrics/metrics.go:54-170`). Prometheus rules define recording rules and alerts for worker down, API unavailable, repeated rate limiting, elevated op failure rate, high queue wait, and retry spikes (`ops/monitoring/prometheus/rules/scraper.yml:1-56`).

Runtime events support multiple backends: off, in-process gochannel, and Redis Streams (`pkg/runtimeevents/backend.go:17-87`). Redis is useful when separate API and worker processes need to feed a shared event stream to the UI.

## Book OCR mapping

### Concrete AITR-794 OCR findings that affect the design

The AITR-794 ticket is the best concrete model for the first book OCR workflow. Its diary and scripts show a practical, operator-driven process:

- `01-extract-pages.py` rendered a 202-page scanned PDF into page PNGs at 200 DPI with PyMuPDF.
- `02-ocr-pages.py` ran `PINOCCHIO_PROFILE=gpt-5-nano-low pinocchio code professional --images <page> <prompt>` and stripped Pinocchio thinking/output markers.
- `03-ocr-batch.py` introduced a SQLite `pages` table with `status`, `started_at`, `finished_at`, `output_path`, `char_count`, `error_msg`, and `attempts` so page OCR could restart and resume.
- `04-ocr-dashboard.py` gave operators a live page-grid dashboard with counts, running pages, recent completions, throughput, ETA, and errors.
- The first queue claim implementation had a race; the fix was `BEGIN IMMEDIATE` around `SELECT pending page -> UPDATE running -> COMMIT`.
- Later scripts reprocessed selected subsets: image-heavy pages, table pages, universal prompt v2, and post-processing. The final universal prompt handled blank pages, full-image pages, mixed text/image pages, tables, code, math, and duplicate-text avoidance.
- Final outputs were filesystem artifacts: per-page markdown versions, page PNGs, figure PNGs, prompt-specific directories, SQLite progress DBs, and one assembled `aitr-794.md`.

This changes the OCR design emphasis. The system should not only know that page 47 has an OCR task; it should know **which prompt version/engine/profile produced which attempt**, which page type was inferred, which artifacts came out, and whether an operator chose to reprocess a page subset with a refined prompt.

### Target OCR workflow

A practical OCR workflow has these stages:

```text
User submits source PDF/images
  -> ingest_source
  -> render_pages
  -> initial page OCR with prompt version v1 or universal-v2
  -> classify/record page type and artifacts
  -> operator QA dashboard
  -> optional subset reprocessing for images/tables/code/blank/error pages
  -> postprocess cross-page breaks and normalize markdown
  -> aggregate book text
  -> generate outputs (markdown, PDF, txt, hOCR, ALTO, EPUB)
  -> QA summary
```

ASCII DAG:

```text
                    +----------------+
                    | ingest_source  |
                    +--------+-------+
                             |
                             v
                    +----------------+
                    | inspect_source |
                    +--------+-------+
                             |
                             v
                    +--------------------+
                    | split/render pages |
                    +----+----------+----+
                         |          |
             +-----------+          +-----------+
             v                                  v
    +----------------+                 +----------------+
    | preprocess p1  |      ...        | preprocess pN  |
    +-------+--------+                 +-------+--------+
            |                                  |
            v                                  v
    +----------------+                 +----------------+
    | OCR page 1     |      ...        | OCR page N     |
    +-------+--------+                 +-------+--------+
            |                                  |
            v                                  v
    +----------------+                 +----------------+
    | cleanup page 1 |      ...        | cleanup page N |
    +-------+--------+                 +-------+--------+
            |                                  |
            +----------------+-----------------+
                             v
                    +----------------+
                    | aggregate_book |
                    +--------+-------+
                             |
                             v
                    +----------------+
                    | export_outputs |
                    +----------------+
```

### OCR package layout

Under the current manifest system, create a new package/site:

```text
sites/bookocr/
  site.yaml
  migrations/
    001_init.sql
  verbs/
    ingest.js
  scripts/
    ingest_source.js
    inspect_source.js
    split_pdf.js
    preprocess_page.js
    ocr_page.js
    cleanup_page.js
    aggregate_book.js
    export_outputs.js
    lib/ids.js
    lib/artifacts.js
```

The site DB should hold queryable, user-facing OCR state:

```sql
CREATE TABLE books (
  book_id TEXT PRIMARY KEY,
  title TEXT,
  source_uri TEXT NOT NULL,
  status TEXT NOT NULL,
  page_count INTEGER,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE TABLE pages (
  book_id TEXT NOT NULL,
  page_number INTEGER NOT NULL,
  status TEXT NOT NULL,
  image_artifact_id TEXT,
  ocr_text_artifact_id TEXT,
  hocr_artifact_id TEXT,
  confidence REAL,
  error_code TEXT,
  updated_at TEXT NOT NULL,
  PRIMARY KEY (book_id, page_number)
);

CREATE TABLE page_attempts (
  book_id TEXT NOT NULL,
  page_number INTEGER NOT NULL,
  attempt INTEGER NOT NULL,
  prompt_version TEXT NOT NULL,
  engine TEXT NOT NULL,
  engine_profile TEXT,
  status TEXT NOT NULL,
  page_type TEXT,
  output_artifact_id TEXT,
  char_count INTEGER,
  error_code TEXT,
  error_message TEXT,
  started_at TEXT,
  finished_at TEXT,
  PRIMARY KEY (book_id, page_number, attempt)
);

CREATE TABLE prompts (
  prompt_version TEXT PRIMARY KEY,
  prompt_hash TEXT NOT NULL,
  prompt_text TEXT NOT NULL,
  created_at TEXT NOT NULL,
  notes TEXT
);

CREATE TABLE exports (
  book_id TEXT NOT NULL,
  kind TEXT NOT NULL,
  artifact_id TEXT NOT NULL,
  created_at TEXT NOT NULL,
  PRIMARY KEY (book_id, kind)
);
```

`page_attempts` and `prompts` are the concrete update from the AITR-794 evidence. The real pipeline improved by changing prompts and reprocessing subsets; without explicit prompt/attempt provenance, production debugging would be guesswork.

### Submit verb sketch

`sites/bookocr/verbs/ingest.js` should validate user input and emit one initial op.

```javascript
doc(`Submit a durable book OCR workflow.`);

__verb__("ingest", {
  short: "Submit a book OCR workflow",
  fields: {
    book_id: { type: "string", help: "Stable book ID" },
    source_uri: { type: "string", help: "PDF path, image folder, or object-store URI" },
    language: { type: "string", default: "eng", help: "OCR language code" },
    ocr_engine: { type: "string", default: "tesseract", help: "OCR backend" }
  }
});

function ingest(ctx) {
  const v = ctx.values || {};
  const workflowID = String(ctx.workflow.id);
  const bookID = String(v.book_id || workflowID);

  ctx.setWorkflowName("OCR " + bookID);
  ctx.setWorkflowMetadata({
    domain: "book-ocr",
    bookID: bookID,
    sourceURI: String(v.source_uri || ""),
    language: String(v.language || "eng"),
    ocrEngine: String(v.ocr_engine || "tesseract")
  });

  const ingestID = workflowID + ":ingest";
  ctx.emit({
    id: ingestID,
    kind: "js",
    queue: "ocr:io",
    dedupKey: "book-ocr:ingest:" + bookID,
    metadata: { script: "ingest_source.js" },
    input: {
      bookID,
      sourceURI: String(v.source_uri || ""),
      language: String(v.language || "eng"),
      ocrEngine: String(v.ocr_engine || "tesseract")
    },
    retry: { maxAttempts: 3, backoffKind: "exponential", initialBackoff: "5s", maxBackoff: "5m" }
  });

  ctx.setTargetOpID(ingestID);
  return { data: { bookID, initialOpID: ingestID } };
}
```

### Dynamic page fan-out sketch

`split_pdf.js` should produce page records and emit one preprocessing task per page.

```javascript
const siteDB = require("site-db");

module.exports = function(ctx) {
  const bookID = String(ctx.input.bookID);
  const pageCount = Number(ctx.input.pageCount || 0);

  siteDB.exec("UPDATE books SET page_count = ?, updated_at = ? WHERE book_id = ?", pageCount, ctx.now, bookID);

  const cleanupIDs = [];
  for (let page = 1; page <= pageCount; page++) {
    const renderID = `${ctx.workflow.id}:page:${page}:render`;
    const ocrID = `${ctx.workflow.id}:page:${page}:ocr`;
    const cleanupID = `${ctx.workflow.id}:page:${page}:cleanup`;

    siteDB.exec(
      "INSERT OR REPLACE INTO pages(book_id, page_number, status, updated_at) VALUES(?, ?, ?, ?)",
      bookID, page, "queued", ctx.now
    );

    ctx.emit({
      id: renderID,
      kind: "pdf/render-page",
      queue: "ocr:cpu",
      dedupKey: `book-ocr:${bookID}:render:${page}`,
      input: { bookID, page, sourceArtifactID: ctx.input.sourceArtifactID }
    });

    ctx.emit({
      id: ocrID,
      kind: "ocr/page",
      queue: "ocr:gpu",
      dependsOn: [{ opID: renderID, required: true }],
      dedupKey: `book-ocr:${bookID}:ocr:${page}`,
      input: { bookID, page, language: ctx.input.language, imageOpID: renderID }
    });

    ctx.emit({
      id: cleanupID,
      kind: "js",
      queue: "ocr:llm",
      dependsOn: [{ opID: ocrID, required: true }],
      metadata: { script: "cleanup_page.js" },
      input: { bookID, page, ocrOpID: ocrID }
    });

    cleanupIDs.push(cleanupID);
  }

  ctx.emit({
    id: `${ctx.workflow.id}:aggregate`,
    kind: "js",
    queue: "ocr:io",
    dependsOn: cleanupIDs.map(opID => ({ opID, required: true })),
    metadata: { script: "aggregate_book.js" },
    input: { bookID, cleanupIDs }
  });

  return { data: { bookID, pageCount, cleanupIDs } };
};
```

### Runner choices for OCR

There are three implementation options for OCR execution:

1. **JS script calls external tools through a new host module.** Add a restricted `require("process")` or `require("exec")` module for controlled subprocess execution. This is fastest to prototype but needs careful sandboxing and cancellation.
2. **Native Go runners.** Add runner kinds such as `pdf/render-page`, `ocr/page`, and `ocr/export`. These can call Go libraries or subprocesses and persist structured artifacts. This is better for production reliability.
3. **External worker bridge.** Add a generic runner that submits to another service (GPU worker, cloud OCR, LLM OCR), then polls or receives callbacks. This is best when OCR needs expensive hardware or provider APIs.

Recommended first production path:

- implement native Go runners for `pdf/render-page` and `ocr/page`;
- keep aggregation/cleanup in JS until the workflow shape stabilizes;
- add an external provider interface inside `ocr/page` so Tesseract, cloud OCR, and LLM OCR can be swapped by configuration.

Go API sketch:

```go
type OCRProvider interface {
    OCRPage(ctx context.Context, input OCRPageInput) (*OCRPageOutput, error)
}

type OCRPageInput struct {
    BookID   string
    Page     int
    Language string
    Image    []byte
}

type OCRPageOutput struct {
    Text       string
    HOCR       string
    Confidence float64
    Warnings   []string
}

type OCRPageRunner struct {
    Provider OCRProvider
}

func (r *OCRPageRunner) Kind() string { return "ocr/page" }

func (r *OCRPageRunner) Run(ctx context.Context, runCtx runner.RunContext) (*model.OpResult, error) {
    input := decodeOCRPageInput(runCtx.Op.Input)
    image := loadImageArtifact(runCtx.Dependencies, input.ImageOpID)
    output, err := r.Provider.OCRPage(ctx, input.withImage(image))
    if err != nil { return nil, classifyOCRError(err) }

    return &model.OpResult{
        OpID: runCtx.Op.ID,
        Data: json({"confidence": output.Confidence}),
        Artifacts: []model.ArtifactWrite{
            textArtifact(runCtx.Op.ID, output.Text),
            hocrArtifact(runCtx.Op.ID, output.HOCR),
            jsonArtifact(runCtx.Op.ID, output.Warnings),
        },
        Records: []model.RecordWrite{
            pageStatusRecord(input.BookID, input.Page, "ocr_complete", output.Confidence),
        },
    }, nil
}
```

## Production/admin analysis

### What works today

The existing system already has a strong baseline for operations:

- CLI workers can be restarted against the same engine DB and site DBs (`pkg/cmd/worker.go:41-48` lists engine DB, sites dir, worker ID, max workers, poll interval, lease duration, HTTP timeout, and proxy flags).
- API server exposes workflow, queue, result, artifact, retry, cancel, and metrics surfaces (`pkg/api/server/routes_engine.go:11-35`).
- Frontend has workflow and artifact APIs, plus retry and cancel mutations (`web/src/api/workflowApi.ts:27-149`).
- Metrics cover submissions, scheduler cycles, ops leased/completed/failed/retried, queue waits, op durations, and worker liveness (`pkg/metrics/metrics.go:54-170`).
- Prometheus alert rules cover worker down, API down, rate limit hot spots, elevated failures, queue wait, and retry spikes (`ops/monitoring/prometheus/rules/scraper.yml:19-56`).
- Runtime events can run over Redis Streams for API/worker separation (`pkg/runtimeevents/backend.go:17-87`).

### Gaps exposed by book OCR

Book OCR stresses the runtime differently from HTML scraping.

#### 1. Long-running tasks need lease heartbeats

`HeartbeatLease` exists in `pkg/engine/store/sqlite/lease_store.go`, but the scheduler path shown above leases work and then executes a runner synchronously. If an OCR page task or PDF processing task runs longer than `--lease-duration`, lease recovery can make a still-running op ready again. This creates duplicate OCR execution unless every task is idempotent.

Minimum fix:

- For OCR, set `--lease-duration` comfortably above expected max page time during the first prototype.
- Then implement a runner wrapper that heartbeats leases while `Run` is active.
- For external subprocesses, stop heartbeating when the subprocess exits or when context is canceled.

Pseudocode:

```go
func WithLeaseHeartbeat(store Store, extendEvery, extendBy time.Duration, next runner.Runner) runner.Runner {
    return runner.Func(next.Kind(), func(ctx context.Context, rc runner.RunContext) (*model.OpResult, error) {
        hbCtx, cancel := context.WithCancel(ctx)
        defer cancel()
        go func() {
            ticker := time.NewTicker(extendEvery)
            defer ticker.Stop()
            for {
                select {
                case <-hbCtx.Done(): return
                case <-ticker.C:
                    _ = store.HeartbeatLease(ctx, rc.Op.ID, rc.Lease, extendBy)
                }
            }
        }()
        return next.Run(ctx, rc)
    })
}
```

#### 2. Cancellation must be cooperative, not only a DB mutation

Current cancel marks pending, ready, and running ops canceled and deletes leases (`pkg/services/engineview/workflow_mutation_service.go:58-72`). That is useful, but a currently executing OCR subprocess may continue until it naturally exits unless its context is canceled or a signal is sent.

Production requirement:

- Worker should periodically observe cancellation state for running ops.
- Runner context should be canceled when the workflow/op is canceled.
- External command runners should send SIGTERM, wait a grace period, then SIGKILL.
- The UI should distinguish `cancel_requested` from `canceled` if cancellation is cooperative and not instant.

#### 3. Large artifacts need storage policy

Engine artifacts are stored as BLOBs in SQLite. This is good for HTML, small logs, and small JSON. Book OCR artifacts can be large: source PDFs, page images, hOCR, per-page debug images, final PDFs, and EPUBs.

Add an artifact backend abstraction:

```go
type ArtifactStore interface {
    Put(ctx context.Context, meta ArtifactMeta, r io.Reader) (ArtifactRef, error)
    Get(ctx context.Context, id model.ArtifactID) (io.ReadCloser, ArtifactMeta, error)
    Delete(ctx context.Context, id model.ArtifactID) error
}
```

Initial backends:

- `sqlite-blob`: current behavior; fine for small local demos.
- `filesystem`: store bytes under `state/artifacts/<workflow>/<op>/<artifact>` and keep metadata/ref in engine DB.
- `s3`/object storage: future production backend for large artifacts and multiple workers.

#### 4. Progress should be stage-aware

The current workflow table can show `opDone/opTotal`, but OCR users want stages:

```text
Uploaded -> Split -> OCR pages 37/248 -> Cleanup 12/248 -> Exporting -> Done
```

Add workflow metadata and/or projection fields:

```json
{
  "stages": {
    "ingest": {"status": "succeeded"},
    "split": {"status": "succeeded", "pageCount": 248},
    "ocr": {"done": 37, "total": 248, "failed": 1},
    "cleanup": {"done": 12, "total": 248},
    "export": {"status": "pending"}
  }
}
```

Implementation options:

- compute on demand from op IDs and site DB page rows;
- persist summary rows in the site DB;
- add generic workflow milestones to the engine DB later if multiple domains need it.

Start with site DB summaries because OCR-specific page counts and confidences belong in the OCR projection.

#### 5. Worker capability routing is needed

OCR jobs may require different worker pools:

- CPU workers for PDF rendering and preprocessing;
- GPU workers for OCR/vision models;
- network-limited workers for cloud OCR/LLM cleanup;
- IO workers for exports.

Queue keys already support this. Add conventions:

```text
ocr:io
ocr:cpu
ocr:gpu
ocr:llm
ocr:export
```

Run workers with queue filters in a future phase. The current scheduler lists all ready queues and processes candidates, so a worker process may lease any queue if registered runners exist. For production, add include/exclude queue flags:

```bash
scraper worker run --queues ocr:gpu --max-workers 1 --lease-duration 30m
scraper worker run --queues ocr:cpu,ocr:io --max-workers 4
```

#### 6. Operators need page-subset reprocessing controls

The AITR-794 workflow improved by reprocessing subsets: initially blank pages, image-heavy pages, table pages, and then all pages with a universal prompt. A production OCR admin surface should expose this directly instead of forcing ad-hoc scripts.

Add API/CLI operations such as:

```http
POST /api/v1/workflows/{workflowID}/ocr/pages:reprocess

{
  "pages": [13, 32, 47, 85, 182],
  "selector": "errors|blank|table|image|code|all",
  "promptVersion": "universal-v2",
  "engineProfile": "gpt-5-nano-low"
}
```

Pseudocode:

```go
func ReprocessPages(workflowID, selector, promptVersion string) error {
    pages := selectPagesFromProjection(workflowID, selector)
    for _, page := range pages {
        enqueue(OpSpec{
            ID: fmt.Sprintf("%s:page:%03d:ocr:%s", workflowID, page, promptVersion),
            Kind: "ocr/vlm-page",
            Queue: "ocr:vlm",
            DedupKey: fmt.Sprintf("%s:%03d:%s", workflowID, page, promptVersion),
            Input: json({"page": page, "promptVersion": promptVersion}),
        })
    }
}
```

#### 7. Debugging needs replay and artifact-first inspection

OCR failures are best debugged from artifacts:

- source page image;
- OCR raw output;
- stderr/stdout from engine;
- cleanup prompt and response if LLM is used;
- structured error details.

The current artifact/result API is a strong starting point. Add conventions:

```text
page-image.png           kind=image, contentType=image/png
ocr-text.txt             kind=text, contentType=text/plain
ocr-hocr.html            kind=hocr, contentType=text/html
execution-log.json       kind=execution-log, contentType=application/json
ocr-debug.json           kind=debug, contentType=application/json
cleanup-prompt.md        kind=prompt, contentType=text/markdown
cleanup-response.md      kind=response, contentType=text/markdown
```

## River/Postgres backend analysis

### What River gives us

The River docs captured in `sources/river/` show these relevant capabilities:

- River requires PostgreSQL and commonly uses `pgx`; it persists jobs to Postgres and ships migrations (`sources/river/01-river-getting-started.md`).
- Jobs are typed Go `JobArgs` plus workers; a River client manages workers and queues (`sources/river/01-river-getting-started.md`).
- Jobs can be scheduled in the future using `ScheduledAt`; River's scheduler moves them to available after the scheduled time (`sources/river/02-river-scheduled-jobs.md`).
- Periodic jobs can run on intervals or cron schedules, but OSS periodic jobs are leader-managed in memory and can skip edge cases across restarts/elections; durable periodic jobs are a Pro feature (`sources/river/03-river-periodic-jobs.md`).
- Unique jobs can enforce uniqueness by args, period, queue, and state, but uniqueness does not mean exactly-once execution (`sources/river/04-river-unique-jobs.md`).
- Retries have defaults and can be customized by job args, insertion options, client policy, or worker policy (`sources/river/05-river-job-retries.md`).
- Multiple queues can isolate priority or high-effort work, and each queue has `MaxWorkers` (`sources/river/06-river-multiple-queues.md`).

### River is a queue backend, not the whole workflow model

River jobs map well to individual ops, but River does not directly replace these scraper concepts:

- workflow records and user-visible workflow status;
- dynamic DAG dependencies between ops;
- per-op artifacts and result envelopes;
- site/package manifests and migrations;
- JS submit verbs and scripts;
- current API/frontend contract for workflows, ops, artifacts, and results.

Therefore, do not design a River migration as `delete scheduler and use River everywhere`. Design it as:

```text
Generic workflow engine
  owns workflows, ops, dependencies, results, artifacts, UI contract

Task backend adapter
  SQLite backend: current LeaseReadyOp/CompleteOp/FailOp
  River backend: River job records wake workers for ready ops
```

### Proposed backend abstraction

Introduce a small scheduling backend interface only after the OCR prototype proves where the seams are.

```go
type TaskBackend interface {
    CreateWorkflow(ctx context.Context, params CreateWorkflowParams) error
    RefreshRunnableOps(ctx context.Context, now time.Time) (int, error)
    ListQueueCandidates(ctx context.Context, now time.Time) ([]QueueCandidate, error)
    LeaseReadyOp(ctx context.Context, req LeaseRequest) (*model.OpSpec, *model.Lease, error)
    CompleteOp(ctx context.Context, opID model.OpID, completion Completion) error
    FailOp(ctx context.Context, opID model.OpID, failure Failure) error
    HeartbeatLease(ctx context.Context, opID model.OpID, lease model.Lease, extendBy time.Duration) error
}
```

Then split responsibilities:

```text
WorkflowStore
  workflows, dependencies, results, artifacts, status projections

TaskQueueBackend
  readiness, leasing, retry timing, scheduled availability, queue limits
```

For the SQLite backend, this is mostly the existing store.

For a River backend:

1. Keep scraper workflow/op tables in Postgres (or SQLite for dev, but Postgres is cleaner).
2. When an op becomes ready, insert a River job with args `{workflowID, opID}`.
3. River worker loads the op from the workflow store and executes it through the existing runner registry.
4. On success/failure, update the workflow store through `CompleteOp`/`FailOp`.
5. For dependencies, either keep `RefreshRunnableOps` as a database transition that inserts River jobs for newly ready ops, or have completion path synchronously enqueue newly unblocked children.

River job args sketch:

```go
type ExecuteOpArgs struct {
    WorkflowID string `json:"workflow_id" river:"unique"`
    OpID       string `json:"op_id" river:"unique"`
}

func (ExecuteOpArgs) Kind() string { return "workflow_execute_op" }

func (a ExecuteOpArgs) InsertOpts() river.InsertOpts {
    return river.InsertOpts{
        Queue: determineQueue(a),
        UniqueOpts: river.UniqueOpts{ByArgs: true},
        MaxAttempts: 1, // let scraper retry policy own retry semantics initially
    }
}
```

Important design decision: start with `MaxAttempts: 1` in River and keep scraper retry state authoritative. Otherwise retry semantics exist in both River and scraper and become hard to explain.

### When to keep SQLite

Keep the current scheduler if:

- deployments are local, single-host, or low concurrency;
- the engine DB and site DB are files under `state/`;
- operators value simple dev setup over distributed robustness;
- OCR workflows are run by one machine or one worker pool at a time;
- the current API/UI contract matters more than River UI.

For this mode, prioritize:

- heartbeats;
- queue filters;
- artifact storage abstraction;
- better progress projections;
- runner cancellation.

### When to add River

Add River if production needs:

- PostgreSQL as the durable system of record;
- multiple machines/workers sharing a queue;
- high throughput queue polling without custom SQL tuning;
- transactional enqueueing with application data in Postgres;
- operational queue features from River and River UI;
- scheduled or periodic jobs that are central to the product.

Even then, preserve the scraper workflow tables and API/UI model unless a separate product decision says otherwise.

## Proposed implementation plan

### Phase 0: Document and name the generic concepts

Goal: make the current system understandable without changing behavior.

Tasks:

1. Add an architecture doc or README section called `Workflow engine concepts`.
2. Add package-level comments that map scraper terms to generic terms.
3. Do not rename database columns yet.
4. Add an ADR: `site` currently means workflow package/domain.

Files to read first:

- `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/engine/model/types.go`
- `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/engine/scheduler/scheduler.go`
- `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/engine/store/store.go`
- `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/doc/topics/scraper-runtime-model.md`

### Phase 1: Implement OCR as a normal package/site

Goal: prove that non-scraping workloads fit before refactoring the whole product.

Tasks:

1. Create `sites/bookocr/site.yaml`.
2. Add OCR site DB migrations.
3. Add `verbs/ingest.js`.
4. Add JS scripts for graph orchestration.
5. Add a simple runner or JS script path for OCR. For a prototype, `ocr/page` may shell out to Tesseract behind a controlled Go runner.
6. Add fixture-based tests with a tiny PDF or two page images.

Validation:

```bash
go test ./pkg/engine/... ./pkg/sites/... ./pkg/cmd/... -count=1

tmpdir=$(mktemp -d)
go run ./cmd/scraper --sites-manifest-dir ./sites \
  site bookocr run ingest \
  --sites-dir "$tmpdir/sites" \
  --engine-db "$tmpdir/engine.db" \
  --book-id smoke-book \
  --source-uri ./sites/bookocr/fixtures/tiny.pdf

go run ./cmd/scraper --sites-manifest-dir ./sites \
  worker run \
  --engine-db "$tmpdir/engine.db" \
  --sites-dir "$tmpdir/sites" \
  --max-cycles 200 \
  --poll-interval 20ms

go run ./cmd/scraper engine status --engine-db "$tmpdir/engine.db"
```

### Phase 2: Add production-critical runtime features

Goal: make OCR safe to run for long tasks and large books.

Tasks:

1. Add lease heartbeat wrapper around runner execution.
2. Add cooperative cancellation checks and context cancellation.
3. Add worker queue include/exclude filters.
4. Add file-backed artifact store.
5. Add OCR progress projection service and API endpoint.
6. Extend frontend workflow detail with stage progress and page table.

API sketch:

```http
GET /api/v1/workflows/{workflowID}/progress

{
  "workflowID": "bookocr-ingest-...",
  "status": "running",
  "stages": [
    {"name": "ingest", "status": "succeeded"},
    {"name": "split", "status": "succeeded", "total": 248},
    {"name": "ocr", "status": "running", "done": 37, "failed": 1, "total": 248},
    {"name": "cleanup", "status": "running", "done": 12, "total": 248},
    {"name": "export", "status": "pending"}
  ]
}
```

### Phase 3: Extract generic package names and interfaces

Goal: reduce scraper-specific naming only after a second workload exists.

Tasks:

1. Introduce generic aliases in Go packages: `domain`, `workflowpkg`, or `workflowspec`.
2. Keep compatibility with `site.yaml` while allowing `workflow.yaml`.
3. Move docs from scraper-specific to workflow-engine-specific language.
4. Add a `TaskBackend` interface if River/Postgres is still desired.
5. Keep HTTP/API compatibility until frontend and docs are updated.

Suggested directory direction:

```text
pkg/engine/model        -> remains generic
pkg/engine/scheduler    -> generic scheduler
pkg/engine/store        -> workflow/task store contracts
pkg/sites               -> pkg/workflowpackages, with sites as compatibility wrapper
pkg/services/engineview -> workflow view service
```

### Phase 4: Optional River backend

Goal: use River where it adds clear operational value.

Tasks:

1. Add Postgres workflow store migrations.
2. Add River migrations and client setup.
3. Add `ExecuteOpArgs` and River worker that calls existing runner registry.
4. Ensure scraper retry policy remains authoritative initially.
5. Implement a readiness-to-River enqueue path for newly ready ops.
6. Add integration tests with Postgres testcontainer or local Docker Compose.
7. Compare UI/API parity with SQLite backend.

Validation:

```bash
DATABASE_URL=postgres://... go test ./pkg/engine/backend/river/... -count=1
DATABASE_URL=postgres://... go run ./cmd/scraper worker run --backend river --queues ocr:cpu
```

## Intern onboarding guide

### Read in this order

1. `scraper/README.md` for the high-level project map.
2. `pkg/doc/topics/scraper-architecture-overview.md` for the current engine/site split.
3. `pkg/doc/topics/scraper-runtime-model.md` for submit-time vs execution-time JS.
4. `pkg/engine/model/types.go` for the core data structures.
5. `pkg/engine/scheduler/scheduler.go` for the control loop.
6. `pkg/engine/store/sqlite/lease_store.go` and `result_store.go` for durable state transitions.
7. `pkg/sites/submitverbs/host.go` and `runtime.go` for workflow submission.
8. `sites/nereval/` for a real fan-out example.
9. `pkg/api/server/routes_engine.go` and `web/src/api/workflowApi.ts` for operator/admin surfaces.
10. `pkg/metrics/metrics.go`, `ops/monitoring/prometheus/rules/scraper.yml`, and `pkg/runtimeevents/backend.go` for production signals.

### Mental model to keep while coding

Always ask:

- Is this data runtime state or domain projection data?
  - Runtime state belongs in the engine DB.
  - OCR/book/page/query data belongs in the OCR site DB.
- Is this work submission or durable execution?
  - Submission verbs validate input and emit initial ops.
  - Workers execute ops.
- Is this artifact small or large?
  - Small artifacts can use current engine BLOBs.
  - Large OCR artifacts need a file/object store.
- Can this op be safely retried?
  - If yes, make it idempotent with stable IDs/dedup keys.
  - If no, isolate side effects and store enough state to resume or compensate.

### Code review checklist for the intern

Before opening a PR:

- [ ] Every emitted op has a stable ID and dedup key.
- [ ] Every long-running runner respects context cancellation.
- [ ] OCR subprocesses produce stdout/stderr artifacts.
- [ ] Page-level status is visible in the OCR site DB.
- [ ] Large files are not blindly inserted into SQLite BLOBs.
- [ ] Failed ops return structured `OpError` codes.
- [ ] Tests cover successful run, failed page, retry, and cancellation.
- [ ] CLI smoke test runs with a temporary engine DB and sites dir.
- [ ] API/frontend still shows workflow, ops, artifacts, and retry/cancel.

## Risks and mitigations

| Risk | Why it matters | Mitigation |
|---|---|---|
| Lease expiry duplicates long OCR tasks | OCR can run longer than default 30s leases. | Add heartbeat wrapper and set longer leases for OCR queues. |
| SQLite BLOBs grow too large | PDFs/images/final exports can make engine DB huge. | Add filesystem/object artifact backend. |
| Cancellation appears instant in UI but process keeps running | DB state can be canceled while subprocess still runs. | Add cooperative cancellation and signal subprocesses. |
| River migration duplicates retry semantics | River and scraper both support retry. | If using River, start with River max attempts 1 and keep scraper retry authoritative. |
| Generic rename breaks existing scraper docs/scripts | Current docs and scripts depend on `site` terms. | Add aliases and compatibility before renaming. |
| OCR workflow creates too many ops at once | Large books can create thousands of page ops. | Batch fan-out, use pagination/chunk ops, and test queue listing performance. |
| Progress is misleading while graph expands | Total op count changes as pages are discovered. | Use OCR site DB stage summary instead of only opDone/opTotal. |

## Open questions

1. Should `book-ocr` live inside `scraper/sites/bookocr` as a proving package, or in its own repo with external manifest loading?
2. Which OCR provider is the first target: Tesseract CLI, local model, cloud OCR, or LLM vision?
3. What artifact size threshold should move data out of SQLite and into filesystem/object storage?
4. Should generic naming use `domain`, `workflow package`, `job package`, or another term?
5. Does production require multi-host workers soon enough to justify River/Postgres now, or is SQLite sufficient for the next milestone?
6. Should periodic/scheduled workflows be part of v1, or should book OCR stay explicitly user-submitted?
7. Should prompt versions be stored as managed database rows, files under the workflow package, or both?
8. Should page-subset reprocessing be a generic workflow feature or an OCR-specific service endpoint first?

## Final recommendation

Use the current scraper engine as the foundation. It already has durable workflows, task dependencies, leases, retries, queue policy, results, artifacts, runtime events, API endpoints, frontend admin operations, metrics, and alerts. Build book OCR as the first non-scraper workflow package to expose real generalization pressure, but model it as an iterative page-processing campaign with prompt versions, page attempts, selective reprocessing, and filesystem/object artifacts.

The AITR-794 scripts make the scheduler recommendation slightly stronger: the current SQLite backend is probably fine for a first port if its transactional lease path remains strict and if OCR adds heartbeat/cancellation/stale-reset behavior. The original Python queue had to move to `BEGIN IMMEDIATE` to avoid duplicate page claims; scraper's existing transactional `LeaseReadyOp` is already closer to the right model than the initial ad-hoc queue.

Add River only when the operational requirement is clearly Postgres/multi-host scheduling. If River is added, integrate it as a task-queue backend beneath the existing workflow model rather than replacing the workflow model.

## References

### Key scraper files

- `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/engine/model/types.go`
- `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/engine/scheduler/scheduler.go`
- `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/engine/store/store.go`
- `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/engine/store/sqlite/lease_store.go`
- `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/engine/store/sqlite/result_store.go`
- `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/sites/submitverbs/host.go`
- `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/sites/submitverbs/runtime.go`
- `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/sites/manifest/manifest.go`
- `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/sites/nereval/scripts/extract_list.js`
- `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/api/server/routes_engine.go`
- `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/services/engineview/workflow_mutation_service.go`
- `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/web/src/api/workflowApi.ts`
- `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/metrics/metrics.go`
- `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/ops/monitoring/prometheus/rules/scraper.yml`
- `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/runtimeevents/backend.go`

### Book OCR current files

- `/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/cmd/XXX/main.go`
- `/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/go.mod`
- `/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/README.md`

### Concrete AITR-794 OCR ticket files

- `/home/manuel/workspaces/2026-05-20/book-ocr/claw-stuff/ttmp/2026/05/20/AITR-794--clean-markdown-transcription-of-aitr-794-pdf/reference/01-diary.md`
- `/home/manuel/workspaces/2026-05-20/book-ocr/claw-stuff/ttmp/2026/05/20/AITR-794--clean-markdown-transcription-of-aitr-794-pdf/scripts/01-extract-pages.py`
- `/home/manuel/workspaces/2026-05-20/book-ocr/claw-stuff/ttmp/2026/05/20/AITR-794--clean-markdown-transcription-of-aitr-794-pdf/scripts/03-ocr-batch.py`
- `/home/manuel/workspaces/2026-05-20/book-ocr/claw-stuff/ttmp/2026/05/20/AITR-794--clean-markdown-transcription-of-aitr-794-pdf/scripts/04-ocr-dashboard.py`
- `/home/manuel/workspaces/2026-05-20/book-ocr/claw-stuff/ttmp/2026/05/20/AITR-794--clean-markdown-transcription-of-aitr-794-pdf/scripts/10-reprocess-universal.py`
- `/home/manuel/workspaces/2026-05-20/book-ocr/claw-stuff/ttmp/2026/05/20/AITR-794--clean-markdown-transcription-of-aitr-794-pdf/scripts/11-postprocess.py`

### River source captures

- `sources/river/01-river-getting-started.md`
- `sources/river/02-river-scheduled-jobs.md`
- `sources/river/03-river-periodic-jobs.md`
- `sources/river/04-river-unique-jobs.md`
- `sources/river/05-river-job-retries.md`
- `sources/river/06-river-multiple-queues.md`
