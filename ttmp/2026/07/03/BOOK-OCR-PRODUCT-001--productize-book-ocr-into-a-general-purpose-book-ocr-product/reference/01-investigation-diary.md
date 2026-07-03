---
Title: Investigation diary
Ticket: BOOK-OCR-PRODUCT-001
Status: active
Topics:
    - book-ocr
    - productization
    - workflow
    - ocr
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ttmp/2026/07/03/BOOK-OCR-PRODUCT-001--productize-book-ocr-into-a-general-purpose-book-ocr-product/scripts/01-dry-run-structured-pipeline.sh
      Note: E2E health-check experiment run in Steps 2–3
    - Path: ttmp/2026/07/03/BOOK-OCR-PRODUCT-001--productize-book-ocr-into-a-general-purpose-book-ocr-product/scripts/book-ocr.go.work
      Note: Build workaround for missing ../scraper replace target
ExternalSources: []
Summary: Chronological diary of the BOOK-OCR-PRODUCT-001 investigation, including the standalone-build failure discovery, dry-run pipeline verification, and report delivery.
LastUpdated: 2026-07-03T10:38:58.275251545-07:00
WhatFor: Understand how the productization assessment was performed and reproduce its experiments.
WhenToUse: Read before continuing productization work or re-running the scripts in scripts/.
---


# Diary

## Goal

Capture the investigation that assesses the current status of the book-ocr repository and produces a productization analysis, design, and implementation guide for a new intern, delivered to reMarkable.

## Step 1: Ticket workspace and repository survey

The task was to assess the state of the book-ocr repo and design a path from "worked once for MIT Report 794" to a proper product. I created the BOOK-OCR-PRODUCT-001 ticket workspace, then surveyed the repository top-down: README, AGENT.md, git history, and package layout, before fanning out deeper reads of every internal package, the scraper workflow runtime dependency, and the historical ttmp tickets.

### Prompt Context

**User prompt (verbatim):** "What's the status of this OCR repo, and how could we turn it into a proper product? Create a new docmgr ticket to do that, and then Create a detailed analysis / design / implementation guide that is for a new intern, explaining all the parts of the system needed to understand what it is, with prose paragraphs and bullet point sand pseudocode and diagrams and api references and file references. It should be very clear and technical. Store in the ticket and the nupload to remarkable. Feel free to run experiments in the ticket's scripts/ directory. Code for scraper and go-go-goja and all is in ~/code/wesen/go-go-golems (do not modify)"

**Assistant interpretation:** Produce a status assessment plus a long-form, intern-oriented analysis/design/implementation guide for productizing book-ocr; keep it in a new docmgr ticket; validate claims with experiments under the ticket's scripts/ directory; upload the deliverable to reMarkable.

**Inferred user intent:** Get a reliable, evidence-based picture of where the project stands after the May 2026 sprint, and a concrete technical roadmap somebody new could execute to turn the one-book pipeline into a reusable product.

### What I did
- `docmgr ticket create-ticket --ticket BOOK-OCR-PRODUCT-001 --title "Productize book-ocr into a general-purpose book OCR product" --topics book-ocr,productization,workflow,ocr`
- Added the design doc and this diary via `docmgr doc add`.
- Read `README.md` (479 lines, current and high quality), `AGENT.md`, git log (last activity 2026-05-27).
- Launched four parallel deep-read investigations: (1) `cmd/book-ocr` + `internal/ocrpipeline`, (2) `internal/ocrquality`/`ocrvalidation`/`bookprofile`/`ocrmvp`/`vlmseparation`, (3) `scraper/pkg/workflow` runtime + geppetto/go-go-goja context, (4) all historical ttmp tickets.

### Why
- The report must be evidence-based and line-anchored; parallel reads cover the ~10k lines of Go quickly without losing depth.

### What worked
- Ticket scaffolding created cleanly under `ttmp/2026/07/03/BOOK-OCR-PRODUCT-001--productize-book-ocr-into-a-general-purpose-book-ocr-product/`.
- README turned out to be an accurate, current map of the system (verified against code reads later).

