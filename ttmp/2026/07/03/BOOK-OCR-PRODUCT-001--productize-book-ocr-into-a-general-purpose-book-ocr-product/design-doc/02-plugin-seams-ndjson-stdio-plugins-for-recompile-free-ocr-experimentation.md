---
Title: 'Plugin seams: NDJSON-stdio plugins for recompile-free OCR experimentation'
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
    - Path: ../../../../../../../go-go-golems/devctl/pkg/protocol/types.go
      Note: Wire format adopted (protocol v2 frames)
    - Path: ../../../../../../../go-go-golems/devctl/pkg/runtime/factory.go
      Note: Host spawn/handshake/shutdown imported per D1
    - Path: internal/bookprofile/profile.go
      Note: D4 — profile gains Plugins section
    - Path: internal/ocrpipeline/client.go
      Note: S1 attach point — StructuredOCRClient interface
    - Path: internal/ocrpipeline/prompts.go
      Note: S2 attach point — prompt rendering
    - Path: internal/ocrpipeline/workflow_executors.go
      Note: S6/S7 attach points and error classification to mirror
    - Path: internal/ocrquality/figures.go
      Note: S5 attach point — figure segmentation to extract behind FigureSegmenter
ExternalSources: []
Summary: Where to cut devctl-style NDJSON-stdio plugin seams into the book-ocr pipeline, what each seam lets an experimenter change without recompiling, the host architecture (importing devctl pkg/protocol and pkg/runtime), decision records, and an implementation plan backed by a working protocol prototype.
LastUpdated: 2026-07-03T11:04:37.840525836-07:00
WhatFor: Design the plugin architecture that lets OCR strategies be swapped per book type via external processes.
WhenToUse: Read together with design doc 01; implement as the plugin track of the productization plan.
---


# Plugin seams: NDJSON-stdio plugins for recompile-free OCR experimentation

## Executive Summary

This document extends the productization plan (design doc 01) with **plugin seams**: points in the pipeline where book-ocr calls out to an external process speaking newline-delimited JSON over stdin/stdout, following the devctl plugin protocol v2. The goal is stated by the ticket: *experiment with different methods of OCR-ing certain types of books without recompiling the project*.

Three facts make this cheap and low-risk:

1. **The transport already exists as importable Go packages.** devctl's `pkg/protocol` (frame types, ~135 LOC) and `pkg/runtime` (process spawn, handshake validation, request/response routing, stream events, process-group shutdown, ~490 LOC) are public, devctl-agnostic packages in `github.com/go-go-golems/devctl`. book-ocr defines only its own op names and input/output schemas — the equivalent of devctl's `pkg/engine` layer.
2. **The seams already exist as Go interfaces.** `ocrpipeline.StructuredOCRClient` (`client.go:34`), `ocrmvp.OCRClient` (`types.go:94`), and `ocrpipeline.FigureResolver` (`types.go:177`) are exactly where a plugin-backed implementation drops in; the remaining seams (prompt, parse, render-transform, figure segmentation, validation, classification) are single functions that need an interface introduced.
3. **The protocol round-trip is proven.** A working prototype (`scripts/02-plugin-protocol-demo/`: stdlib-only Go host + Python plugin) exercises handshake, capability discovery, `prompt.render`, `ocr.page` on a real Report-794 page image, interleaved progress events, and `E_UNSUPPORTED` fallback — exit 0 on this machine today.

The recommended seam set, ranked by experimentation value: `ocr.page` (whole-strategy swap — the headline seam), `figures.segment` (replace the pixel heuristics with Python/OpenCV — the strongest concrete use case), `prompt.render` and `response.parse` (cheap per-book prompt/parser experiments), `page.classify` (route pages to different strategies), `validate.page`/`validate.book` (book-type-specific QA), and `markdown.transform` (post-render tweaks). Plugins are declared in the **book profile**, so "try a different OCR method for this book type" is a YAML edit plus any-language script — no Go toolchain involved.

The critical design rule: **plugins replace strategies, never invariants.** The host keeps enforcing the single-target-image rule, schema validation, artifact writing, retry classification, and provenance recording regardless of what a plugin returns. A plugin can change *how* a page is OCR'd; it cannot change *what the pipeline guarantees*.

