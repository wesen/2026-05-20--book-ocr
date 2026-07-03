---
Title: Book OCR productization analysis, design, and implementation guide
Ticket: BOOK-OCR-PRODUCT-001
Status: active
Topics:
    - book-ocr
    - productization
    - workflow
    - ocr
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: cmd/book-ocr/main.go
      Note: CLI dispatch
    - Path: go.mod
      Note: Finding F1 — replace ../scraper breaks clean-clone builds
    - Path: internal/bookprofile/profile.go
      Note: Profile layer — the productization seam (Phase 2 target)
    - Path: internal/ocrpipeline/prompts.go
      Note: Report-794 lexicon and Lisp assertion hardcoded in prompt (F2
    - Path: internal/ocrpipeline/renderer.go
      Note: common-lisp fence and figure-suppression cues hardcoded (F3)
    - Path: internal/ocrpipeline/structured_ocr.go
      Note: Per-page orchestration
    - Path: internal/ocrpipeline/workflow_executors.go
      Note: Structured workflow DAG
    - Path: internal/ocrquality/figures.go
      Note: Ink-band figure crop engine; page naming hardcode (F4)
ExternalSources: []
Summary: Full status assessment of the book-ocr repository and an intern-ready analysis, design, and phased implementation guide for turning the one-book structured OCR pipeline into a general-purpose book OCR product.
LastUpdated: 2026-07-03T10:38:58.157441727-07:00
WhatFor: Onboard a new engineer onto the book-ocr system and give them an executable productization roadmap.
WhenToUse: Read before doing any productization work on book-ocr; use the References section to navigate the code.
---


# Book OCR productization analysis, design, and implementation guide

## Executive Summary

`book-ocr` is a Go application that converts a scanned technical book (a directory of page PNGs) into durable, reviewable artifacts: structured per-page JSON, deterministic Markdown, extracted figure crops, validation reports, and a rendered PDF. It was built during an intense one-week sprint (2026-05-20 to 2026-05-27) around a single book — MIT AI Lab Technical Report 794, *Presentation Based User Interfaces* — and it successfully OCR'd all 202 pages of that book through a durable, resumable workflow engine.

The system is in genuinely good shape for a one-book project: the architecture is layered correctly (generic workflow runtime below, OCR policy above), the failure history is documented in nine ticket workspaces, and the core design rule — *the model sees exactly one page image and returns structured JSON; Go renders the Markdown deterministically* — is empirically justified by a purpose-built benchmark. All tests pass and a full dry-run of the pipeline (workflow engine → page OCR → assembly → figure embedding → PDF render → validation) completes successfully today (verified 2026-07-03; see Appendix A).

It is, however, not yet a product. The five blocking problems, in order of severity:

1. **It does not build from a clean clone.** `go.mod` pins the workflow runtime with `replace github.com/go-go-golems/scraper => ../scraper`, a path that only exists on the original dev machine layout (finding F1).
2. **Book-specific policy is hardcoded in three separate places.** The Report 794 lexicon, page-type map, and QA expectations live independently in the prompt strings, in `bookprofile.Report794()`, and in `ocrquality` defaults. Running a *different* book today means editing Go source (findings F2–F4).
3. **The prompt and renderer assume Common Lisp.** Every code block is fenced as `common-lisp`; figure-suppression heuristics key on strings like `ppscalc` (finding F3).
4. **Operational surfaces are developer-grade.** Live inference is opt-out (`--dry-run` defaults to true), the CLI is a hand-rolled `switch` that still prints `ocr-mvp` in usage, and targeted page rerun performs raw SQL surgery on the engine database (findings F5–F7).
5. **There is no ingestion story.** The product input is "a directory of `page_*.png` files"; a real user starts from a PDF or a stack of photos (finding F8).

The proposed path to a product keeps the current architecture — it is fundamentally right — and generalizes it in four phases: (1) make it buildable, installable, and CI-verified; (2) move all book policy into the `bookprofile` layer so a new book is a YAML file, not a code change; (3) add PDF ingestion and a `book-ocr init` bootstrap flow that drafts a profile automatically; (4) grow operational maturity (first-class rerun API, run inspection, cost/progress reporting, and optionally a review web UI). Each phase is specified with file-level guidance, API sketches, and pseudocode below.

## Problem Statement and Scope

Two questions from the ticket:

1. **What is the status of this OCR repo?** Answered in "Current Status" with build/test/experiment evidence.
2. **How could we turn it into a proper product?** Answered by the gap analysis, the product definition, the proposed architecture, the decision records, and the phased implementation plan.

The document doubles as an onboarding guide: a new intern should be able to read it top to bottom, understand every moving part, navigate to the right file for each subsystem, and start executing Phase 1 without further help.

Out of scope: building the product itself; modifying the `~/code/wesen/go-go-golems` repositories (scraper, geppetto, pinocchio, go-go-goja) — they are treated as read-only dependencies whose evolution we can request but not perform here.

## Part I — What the System Is

### The one-paragraph version

You give `book-ocr` a directory of scanned page images. It creates a durable workflow run in SQLite, fans out one OCR step per page, and each step sends *exactly one* page image to a vision-language model (through Geppetto/Pinocchio profile resolution), asking for structured JSON — headings, paragraphs, lists, tables, code, figures — rather than free-form Markdown. Go code then renders that JSON into deterministic Markdown, extracts figure crops from the source image with a pixel-analysis heuristic, assembles all pages into one book Markdown, validates the result (page continuity, duplicated captions, suspiciously short pages), and renders a review PDF via Pandoc/XeLaTeX. Every intermediate artifact — raw model response, parsed JSON, rendered Markdown, model conversation turns — is preserved on disk and in SQLite, so any page can be inspected, diagnosed, and rerun individually without touching the rest of the book.

### Why it looks the way it does (the history that matters)

The repository's shape is the fossil record of one week of iteration, preserved in `ttmp/`. The sequence matters because each layer exists as a reaction to a concrete failure:

1. **SCRAPER-JOBS-001** (05-24): a generic durable workflow runtime (`scraper/pkg/workflow`) was designed as a facade over the scraper repo's engine — durable runs, per-step leases, retries, artifacts, projections — explicitly so an OCR pipeline could run on it.
2. **OCR-MVP-001** (05-24): the first pipeline — *freeform* OCR, one multimodal LLM call per page returning a Markdown blob — proved the runtime end-to-end.
3. **BOOK-OCR-HQ-001 / OCR-QUALITY-WORKERS-001** (05-24): quality experiments over the first 30 pages produced the QA/normalization/figure-extraction workers and the `bookprofile` config layer.
4. **BOOK-OCR-EXTERNALIZE-001** (05-24/25): all OCR code moved out of the scraper repo into this one. Scraper stayed a pure workflow runtime. The boundary is exactly right for productization; the `replace ../scraper` directive is its unfinished edge.
5. **BOOK-OCR-FULL-001** (05-25): the full 202-page run *worked* but exposed the pivotal failure — with `--context-window 1`, neighboring page images sent as context caused the model to copy adjacent-page content into the target page ("context bleed"): duplicate figure captions, false figures, phantom prose.
6. **BOOK-OCR-VLM-SEPARATION-001** (05-25): instead of guessing, a benchmark harness measured six multimodal prompt layouts against page-level "anchor oracles". Verdict: `target-only` layouts scored ~0.94 average isolation; every layout including context images leaked. This turned a debugging hunch into a design rule.
7. **BOOK-OCR-PIPELINE-REDESIGN-001** (05-25): the structured pipeline — target-page-only input, structured-JSON output, deterministic Go rendering.
8. **BOOK-OCR-STRUCTURED-WORKFLOW-001** (05-25→27): the structured pipeline promoted from a CLI loop onto the workflow runtime, plus PDF rendering, targeted rerun, and Common Lisp rendering. This is the current production path.

The two rules that fall out of this history, quoted from `README.md:423-424` because every future change must respect them:

> 1. **Primary OCR uses exactly one target page image.** This prevents neighboring-page context bleed.
> 2. **The model returns structured JSON, not final Markdown.** Go owns rendering.

### The big picture

```text
                                   ┌──────────────────────────────────────────────┐
                                   │                 cmd/book-ocr                  │
                                   │  structured-run · structured-rerun-pages ·   │
                                   │  status · resume · retry · cancel ·          │
                                   │  structured-pages · quality-pass · run(mvp)  │
                                   └───────┬──────────────────────────┬───────────┘
                                           │ registers packages       │ queries
                                           ▼                          ▼
 ┌─────────────────────────┐   ┌───────────────────────────┐   ┌──────────────────┐
 │  internal/ocrpipeline   │   │ scraper/pkg/workflow      │   │  projections/    │
 │  (STRUCTURED, current)  │──▶│ Runtime · StepContext ·   │──▶│  *.db (SQLite    │
 │  prompts · types ·      │   │ Executor · Package ·      │   │  read models)    │
 │  renderer · client ·    │   │ ArtifactStore ·           │   └──────────────────┘
 │  executors · projection │   │ ProjectionStore           │
 └──────┬──────┬───────────┘   │ (engine.db: workflows,    │
        │      │               │  ops, leases, results,    │
        │      │  reuses       │  artifacts, deps)         │
        ▼      ▼               └───────────────────────────┘
 ┌───────────────┐ ┌──────────────────┐ ┌──────────────────┐ ┌────────────────────┐
 │ ocrvalidation │ │ ocrquality       │ │ ocrmvp (legacy   │ │ bookprofile        │
 │ caption/      │ │ QA · normalize · │ │ freeform path;   │ │ per-book policy    │
 │ anchor checks │ │ figure crops     │ │ page discovery   │ │ YAML + discovery   │
 └───────────────┘ └──────────────────┘ │ still reused)    │ │ patch proposals    │
                                        └──────────────────┘ └────────────────────┘
        │
        ▼ model calls (exactly 1 image per call)
 ┌────────────────────────────────────────────────┐
 │ geppetto (inference engines, turns)            │
 │ pinocchio (profile registry: --profile slug →  │
 │ provider/model/settings; turns.db chat store)  │
 └────────────────────────────────────────────────┘
```

The dependency direction is strict and load-bearing: `book-ocr` imports `scraper/pkg/workflow`; scraper contains zero OCR knowledge. Inside book-ocr, `ocrpipeline` (the current path) reuses `ocrmvp.DiscoverPageImages` for page discovery (`internal/ocrpipeline/workflow_executors.go:33`), `ocrquality.EmbedExtractedFigures` for figure crops (`workflow_executors.go:167`), and `ocrvalidation` for deterministic checks (`workflow_executors.go:251`, `structured_ocr.go:301-317`).

## Part II — Current Status (evidence-based)

### What was verified on 2026-07-03

| Check | Result | Evidence |
|---|---|---|
| Build from repo as-is | **FAILS** | `replacement directory ../scraper does not exist` for `ocrmvp`, `ocrpipeline`, `ocrquality` |
| Build with corrected module resolution | PASSES | `scripts/book-ocr.go.work`, `go build ./...` → OK |
| Full test suite | PASSES (6/6 packages with tests) | `go test ./... -count=1`: `bookprofile`, `ocrmvp`, `ocrpipeline`, `ocrquality`, `ocrvalidation`, `vlmseparation` all `ok` |
| End-to-end dry-run (3 pages, figures, PDF) | PASSES | `scripts/01-dry-run-structured-pipeline.sh`: all 9 checked artifacts present, `validation-report.json` shows `page_count: 3, expected: 3, warnings: 0`, `book.pdf` 10.9 KB |
| Last commit | 2026-05-27 (`cd93b7b` "Write Book OCR README") | `git log` |
| Working tree | clean | `git status` |
| Docs | current, high quality | 479-line README verified against code; 9 ticket workspaces with diaries |

### What works today (feature inventory)

- **Full structured book run**: `structured-run` executes discover → N parallel page steps → assemble → validate as a durable workflow with per-step retry (3 attempts, exponential backoff 1s→30s, `internal/ocrpipeline/workflow_package.go:90-92`).
- **Crash-safe resume**: all state is in `engine.db`; `resume` re-opens the run and continues; expired leases recover automatically.
- **Targeted page repair**: `structured-rerun-pages --pages 132,140` reprocesses only the named pages and rebuilds downstream artifacts — the practical review loop that fixed the real book.
- **Figure extraction**: pixel-space "ink band" analysis (`internal/ocrquality/figures.go:239-335`) crops figures from source pages, with sidecar JSON and red-rectangle debug overlays.
- **Deterministic rendering**: structured JSON → Markdown with wrapped prose, real pipe tables, fenced code, figure links (`internal/ocrpipeline/renderer.go`).
- **Validation**: page-count gates, adjacent-duplicate-caption detection, short-page warnings, per-page block warnings (`workflow_executors.go:238-284`, `structured_ocr.go:300-321`).
- **PDF review artifact**: Pandoc/XeLaTeX render wired into the assemble step (`workflow_executors.go:344-366`).
- **Turn persistence**: every model conversation saved as YAML and into a Pinocchio-compatible `turns.db` (`internal/ocrpipeline/session.go`).
- **Diagnostics**: the `vlm-separation` benchmark for measuring context-bleed under different prompt layouts — the best-tested package in the repo.

### Honest quality assessment

The May sprint ended with the structured pipeline *validated* but with the authors' own acceptance criteria not fully closed. From the tickets' open items: prose completeness of structured output was measured as lower-volume than the freeform path on the first-50 run; Phase 6 "production hardening" of the redesign ticket (figure QA integration, full-book acceptance gates) is the one explicitly open phase; and Common Lisp OCR fidelity still requires manual review (parentheses, package prefixes, `*earmuffs*` — `README.md:474`). The README's own "Current limitations" section (`README.md:468-474`) is accurate and maps 1:1 onto findings F5–F7 below.

## Part III — Architecture Deep Dive

This section is the onboarding core: each subsystem, what it does, how to call it, and where it lives.

### 1. The workflow runtime (`scraper/pkg/workflow`)

The runtime is an embeddable durable-execution engine: ~1,830 lines facading `scraper/pkg/engine/{model,runner,scheduler,store}`. Everything below persists in a single SQLite file (book-ocr names it `engine.db`).

**Core API surface** (all paths relative to `~/code/wesen/go-go-golems/scraper/pkg/workflow/`):

| Symbol | Where | What |
|---|---|---|
| `Runtime`, `NewRuntime(ctx, Config)` | `runtime.go:80-124` | The engine handle. `Config` carries store, artifact store, projection store, worker counts, per-queue limits. |
| `Config.Queues map[model.QueueKey]QueueConfig` | `runtime.go:60-76` | Per-queue `MaxWorkers` and optional token-bucket `RateLimit` — how you cap concurrent model calls. |
| `Package`, `NewPackage(name).Entrypoint(...)` | `package.go:12-39` | A named workflow definition. The entrypoint receives a `RunBuilder` and seeds initial steps. |
| `Executor`, `NewTypedExecutor[I](kind, fn)` | `executor.go:34-74` | Business logic per step kind; typed input decoding. |
| `StepContext` | `context.go:18-299` | Executor-facing durable step: `Input`, `DependencyData`, `Result`, `Artifact`/`StoreArtifact`, `Projection(name)`, `Emit` (dynamic fan-out), `Record`. |
| `Error`, `Retryable(code, err)`, `Permanent(code, err)` | `errors.go:15-83` | Error classification drives the scheduler's retry decision. |
| `rt.StartRun / RunOnce / StartWorkers` | `runtime.go:219-311` | Create a run; execute one scheduling cycle; poll loop. |
| `rt.RetryStep / CancelRun` | `operators.go:21-36` | Operator controls. |
| `FileArtifactStore` | `artifact_store.go:42-105` | Large artifacts on the filesystem with JSON metadata sidecars (avoids SQLite BLOB bloat). |
| `SQLiteProjectionStore` | `projection_store.go:30-84` | One SQLite DB per named read model under `projections/`. |

**Persistence model** (`engine.db`, migrations `001_engine_core.sql` / `002_engine_runtime.sql`):

```text
workflows(id, site, name, status, input_json, metadata_json, ...)
ops(id, workflow_id, parent_id, kind, queue_key, dedup_key, input_json,
    retry_json,            -- retry policy
    retry_state_json,      -- live attempt counter + last error
    next_attempt_at, status, ...)
op_dependencies(op_id, depends_on_op_id, required)
leases(op_id, worker_id, token, acquired_at, expires_at)
queue_limit_state(site, queue_key, tokens, last_refill_at)   -- token bucket
results(op_id, data_json, records_json, emitted_json, error_json, ...)
artifacts(id, op_id, name, kind, content_type, metadata_json, body BLOB, ...)
```

**Execution model** (`engine/scheduler/scheduler.go:197-294`): each `RunOnce` cycle (a) recovers expired leases back to `ready`, (b) cancels pending ops whose required dependencies failed (propagation to fixpoint), (c) promotes `pending → ready` ops whose dependencies are complete, (d) leases and executes ready ops up to global and per-queue caps. Coordination is entirely through the leases table, so multiple worker processes can share one `engine.db`. Retry scheduling honors the `Retryable` flag on the classified error and the per-op backoff policy (`scheduler.go:472-529`).

Two known runtime gaps, documented since SCRAPER-JOBS-001 and still open: no lease heartbeat (a page step running longer than the lease duration can be double-executed — currently mitigated by idempotent artifact writes), and cancellation is a DB mutation without cooperative in-flight interruption (`operators.go:29-35`).

### 2. The structured OCR pipeline (`internal/ocrpipeline`) — the production path

~1,850 lines across 11 files plus tests. Five concerns, cleanly separated:

**a) The data contract** (`types.go`, 190 lines). The model must return:

```jsonc
{
  "schema_version": "structured-ocr/v1",
  "book_id": "report-794",
  "page_number": 32,
  "page_type": "body",          // blank|title|front_matter|table_of_contents|
                                 // table_of_figures|body|figure|table|bibliography
  "blocks": [
    {"id": "p032-b001", "type": "heading",   "text": "2.1 PPSCalc", "level": 2},
    {"id": "p032-b002", "type": "paragraph", "text": "..."},
    {"id": "p032-b003", "type": "table",
     "table": {"headers": ["", "A", "B"], "rows": [["1", "10", "20"]]}},
    {"id": "p032-b004", "type": "code",   "text": "(defmethod present ...)"},
    {"id": "p032-b005", "type": "figure", "caption": "Figure 2-2: ...",
     "description": "...", "diagram_text": ["..."]},
    {"id": "p032-b006", "type": "page_footer", "text": "32"}
  ],
  "warnings": [{"code": "low_confidence", "message": "...", "block_id": "p032-b004"}]
}
```

The Go structs mirror this (`StructuredPageOCR` `types.go:11-18`, `OCRBlock` `types.go:78-90`). A crucial productization lesson is encoded in the custom `UnmarshalJSON` methods: real models violate schemas in predictable ways, so `page_number` is coerced from float/string (`types.go:20-48`), a nested `figure:{...}` object gets flattened (`types.go:92-139`), `diagram_text` accepts string or array, and list items accept bare strings (`types.go:146-159`). Parsing is layered the same way: strict JSON → strip fences and extract the first balanced `{...}` (`structured_ocr.go:115-156`) → `jsonsanitize.Sanitize` retry → regex repair of leading-zero numbers (`structured_ocr.go:158-163`). Raw responses are written to disk *before* parsing so failed parses stay inspectable.