### What didn't work
- N/A for this step.

### What I learned
- The repo has 9 prior tickets and a disciplined diary culture; most historical context is recoverable from `ttmp/`.
- The production path is the structured OCR workflow; the freeform `ocrmvp` path is legacy.

### What was tricky to build
- Nothing yet; pure discovery.

### What warrants a second pair of eyes
- N/A.

### What should be done in the future
- N/A.

### Code review instructions
- Ticket root: `ttmp/2026/07/03/BOOK-OCR-PRODUCT-001--productize-book-ocr-into-a-general-purpose-book-ocr-product/`.

### Technical details
- `docmgr status --summary-only`: 9 tickets, 26 docs, all stale >30 days before this ticket.

## Step 2: Discovered the repo does not build standalone (missing ../scraper replace target)

The first health-check experiment failed immediately: `go build ./...` errors because `go.mod:125` declares `replace github.com/go-go-golems/scraper => ../scraper`, and no scraper checkout exists at `/home/manuel/code/wesen/scraper` (the real one lives at `/home/manuel/code/wesen/go-go-golems/scraper`). This is finding F1 of the productization report: the repo has an unpublished, path-coupled dependency on the workflow runtime and does not build from a clean clone.

### Prompt Context

**User prompt (verbatim):** (see Step 1)

**Assistant interpretation:** Running experiments is part of the assessment; a failing build is a first-class status finding.

**Inferred user intent:** Know what actually works today, not what the docs claim.

### What I did
- Ran `go build ./... && go test ./... -count=1`. Three packages failed setup: `ocrmvp`, `ocrpipeline`, `ocrquality` — all with `github.com/go-go-golems/scraper@v0.0.0: replacement directory ../scraper does not exist`. Packages without the scraper dependency passed: `bookprofile`, `ocrvalidation`, `vlmseparation`.
- Tried `ln -s ~/code/wesen/go-go-golems/scraper ~/code/wesen/scraper` — failed: `Read-only file system` (sandbox restricts writes outside the workspace).
- Wrote `scripts/book-ocr.go.work` (a Go workspace file with `use` + absolute-path `replace`) and exported `GOWORK` pointing at it. Two version bumps were needed: go.work `go 1.24.3` → rejected because book-ocr's go.mod requires 1.26.3; then `go 1.26.3` → rejected because scraper's go.mod requires 1.26.4.

### Why
- A `go.work` override fixes module resolution without modifying either repository (the user forbade changes to go-go-golems, and a stray symlink is untracked state).

### What worked
- `GOWORK=<ticket>/scripts/book-ocr.go.work` with `go 1.26.4` resolves the scraper module from its real location.

### What didn't work
- Symlink approach: `ln: failed to create symbolic link '/home/manuel/code/wesen/scraper': Read-only file system`.
- go.work with `go 1.24.3`: `go: module . listed in go.work file requires go >= 1.26.3, but go.work lists go 1.24.3`.
- go.work with `go 1.26.3`: `go: module /home/manuel/code/wesen/go-go-golems/scraper requires go >= 1.26.4 (running go 1.26.3)`.

### What I learned
- The scraper dependency is pinned as `v0.0.0` with a relative-path replace, i.e. the workflow runtime has never been versioned or published. Dependency hygiene is step zero of any productization plan.
- The go.work `go` directive must be >= the highest `go` directive among all used/replaced modules.

### What was tricky to build
- The failure mode is invisible in the repo's own docs: the README build instructions assume a sibling `../scraper` checkout that only exists on the original dev machine layout.

### What warrants a second pair of eyes
- Whether to publish, vendor, or extract `scraper/pkg/workflow` — covered by a decision record in the design doc.

### What should be done in the future
- Publish or extract the workflow runtime so `go build` works from a clean clone; add CI that builds from a bare checkout.