## Problem Statement

Design doc 01 makes book policy declarative via `bookprofile` (Phase 2). Declarative policy covers vocabulary, rendering, and QA thresholds — but not *method*. Concrete experiments the current architecture forces through a recompile:

- OCR a math-heavy book with a two-pass strategy (layout detection, then region-by-region transcription).
- Try `tesseract` + LLM cleanup instead of pure VLM OCR for clean-print books (cheaper, possibly more accurate for body text).
- Run an ensemble: three model calls per page, majority-vote the blocks.
- Replace the `ink-band-v1` figure segmentation (`ocrquality/figures.go:239-335` — hand-tuned row-scan heuristics) with OpenCV contour detection or a layout-analysis model, which want to live in Python.
- Add validation that checks LaTeX math balance for equation-dense books.
- Try a different prompt phrasing for novels vs technical reports without touching `prompts.go`.

Each of these is a *strategy* experiment on one pipeline stage. The devctl plugin model — local process, NDJSON stdio, capability handshake, language-agnostic — fits: experiments are scripts sitting next to the book profile, iterated at edit-rerun speed via `structured-rerun-pages`.

## Protocol Background (what we adopt from devctl)

Reference: devctl `pkg/protocol/types.go` and `pkg/doc/topics/devctl-plugin-authoring.md:199-301`. Summary of the v2 contract as book-ocr will use it:

```jsonc
// plugin → host, FIRST stdout line (handshake; stdout is protocol-only after this)
{"type": "handshake", "protocol_version": "v2", "plugin_name": "my-strategy",
 "capabilities": {"ops": ["ocr.page", "figures.segment"]},
 "declares": {"engine": "tesseract+gpt", "version": "1.2.0"}}

// host → plugin (stdin), one JSON object per line
{"type": "request", "request_id": "book-ocr-42", "op": "ocr.page",
 "ctx": {"repo_root": "/work/dir", "cwd": "", "deadline_ms": 120000, "dry_run": false},
 "input": { /* op-specific, see schemas below */ }}

// plugin → host: exactly one response per request; events may interleave
{"type": "event", "stream_id": "", "event": "log", "level": "info", "message": "…"}
{"type": "response", "request_id": "book-ocr-42", "ok": true, "output": { … }}
{"type": "response", "request_id": "book-ocr-42", "ok": false,
 "error": {"code": "E_UNSUPPORTED", "message": "…", "details": {"retryable": false}}}
```

Host-side mechanics we inherit by importing `devctl/pkg/runtime` (file:line references into `~/code/wesen/go-go-golems/devctl`):

- **Spawn + handshake validation** with timeout and stderr capture on failure: `Factory.Start` (`pkg/runtime/factory.go:51-121`); the plugin gets its own process group (`Setpgid`) so shutdown can SIGTERM→SIGKILL the whole tree (`factory.go:183-208`).
- **Concurrency-safe request correlation**: atomic request-id counter (`client.go:200-203`), mutex-guarded stdin writes (`client.go:205-214`), router mapping id → response channel (`router.go:26-58`). Multiple page workers can share one plugin process and call concurrently.
- **Host-authoritative timeouts**: `Call` selects on Go context vs response (`client.go:104-130`); `deadline_ms` is advisory for the plugin.
- **Protocol policing**: non-JSON stdout ⇒ `E_PROTOCOL_STDOUT_CONTAMINATION` and fail-all (`client.go:216-264`, `router.go:85-115`); stderr is forwarded to zerolog as the plugin's logging channel (`client.go:266-285`).
- **Capabilities as allowlist**: `SupportsOp` checked before dispatch (`client.go:70-81`) — this is what makes graceful fallback to built-in behavior trivial.

devctl has no plugin auto-restart (each CLI invocation spawns fresh); the analogous model for book-ocr is plugin lifetime = one workflow-run drain (Decision D3).

## Seam Analysis

The structured pipeline with every candidate seam marked. Stages are host-owned; a seam is a point where the host asks a plugin *instead of* (or *in addition to*) its built-in implementation:

```text
 ingest ─▶ discover ─▶ [classify] ─▶ page OCR ─▶ parse ─▶ render ─▶ assemble ─▶ validate ─▶ pdf
   S8         (host)      S6      ┌────S1────┐    S3        S4      + figures     S7
                                  │ prompt S2 │                        S5
                                  └───────────┘
   host-enforced regardless of plugins: single-image invariant · raw-response artifact ·
   schema validation · page-number gate · artifact sequence 01–07 · retry classification
```

### S1 — `ocr.page`: the whole-strategy seam (highest value)

- **Attaches to:** `ocrpipeline.StructuredOCRClient` (`internal/ocrpipeline/client.go:34-36`) — the interface already has live, dry-run, and test implementations, so a fourth (`PluginStructuredOCRClient`) is idiomatic.
- **What it lets you tweak:** everything about how a page image becomes `StructuredPageOCR` JSON — provider/model choice outside Geppetto, multi-pass strategies (layout → regions → merge), classical OCR + LLM cleanup, ensembles/voting, per-page-type strategies, local models. The plugin returns the same `structured-ocr/v1` JSON the model returns today; the host still validates, repairs, renders, and writes artifacts.
- **Input schema:** `{book_id, page_number, image_path, page_type_hint?, profile: {lexicon, prompt_policy, code_policy}}` — image passed **by path**, not base64 (local-first; a Python plugin opens it directly; frames stay small — verified in the prototype with a 480 KB page).
- **Output schema:** `StructuredPageOCR` JSON (as `types.go:11-18`) plus an `engine` provenance object; optionally `raw_response` so `03-raw-response.json` stays meaningful for LLM-based plugins.
- **Invariants kept host-side:** page-number mismatch gate (`structured_ocr.go:273-277`), schema validation/repair (`structured_ocr.go:81-203`), artifact writes, projection updates. The single-image rule becomes: *the host hands the plugin exactly one target page image path*; the host never supplies neighbor images.

### S2 — `prompt.render`: cheap prompt experiments

- **Attaches to:** `RenderStructuredOCRPrompt` (`internal/ocrpipeline/prompts.go:10-77`), consulted by the built-in Geppetto client (an S1 plugin owns its own prompting).
- **What it lets you tweak:** system/user prompt text per book type — phrasing, few-shot examples, lexicon injection strategy, "prose-first" vs "table-first" instruction ordering — the exact experiments the HQ-001 ticket did by editing Go constants.
- **Input:** `{book_id, page_number, schema_version, profile: {...}}`; **Output:** `{system, user}`. The host appends the non-negotiable schema contract itself so a plugin cannot accidentally drop the JSON-only instruction.

### S3 — `response.parse`: model-quirk repair

- **Attaches to:** `ParseStructuredOCRResponse` (`internal/ocrpipeline/structured_ocr.go:81-104`). The built-in layered parser (strict → fence-strip → sanitize → regex repair) stays as fallback when the plugin declines (`ok:false, code:E_DECLINED`) or is absent.
- **What it lets you tweak:** repair strategies for new models with different failure signatures — the repo's history shows each model generation needs its own repairs (leading-zero numbers, nested figure objects, `diagram_text` as string). Iterating these in Python against the `03-raw-response.json` corpus beats recompiling.
- **Input:** `{raw_response, book_id, page_number}`; **Output:** `StructuredPageOCR` JSON. Host still runs `repairStructuredPage` and validation on the result.

### S4 — `markdown.transform`: post-render adjustments

- **Attaches to:** immediately after `RenderPageMarkdown` (`internal/ocrpipeline/renderer.go:19-40`), per page.
- **What it lets you tweak:** house-style variations without forking the renderer — different figure-marker syntax, heading normalization, book-specific text substitutions. The `textualFigureFallback` hack (`renderer.go:102-114`) is exactly the kind of one-book transform that should have been a plugin.
- **Input:** `{page_number, markdown, structured: {...}}`; **Output:** `{markdown}`. Determinism caveat recorded in provenance: a transform plugin makes rendering only as deterministic as the plugin is.

### S5 — `figures.segment`: the strongest concrete use case