**b) The prompt** (`prompts.go`, 77 lines). System: *"You are a precise structured OCR engine for scanned technical books. Return strict JSON only."* (`prompts.go:7-8`). The user prompt contains the schema contract, the single-page rule (*"Transcribe exactly one target page image... Do not infer text from neighboring pages"*, `prompts.go:11-17`), table rules (grids must become table blocks), code rules (readable screenshots must be transcribed, never left as empty code blocks), figure rules (a figure block is only for content that remains meaningful as an image after text is extracted) — and, hardcoded today, the Report 794 vocabulary (*"data base", "PSBase", "PPSCalc", "Dired", "Steamer", "Zmacs", "Xerox Star"*, `prompts.go:19`), the Common Lisp assertion (`prompts.go:20`), and a PPSCalc few-shot example (`prompts.go:62-70`).

**c) The model client** (`client.go`, 165 lines; `session.go`, 99 lines). `GeppettoStructuredOCRClient.OCRPage` (`client.go:44-76`) builds a two-block turn (system + user-multimodal with one image payload carrying `role:"target"`, `detail:"high"`), resolves the `--profile` slug through pinocchio's `profilebootstrap` into a concrete provider/model engine, and calls `eng.RunInference(ctx, inputTurn)`. `CountTurnImages` (`client.go:112-149`) enforces the exactly-one-image invariant; `RunStructuredPage` refuses to proceed otherwise (`structured_ocr.go:247-249`). Both input and final turns are persisted as YAML files and into a SQLite turn store (`session.go:29-80`) keyed `book-ocr:<book>:<run>` / `page:%03d`. `DryRunStructuredOCRClient` (`structured_ocr.go:38-57`) fabricates deterministic pages for offline testing — currently with Report-794-specific fixtures for pages 12/13/32 (`structured_ocr.go:59-79`).

**d) The renderer** (`renderer.go`, 241 lines). `RenderPageMarkdown` emits an HTML comment marker `<!-- page:NNN -->` (the seam every downstream tool splits on), then maps blocks: headings → `#`×level; paragraphs → 88-column wrapped prose; lists → nested bullets; tables → GitHub pipe tables with synthesized `Column N` headers and ragged-row padding (`renderer.go:175-204`); code → fenced blocks **hardcoded as `common-lisp`** (`renderer.go:121`); footnotes → `[^id]:`; page footers → suppressed by default. Figures get the most interesting treatment: the image is *suppressed* when the caption/description indicates it is really textual content already transcribed — cues like `code listing`, `lisp-like definition` (`renderer.go:82-90`) or, when a table block follows, `ppscalc`, `spreadsheet`, `grid` (`renderer.go:92-100`). This "don't ship pictures of text" policy is a genuine product feature; its keyword lists are book-specific today.

**e) The workflow package** (`workflow_*.go`, ~740 lines). Four executors on four queues:

```text
discover-structured-pages   (queue structured-control, MaxAttempts 1)
   │  DiscoverPageImages(imageDir, glob, start, end)          workflow_executors.go:33
   │  seed projection rows; Emit one op per page              workflow_executors.go:40-53
   ▼
structured-page-NNN × N     (queue structured-vision, retry 3× exp backoff)
   │  RunStructuredPage → 6 artifacts per page                workflow_executors.go:69-123
   │  errors classified: parse/mismatch/missing-image = Permanent;
   │  TLS/rate-limit/provider = Retryable                     workflow_executors.go:368-399
   ▼
assemble-structured-markdown (queue structured-assemble, DependsOn all pages)
   │  concat rendered pages → assembled.md                    workflow_executors.go:125-236
   │  optional: EmbedExtractedFigures → embedded-figures.md
   │  optional: pandoc/xelatex → book.pdf                     workflow_executors.go:344-366
   ▼
validate-structured-run      (queue structured-validation, DependsOn assemble)
      page-count gate (Permanent error on mismatch), duplicate-caption scan,
      short-page query from projection → validation-report.json
                                                              workflow_executors.go:238-284
```