### Code review instructions
- `scripts/book-ocr.go.work` and `scripts/01-dry-run-structured-pipeline.sh` reproduce the failure and the workaround.

### Technical details
```text
internal/ocrmvp/discover.go:11:2: github.com/go-go-golems/scraper@v0.0.0: replacement directory ../scraper does not exist
FAIL github.com/go-go-golems/book-ocr/internal/ocrmvp [setup failed]
FAIL github.com/go-go-golems/book-ocr/internal/ocrpipeline [setup failed]
FAIL github.com/go-go-golems/book-ocr/internal/ocrquality [setup failed]
ok   github.com/go-go-golems/book-ocr/internal/bookprofile
ok   github.com/go-go-golems/book-ocr/internal/ocrvalidation
ok   github.com/go-go-golems/book-ocr/internal/vlmseparation
```

## Step 3: Verified the full pipeline works end-to-end (dry-run experiment)

With module resolution fixed, the whole suite passes and a full three-page dry-run of the structured workflow — including figure embedding, PDF rendering, and validation — completes cleanly. The repo's engine room is healthy; the productization gaps are packaging and configuration, not correctness.

### Prompt Context

**User prompt (verbatim):** (see Step 1)

**Assistant interpretation:** Establish exactly what works today before writing recommendations.

**Inferred user intent:** Ground the productization plan in verified behavior.

### What I did
- `go test ./... -count=1` with the go.work override: all six test packages `ok` (`bookprofile`, `ocrmvp`, `ocrpipeline`, `ocrquality`, `ocrvalidation`, `vlmseparation`).
- Ran `scripts/01-dry-run-structured-pipeline.sh`: 3-page structured run with `--embed-figures --render-pdf --expected-pages 3` against the Report-794 page images.

### Why
- The dry-run exercises the workflow engine, page executors, assembly, figure embedding, Pandoc/XeLaTeX rendering, and validation without model spend.

### What worked
- Workflow `book-ocr/structured-6fdeff74-…` succeeded; all nine checked artifacts present (`assembled.md` 206 B, `book.pdf` 10.9 KB, `engine.db`, `turns.db` 3.0 MB, full `pages/page_001/01–06` sequence); `validation-report.json`: `page_count 3, expected 3, warnings 0`.

### What didn't work
- First execution failed on the missing `../scraper` (Step 2); nothing else.

### What I learned
- pandoc 3.1.3 + XeTeX (TeX Live 2023) are present on this machine, so the PDF path is fully exercisable locally.
- Dry-run output is Report-794-specific for pages 12/13/32 (hardcoded fixtures in `structured_ocr.go:59-79`) — recorded as part of finding F3.

### What was tricky to build
- Nothing; the script is a thin harness.

### What warrants a second pair of eyes
- N/A.

### What should be done in the future
- CI should run this same script (minus or with pandoc) on every PR — Phase 1 item 2 in the design doc.

### Code review instructions
- `scripts/01-dry-run-structured-pipeline.sh`; run it directly, exit 0 = healthy.

### Technical details
- Work dir pattern `/tmp/book-ocr-product-001-dryrun-XXXX`; see Appendix A of the design doc for the full artifact listing.

## Step 4: Wrote the productization analysis / design / implementation guide

Synthesized four parallel deep-read investigations (structured pipeline + CLI; quality/validation/profile/legacy/benchmark packages; scraper workflow runtime; nine-ticket history) into the primary design doc. The document is intern-oriented: system overview, evidence-based status, architecture deep dive with file:line anchors, findings F1–F9, six decision records, target architecture with API sketches and pseudocode, a four-phase implementation plan, and a testing strategy.

### Prompt Context

**User prompt (verbatim):** (see Step 1)

**Assistant interpretation:** The main deliverable — long-form, clear, technical, with prose, bullets, pseudocode, diagrams, API and file references.

**Inferred user intent:** A document a new intern can execute from.

