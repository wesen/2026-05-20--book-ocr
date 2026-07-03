---
Title: Goja scripting as the in-process extension surface
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
    - Path: ../../../../../../../go-go-golems/geppetto/pkg/js/modules/geppetto/module.go
      Note: The LLM/VLM JS module to wire (Options injection)
    - Path: ../../../../../../../go-go-golems/go-go-goja/pkg/engine/factory.go
      Note: Runtime factory + allowlist posture to copy
    - Path: ../../../../../../../go-go-golems/go-go-goja/pkg/replsession/evaluate.go
      Note: Interrupt-watchdog pattern for script timeouts
    - Path: ../../../../../../../go-go-golems/goja-text/pkg/xgoja/providers/text/text.go
      Note: Text toolkit provider to allowlist
    - Path: ../../../../../../../go-go-golems/pinocchio/cmd/pinocchio/cmds/js.go
      Note: Working wiring recipe incl. turn store
ExternalSources: []
Summary: Analysis and design for a third extension surface — goja JavaScript running in-process with a capability-allowlisted module set — covering the go-go-goja engine/owner model, the existing geppetto JS module (LLM/VLM calls with the user's credentials, already built), the goja-text toolkit, minitrace and pinocchio as reference embedders, script hooks mirroring the plugin seams, interrupt-based limits, and the relationship to the NDJSON plugin surface.
LastUpdated: 2026-07-03T17:16:25.733834857-04:00
WhatFor: Design the script extension surface that lets users and LLM drivers write OCR strategies as sandboxable in-process JavaScript.
WhenToUse: Read before implementing internal/scripting or deciding between the script and NDJSON-plugin surfaces for a given extension.
---


# Goja scripting as the in-process extension surface

## Executive Summary

The NDJSON-stdio plugin surface built earlier in this ticket is language-agnostic and process-isolated, but its isolation is also its cost: every capability crosses a serialization boundary, sandboxing means OS-level machinery (namespaces, bind mounts), and the host cannot hand a plugin a live Go object — only paths and JSON. This document analyzes the alternative the ecosystem already contains: embedding goja (a pure-Go JavaScript interpreter) through go-go-goja, and letting scripts implement the same strategy seams in-process.

The investigation's central finding is that **almost everything required already exists as maintained code in the go-go-golems tree, and all of it is already in book-ocr's dependency graph**:

- **go-go-goja** (v0.8.3, already an indirect dependency via scraper) provides the runtime factory with an explicit module allowlist — a capability sandbox by construction — plus a runtime-owner model that serializes VM access, associates Go contexts with JS execution, and recovers panics. `pkg/xgoja/app/factory.go:123-135` shows the exact hardened posture: implicit and data-only module auto-loading disabled, explicit modules only.
- **geppetto** (v0.11.28, book-ocr's existing pinned version, which already ships the module) provides `require("geppetto")`: a complete JS binding over the same inference machinery the built-in OCR client uses — profile-registry resolution (`inferenceProfiles.load/resolve`), multimodal turn building including `imageFile(path)`, synchronous and Promise-based `run({timeoutMs})`, and turn-store persistence. "Scripts doing LLM/VLM calls with the user's credentials" is therefore not something to build; it is something to *wire and constrain*. pinocchio's `js` command (`cmd/pinocchio/cmds/js.go:305-343`) is the working wiring recipe, including turn-store injection.
- **goja-text** provides four pure-Go, TypeScript-declared modules — `markdown` (goldmark AST parse/walk/build), `extract` (find structured data in noisy text, with positions and confidence), `sanitize` (tree-sitter YAML/JSON repair), `template` — which are precisely the toolkit for script-authored parsing, validation, and transform strategies.
- **go-minitrace** demonstrates every embedding pattern book-ocr needs: a domain module whose loader closure captures host state, fresh runtime per invocation, host-enforced limits builders, generated `.d.ts`, and Glazed help topics for the JS API.

The design: scripts implement the *same seams* as plugins (one mental model — a script exports `ocrPage`, `renderPrompt`, `parseResponse`, `validatePage`, `classifyPage`, `transformMarkdown` functions; exported-symbol presence is the capability handshake), executed by a script host in `internal/scripting` that builds one frozen runtime factory with an allowlisted module set and spins a fresh runtime per page call under the step's context, with an interrupt watchdog. The invariant rule is unchanged: scripts replace strategies, never invariants. For the agent-first product, this surface is stronger than NDJSON plugins on every axis except polyglotism: LLM drivers write JavaScript natively, the generated TypeScript declarations are a machine-readable API contract, and the capability allowlist replaces most of the OS sandbox.

## Problem Statement and Scope

The pilot ticket's agent-first design (BOOK-OCR-WILENSKY-PILOT-001 doc 02) committed to LLM drivers writing their own extension code, executed under a sandbox. NDJSON plugins answer that with process isolation; the open questions this document answers are: (1) what does an in-process goja surface offer that stdio plugins cannot, and vice versa; (2) how do scripts get safe access to model inference under the user's credentials, page context, and text tooling; (3) what are the concrete limits (time, stack, memory) and their honest gaps; (4) what is the implementation shape. Out of scope: implementing the surface (this is the analysis/design), and replacing the plugin surface (they are complementary; see the comparison).

## Part I — The Building Blocks, As They Exist

### The engine and its capability model (go-go-goja)

Runtime construction is a two-phase builder: `engine.NewRuntimeFactoryBuilder(opts...)` accumulates module registrars and options, `Build()` freezes an immutable `RuntimeFactory` (`pkg/engine/factory.go:49-180`), and `factory.NewRuntime(engine.WithStartupContext(ctx), engine.WithLifetimeContext(ctx))` creates a `goja.Runtime` plus event loop plus owner (`factory.go:184-288`). The factory is reusable: build once, create many runtimes.

The capability question — *which* modules a runtime can `require` — has a dangerous default and a clean override. By default the builder exposes everything in the global module registry, which includes `fs` (full host filesystem), `exec` (arbitrary commands), `os`, and `database` (`pkg/engine/options.go:48-53`, blank imports at `pkg/engine/runtime.go:20-29`). The override is exactly what xgoja's own app layer does:

```go
engine.NewRuntimeFactoryBuilder(
    engine.WithImplicitDefaultRegistryModules(false),  // no "everything" fallback
    engine.WithDataOnlyDefaultRegistryModules(false),  // not even the safe auto-set
).WithModules(/* explicit allowlist only */)
```

With this posture (`pkg/xgoja/app/factory.go:123-135`), a script's entire world is the modules the host lists. There is no ambient filesystem, no process spawning, no network — not because a kernel forbids it, but because no code path to it exists inside the VM. This is the structural difference from process sandboxing: capabilities are granted by linking, not restricted by confinement.

Concurrency follows one rule: a `goja.Runtime` is single-threaded, and all access goes through `rt.Owner.Call(ctx, opName, func(ctx, vm) (any, error))`, which marshals the closure onto the runtime's event-loop goroutine, detects reentrancy, honors context cancellation for the *wait*, and optionally recovers panics into errors (`pkg/runtimeowner/runner.go:90-134, 189-212`). The named-context model (startup, lifetime, owner-entry, custom-per-call, cleanup) is documented in depth in the vault article "Runtime Context Ownership in go-go-goja".

### Limits: what exists and what does not

Three bounding mechanisms exist, and one gap:

- **Interrupts.** `vm.Interrupt(cause)` aborts running JS with an `InterruptedError`; goja supports it natively (`goja/runtime.go:1518`). go-go-goja itself never wires context cancellation to interrupts for calls — a CPU-spinning script is *not* stopped by canceling the `Owner.Call` context — but its REPL package shows the canonical watchdog: a goroutine selects on `ctx.Done()`, calls `VM.Interrupt(cause)`, then `ClearInterrupt()` (`pkg/replsession/evaluate.go:566-589`). `Runtime.Close` also force-interrupts stuck JS (`pkg/engine/runtime.go:130-152`). **book-ocr's script host must implement this watchdog per call.**
- **Stack depth.** `vm.SetMaxCallStackSize(n)` bounds recursion; no code in go-go-goja calls it, so the host sets it explicitly.
- **Owner wait bound.** `runtimeowner.Options.MaxWait` caps how long a `Call` waits for its result.
- **The gap: memory.** goja has no heap limit API. A script can allocate unboundedly. In-process, the only mitigations are the interrupt watchdog (bounding wall time bounds most allocation loops), process-level rlimits/cgroups in a hosted deployment, and code review/pinning for the local case. This is the one axis where NDJSON plugins retain a hard advantage.

Error propagation is well-defined in both directions: thrown JS values arrive in Go as `*goja.Exception` (with a parseable stack — `pkg/inspector/runtime/errors.go:27-44` extracts message plus file/line frames, reusable for attributing a page-script failure to a source line), and Go panics inside native modules become `ErrPanicked`-wrapped errors rather than host crashes when `RecoverPanics` is set.

### The LLM/VLM surface already exists: geppetto's JS module

`geppetto/pkg/js/modules/geppetto` (~6,800 lines, present in book-ocr's pinned v0.11.28) exposes the full inference stack to JavaScript behind a deliberately narrowed "hard-cut" API whose public surface is locked by contract tests. The flow a script uses:

```js
const gp = require("geppetto");
const registry = gp.inferenceProfiles.load(sources);   // the user's profile registries
const settings = registry.resolve("gpt-5-mini-low");    // the user's credentials/config
const agent    = gp.agent().name("ocr").inference(settings).build();
const session  = agent.session().id("page-042").build();
const result   = session.next()
    .system("You are a precise structured OCR engine…")
    .user((msg) => msg.text(prompt).imageFile(imagePath))   // multimodal: VLM call
    .run({ timeoutMs: 120000 });
const text = result.text();
```

Three properties make this the right reuse rather than a custom `llm` module:

1. **Credential parity.** `inferenceProfiles.load/resolve` uses the same chained-registry resolution as the CLI's `--profile`/`--profile-registries` flags (`api_inference_profiles.go:22-79`), so a script runs under exactly the credentials and model configuration the user already manages. Better: the host can inject a pre-resolved registry or default settings via `geppetto.Options{EngineProfileRegistry, DefaultInferenceSettings, UseDefaultProfileResolve}` (`module.go:38-59`) so scripts need not name registries at all.
2. **Artifact parity.** `Options.DefaultTurnStore/DefaultPersister` (the pinocchio wiring, `js.go:336-342`) means every JS-initiated inference persists turns into the run's `turns.db` exactly like the built-in client — the audit trail survives the extension surface.
3. **Owner-thread correctness is handled.** Synchronous `run` executes the real `runner.RunInference` on the owner with proper context plumbing; `runAsync` settles a JS Promise back on the owner (`api_session.go:608-780`). The hard concurrency problems are upstream's, already solved and tested.

One honest limitation: there is no host-enforced *call budget* hook in `geppetto.Options` — a script inside its time limit can make several model calls. The interrupt watchdog bounds wall time (and therefore spend, coarsely), and the turn store audits every call after the fact; a first-class per-call-site budget hook is an upstream request, recorded as an open question.

### The text toolkit: goja-text

Four pure-Go modules, each with tests and TypeScript declarations, packaged as an xgoja provider (`pkg/xgoja/providers/text/text.go`):

| Module | What a book-ocr script uses it for |
|---|---|
| `extract` | find fenced/tagged/raw JSON-YAML candidates in noisy model output, with byte/row/column spans and confidence (`pkg/extract/module.go:35-62`) — the substrate for script-authored `parseResponse` strategies |
| `sanitize` | tree-sitter-backed repair of near-valid YAML/JSON (`pkg/sanitize/module.go:45-142`) — a stronger, configurable version of the host's built-in `jsonsanitize` pass |
| `markdown` | goldmark AST: `parse`, `walk` with JS visitors, `textContent`, builders (`pkg/markdown/module.go:192-255`) — validation rules and `transformMarkdown` strategies operate on a real AST instead of regexes |
| `template` | text/html templating with sprig+glazed function sets — report/output shaping |

Notably absent (and fine): no PDF extraction, no fuzzy matching, no tokenization. PDF text-layer access belongs in book-ocr's own host module (below); anything else missing follows the same `modules.NativeModule` + `TypeScriptDeclarer` pattern.

### The reference embedders

**go-minitrace** shows the domain-module pattern: one native module whose `NewLoader(ctx, conn, name, settings)` closure captures host state and exposes builder objects plus a `runtime` info object (`pkg/minitracejs/module.go:22-69`); a fresh runtime per command invocation with both contexts bound and `defer runtime.Close()` (`cmds/query/js_runtime.go:60-83`); host-enforced query limits as first-class builder methods (`MaxRows`, `Timeout`, `RequireOrderBy`); and a `HostServices` interface pulled from `providerapi.ModuleSetupContext.Host` for xgoja binaries (`provider.go:21-25, 62-75`). **pinocchio's `js` command** is the geppetto-specific recipe: factory builder + startup/lifetime contexts + `gp.Register(reg, Options{DefaultTurnStore…})` (`cmds/js.go:305-343`).

## Part II — Design: The book-ocr Script Surface

### Script contract: the seams, as exported functions

A strategy script is one JS file exporting hook functions named after the seams. Export presence is the handshake — the in-process analogue of the plugin protocol's capability list:

```js
// wilensky-strategy.js — implements three seams
const gp = require("geppetto");
const extract = require("extract");

exports.classifyPage = ({ pageNumber, imagePath, textLayer }) => {
  const chars = (textLayer() || "").length;
  return { pageType: chars > 800 ? "body" : "figure",
           strategy: chars > 800 ? "textlayer" : "" };   // route prose to the free path
};

exports.ocrPage = ({ bookId, pageNumber, imagePath, textLayer, llm }) => {
  const result = llm.session(`page-${pageNumber}`).next()
      .system(SYSTEM).user(m => m.text(prompt(textLayer())).imageFile(imagePath))
      .run({ timeoutMs: 120000 });
  return JSON.parse(result.text());     // structured-ocr/v1 page object
};

exports.validatePage = ({ page, markdown }) => {
  const bad = page.blocks.filter(b => b.type === "heading" && !b.text);
  return bad.map(b => ({ code: "empty_heading", message: `block ${b.id}` }));
};
```

The host detects hooks after evaluating the script once per runtime (`Owner.Call` → run program → inspect `exports`), then invokes them per page. Return shapes mirror the plugin op schemas exactly — same host-side validation, same repair pass, same page-number gate, same artifacts. One script may implement any subset; scripts and plugins and built-ins compose per seam through the same adapter slots (`StructuredWorkflowConfig`).

### The module allowlist (the capability grant)

| Module | Source | Grant rationale |
|---|---|---|
| `geppetto` | geppetto pkg/js, host-wired | inference under the user's credentials; turn store = the run's `turns.db`; default settings = the run's `--profile` |
| `extract`, `sanitize`, `markdown`, `template` | goja-text | pure text tooling |
| `bookocr` | new, ~300 lines | the layered context object — `page`, `book`, and (in post stages) `run` — detailed in the next section |
| `yaml`, `crypto`, `path`, `time` | go-go-goja safe set | data plumbing |
| — excluded — | `fs`, `exec`, `os`, `database`, `process` | no ambient filesystem/process/network capability; anything a strategy legitimately needs arrives through `bookocr` scoped accessors |

The `bookocr` module follows the minitrace loader-closure pattern: the executor constructs the loader with the current work item's context captured.

### The context model: page, book, run

A page-only context undersells what the host knows. The system holds context at three scopes, and the workflow DAG — not caution — dictates which scope each hook may see:

- **`page`** — the current work item: `number`, `imagePath`, `readImage()`, `typeHint`, `warn(code, msg)`.
- **`book`** — *source-derived* context, immutable from the moment discovery completes and therefore safe everywhere: the compiled profile view (lexicon, policies), the ingest manifest (source hash, DPI, page count), and crucially `book.textLayer(n)` / `book.imageInfo(n)` for **any** page — the text layer comes from the source PDF, not from the run, so reading page 41's text layer while OCRing page 42 is deterministic regardless of execution order.
- **`run`** — *run-derived* context: other pages' structured JSON, rendered Markdown, quality signals, and warnings. This scope exists only in post-assembly hooks.

The determinism argument is the heart of it. Page steps are parallel siblings in the DAG; during `ocrPage(42)`, page 41 may be running, done, or not started, and a targeted rerun of page 42 months later sees a *different* neighborhood than the first run did. If page-stage hooks could read run-derived neighbor output, results would depend on execution order and rerun history — precisely what the artifact model promises they do not. So the capability grant is staged by seam:

| Hook (seam) | `page` | `book` (source) | `run` (other pages' output) |
|---|---|---|---|
| `classifyPage`, `renderPrompt`, `ocrPage`, `parseResponse` | yes | yes | **no** |
| `validatePage` | yes | yes | no |
| `validateBook`, `transformMarkdown` (assemble stage) | — | yes | **yes, read-only** |
| `postProcessBook` (new seam, below) | — | yes | yes, read-only |

Two consequences are worth naming. First, the May invariant survives intact and gets sharper: the single-image rule concerned neighbor *images*, whose bleed the vlm-separation benchmark measured; neighbor *text layers* are source material, benchmarked separately (the `target-plus-text-context` scenario), and their use in a prompt is now an explicit, profile-visible script decision rather than a hidden host behavior. A script that enriches page 42's prompt with the tail of page 41's text layer is running the continuity experiment the May redesign deferred — deterministically, because text layers precede the run.

Second, the `run` scope motivates a seam the plugin design never had: **`postProcessBook`**, a book-stage hook running after assembly with the full run context. This is the natural home for the second-pass cleanup work that has sat in future-work since the HQ-001 ticket: cross-page hyphenation repair, running-header removal informed by page-frequency statistics (the W7 detector — a heading recurring on many pages is a header), figure-numbering continuity checks, and glossary-consistency sweeps. Every one of those is a cross-page computation over run-derived text, which is exactly what the goja surface handles well (goja-text's markdown AST + the `run` scope) and what the per-page plugin protocol handles awkwardly (it would need the host to serialize the whole book across stdio).

```js
exports.postProcessBook = ({ book, run }) => {
  const headingCounts = {};
  for (const p of run.pages()) {
    for (const h of md.headings(p.markdown())) headingCounts[h] = (headingCounts[h] || 0) + 1;
  }
  const headers = Object.keys(headingCounts).filter(h => headingCounts[h] > run.pageCount() / 3);
  return run.pages().map(p => ({ page: p.number, markdown: stripHeadings(p.markdown(), headers) }));
};
```

### Execution model and limits

```text
internal/scripting host:

startup (once per run):
    factory = NewRuntimeFactoryBuilder(
                  WithImplicitDefaultRegistryModules(false),
                  WithDataOnlyDefaultRegistryModules(false),
              ).WithModules(allowlist…).Build()
    program = goja.Compile(scriptPath, source)        # compile once
    pin source sha256 into run manifest               # same provenance rule as plugins

per page call (inside the page executor, ctx = step context):
    rt = factory.NewRuntime(WithStartupContext(ctx), WithLifetimeContext(ctx))
    defer rt.Close(background)
    rt.VM.SetMaxCallStackSize(4096)
    stop = watchdog(ctx, rt.VM)        # goroutine: <-ctx.Done() → VM.Interrupt(cause)
    defer stop()
    out, err = rt.Owner.Call(ctx, "ocrPage", func(ctx, vm) (any, error) {
        run program; find exports.ocrPage; call with page context object
    })
    classify err:  *goja.Exception   → Permanent (script bug; parse stack for file:line)
                   InterruptedError  → per cause (timeout vs shutdown)
                   geppetto run error → existing provider classification
```

Fresh-runtime-per-page is the default (decision record below): strongest isolation, deterministic globals, and the ~millisecond construction cost is noise against an OCR call. The frozen factory amortizes everything shareable.

### What this surface gives the agent-first product

- LLM drivers emit JavaScript more reliably than any bespoke format; the strategy script *is* the natural output of "write me a better OCR strategy for this book".
- Every allowlisted module carries TypeScript declarations (`modules.TypeScriptDeclarer`; geppetto ships a generated `geppetto.d.ts` with parity tests). Concatenated, they are the machine-readable API contract the agent-first design wanted — `book-ocr scripts dts` replaces hand-written schema docs, and `book-ocr scripts check` (compile + hook detection + fixture round-trip) is the in-process twin of `plugin_check`.
- The W3 failure class shrinks structurally: a script returns a JS object, not prose-embedded JSON — no fence-stripping, no field-name drift from prompt thinness; host validation still applies but has less to repair.

## Part III — Choosing Between the Surfaces

| Axis | Goja scripts | NDJSON plugins |
|---|---|---|
| Language | JavaScript only | any |
| Python CV ecosystem (the S5 case) | no | **yes — keep plugins for figures.segment-class work** |
| Capability control | allowlist by construction; no ambient FS/exec/net | OS sandbox required (namespaces, bind mounts) |
| Host objects (live engine, turn store) | direct injection | impossible; paths+JSON only |
| Per-call overhead | ~µs (owner dispatch) | ~ms (frame round-trip), amortized process |
| Runaway CPU | interrupt watchdog (host-built) | SIGKILL process group |
| Runaway memory | **no cap** (goja limitation) | rlimits/cgroups |
| Crash isolation | in-process (panic recovery, but shared fate for native bugs) | full |
| Typed API for agents | generated .d.ts | JSON Schemas (to be written) |
| Turn/artifact parity | native (shared turn store) | plugin must echo raw_response |

The honest division: **scripts become the default surface for logic, prompting, parsing, validation, classification, and LLM/VLM composition; plugins remain the surface for polyglot and OS-heavy strategies (Python CV, external OCR engines) and for hosted execution of fully untrusted code where memory isolation is non-negotiable.** Both implement the same seams through the same config slots, so a book profile can mix them.

## Decision Records

### Decision: Reuse geppetto's JS module rather than building a book-ocr `llm` module

- **Context:** Scripts need model calls under the user's credentials; a thin custom wrapper (`llm.complete(...)`) was the alternative.
- **Options considered:** (1) wire `geppetto/pkg/js/modules/geppetto` with host-injected options; (2) custom minimal module wrapping `RunInference`; (3) both (custom facade over the geppetto module).
- **Decision:** Option 1, possibly adding a convenience `llm` handle on the `bookocr` module that pre-binds settings and session naming (a facade in JS, not a second Go binding).
- **Rationale:** The module already solves profile resolution, multimodal turns, owner-thread execution, Promise settlement, persistence, and locks its surface with parity tests; v0.11.28 — book-ocr's current pin — already contains it. A custom module would re-implement the hard parts and drift.
- **Consequences:** book-ocr inherits the hard-cut API's shape (agent/session, no bare `runInference`) — fine, and its stability tests are upstream's. API drift between v0.11.28 and geppetto HEAD must be validated when bumping. The missing per-run call-budget hook becomes an upstream request.
- **Status:** proposed

### Decision: Explicit-allowlist runtimes; never the implicit module registry

- **Context:** The builder's default exposes `fs`/`exec`/`os`/`database` to any script.
- **Options considered:** (1) explicit `WithModules` with both implicit flags disabled (the xgoja app posture); (2) `MiddlewareExclude("fs","exec",…)` denylist; (3) default modules plus guarded host-provider variants.
- **Decision:** Option 1.
- **Rationale:** Denylists fail open when upstream adds a module; the allowlist fails closed. The xgoja app layer itself models this posture (`pkg/xgoja/app/factory.go:123-135`).
- **Consequences:** every capability a strategy needs must be deliberately granted via `bookocr` or an allowlisted module; that friction is the feature.
- **Status:** proposed

### Decision: Fresh runtime per page call, from one frozen factory

- **Context:** Page steps run in parallel (`--max-workers`); a goja runtime is single-threaded.
- **Options considered:** (1) runtime per page call; (2) runtime pool, one per worker; (3) one runtime, serialized calls.
- **Decision:** Option 1; revisit toward (2) only if profiling shows construction cost matters.
- **Rationale:** Fresh globals per page eliminate cross-page state leakage (a correctness property for targeted reruns, which must reproduce a page independent of visit order); construction is milliseconds against multi-second OCR calls; minitrace validates the per-invocation pattern.
- **Consequences:** scripts cannot cache across pages (models/registries re-resolve per page — mitigated by host-side caching inside the `bookocr`/geppetto options, which are host objects and *can* be shared safely).
- **Status:** proposed

### Decision: Script hooks mirror the plugin seams one-to-one

- **Context:** Two extension surfaces risk two mental models and two documentation sets.
- **Options considered:** (1) same seam names, same input/output shapes, same adapter slots; (2) a script-native API designed independently.
- **Decision:** Option 1.
- **Rationale:** Everything already written — seam docs, host-side invariants, validation, the A/B comparison loop, profile bindings — applies unchanged; a profile can swap a plugin for a script per seam without retraining anyone (or any agent).
- **Consequences:** Some JS-side ergonomics are sacrificed to shape parity (e.g. `ocrPage` returns the wire-shaped page object rather than a builder). The `bookocr` module can layer conveniences without changing the contract.
- **Status:** proposed

## Implementation Plan

1. **G1 — Script host** (`internal/scripting`, est. 3–4 days): frozen factory with the allowlist; program compile + source pinning; hook detection; per-call runtime + interrupt watchdog + `SetMaxCallStackSize`; `*goja.Exception` → file:line classification; adapters implementing the existing seam interfaces (`StructuredOCRClient`, `PromptRenderer`, `ResponseParser`, validators, classifier) so `StructuredWorkflowConfig` needs no changes; `--script seam=path` and profile `scripts:` bindings beside `plugins:`.
2. **G2 — Modules** (est. 2–3 days): wire goja-text's four modules; register geppetto's module with `Options{EngineProfileRegistrySpec: run registries, DefaultInferenceSettings: run profile, DefaultTurnStore: run turns.db, UseDefaultProfileResolve: true}` (the pinocchio recipe); implement `bookocr` (page context, `textLayer()` via the ingest manifest, `warn`).
3. **G3 — Agent affordances** (est. 2 days): `book-ocr scripts dts` (concatenate TypeScript declarations), `book-ocr scripts check` (compile, hook inventory, fixture round-trip), Wilensky-textlayer strategy re-implemented as the example script — the direct A/B against the pilot's plugin.
4. **G4 — Hardening**: rlimit/cgroup guidance for hosted mode (the memory-gap mitigation), per-run LLM call-budget upstream request, `.d.ts`-driven driver-guide section.

## Risks and Open Questions

- **No memory cap** is the structural weakness; scripts are bounded in time but not heap. Local single-user mode accepts it (same trust as running the CLI); hosted agent-first mode either wraps the whole worker in cgroup limits or routes fully-untrusted strategies to the plugin surface. Named, not solved.
- **Same-process fate sharing:** panic recovery covers Go panics, but a bug in a native module can still corrupt the host in ways a subprocess cannot. The allowlist keeps the native surface small.
- **Version drift:** the geppetto JS module exists at the pinned v0.11.28, but the hard-cut API stabilized across later versions; bumping geppetto needs the module's parity tests re-run against book-ocr's usage. Validation step in G2.
- **Budget enforcement** for LLM calls is time-bounded, not count-bounded, until an upstream hook exists.
- Open: `postProcessBook` output semantics — replace pages' rendered Markdown (making assembly two-phase) or emit a separate cleaned artifact beside `assembled.md`? (Leaning: separate artifact first — non-destructive, diffable against the per-page renders — with in-place replacement as a profile opt-in later.)
- Open: expose `runAsync`/Promises to strategy scripts, or keep hooks synchronous for v1? (Leaning: synchronous; the host parallelizes across pages already, and sync hooks keep the interrupt story simple.)

## References

**go-go-goja** (`~/code/wesen/go-go-golems/go-go-goja`): `pkg/engine/factory.go:49-288` (builder/factory/NewRuntime), `pkg/engine/options.go:48-90` (implicit/data-only flags), `pkg/engine/module_specs.go:52-239` (registrars, data-only set, process opt-in), `pkg/engine/module_middleware.go:23-66` (Safe/Only/Exclude), `pkg/runtimeowner/runner.go:90-212` (Call/Post/reentrancy/panic recovery), `pkg/engine/runtime.go:87-152` (Close + force-interrupt), `pkg/replsession/evaluate.go:566-589` (the interrupt-watchdog pattern to copy), `pkg/inspector/runtime/errors.go:27-44` (exception→file:line), `pkg/xgoja/app/factory.go:123-179` (hardened posture + host-service wiring), `pkg/xgoja/providerapi/{module.go:15-51, capabilities.go:74-115}` (host-service injection), `pkg/xgoja/providers/http/http.go:22-27,197-214` (host-service retrieval template), `modules/{fs,exec,os,database}` (the excluded set).

**geppetto** (`~/code/wesen/go-go-golems/geppetto`): `pkg/js/modules/geppetto/module.go:28-192` (module, Options, exports), `api_inference_profiles.go:22-129`, `api_agent.go:68-295`, `api_session.go:86-780` (session/turn builder, `imageFile`, run/runAsync on the owner), `pkg/js/runtime/runtime.go:17-97` (bootstrap incl. `WithDataOnlyDefaultRegistryModules(IncludeDefaultModules)`), `pkg/doc/types/geppetto.d.ts`, `module_hardcut_test.go` (surface lock).

**goja-text** (`~/code/wesen/go-go-golems/goja-text`): `pkg/{markdown,extract,sanitize,template}/module.go`, `pkg/xgoja/providers/text/text.go:17-63`.

**Reference embedders:** go-minitrace `pkg/minitracejs/module.go:13-69`, `provider.go:21-77`, `cmds/query/js_runtime.go:23-88`, `pkg/minitracejs/typescript.go`; pinocchio `cmd/pinocchio/cmds/js.go:197-343`.

**Companions:** this ticket's design doc 02 (plugin seams — the shapes scripts mirror), BOOK-OCR-WILENSKY-PILOT-001 doc 02 (agent-first design this surface serves), vault article "Runtime Context Ownership in go-go-goja" (the context/ownership model).