The projection (`workflow_projection.go:11-50`) maintains `structured_pages` (status, artifact IDs, warning/table/figure counts, rendered bytes, error) and `structured_warnings` in `projections/book_ocr_structured.db` — this is what `book-ocr structured-pages` queries and what the validator uses to find short pages.

**Per-page artifact contract** (written by `RunStructuredPage`, `structured_ocr.go:233-241`):

```text
pages/page_NNN/01-turn-input.yaml     what we sent (audit)
pages/page_NNN/02-turn-final.yaml     full conversation incl. response
pages/page_NNN/03-raw-response.json   verbatim model text, pre-parse
pages/page_NNN/04-structured.json     parsed+repaired StructuredPageOCR
pages/page_NNN/05-rendered.md         deterministic Markdown
pages/page_NNN/06-validation.json     per-page validation
pages/page_NNN/07-error.txt           only on failure
```

The numbered-prefix convention makes the debugging workflow self-documenting: if `04` is wrong, fix prompt/model and rerun the page; if `05` is wrong given a correct `04`, fix the renderer; if only `book.pdf` is wrong, fix Pandoc/LaTeX settings (`README.md:352-359`).

### 3. Supporting packages

**`internal/ocrquality`** (~1,900 lines) — post-OCR QA as its own workflow package (`ocr-quality/*` step kinds, `package.go:38-126`): QA-before → normalize → embed-figures → QA-after → import-log → write-discovery → assemble-report. Its two durable contributions to the product: (a) the **figure crop engine** `EmbedExtractedFigures` (`figures.go:47-98`) — for each `[FIGURE: ...]` marker or figure caption, it scans the source PNG for horizontal "ink bands" (rows with ≥ max(8, width/250) non-white pixels, `figures.go:301-335`), drops header/footer furniture zones, crops with a 24px margin, and writes crop + JSON sidecar + debug overlay; (b) the **discovery mechanism** (`package.go:300-377`) which converts run observations into `bookprofile.DiscoveryState` and a machine-proposed profile patch — the seam that productization Phase 3 builds on. Its QA checks (page continuity, known-bad terms, expected strings, list-page style; `qa.go:16-135`) currently default to Report-794 constants (`markdown.go:82-114`).

**`internal/ocrvalidation`** (~130 lines) — small, fully generic, well-tested deterministic validators: figure-caption extraction (`adjacent.go:12-26`), adjacent-duplicate-caption detection across consecutive pages (`adjacent.go:28-56`), and the anchor oracle `EvaluateAnchors` (`anchors.go:5-31`) whose key trick is whitespace-normalized matching so line breaks don't defeat phrase detection. The design distinction it encodes: a page may *mention* a neighbor's figure in prose, but must not *render its caption* — that's the bleed signal.

**`internal/bookprofile`** (~330 lines) — the per-book policy layer and the single most important package for productization. `Profile` (`profile.go:40-107`) already models page-image globs and number regexes, prompt policy, vocabulary (preferred/protected/historical terms), known page types, QA expectations, normalization, figure policy, and context policy. `Resolve(bookID, profilePath)` (`profile.go:132-143`) loads YAML or falls back to the built-in `Report794()` for three hardcoded book IDs. `BuildPatch` (`discovery.go:113-134`) proposes profile updates from run observations. The gap: `ocrpipeline` and `ocrmvp` don't consume it — only `ocrquality` does.

**`internal/ocrmvp`** (~1,100 lines) — the legacy freeform path. Still the source of two reused pieces: `DiscoverPageImages` (`discover.go:17-52`; glob → sort → infer page number from the last digit-run in the filename) and the dynamic fan-out pattern. Its five versioned prompt strings (`prompt.go`) tell the productization story in miniature: v1 generic → v5 with the whole Report-794 lexicon pasted into the prompt. Superseded by `ocrpipeline`, retained for comparison runs.

**`internal/vlmseparation`** (~2,300 lines, best test coverage in the repo) — the diagnostic benchmark: six multimodal layouts × risky pages, scored against hand-built anchor oracles for 16 Report-794 pages (`oracle.go:101-124`), with layered response parsing, SQLite + turns persistence, a rescore command that re-measures old runs with improved parsers (`rescore.go`), and a report merger that combines runs picking best-per-cell (`report.go:154-169`). Keep it: for a product it becomes the model-qualification harness ("is model X safe for OCR at temperature Y?").

### 4. The AI substrate (geppetto / pinocchio)

book-ocr never talks to a provider SDK directly. It builds a `turns.Turn` (typed blocks: system, user-multimodal with image payload maps, LLM text) and calls `eng.RunInference(ctx, turn)` on an engine resolved from a **profile registry**: `--profile gpt-5-mini-low --profile-registries /path/profiles.yaml` → pinocchio `profilebootstrap.ResolveCLIEngineSettings` → concrete provider/model/settings (`client.go:49-67`). Turns persist into a Pinocchio-compatible chat store, so past model conversations are replayable/inspectable with existing tooling. Versions: `geppetto v0.11.28`, `pinocchio v0.10.26` — properly pinned, unlike scraper. For an eventual scripting/extension story, `go-go-goja` (same monorepo) provides a `require()`-based way to expose Go services to embedded JavaScript, but nothing in book-ocr uses it today and nothing in this plan requires it.

### 5. End-to-end pseudocode

The whole system in one page, for the intern who wants the runtime shape without reading 5,000 lines:

```text
structured-run(bookID, imageDir, range, workDir, profile, options):
    rt = workflow.NewRuntime(SQLiteStore(workDir/engine.db),
                             FileArtifactStore(workDir/artifacts),
                             SQLiteProjectionStore(workDir/projections),
                             queues={control:1, vision:maxWorkers, assemble:1, validation:1})
    client = dryRun ? DryRunClient : GeppettoClient(profile, registries)
    RegisterStructuredWorkflow(rt, client, workDir)
    run = rt.StartRun("book-ocr/structured", input)
    loop: rt.RunOnce() until run terminal          # poll loop, prints progress

executor discover(step):
    pages = DiscoverPageImages(imageDir, glob, start, end)
    for p in pages:
        seedProjectionRow(p, "pending")
        h[p] = step.Emit("structured-page-%03d" % p, PageInput(p),
                         queue=vision, retry={3, exp 1s→30s})
    a = step.Emit("assemble-structured-markdown", ..., dependsOn=Require(h...))
    step.Emit("validate-structured-run", ..., dependsOn=Require(a))

executor page(step, in):
    markRunning(projection, in.page)
    img = read(in.imagePath)
    turn = systemBlock(SYSTEM_PROMPT) + userBlock(RenderPrompt(bookID, page), img)
    assert CountTurnImages(turn) == 1              # THE invariant
    write(01-turn-input.yaml); resp = client.OCRPage(turn)
    write(02-turn-final.yaml, 03-raw-response.json)   # BEFORE parsing
    page = Parse(resp)          # strict → fence-strip → sanitize → regex repair
    page = Repair(page)         # fill block IDs, backfill captions, dedupe headings
    if page.number != in.page: fail Permanent("structured_page_mismatch")
    md  = RenderPageMarkdown(page)                 # deterministic
    val = ValidateStructuredPage(page, md)
    write(04-structured.json, 05-rendered.md, 06-validation.json)
    step.StoreArtifact(json, md, validation); markSucceeded(projection, counts)

executor assemble(step, in):
    results = [step.DependencyData(op) for op in pageOps]; sort by page
    assembled.md = concat(r.renderedMarkdown)
    if embedFigures: embedded-figures.md = EmbedExtractedFigures(assembled, imageDir)
    if renderPDF:    book.pdf = pandoc(xelatex, margins, DejaVu fonts)

executor validate(step, in):
    pages = splitRenderedPages(assembled)          # on "<!-- page:" markers
    warn += DetectAdjacentDuplicateFigureCaptions(pages)
    if len(pages) != expectedPages: fail Permanent  # acceptance gate
    warn += shortPages(projection, minRenderedBytes)
    write(validation-report.json)
```