- **Attaches to:** the crop-computation core of `EmbedExtractedFigures` (`internal/ocrquality/figures.go:47-98`); requires extracting today's `cropNonWhite`/`inkBands`/`meaningfulInkUnion` pipeline (`figures.go:239-335`) behind a new `FigureSegmenter` interface with `ink-band-v1` as the built-in.
- **What it lets you tweak:** the whole computer-vision layer — OpenCV contour detection, deep layout models (LayoutParser-style detectors), per-book-family tuning. This is Python-ecosystem territory; a Go recompile loop is the wrong iteration medium for it, which makes this seam the clearest justification for stdio plugins over Go interfaces alone.
- **Input:** `{image_path, page_number, markers: [{caption, description}], hints: {margin_px, header_zone, footer_zone}}`; **Output:** `{figures: [{marker_index, crop: {x,y,w,h}, confidence, warnings: []}]}`. The host performs the actual cropping/writing so sidecars and debug overlays stay uniform.

### S6 — `page.classify`: routing pages to strategies

- **Attaches to:** a new step between discovery and page fan-out in `StructuredDiscoverExecutor` (`internal/ocrpipeline/workflow_executors.go:20-67`).
- **What it lets you tweak:** per-page strategy selection — "this page is a full-page diagram → vision-heavy plugin; this is body prose → cheap path". Output feeds `page_type_hint` into S1 input and can select among multiple configured `ocr.page` plugins. This is how "different methods for certain types of books" becomes "different methods for certain types of *pages*".
- **Input:** `{image_path, page_number, profile}`; **Output:** `{page_type, strategy?: string, confidence}`.

### S7 — `validate.page` / `validate.book`: pluggable QA rules