### What I did
- Wrote `design-doc/01-book-ocr-productization-analysis-design-and-implementation-guide.md` (~500 lines).
- Key structural choices: findings numbered F1–F9 so phases can reference them; decision records for product shape, dependency strategy, profile-driven policy, structured-only OCR, ingestion, and rerun API; ASCII diagrams (PDF-safe through pandoc, unlike mermaid).

### Why
- The skill's writing-style reference mandates evidence-anchored claims and compact decision records; findings-to-phases traceability keeps the plan executable.

### What worked
- The four investigations converged: every "limitation" in the README maps to a concrete finding with file:line evidence, and the ticket history independently confirms the open items.

### What didn't work
- N/A.

### What I learned
- The central assessment: nothing wrong is architectural. The layering (runtime below, policy above) and the two invariants (single target image; JSON-not-Markdown) are right and empirically justified; the product work is packaging, a configuration seam (bookprofile), ingestion, and operational polish.

### What was tricky to build
- Keeping the doc navigable at this length: Part I–IX structure with the gap analysis as the hinge between "what is" and "what to build".

### What warrants a second pair of eyes
- The decision to stay CLI-first (no hosted service) and the estimate ranges on Phases 2–4.

### What should be done in the future
- Convert Phase 1 into concrete tasks once the go-go-golems owner decides tag-vs-extract for the workflow runtime (open question 1).

### Code review instructions
- Read the design doc top to bottom; verify F1–F9 anchors by opening the cited file:line locations; run `scripts/01-dry-run-structured-pipeline.sh` to reproduce Appendix A.

### Technical details
- Sources: four Explore-agent reports (structured pipeline/CLI; support packages; workflow runtime; ticket history), plus direct reads of README/go.mod and the experiments in Steps 2–3.

## Step 5: Prototyped the NDJSON-stdio plugin protocol round-trip

The user extended the ticket: add devctl-style JSONL-stdio plugin seams so OCR strategies can be swapped per book type without recompiling. Before designing the seams, I proved the transport: a stdlib-only Go host (`scripts/02-plugin-protocol-demo/host.go`) spawning a Python plugin (`ocr_plugin.py`) that speaks the devctl v2 frame shapes — handshake with capability list, request/response correlated by `request_id`, interleaved `event` progress frames, and `E_UNSUPPORTED` for unknown ops.

### Prompt Context

**User prompt (verbatim):** "Ok, wealso want to add plugin seams where we can call jsonl stdio plugins, a bit like in ~/code/wesen/go-go-golems/devctl for example. Examine where these seams would make sense and what they would allow us to tweak, because we could then experiment with different methods of ocr-ing certain types of books without having to recompiile the project all the time."

**Assistant interpretation:** Analyze book-ocr for the right plugin seam locations, specify what each seam lets an experimenter change, and design the plugin protocol/host after devctl's NDJSON-stdio model; document it in the ticket.

**Inferred user intent:** Fast, recompile-free experimentation with OCR methods (prompts, parsers, segmentation, whole OCR strategies) per book type, in any language (especially Python for CV/ML).

### What I did
- Loaded the devctl-plugin-authoring skill and its protocol quick reference (handshake / request / response frames, `E_UNSUPPORTED`, ctx with `repo_root`/`dry_run`/`deadline_ms`).
- Launched a deep-read of `~/code/wesen/go-go-golems/devctl` for the host-side runner, framing, and capability model (file:line anchors for the design doc).
- Wrote and ran the demo: `run-demo.sh` → handshake ok, `prompt.render` ok, `ocr.page` returned StructuredPageOCR-shaped JSON for the real 480 KB page_012.png with a progress event, unsupported op answered `E_UNSUPPORTED`, clean exit (`PLUGIN_PROTOCOL_DEMO_OK`).
- Confirmed the existing Go interfaces the seams should attach to: `ocrpipeline.StructuredOCRClient` (`client.go:34`), `ocrpipeline.FigureResolver` (`types.go:177`), `ocrmvp.OCRClient` (`types.go:94`); `ocrquality.EmbedExtractedFigures` (`figures.go:47`) is a bare function that needs an interface introduced.