## Part IV — Gap Analysis: What Stands Between This and a Product

Each finding names the evidence and the phase that fixes it.

**F1 — Unpublished, path-coupled runtime dependency.** `go.mod:125`: `replace github.com/go-go-golems/scraper => ../scraper`, required version `v0.0.0`. A clean clone does not build (verified; Appendix A). The workflow runtime has never been tagged or published, and book-ocr silently tracks its working tree. *Fix: Phase 1.*

**F2 — Three sources of truth for book policy.** The Report-794 lexicon lives in (a) `ocrpipeline/prompts.go:19-20` (prompt text), (b) `bookprofile/profile.go:145-174` (`Report794()`), (c) `ocrquality/markdown.go:82-114` (QA defaults) — and a fourth shadow copy in the legacy `ocrmvp/prompt.go:142-193`. `bookprofile.Resolve`'s bookID switch only knows Report-794 aliases (`profile.go:137-143`). The structured pipeline — the production path — does not read `bookprofile` at all. *Fix: Phase 2.*

**F3 — Language and content-heuristic hardcoding.** Every code block renders as ` ```common-lisp ` (`renderer.go:121`); the prompt asserts all listings are Common Lisp (`prompts.go:20`); figure-suppression cue lists include `ppscalc` (`renderer.go:94`); `textualFigureFallback` rewrites "items A, B and C" into a Lisp-style set (`renderer.go:102-114`); dry-run fixtures are pages 12/13/32 of this one book (`structured_ocr.go:59-79`). *Fix: Phase 2.*

**F4 — Page-naming convention hardcoded outside the profile.** `page_%03d.png` and `<!-- page:%03d -->` (3-digit) are baked into `ocrquality/figures.go:212`, `ocrmvp/discover.go`, `vlmseparation/runner.go:197-198`, while `bookprofile.PageImagePolicy` already has `Glob`/`PageNumberRegex` fields that these code paths ignore. Books over 999 pages, or scans named differently, break silently. *Fix: Phase 2.*

**F5 — Rerun operator bypasses the runtime.** `structured-rerun-pages` → `requeueStructuredPages` (`cmd/book-ocr/main.go:590-665`) issues raw SQL against `workflows`/`ops`/`leases`, including a hardcoded `retry_state_json` literal and in-place mutation of the assemble op's `input_json`. It works, but couples the CLI to the engine's private schema; any scraper migration breaks it. Acknowledged in `README.md:470`. *Fix: Phase 4 (needs a scraper-side API; F1 makes that coordinatable).*

**F6 — Developer-grade CLI surface.** Hand-rolled `switch` dispatch with stdlib `flag` (`main.go:49-87`); usage text still says `ocr-mvp` (`main.go:90-106`); `--dry-run` defaults to `true` on `structured-run`/`structured-page` so a user's first real run silently produces fake output (`main.go:226,299`); queue configs for all three workflow families are registered unconditionally (`main.go:894-903`). The go-go-golems house standard (Cobra + Glazed, help system, structured output) is already used by `vlm-separation` (`command.go:309-346`) but not the main commands. *Fix: Phases 1–2.*

**F7 — Validation gates are byte-thresholds, not page-type-aware.** Short-page detection is `--min-rendered-bytes` (README limitation `README.md:473`); code-fence auditing (Lisp-looking lines outside fences) is a manual `rg` playbook (`README.md:380-398`); page markers are invisible in the PDF, making manual review navigation harder (`README.md:471`). *Fix: Phases 2 and 4.*

**F8 — No ingestion or packaging story.** Input must already be a directory of page PNGs (the Report-794 pages were prepared outside this repo); output is Markdown + PDF only (no EPUB/HTML), and there's no `book-ocr init` to start a new book. Cost/token reporting per run is absent (turns are persisted but not aggregated). *Fix: Phase 3.*

**F9 — No CI, no releases.** `.github/` exists with go-template boilerplate but there is no green pipeline proving a clean-machine build (F1 makes that impossible today), no versioned binary releases (`.goreleaser.yaml` exists, unused since the module rename), no `docker run` path. *Fix: Phase 1.*

None of these findings are architectural. That is the central assessment: **the layering is right, the invariants are right and empirically justified, the artifacts are right; what's missing is the packaging, the configuration seam, and the operational polish.**

## Part V — Product Definition

### Decision: What "product" means for this codebase

- **Context:** "Product" could mean a hosted OCR service, a desktop app, a Go library, or a polished CLI. The codebase is a Go CLI over local SQLite state with per-book YAML policy; the user base to date is one power user, but the artifact model (inspectable pages, targeted rerun) is its differentiator.
- **Options considered:** (1) CLI-first product ("the book OCR pipeline you can audit"), installable binary + profiles + docs; (2) hosted web service with upload/review UI; (3) Go library (`pkg/bookocr`) for embedding; (4) all three at once.
- **Decision:** CLI-first (option 1), with the library extraction happening naturally as part of Phase 2 (policy/config seam) and a review web UI deferred to Phase 4 as an optional local server. No hosted multi-tenant service in this plan.
- **Rationale:** The system's strengths — durable local runs, full artifact audit trails, targeted rerun — are exactly what a technically sophisticated archivist/researcher wants from a CLI, and none of them require a server. A hosted service adds auth, storage, tenancy, and API-key custody problems orthogonal to every finding F1–F9. The runtime already supports multi-process workers on one SQLite file, so a local web UI later is cheap.
- **Consequences:** Product surface = `book-ocr` binary + book-profile YAML + documentation. Success criterion: *a stranger with a directory of page scans and an API key produces a validated book PDF without editing Go code.* A hosted offering remains possible later because Phase 2 forces all policy through data, not code.
- **Status:** proposed

### The target user and the promise

The user is someone digitizing a technical book: a retro-computing archivist, a researcher with a scanned thesis, a publisher resurrecting an out-of-print manual. The promise, phrased as product copy: *"Point it at your scans. Every page becomes structured, searchable Markdown — tables as tables, code as code, figures as cropped images — with a paper trail for every model call, validation that tells you which pages to distrust, and one-command repair for the pages you fix."*

That promise is already 80% implemented. The remaining 20% is making it true for books that are not MIT Report 794.

### Product-shaping decision records

### Decision: Publish the workflow runtime rather than vendor or in-line it

- **Context:** F1. book-ocr needs `scraper/pkg/workflow` to build. The user's constraint: don't modify go-go-golems repos (here) — but publishing/tagging is a repo-owner action we can specify.
- **Options considered:** (1) Tag and require `github.com/go-go-golems/scraper` at a real version; (2) extract `pkg/workflow`+`pkg/engine` into a new standalone `go-go-golems/workflow` module and depend on that; (3) vendor the ~5k lines into book-ocr's `vendor/` or fork into `internal/`; (4) keep the replace and document the required checkout layout.
- **Decision:** Option 2 as the target (a scraper-independent `workflow` module), with option 1 as the immediate Phase-1 unblocker if extraction can't happen right away.
- **Rationale:** The runtime is already domain-free (BOOK-OCR-EXTERNALIZE-001 verified scraper contains no OCR); "scraper" is a misleading home for a generic durable-workflow engine and drags scraper's dependency tree (goja modules, engineview) into every consumer. Vendoring (3) forks the code and forfeits upstream fixes; (4) is the status quo that broke the build.
- **Consequences:** Requires one go-go-golems-side change (tag or extract), performed outside this ticket. book-ocr's go.mod drops the `replace`, CI can build from a clean clone. Until then, developers use the documented `go.work` pattern (`scripts/book-ocr.go.work`).
- **Status:** proposed

### Decision: The book profile becomes the single source of all book policy

- **Context:** F2–F4. Four places currently encode Report-794 knowledge; the production pipeline reads none of the profile layer.
- **Options considered:** (1) Extend `bookprofile.Profile` and thread it through prompt building, rendering, validation, discovery, and dry-run fixtures; (2) template the prompts only (Go `text/template` with per-book variable files) and leave renderer/QA as flags; (3) per-book Go plugins/scripting (go-go-goja).
- **Decision:** Option 1. `Resolve(bookID, profilePath) → Profile` becomes mandatory in the structured path; every book-specific string in `prompts.go`, `renderer.go`, and `ocrquality` defaults moves behind a profile field with sensible generic defaults; `Report794()` becomes `profiles/report-794.yaml`, a data file and regression fixture.
- **Rationale:** The profile type already models 90% of what's needed (`profile.go:40-107`) and the discovery/patch loop already targets it — this is completing an existing design, not inventing one. Prompt-only templating (2) leaves the renderer's `common-lisp` fence and `ppscalc` cues hardcoded, which F3 shows are just as book-specific as the prompt. Scripting (3) is power without need: no current requirement is beyond declarative data, and a JS surface multiplies the testing matrix.
- **Consequences:** New profile fields needed: `CodePolicy{DefaultLanguage, LanguageHints}`, `RenderPolicy{WrapWidth, FigureSuppressionCues, PageMarkerFormat}`, `LexiconPolicy` feeding the prompt, plus prompt-template selection. Every consumer signature grows a `*bookprofile.Profile` (or a compiled `RunPolicy`) parameter. Report-794 outputs must be byte-identical before/after the migration (golden tests).
- **Status:** proposed

### Decision: Keep structured-JSON-plus-Go-rendering as the only production OCR mode

- **Context:** One could "simplify" the product by letting the model emit Markdown directly (the freeform path still exists in `ocrmvp`), or chase higher fidelity by sending neighbor pages for context.
- **Options considered:** (1) Structured-only (status quo); (2) offer freeform as a user-selectable mode; (3) reintroduce bounded neighbor context for continuity.
- **Decision:** Structured-only for production output. Freeform (`ocrmvp`) stays as an internal comparison harness; neighbor context stays confined to diagnostics (`vlm-separation`) and future *text-only* continuity passes (neighbor rendered text, never neighbor images).
- **Rationale:** The VLM separation benchmark measured the alternatives: every layout that included context images leaked adjacent-page content; `target-only` scored ~0.94. The 202-page freeform run produced the context-bleed defect class that cost the most review time. Determinism of Go rendering is also what makes golden-file testing and targeted rerun meaningful.
- **Consequences:** Prose-completeness (structured output was lower-volume than freeform on the first-50 comparison) must be addressed by prompt iteration and validation gates, not by abandoning the architecture. The product docs state the invariant as a guarantee.
- **Status:** accepted (inherited from BOOK-OCR-PIPELINE-REDESIGN-001; re-affirmed)

### Decision: Ship ingestion as `book-ocr ingest` wrapping pdftoppm

- **Context:** F8. Real users start from a PDF (or a scanner run), not a prepared `page_*.png` directory.
- **Options considered:** (1) Shell out to poppler's `pdftoppm` (and `magick` for image normalization) behind a new `ingest` command; (2) pure-Go PDF rasterization; (3) document manual preparation only.
- **Decision:** Option 1, with tool presence checked at startup and a clear error naming the missing binary (same pattern as the existing Pandoc dependency).
- **Rationale:** The pipeline already accepts external binaries for output (pandoc/xelatex, `workflow_executors.go:344-366`); mirroring that for input is consistent and cheap. Pure-Go rasterizers are a maintenance burden with worse output fidelity. DPI (300 default), grayscale, and page-naming then conform to the profile's `PageImagePolicy` by construction, which retires most of F4's risk for new books.
- **Consequences:** `ingest` writes `pages/page_NNNN.png` (4-digit — see Phase 2 note on marker width) plus an `ingest-manifest.json` (source PDF hash, DPI, page count) that `init` and `structured-run` can validate against. Photographed-book dewarping is explicitly out of scope; document it.
- **Status:** proposed

### Decision: Replace rerun-by-SQL with a runtime `RequeueSteps` API

- **Context:** F5. Targeted rerun is the flagship review-loop feature and currently the most fragile code in the repo.
- **Options considered:** (1) Add `Runtime.RequeueSteps(runID, stepIDs, ResetOptions)` to the workflow runtime (scraper-side change) and call it from book-ocr; (2) keep SQL surgery but pin the engine schema with a version check; (3) model rerun as a *new* workflow that depends on existing artifacts.
- **Decision:** Option 1, requested from the runtime owner alongside the F1 tagging work; option 2's version check as an interim safety (refuse to requeue if `schema_migrations` contains unknown versions).
- **Rationale:** Resetting op state, clearing leases/retry state, and flipping downstream ops to `pending` is scheduler business — the current implementation re-implements `RefreshRunnableOps`' invariants by hand (`main.go:590-665` vs `op_store.go:104-210`). A first-class API is also what a future web UI needs.
- **Consequences:** One cross-repo API negotiation; interim schema guard is ~20 lines. Rerun semantics (downstream to `pending`, not `ready`) get runtime tests instead of being a CLI comment (`main.go:627-629`).
- **Status:** proposed

## Part VI — Proposed Target Architecture

Phases 1–4 transform the repo without moving its skeleton. The target shape:

```text
book-ocr/
  cmd/book-ocr/            Cobra root; subcommands in cmd/book-ocr/cmds/
  pkg/bookocr/             public embedding API (extracted during Phase 2)
      profile/             (moved from internal/bookprofile, now public)
      pipeline/            structured OCR: types, prompts, renderer, client
      quality/             QA, figures, normalization
      validation/          deterministic validators
  internal/ocrmvp/         legacy freeform path (comparison harness only)
  internal/vlmseparation/  model-qualification benchmark
  profiles/                report-794.yaml, generic-technical-book.yaml, ...
  docs/                    user guide, profile reference, operations runbook