- **Attaches to:** `ValidateStructuredPage` (`structured_ocr.go:300-321`) and `StructuredValidateExecutor` (`workflow_executors.go:238-284`), *additive* — plugin warnings append to built-in warnings, never replace them.
- **What it lets you tweak:** book-type-specific acceptance rules — LaTeX balance for math books, code-fence audits for programming books (finding F7's audit command could debut as a plugin), bibliography format checks. Findings land in `06-validation.json` / `validation-report.json` with a `source: plugin/<name>` tag.
- **Input:** page level `{page_number, structured, markdown}`; book level `{pages: [{page_number, markdown_path}], assembled_path}` (paths, not content). **Output:** `{warnings: [{code, message, page?, block_id?}]}`.

### S8 — `ingest.pages`: preprocessing

- **Attaches to:** the Phase-3 `book-ocr ingest` command (design doc 01) as an optional pre/post hook.
- **What it lets you tweak:** dewarping, deskewing, contrast normalization, double-page splitting — again Python/CV territory. Lower priority: ingest is already an external-tool wrapper by design.
- **Input:** `{source, out_dir, dpi}`; **Output:** `{pages: [{path, page_number}]}`.

### What deliberately has **no** seam

- **Workflow orchestration** (step graph, retries, leases): the runtime is the product's reliability core; plugins are called *from* executors, never define them.
- **Artifact layout** (`01–07` files, `engine.db`, projections): the audit trail must stay uniform across strategies or the review loop and `structured-rerun-pages` break.
- **The assembler**: pure concatenation ordered by page number (`workflow_executors.go:125-236`); nothing to experiment with that S4 doesn't cover.

### Seam summary

| Seam | Op | Attach point (today) | You can swap… | Effort | Priority |
|---|---|---|---|---|---|
| S1 | `ocr.page` | `StructuredOCRClient` iface, `client.go:34` | entire page-OCR strategy | S (iface exists) | **P1** |
| S2 | `prompt.render` | `prompts.go:10` | prompt text per book type | S | **P1** |
| S5 | `figures.segment` | `figures.go:47-98` | CV segmentation (Python) | M (extract iface) | **P1** |
| S3 | `response.parse` | `structured_ocr.go:81` | parser/repair per model | S | P2 |
| S7 | `validate.*` | `structured_ocr.go:300`, `workflow_executors.go:238` | QA rules per book type | S | P2 |
| S6 | `page.classify` | `workflow_executors.go:20-67` | per-page strategy routing | M (new step) | P2 |
| S4 | `markdown.transform` | after `renderer.go:19` | house-style transforms | S | P3 |
| S8 | `ingest.pages` | Phase-3 ingest cmd | scan preprocessing | S | P3 |

## Host Architecture

### New package: `internal/plugin`

```text
internal/plugin/
  manager.go      Manager: load specs from profile, spawn via devctl runtime.Factory,
                  handshake, capability index, Close (process-group shutdown)
  ops.go          op name constants + typed input/output structs per seam (the
                  book-ocr equivalent of devctl pkg/engine)
  client_ocr.go   PluginStructuredOCRClient  (implements ocrpipeline.StructuredOCRClient)
  segmenter.go    PluginFigureSegmenter      (implements new ocrquality.FigureSegmenter)
  hooks.go        PromptRenderer / ResponseParser / Validator / Transformer adapters
  provenance.go   handshake + declares recorded into run metadata and page artifacts
```

Imports: `github.com/go-go-golems/devctl/pkg/protocol` and `.../pkg/runtime` (Decision D1). No book-ocr package below `internal/plugin` knows about frames or processes — they see ordinary Go interfaces.

### Profile wiring (the experimentation UX)

Plugins are declared in the book profile, because "per book type" is exactly what the profile models:

```yaml
# my-math-book.profile.yaml (excerpt)
plugins:
  - id: layout-ocr
    path: ./plugins/layout_ocr.py        # relative to profile file
    args: ["--model", "gpt-5-mini"]
    env: { OMP_NUM_THREADS: "4" }
    seams: [page.classify, ocr.page]     # which seams this plugin claims
  - id: cv-figures
    path: ./plugins/opencv_segmenter.py
    seams: [figures.segment]
  - id: math-checks
    path: ./plugins/latex_validator.py
    seams: [validate.page]
```

Resolution rule per seam: the first declared plugin that (a) lists the seam in `seams:` and (b) advertises the op in its handshake `capabilities.ops` wins; otherwise built-in. `seams:` must be a subset of handshake ops at startup or the run refuses to start — fail-fast beats silently falling back during a 202-page live run. CLI override for quick experiments: `--plugin ocr.page=./try_this.py` without editing the profile.

### Execution wiring (pseudocode)

```text
structured-run startup (cmd/book-ocr, after profile Compile):
    mgr = plugin.NewManager(profile.Plugins, workDir)      # spawn + handshake all
    defer mgr.Close()                                       # SIGTERM→SIGKILL pgroups
    record run metadata: [{plugin_name, version, declares, ops}, ...]
    client = mgr.Has(OpOCRPage)
             ? plugin.NewOCRClient(mgr, fallback=builtinClient)
             : builtinClient
    segmenter = mgr.Has(OpFiguresSegment) ? plugin.NewSegmenter(mgr) : inkBandV1
    RegisterStructuredWorkflow(rt, client, segmenter, validatorChain(mgr), ...)

PluginStructuredOCRClient.OCRPage(ctx, in, imageBytes):
    if !mgr.Supports("ocr.page"): return fallback.OCRPage(ctx, in, imageBytes)
    out, err = mgr.Call(ctx, "ocr.page", OCRPageInput{
        BookID: in.BookID, PageNumber: in.PageNumber,
        ImagePath: in.ImagePath,                 # path, not bytes
        Profile:   in.PromptSpec,                # compiled policy view
    })                                           # ctx carries the step deadline
    if err (protocol / timeout / E_*):
        return classify(err)                     # retryable flag → workflow.Retryable/Permanent
    return StructuredOCRResult{
        Page: out.Page, RawResponse: out.RawResponse or marshal(out.Page),
        Provenance: out.Engine,                  # → 04-structured.json "engine" + run metadata
    }
    # host continues unchanged: page-number gate, repair, render, artifacts 01–07
```

Error mapping: plugin `error.details.retryable=true` (or `E_TIMEOUT`, transient transport failures) → `workflow.Retryable`, feeding the existing 3-attempt policy (`workflow_package.go:90-92`); everything else → `workflow.Permanent` — the same philosophy as `classifyStructuredPageError` (`workflow_executors.go:368-399`). A plugin *process death* mid-run fails pending calls (runtime router `failAll`) → those page steps retry; the manager respawns the plugin lazily on the next call (one respawn attempt, then permanent failure).

### Concurrency, lifecycle, provenance

- One plugin process per declared plugin per run, shared across page workers. devctl's `runtime.Client` is safe for concurrent `Call`s (mutex-guarded writes, id-routed responses). Plugins that can't handle concurrent requests just process serially — requests queue on their stdin loop; `--max-workers` still bounds host-side parallelism.
- Lifetime: spawned before the run/resume drain, killed after the drain loop exits (the same session model as devctl commands). `structured-rerun-pages` spawns them the same way.
- Provenance: the handshake (`plugin_name`, `declares`) plus per-call `engine` output is written into run metadata and echoed into each page's `04-structured.json`; a reviewer can always answer "which strategy produced this page?" — required for meaningful A/B experiments and for trusting `structured-rerun-pages` diffs.
- `ctx.dry_run` passes through: `--dry-run` runs invoke plugins with `dry_run: true`, letting plugin authors implement their own deterministic fixtures.

## Decision Records

### Decision: D1 — Adopt devctl protocol v2 by importing its packages

- **Context:** book-ocr needs a plugin transport; devctl already ships one with docs, examples, adversarial tests, and operator familiarity.
- **Options considered:** (1) import `devctl/pkg/protocol` + `pkg/runtime`; (2) copy those ~630 LOC into book-ocr and drift freely; (3) design a book-ocr-specific protocol; (4) gRPC / hashicorp go-plugin.
- **Decision:** Option 1. Fall back to option 2 only if the devctl module dependency proves heavy — the packages are cleanly copyable since they depend only on `pkg/errors` and zerolog.
- **Rationale:** The packages are devctl-agnostic by construction (op names are opaque strings; the engine layer is separate — `pkg/runtime` knows nothing about `config.mutate`). Same wire format means one mental model and one authoring guide covers both hosts, and devctl's example/testdata plugins are reusable templates. hashicorp go-plugin is Go-to-Go biased; the whole point here is Python-friendly.
- **Consequences:** book-ocr inherits POSIX-only process-group handling (`Setpgid`) — acceptable, the product already assumes pandoc/xelatex on a Unix box. Protocol evolution is coupled to devctl unless we later copy. One new `require` in go.mod.
- **Status:** proposed

### Decision: D2 — Plugins attach at existing Go interfaces; the host keeps all invariants

- **Context:** A plugin system can either replace pipeline *stages* (strategy) or restructure the *pipeline* (orchestration). The review loop and artifact model depend on the pipeline shape staying fixed.
- **Options considered:** (1) seams = adapter implementations of existing interfaces (`StructuredOCRClient`, new `FigureSegmenter`, hook interfaces), invariants host-side; (2) plugins as workflow executors (a plugin owns a whole step including artifacts); (3) plugins that define new workflow graphs.
- **Decision:** Option 1.
- **Rationale:** The single-image rule, artifact sequence, schema gates, and retry classification are the product's guarantees (design doc 01, Part III); letting plugins own them turns every experiment into a potential audit-trail regression. Adapter seams also mean zero change to the workflow package, and testability via the same fake-client patterns already used throughout the repo.
- **Consequences:** Plugins cannot add step types or graphs (S6 classification is a host-added step *calling* a plugin). If a future experiment genuinely needs a different graph shape, that's a new workflow package, not a plugin.
- **Status:** proposed

### Decision: D3 — One long-lived plugin process per run, concurrent calls, no per-call spawn

- **Context:** A 202-page run makes ~202 `ocr.page` calls from up to `--max-workers` goroutines; spawn cost and Python interpreter startup (100ms+, much more with CV imports or model loading) matter.
- **Options considered:** (1) long-lived process per run, shared, id-correlated concurrent calls; (2) spawn per call (devctl's per-CLI-invocation model taken literally); (3) process pool per seam.
- **Decision:** Option 1, with one lazy respawn on crash. A `--plugin-debug-spawn-per-call` escape hatch can come later for stateless debugging.
- **Rationale:** Per-call spawn costs tens of seconds of pure startup across a book for a CV-heavy Python plugin and defeats warm model state. The imported `runtime.Client` is already concurrency-safe; serial plugins degrade gracefully (queueing), consistent with how `--max-workers` already throttles the vision queue.
- **Consequences:** Plugin authors own their internal concurrency story; docs must state "requests may arrive concurrently; respond per request_id in any order". Long-lived state in plugins (model loaded once) becomes possible — that's the feature.
- **Status:** proposed

### Decision: D4 — Plugins are declared in the book profile, with a CLI override

- **Context:** Where plugin bindings live determines the experimentation UX ("per book type, no recompile").
- **Options considered:** (1) `plugins:` section in the book profile + `--plugin seam=path` CLI override; (2) a global config file (like `.devctl.yaml`); (3) CLI flags only; (4) auto-discovery directory (`plugins/book-ocr-*`).
- **Decision:** Option 1.
- **Rationale:** The profile is already "everything book-specific as data" (design doc 01, Phase 2); an OCR method *is* book-type policy. Profiles travel with the book's workspace, so an experiment is reproducible from the profile file alone. The CLI override covers ad-hoc trials; auto-discovery invites accidental activation during a costly live run.
- **Consequences:** `bookprofile.Profile` gains a `Plugins []PluginSpec` field (absorbed into the Phase-2 schema work). Fail-fast validation at startup: every declared seam must appear in the plugin's handshake capabilities.
- **Status:** proposed

## Implementation Plan (plugin track)

Slots into design doc 01's phases as a parallel track; depends on Phase 1 (buildable/CI) and shares Phase 2's profile schema work.

**P1 — Transport + headline seams (est. 3–5 days).**
1. `internal/plugin/manager.go` around `devctl/pkg/runtime` (`Factory.Start`, `Client.Call`); promote the prototype host patterns (`scripts/02-plugin-protocol-demo/host.go`) into it.
2. `ops.go` with `ocr.page` and `prompt.render` schemas; `client_ocr.go` adapter with fallback; wire into run/resume/rerun setup in `cmd/book-ocr`; provenance into run metadata + `04-structured.json`.
3. Extract a `FigureSegmenter` interface in `ocrquality` (built-in `ink-band-v1` unchanged); `figures.segment` adapter.
4. Profile `plugins:` schema + `--plugin` override + startup capability validation.
5. Reference plugins under `examples/plugins/`: `identity-prompt.py` (golden pass-through), `tesseract-ocr.py` (real alternative strategy), `opencv-segmenter.py` sketch.

**P2 — Quality seams (est. 2–3 days).** `response.parse` with decline-to-builtin fallback; `validate.page`/`validate.book` additive chain with `source:` tagging; `page.classify` step in the discover executor feeding strategy routing.

**P3 — Long tail.** `markdown.transform`; `ingest.pages` hook once Phase-3 ingest exists; `--plugin-debug-spawn-per-call`; a plugin cookbook doc (point authors at devctl's authoring guide for the wire format, plus book-ocr op schemas).

## Testing Strategy

1. **Protocol/manager tests** — port devctl's adversarial-fixture approach (`devctl/testdata/plugins/`: `noisy-handshake`, `streams-only-never-respond`) as book-ocr fixtures: contaminated stdout, missing handshake, never-responds (timeout), death mid-call (respawn-once), concurrent-call interleaving.
2. **Identity-plugin golden test** — an `ocr.page` plugin that replays the dry-run fixture generator must produce byte-identical `05-rendered.md` versus the built-in dry-run path; same for identity `prompt.render` and identity `markdown.transform`. This proves seams are transparent when pass-through.
3. **Contract tests per op** — schema round-trip of every input/output struct; unknown-field and missing-field behavior pinned; op schema version (`ocr.page/v1`) asserted.
4. **The demo as CI smoke** — `scripts/02-plugin-protocol-demo/run-demo.sh` (already green) migrates to a Go test once `internal/plugin` exists.
5. **A/B harness playbook** — run the same page range with built-in vs plugin strategy into two work dirs and diff `04-structured.json`/`05-rendered.md`. This *is* the experiment loop, so it gets documented as a ticket playbook rather than a test.

## Risks and Open Questions

**Risks**

- *Schema drift between host and plugins.* Op inputs embed profile-policy views; renaming a field breaks external scripts silently. Mitigation: version op schemas (`"op_schema": "ocr.page/v1"` in input), contract tests, and a `book-ocr plugins check <profile>` command that handshakes and validates capabilities without running a book.
- *Non-determinism enters the deterministic half.* S4 and S3 weaken the "Go owns rendering deterministically" guarantee if plugins are stochastic. Mitigation: provenance tagging makes plugin involvement visible per page; golden tests cover built-in paths; the docs state the trade explicitly.
- *Zombie/leaked processes.* Inherited mitigations: process groups + SIGTERM→SIGKILL (`factory.go:183-208`); manager `Close` deferred at the cmd layer so it runs on all exits.
- *Latency regression from serial plugins.* A serial Python plugin caps effective parallelism at 1 regardless of `--max-workers`. Mitigation: document it; a per-seam `instances: N` process pool is a cheap later extension of D3.
- *devctl module dependency weight* (D1 fallback): if `go mod graph` shows heavy transitive additions, copy the two packages; they are small and stable.

**Open questions**

1. Should `ocr.page` plugins be able to return conversation turns for the turns DB, or is `03-raw-response.json` provenance enough? (Leaning: optional `raw_response` field; turns DB stays Geppetto-only.)
2. Per-seam plugin *chains* (e.g. two validators) from day one, or first-wins? (Leaning: chains for `validate.*` only — they're additive by nature; first-wins elsewhere.)
3. Should the vlm-separation benchmark learn to drive `ocr.page` plugins so alternative strategies get the same qualification treatment as models? (Attractive; defer to P2.)
4. Windows support: `pkg/runtime` uses POSIX process groups — accept Unix-only plugins, or contribute a Windows path upstream?

## References

**Prototype (this ticket):** `scripts/02-plugin-protocol-demo/{host.go, ocr_plugin.py, run-demo.sh}` — green on 2026-07-03 (`PLUGIN_PROTOCOL_DEMO_OK`: handshake, capability check, `prompt.render`, `ocr.page` on the real 480 KB `page_012.png`, progress event interleaving, `E_UNSUPPORTED` fallback, clean shutdown).

**devctl (read-only, `~/code/wesen/go-go-golems/devctl`):** `pkg/protocol/types.go:5-88` (frames), `pkg/protocol/errors.go:3-12` (codes), `pkg/protocol/validate.go:9-11` (handshake validation); `pkg/runtime/factory.go:51-208` (spawn/handshake/shutdown), `pkg/runtime/client.go:73-304` (Call, read loops, SupportsOp), `pkg/runtime/router.go:26-152` (correlation, fail-all, event buffering); `pkg/engine/pipeline.go` (the op-dispatch pattern to imitate); `pkg/config/config.go:35-42` (PluginSpec shape); `pkg/doc/topics/devctl-plugin-authoring.md:199-301` (canonical frame examples); `examples/plugins/python-minimal/plugin.py` and `testdata/plugins/` (adversarial fixtures).

**book-ocr attach points:** `internal/ocrpipeline/client.go:34-36` (S1), `internal/ocrpipeline/prompts.go:10-77` (S2), `internal/ocrpipeline/structured_ocr.go:81-104,300-321` (S3, S7), `internal/ocrpipeline/renderer.go:19-40,102-114` (S4), `internal/ocrquality/figures.go:47-98,239-335` (S5), `internal/ocrpipeline/workflow_executors.go:20-67,238-284,368-399` (S6, S7, error classification), `internal/bookprofile/profile.go:40-107` (D4 config home), `internal/ocrpipeline/types.go:177-190` (FigureResolver precedent).

**Companion:** design doc `01-book-ocr-productization-analysis-design-and-implementation-guide.md` (findings F1–F9, phases 1–4; this document is the plugin track of that plan).