### Why
- A 150-line host + 100-line Python plugin retires the main protocol risks (stdout contamination, correlation, event interleaving, large-frame scanning) before any design commitment.

### What worked
- Full round-trip on first run after wiring; `bufio.Scanner` with a 16 MB buffer handles large structured pages.

### What didn't work
- N/A (earlier attempt at a watch-loop for the background agent was pointless — notifications arrive automatically).

### What I learned
- Passing `image_path` instead of base64 bytes keeps frames small and lets Python plugins mmap/open images directly — right default for a local-first product.
- Event frames interleaving with responses means the host needs a frame loop per request, not a single read — reflected in `awaitTerminal`.

### What was tricky to build
- Nothing protocol-wise; the devctl shapes are simple by design.

### What warrants a second pair of eyes
- Concurrency model choice (one plugin process shared across page workers vs per-worker) — covered by a decision record in design doc 02.

### What should be done in the future
- Fold the host prototype into `internal/plugin` when Phase P1 (plugin seams) is implemented.

### Code review instructions
- Run `scripts/02-plugin-protocol-demo/run-demo.sh`; exit 0 and `PLUGIN_PROTOCOL_DEMO_OK` = protocol viable.

### Technical details
```text
handshake ok: plugin=demo-ocr-strategy ops=[prompt.render ocr.page]
prompt.render ok: {"system": "You are a precise structured OCR engine. JSON only.", ...}
  event(r-002): progress {"stage": "segmenting", "page": 12}
ocr.page ok: page=12 blocks=1 text="Demo plugin transcription of page 12 (source image: 480576 bytes)."
unsupported-op handling ok: E_UNSUPPORTED
PLUGIN_PROTOCOL_DEMO_OK
```

## Step 6: Wrote the plugin-seams design doc

Synthesized the devctl protocol deep-read and the seam analysis into design doc 02. The pivotal discovery: devctl's `pkg/protocol` and `pkg/runtime` are public, devctl-agnostic Go packages (~630 LOC combined, deps only pkg/errors + zerolog), so book-ocr can import the transport wholesale and only define its own ops — the analogue of devctl's `pkg/engine` layer.

### Prompt Context

**User prompt (verbatim):** (see Step 5)

**Assistant interpretation:** Deliver the seam analysis as a companion design doc to the productization guide.

**Inferred user intent:** A concrete, implementable plugin architecture, not just a list of ideas.

### What I did
- Wrote `design-doc/02-plugin-seams-ndjson-stdio-plugins-for-recompile-free-ocr-experimentation.md`: eight seams (S1–S8) ranked with attach points, op schemas, and what each lets you tweak; host architecture (`internal/plugin` importing devctl runtime); profile-based plugin declaration; four decision records (D1–D4); P1–P3 implementation track; testing strategy including adversarial plugin fixtures and identity-plugin golden tests.

### Why
- The seams map 1:1 onto existing Go interfaces/functions, so anchoring each op to a file:line makes the design directly implementable.

### What worked
- The devctl layering (generic protocol/runtime vs devctl-specific engine) is exactly the reuse boundary we need; no protocol design work required.

### What didn't work
- N/A.

### What I learned
- devctl has no plugin auto-restart and is fine — mapping "plugin lifetime = one workflow-run drain" onto book-ocr keeps the model simple; crash recovery composes with the existing page-step retry policy instead of needing supervisor logic.
- The strongest argument for stdio plugins over plain Go interfaces is S5 (figure segmentation): the CV iteration loop belongs in Python.

### What was tricky to build
- Deciding what gets NO seam (workflow graph, artifact layout, assembler) mattered as much as the seams themselves — "plugins replace strategies, never invariants" is the rule that keeps the audit trail trustworthy.

### What warrants a second pair of eyes
- D1 (import devctl vs copy the two packages) — depends on tolerance for the transitive dependency; and D3's concurrency contract for plugin authors.