```

### The generalized book profile (API sketch)

```yaml
# profiles/generic-technical-book.yaml
id: generic-technical-book
family: technical-report
page_images:
  glob: "page_*.png"
  page_number_regex: "page_(\\d+)\\.png"
  marker_format: "<!-- page:%04d -->"      # width now profile-owned (F4)
prompt:
  template: structured-ocr/v1               # named, versioned prompt template
  language_note: ""                         # e.g. "Code listings are Common Lisp..."
  lexicon:                                   # injected as "preserve exactly" terms
    protected_terms: []
    historical_spellings: []
  examples: []                               # optional few-shot blocks
code:
  default_language: ""                       # "" => plain ``` fence
  language_hints: {}                         # e.g. {"defmethod": common-lisp}
render:
  wrap_width: 88
  figure_suppression_cues: [code listing, code sample, spreadsheet, grid]
  include_footers: false
qa:
  expected_pages: 0                          # 0 => derived from discovery
  known_bad_terms: []
  expected_strings: []
  min_rendered_bytes_by_page_type:           # F7: type-aware thresholds
    body: 200
    figure: 40
    blank: 0
figures:
  segmentation: ink-band-v1
  margin_px: 24
```

Loading pseudocode (replacing today's scattered constants):

```text
RunPolicy = Compile(Resolve(bookID, profilePath))
  # Compile validates, fills defaults, and produces the three consumers' views:
  #   PromptSpec  -> RenderStructuredOCRPrompt(spec, page)
  #   RenderOpts  -> RenderPageMarkdown(page, opts)
  #   QASpec      -> ValidateStructuredPage / validate-structured-run
StructuredRunInput gains: ProfilePath string   (workflow input carries the policy)
```

### CLI after productization

```text
book-ocr init      --pdf book.pdf --book-id my-book [--profile-from generic-technical-book]
                     → ingest pages, draft my-book.profile.yaml via discovery dry pass
book-ocr run       --book-id my-book --work-dir runs/my-book --live   # structured-run, renamed;
                                                                      # --live replaces --dry-run=false (F6)
book-ocr status|pages|resume|retry|cancel --work-dir ...              # unchanged semantics
book-ocr rerun     --pages 12,40-45 --render-pdf                      # first-class requeue (F5)
book-ocr report    --work-dir ...                                     # cost/tokens/warnings summary (F8)
book-ocr audit     --work-dir ...                                     # code-fence & figure audits, was rg playbook (F7)
book-ocr qualify-model --profile X                                    # vlm-separation, repositioned
```

### The review loop as a product feature

The manual validation workflow in `README.md:337-363` is the actual product core and survives unchanged — the phases just automate its sharp edges:

```text
   run ──▶ book.pdf ──▶ human review ──▶ pages look wrong?
    ▲                                        │
    │                                        ▼
    │                          inspect pages/page_NNN/03..06
    │                          (04 wrong → prompt/model | 05 wrong → renderer)
    │                                        │
    └──── rerun --pages ... ◀── fix profile / prompt / renderer
                    │
                    └──▶ repeated finding? add a validation rule (F7 audit command)