### What should be done in the future
- Implement track P1 (manager + ocr.page + prompt.render + figures.segment) after productization Phase 1 lands.

### Code review instructions
- Read design doc 02; verify seam anchors (S1–S8) against the cited book-ocr lines and the devctl references; run `scripts/02-plugin-protocol-demo/run-demo.sh` for the transport proof.

### Technical details
- Op set: ocr.page, prompt.render, response.parse, markdown.transform, figures.segment, page.classify, validate.page, validate.book, ingest.pages; profile `plugins:` section binds plugin → seams with fail-fast capability validation.

## Step 7: Fixed F1 — dropped the ../scraper replace, pinned published scraper v0.0.4

The user green-lit step 1 of the plan. It turned out better than expected: scraper is already published on GitHub with tags v0.0.1–v0.0.4, and the local checkout at `~/code/wesen/go-go-golems/scraper` sits exactly at v0.0.4 (`git describe --tags` → `v0.0.4`, clean). So "drop the replace, require @latest" was API-neutral: the repo now builds standalone from the module proxy, no workspace or sibling checkout required.

### Prompt Context

**User prompt (verbatim):** "1. drop ../scraper, get the @latest version . explain the cross-repo things?"

**Assistant interpretation:** Execute the F1 fix now (remove the replace directive, depend on the latest published scraper), and explain what the remaining cross-repo asks are.

**Inferred user intent:** Start executing the roadmap, cheapest unblocker first.

### What I did
- `go list -m -versions github.com/go-go-golems/scraper` → v0.0.1–v0.0.4; confirmed local checkout == v0.0.4 tag.
- `go mod edit -dropreplace github.com/go-go-golems/scraper`.
- `go get @latest` failed first (`unknown revision v0.0.0` — the stale v0.0.0 require blocks graph resolution), so: `go mod edit -require=github.com/go-go-golems/scraper@v0.0.4` then `go mod tidy` (which bumped the `go` directive 1.26.3 → 1.26.4 to satisfy scraper).
- `GOWORK=off go build ./... && go test ./... -count=1` → BUILD_OK, all packages `ok`.
- Reran `scripts/01-dry-run-structured-pipeline.sh` with `GOWORK=off` to confirm the E2E path against the published module; updated the script (go.work now optional, kept as a template for local-scraper development).

### Why
- F1 was the single blocker for clean-clone builds and CI; the published tag made it a five-minute fix instead of a cross-repo negotiation.

### What worked
- v0.0.4 is API-identical to the local tree, so no code changes were needed.

### What didn't work
- `go get github.com/go-go-golems/scraper@latest` with `v0.0.0` still in require: `reading github.com/go-go-golems/scraper/go.mod at revision v0.0.0: unknown revision v0.0.0`. Setting the require version explicitly first is the workaround.

### What I learned
- The replace directive was hiding an already-solved problem — nobody had checked whether scraper was published since the externalization sprint.
- `go build ./...` now also compiles the demo host under ttmp scripts (it is inside the module); harmless, but ttmp scripts with `package main` become part of the build surface.

### What was tricky to build
- Only the v0.0.0 resolution ordering quirk above.

### What warrants a second pair of eyes
- go.mod/go.sum diff (11 lines / 12 lines) — verify no unintended dependency bumps beyond the scraper pin and go 1.26.4.

### What should be done in the future
- Future workflow-runtime changes must go through scraper releases (or the pkg/workflow extraction) now that book-ocr consumes published versions.
- Consider excluding ttmp from the module (or build-tagging demo mains) if `go build ./...` surface matters for CI.

### Code review instructions
- `git diff go.mod go.sum`; then `GOWORK=off go build ./... && go test ./... -count=1`; then `scripts/01-dry-run-structured-pipeline.sh`.

### Technical details
```text
go.mod before: go 1.26.3 / require scraper v0.0.0 / replace scraper => ../scraper
go.mod after:  go 1.26.4 / require scraper v0.0.4   (replace removed)
```