```

## Part VII — Phased Implementation Plan

Each phase is landable independently, ordered by risk reduction per unit effort. File paths are current-layout; Phase 2's package moves adjust them.

### Phase 1 — Buildable, installable, honest (est. 2–4 days)

1. **Resolve F1.** Preferred: extract/tag the workflow module (owner action in go-go-golems); then in `go.mod` drop line 125 and require the real version. Interim: commit `docs/development.md` describing the `go.work` pattern with `scripts/book-ocr.go.work` as the canonical copy.
2. **CI (F9).** GitHub Actions: `go build ./...`, `go test ./... -count=1`, `golangci-lint run`, plus the 3-page dry-run smoke (`scripts/01-dry-run-structured-pipeline.sh` minus PDF, or with `pandoc` installed in the runner). This is the regression net for every later phase.
3. **Fix F6's sharp edge.** `--dry-run` default flips to `false` with an explicit `--dry-run` opt-in (`cmd/book-ocr/main.go:226,299`); add a startup banner stating profile, model, page count, and estimated call count before live runs; fix the `ocr-mvp` usage strings (`main.go:90-106`).
4. **Releases.** Re-point `.goreleaser.yaml` at `cmd/book-ocr`, tag `v0.1.0`, verify `go install github.com/go-go-golems/book-ocr/cmd/book-ocr@latest`.

Exit criterion: a clean CI machine builds, tests, dry-runs, and publishes a binary.

### Phase 2 — Profile-driven generalization (est. 1–2 weeks)

1. **Extend `bookprofile`** with `CodePolicy`, `RenderPolicy`, `LexiconPolicy`, `marker_format`, and type-aware QA thresholds (fields sketched above); add `Compile() (RunPolicy, error)` with validation.
2. **Thread the profile through the structured path:**
   - `prompts.go`: `RenderStructuredOCRPrompt(spec PromptSpec, bookID, page)` — the Report-794 lexicon sentence and Lisp assertion become template data (`prompts.go:19-20`); the PPSCalc example moves to `profiles/report-794.yaml` `examples:`.
   - `renderer.go`: `RenderOptions` gains `CodeLanguage`, `SuppressionCues`, `MarkerFormat`; delete the hardcoded `common-lisp` (`renderer.go:121`) and cue lists (`renderer.go:84,94`); `textualFigureFallback` (`renderer.go:102-114`) becomes a profile-gated transform, default off.
   - `workflow_types.go`/`workflow_package.go`: `StructuredRunInput.ProfilePath`; executors compile the policy once and pass it down.
   - `ocrquality`: replace `defaultKnownBadTerms`/`defaultExpectedStrings`/`defaultListPages` (`markdown.go:82-114`) with profile values (plumbing already half-exists in `package.go:128-171`).
   - Retire the page-naming hardcodes (F4): `figures.go:212`, `vlmseparation/runner.go:197-198` read the profile's glob/regex/marker format.
3. **Move `Report794()` to `profiles/report-794.yaml`**; keep the function as a loader of that file for one release.
4. **Golden regression tests** (see Testing) proving Report-794 rendering is byte-identical pre/post migration — this is the phase's safety net.
5. **Generalize dry-run fixtures** (`structured_ocr.go:59-79`): a generic synthetic book (title page, prose page, table page, code page, figure page) so dry runs demonstrate every block type for any book ID.
6. **Cobra/Glazed migration** of the CLI (F6), preserving flag names; `structured-run` → `run` with alias.

Exit criterion: a second real book (any scanned technical PDF) runs live to a validated artifact set with **zero Go changes** — only a profile YAML. This is the product's definition of done.

### Phase 3 — Ingestion and onboarding (est. 1 week)

1. **`book-ocr ingest`** (decision record above): pdftoppm wrapper, DPI/grayscale flags, manifest, profile-conformant naming.
2. **`book-ocr init`**: ingest (if given a PDF) → run a cheap discovery pass (dry-run + first/middle/last live sample pages) → emit a drafted `<book>.profile.yaml` using `bookprofile.BuildPatch` (`discovery.go:113-134`) → print what the user should review (expected pages, detected TOC pages, language guess).
3. **`book-ocr report` (F8):** aggregate the turns DB and projection into tokens/calls/failures/durations per run; print cost using per-profile pricing from a small static table.
4. **User docs:** quickstart (PDF → book.pdf in four commands), profile reference generated from the schema, operations runbook distilled from `README.md:268-405`.

Exit criterion: the stranger-with-a-PDF scenario from the product definition completes in under 30 minutes of wall-clock attention.

### Phase 4 — Operational maturity (est. 1–2 weeks, partially cross-repo)

1. **First-class rerun (F5):** land `Runtime.RequeueSteps` upstream; replace `requeueStructuredPages` (`main.go:590-665`); interim schema-version guard immediately.
2. **`book-ocr audit` (F7):** codify the README's `rg` playbooks — Lisp-looking lines outside fences, table-like image captions, figure/caption count mismatches — as validation rules feeding `validation-report.json`.
3. **Review PDF mode:** visible page-source markers in the PDF (currently HTML comments vanish — `README.md:471`); optional two-up source-image-vs-render layout for spot checking.
4. **Runtime hardening requests upstream:** lease heartbeats for long page steps and cooperative cancellation (documented gaps since SCRAPER-JOBS-001).
5. **Optional local review UI:** a `book-ocr serve` that reads `engine.db`/projections read-only: page grid colored by status/warnings, click-through to `03/04/05` artifacts, rerun buttons calling the Phase-4.1 API. Deferred until the CLI product has users; the projection layer makes it a rendering exercise.

## Part VIII — Testing Strategy

The repo's existing discipline (fake clients, deterministic fixtures, integration tests over temp SQLite — e.g. `workflow_retry_test.go`'s fail-once-then-succeed client) extends naturally:

1. **Golden-file rendering tests (Phase 2 gate).** Fixture set of `04-structured.json` files from the real Report-794 run → rendered Markdown compared byte-for-byte before/after profile threading. Add per-block-type fixtures (table edge cases: ragged rows, missing headers; figure suppression on/off).
2. **Profile compile tests.** Valid/invalid YAML, default filling, unknown fields rejected (typo in `figure_suppression_cues` should fail loudly, not silently no-op).
3. **Second-book integration test.** A tiny synthetic "book" (5 generated PNGs with known text via the dry-run fixture generator) with a *non-Report-794* profile, dry-run through the full workflow in CI; asserts F2/F3/F4 stay fixed.
4. **Parser adversarial corpus.** The `03-raw-response.json` files from real runs are a free corpus; snapshot the parse results so `jsonsanitize`/repair changes are visible diffs. (The vlmseparation `rescore` command is the model for this.)
5. **Requeue semantics tests (Phase 4).** Runtime-level: requeue a mid-graph step, assert downstream ops flip to pending and re-derive readiness; assert lease/retry-state clearing.
6. **Live smoke, guarded.** Keep the existing pattern (env-var-gated live tests, e.g. `geppetto_ocr_test.go`) — one page through a cheap profile in a nightly job, never in PR CI.
7. **Model qualification.** `vlm-separation` runs against any new default profile/model before it's recommended in docs; its target-only score and parse-strategy distribution are the acceptance numbers.

## Part IX — Risks, Alternatives, Open Questions

**Risks**

- *Cross-repo coupling remains the long pole.* Phases 1 and 4.1 need go-go-golems-side actions (tag/extract, RequeueSteps). Mitigations are specified (go.work docs, schema guard), but the product's clean-clone story is hostage to F1 until the owner acts.
- *Prompt generalization can regress Report-794 quality.* Removing the lexicon from the prompt and re-injecting it as template data changes token order; models are sensitive to phrasing. Mitigation: golden tests catch renderer changes, but prompt changes need an A/B rerun of the 16 oracle pages (`vlmseparation/oracle.go:101-124`) before merging.
- *Prose completeness is still unproven at book scale.* The first-50 structured run produced less prose than freeform. If prompt iteration can't close this, the product needs a documented "completeness review" step; the type-aware byte thresholds (Phase 2) are detection, not cure.
- *SQLite as the only backend* caps multi-machine scale. Fine for the CLI product; the River/Postgres option investigated in SCRAPER-JOBS-001 remains the escape hatch if a hosted offering ever happens.
- *Model/provider drift.* `gpt-5-mini-low` behavior (JSON discipline, schema quirks the repair layer compensates for) will change under us. The adversarial parse corpus and qualification benchmark are the detection mechanism.

**Alternatives set aside** (beyond the decision records): a general-document product (invoices, forms — different competition and structure model); fine-tuning a dedicated OCR model (data volume nowhere near sufficient; prompt+repair is working); go-go-goja scripting for per-book hooks (declarative profiles cover all observed needs; revisit only if a real book demands imperative logic).

**Open questions**

1. Who owns tagging/extracting the workflow runtime, and on what timeline? (Blocks Phase 1 exit.)
2. Should `pkg/bookocr` be public API in v0.2, or stay `internal/` until a second consumer exists? (Leaning: keep internal until asked; "pkg extraction" in Phase 2 can mean layout only.)
3. Is EPUB/HTML output in scope for the product's v1, or is Markdown+PDF enough? (Markdown+pandoc makes EPUB nearly free; deciding factor is validation effort.)
4. Do we keep the legacy `ocrmvp` runnable (`run` command) in the released binary, or demote it to test-only? (Leaning: test-only; it confuses the product surface.)
5. What is the pricing-table maintenance story for `book-ocr report` cost estimates — static table in-repo, or omit costs and report tokens only?

## References

**This repo (evidence anchors used above)**

- `cmd/book-ocr/main.go` (997 lines) — CLI dispatch `:49-87`, runtime wiring `:886-905`, rerun SQL `:590-665`
- `internal/ocrpipeline/` — `types.go:11-190`, `prompts.go:5-77`, `renderer.go:19-241`, `structured_ocr.go:38-321`, `client.go:14-165`, `session.go:29-99`, `workflow_executors.go:20-399`, `workflow_package.go:18-92`, `workflow_projection.go:11-100`
- `internal/ocrquality/` — `figures.go:47-456`, `qa.go:16-222`, `package.go:38-377`, `markdown.go:11-114`, `normalize.go:24-150`
- `internal/ocrvalidation/` — `adjacent.go:12-60`, `anchors.go:5-40`, `types.go:3-33`
- `internal/bookprofile/` — `profile.go:11-192`, `discovery.go:9-134`
- `internal/ocrmvp/` — `discover.go:17-172`, `prompt.go:17-216`, `geppetto_ocr.go:16-118`, `package.go:18-103`
- `internal/vlmseparation/` — `oracle.go:101-124`, `scoring.go:30-426`, `runner.go:28-255`, `report.go:154-169`, `command.go:309-369`
- `go.mod:125` — the `../scraper` replace directive (F1)
- `README.md` — operations guide `:268-405`, design rules `:421-429`, limitations `:468-474`

**Dependency (read-only): `~/code/wesen/go-go-golems/scraper/pkg/workflow/`** — `runtime.go:60-340`, `package.go:12-128`, `executor.go:13-166`, `context.go:18-299`, `errors.go:15-83`, `operators.go:15-52`, `artifact_store.go:14-105`, `projection_store.go:18-84`; engine: `scheduler/scheduler.go:197-551`, `store/sqlite/migrations/00{1,2}_*.sql`, `store/sqlite/{op_store.go:104-210, result_store.go:14-102, lease_store.go:14+}`

**History (ttmp tickets, 2026-05):** SCRAPER-JOBS-001, OCR-MVP-001, BOOK-OCR-HQ-001, OCR-QUALITY-WORKERS-001, BOOK-OCR-EXTERNALIZE-001, BOOK-OCR-FULL-001, BOOK-OCR-VLM-SEPARATION-001, BOOK-OCR-PIPELINE-REDESIGN-001, BOOK-OCR-STRUCTURED-WORKFLOW-001 — each under `ttmp/2026/05/2{4,5}/…/` with `index.md`, `design-doc/`, `reference/01-diary.md`

**This ticket:** `scripts/book-ocr.go.work` (build workaround), `scripts/01-dry-run-structured-pipeline.sh` (E2E health check), `reference/01-investigation-diary.md`

## Appendix A — Experiments run for this assessment (2026-07-03)

1. **Standalone build check.** `go build ./...` in a clean environment → FAIL: `github.com/go-go-golems/scraper@v0.0.0: replacement directory ../scraper does not exist` (packages `ocrmvp`, `ocrpipeline`, `ocrquality`; `bookprofile`, `ocrvalidation`, `vlmseparation` unaffected). Confirms F1.
2. **Corrected build + full tests.** With `GOWORK=scripts/book-ocr.go.work` (note: the workspace `go` directive had to be raised to 1.26.4 to satisfy scraper's go.mod): `go build ./...` OK; `go test ./... -count=1` → all 6 test packages `ok` (`ocrpipeline` 1.26s is the slowest — the workflow integration tests).
3. **End-to-end dry-run** (`scripts/01-dry-run-structured-pipeline.sh`): 3 pages of Report 794, `--embed-figures --render-pdf --expected-pages 3`, deterministic fake client. Workflow `book-ocr/structured-6fdeff74…` succeeded; all artifacts present: `assembled.md` (206 B), `embedded-figures.md`, `book.pdf` (10.9 KB via pandoc 3.1.3 / XeTeX 3.141592653), `validation-report.json` (`page_count: 3, expected_pages: 3, warning_count: 0`), `engine.db` (112 KB), `turns.db` (3.0 MB), and the full `pages/page_NNN/01–06` sequence. Confirms the pipeline, workflow engine, figure embedding, PDF toolchain, and validation all function today.
