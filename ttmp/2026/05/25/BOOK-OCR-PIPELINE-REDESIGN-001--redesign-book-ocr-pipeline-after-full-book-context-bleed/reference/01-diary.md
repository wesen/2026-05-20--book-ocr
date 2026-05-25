---
Title: Diary
Ticket: BOOK-OCR-PIPELINE-REDESIGN-001
Status: active
Topics:
    - ocr
    - book-processing
    - workflow
    - geppetto
    - pinocchio
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/ocrmvp/geppetto_ocr.go
      Note: Context-image root cause reference
    - Path: internal/ocrpipeline/renderer.go
      Note: Deterministic Markdown renderer
    - Path: internal/ocrpipeline/session.go
      Note: Pinocchio turn store wrapper
    - Path: internal/ocrpipeline/types.go
      Note: Structured OCR data contracts
    - Path: internal/ocrvalidation/adjacent.go
      Note: Adjacent caption validation
    - Path: internal/ocrvalidation/anchors.go
      Note: Expected/forbidden anchor validation
    - Path: internal/vlmseparation/oracle.go
      Note: Risky-page oracle evidence feeding validation design
    - Path: internal/vlmseparation/report.go
      Note: Benchmark evidence source for report workflow
    - Path: ttmp/2026/05/25/BOOK-OCR-PIPELINE-REDESIGN-001--redesign-book-ocr-pipeline-after-full-book-context-bleed/design-doc/01-structured-book-ocr-pipeline-redesign-and-implementation-guide.md
      Note: |-
        Main intern-facing redesign guide
        Benchmark-informed redesign update
        Updated implementation progress
ExternalSources: []
Summary: Diary for the Book OCR pipeline redesign after the full-book context-image bleed regression.
LastUpdated: 2026-05-25T00:00:00-04:00
WhatFor: Use this diary to understand why the structured, target-page-only, turn-persisted OCR redesign ticket was created and what documentation was produced.
WhenToUse: Read before implementing the redesigned pipeline or continuing the ticket.
---




# Diary

## Goal

Capture the design and planning work for the next Book OCR pipeline after the full 202-page run exposed context-image bleed, inconsistent Markdown, false figure markers, and weak turn-level observability.

## Step 1: Create the pipeline redesign ticket and write the intern-facing guide

I created a new docmgr ticket for the work implied by the full-book OCR regression and the follow-up design discussion. The scope is broader than only adding a turns database: it covers target-page-only OCR, structured output contracts, deterministic Markdown rendering, Geppetto turn/session scoping, Pinocchio turn persistence, figure QA, duplicate-caption detection, and a safer full-book rerun plan.

I then wrote a long-form design and implementation guide for a new intern. The guide explains the current Book OCR system, the files involved, what went wrong in the last full-book run, which existing Geppetto and Pinocchio APIs should be reused, and how to implement the redesigned pipeline in incremental phases.

### Prompt Context

**User prompt (verbatim):** "Ok, create anew ticket for all this work (not just the turns stuff, but all since the last run).

Create  a detailed analysis / design / implementation guide that is for a new intern, explaining all the parts of the system needed to understand what it is, with prose paragraphs and bullet point sand pseudocode and diagrams and api references and file references. It should be very clear and technical. Store in the ticket and the nupload to remarkable.


Keep a diary as you work"

**Assistant interpretation:** Create a new docmgr ticket for the full OCR pipeline redesign that emerged after the full-book run, write a detailed intern-facing guide with diagrams, pseudocode, API/file references, upload it to reMarkable, and keep a diary.

**Inferred user intent:** The user wants the next implementation phase to be well-scoped and teachable, so an intern can understand the current OCR system, the regression, and the redesigned architecture before changing code.

**Commit (code):** N/A — this step creates documentation and planning artifacts.

### What I did

- Created ticket `BOOK-OCR-PIPELINE-REDESIGN-001`.
- Added tasks for:
  - analyzing the full-book OCR regression,
  - designing structured target-page-only OCR with Geppetto turns and Pinocchio persistence,
  - writing the intern-facing guide,
  - uploading the guide to reMarkable.
- Reviewed current implementation files:
  - `internal/ocrmvp/geppetto_ocr.go`
  - `internal/ocrmvp/prompt.go`
  - `internal/ocrmvp/types.go`
  - `internal/ocrquality/figures.go`
- Reviewed Geppetto and Pinocchio references:
  - `geppetto/pkg/doc/topics/08-turns.md`
  - `pinocchio/pkg/persistence/chatstore/turn_store.go`
  - `pinocchio/pkg/persistence/chatstore/turn_store_sqlite.go`
  - `pinocchio/pkg/cmds/cmdlayers/helpers.go`
- Wrote the design guide:
  - `design-doc/01-structured-book-ocr-pipeline-redesign-and-implementation-guide.md`

### Why

- The previous full-book run proved the pipeline can run end-to-end, but not that it produces a final-quality book artifact.
- Neighboring page PNG context caused target-page contamination.
- Freeform Markdown generation allowed style drift across pages.
- The new work needs a durable turn history for debugging and replay.
- The existing Pinocchio `chatstore` should be reused instead of inventing a new turns database.

### What worked

- The current code clearly shows where context images enter the OCR turn: `multimodalImages` in `internal/ocrmvp/geppetto_ocr.go`.
- Geppetto already provides a turn/block abstraction that fits chained page calls.
- Pinocchio already provides `--turns-dsn` / `--turns-db` patterns and a SQLite `chatstore.TurnStore` suitable for Book OCR turn snapshots.
- The design guide now gives an implementation sequence that starts with deterministic types/rendering and turn persistence before live OCR.

### What didn't work

- N/A for this documentation step. The underlying pipeline problem remains unresolved until implementation starts.

### What I learned

- The fix should not be only “make the prompt stricter.” A stricter prompt helps, but the primary design change is to stop sending neighboring page PNGs in primary OCR.
- Session scoping should be one page pipeline per session, not one giant book session and not one session per call.
- The turn store identifiers should be queryable: `convID` for book run, `sessionID` for page/chapter, `turnID` for the exact inference call.

### What was tricky to build

- The design had to preserve useful continuity without reintroducing the same visual bleed. The solution is to allow growing textual/structured history while forbidding neighboring page images in the primary OCR call.
- The guide also had to distinguish final reader-facing Markdown from debug/provenance information. Diagram labels may be useful, but they should not necessarily render inline after an embedded image link.

### What warrants a second pair of eyes

- Whether normalization should return structured JSON or directly return Markdown.
- Whether figure QA should run for every figure block or only suspicious crops.
- Whether `turns-db` should default to `<work-dir>/turns.db` even when the user does not explicitly request persistence.
- Whether the new package should be named `internal/ocrpipeline`, `internal/pageocr`, or something else.

### What should be done in the future

- Upload the guide to reMarkable.
- Commit the ticket documentation.
- Implement Phase 2 from the guide: `OCRTurnStore` wrapper and `--turns-db` / `--turns-dsn` support.
- Implement structured block types and deterministic renderer before another full live run.

### Code review instructions

- Start with the guide:
  - `/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/25/BOOK-OCR-PIPELINE-REDESIGN-001--redesign-book-ocr-pipeline-after-full-book-context-bleed/design-doc/01-structured-book-ocr-pipeline-redesign-and-implementation-guide.md`
- Check that the file/API references match current source code.
- Check that the implementation plan starts with deterministic contracts and tests before live provider calls.

### Technical details

Primary proposed scoping:

```text
convID    = book-ocr:<book-id>:<run-id>
sessionID = page:<NNN> or chapter:<NN>:continuity
turnID    = page:<NNN>:01-structured-ocr, page:<NNN>:02-normalize, page:<NNN>:03-figure-qa
phase     = input, final, parse-error, qa
```

Primary proposed chain:

```text
Call 1: target page PNG only -> structured OCR blocks
Call 2: structured blocks + text-only nearby context -> normalized structured blocks / Markdown
Call 3: target page PNG + figure block metadata -> figure QA / crop validation
Call 4: assembled text only -> chapter/book continuity patch suggestions
```

## Step 2: Upload the redesign guide to reMarkable and close the documentation loop

I uploaded the new redesign guide and diary as a bundled PDF to reMarkable. This makes the analysis portable for review away from the terminal and gives the next implementer a stable reading package before starting code changes.

After the upload succeeded, I checked off the reMarkable task and updated the ticket changelog so the ticket records both the written guide and the external delivery location.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Finish the documentation deliverable by sending it to reMarkable and recording that in the ticket.

**Inferred user intent:** The user wants the design package to be reviewable on the reMarkable, not only stored in the repository.

**Commit (code):** N/A — documentation/upload bookkeeping only.

### What I did

- Uploaded a bundle containing:
  - `design-doc/01-structured-book-ocr-pipeline-redesign-and-implementation-guide.md`
  - `reference/01-diary.md`
- Destination:
  - `/ai/2026/05/25/BOOK-OCR-PIPELINE-REDESIGN-001/BOOK OCR PIPELINE REDESIGN 001 Guide.pdf`
- Checked task 4.
- Updated the changelog with the upload result.

### Why

- The user explicitly requested a reMarkable upload.
- The guide is long enough that it benefits from being reviewed as a PDF with a table of contents.

### What worked

- `remarquee upload bundle` succeeded on the first attempt.

### What didn't work

- N/A.

### What I learned

- The current guide's Markdown is compatible with the reMarkable upload pipeline.

### What was tricky to build

- N/A for upload; the only important constraint was avoiding extra reMarkable status/list calls and using the direct upload command.

### What warrants a second pair of eyes

- Review the PDF for formatting/readability on the device, especially diagrams and long code blocks.

### What should be done in the future

- Start implementing the guide's Phase 2: turn persistence plumbing and structured OCR types.

### Code review instructions

- Validate with:
  - `docmgr doctor --ticket BOOK-OCR-PIPELINE-REDESIGN-001 --stale-after 30`
- Review the uploaded source Markdown in the ticket workspace.

### Technical details

Upload command shape:

```bash
remarquee upload bundle DESIGN.md DIARY.md \
  --name "BOOK OCR PIPELINE REDESIGN 001 Guide" \
  --remote-dir "/ai/2026/05/25/BOOK-OCR-PIPELINE-REDESIGN-001" \
  --toc-depth 2 \
  --non-interactive
```

## Step 3: Fold VLM benchmark evidence into the structured OCR redesign

The broad VLM separation benchmark changed the redesign from a prompt-only reaction into an evidence-backed pipeline policy. I updated the redesign guide to explain what the benchmark proved, what it did not prove, and how the production pipeline should use benchmark evidence without letting diagnostic image-context calls write final OCR text.

The updated design now separates production stages from diagnostic stages. Primary OCR remains target-page-only. Neighboring page images are allowed only in benchmark/diagnostic paths. Validation becomes a first-class pipeline concept that distinguishes context bleed, coverage misses, provider/schema failures, and false figure positives.

### Prompt Context

**User prompt (verbatim):** "improve oracles, then improve the OCR redesign. Commit at appropriate intervals, keep a diary as you work"

**Assistant interpretation:** After tightening the VLM benchmark oracles, revise the structured OCR redesign so it reflects the benchmark evidence and is more actionable for implementation.

**Inferred user intent:** The user wants the redesign to move from a high-level document to an implementation-ready guide that incorporates real benchmark findings.

**Commit (code):** N/A — design documentation update only.

### What I did

- Updated `design-doc/01-structured-book-ocr-pipeline-redesign-and-implementation-guide.md`.
- Added benchmark-related related files for:
  - `internal/vlmseparation/scenarios.go`,
  - `internal/vlmseparation/scoring.go`,
  - `internal/vlmseparation/oracle.go`,
  - `internal/vlmseparation/report.go`.
- Added a new benchmark evidence section with the broad-run findings.
- Added a benchmark-informed pipeline diagram that separates:
  - production structured OCR,
  - target-page figure QA,
  - text-only normalization,
  - diagnostic VLM separation benchmark calls.
- Revised the recommended first PR so it includes deterministic validation types and regression fixtures, not only renderer/session plumbing.

### Why

- The redesign guide was written before the broad risky-page benchmark and therefore did not include the strongest available evidence.
- The broad run did not reproduce forbidden-caption bleed, but that should not weaken the production page-boundary rule. Instead, it should refine how diagnostics and validation fit into the pipeline.

### What worked

- The guide now states that neighboring page images may be used in diagnostics, but not as a source of final OCR text.
- The guide now names concrete validation failure classes:
  - context bleed,
  - coverage miss,
  - provider/schema failure,
  - figure false positive.
- The first implementation PR now includes `internal/ocrvalidation` concepts and fixtures for page 12/13 and 115/116 regressions.

### What didn't work

- N/A

### What I learned

- Benchmark evidence should not be copied directly into production behavior. It should become validation policy and diagnostic tooling unless it proves a production path is safe across a larger envelope.
- The structured OCR redesign needs a separate validation package so production OCR does not depend on benchmark command code.

### What was tricky to build

- The benchmark result is nuanced: it weakens a blanket claim that all multi-image prompts inevitably bleed, but it does not justify using neighboring PNGs in primary OCR. The design text needed to preserve both facts.
- The dependency direction matters. Production `ocrpipeline` should not import a diagnostic benchmark package just to reuse scoring/oracle helpers.

### What warrants a second pair of eyes

- Review whether the proposed `internal/ocrvalidation` package boundary is the right place for shared oracle and adjacent-caption logic.
- Review whether the first implementation PR is still small enough after adding validation fixtures.

### What should be done in the future

- Implement the recommended first PR: structured types, renderer, turn store wrapper, and deterministic validation gates.
- Convert the benchmark oracles into reusable fixtures if they are needed by production QA.

### Code review instructions

- Start with the new `Benchmark Evidence Update` section in:
  - `/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/25/BOOK-OCR-PIPELINE-REDESIGN-001--redesign-book-ocr-pipeline-after-full-book-context-bleed/design-doc/01-structured-book-ocr-pipeline-redesign-and-implementation-guide.md`
- Then review the revised `Recommended First PR` section.
- Validate with:
  - `docmgr doctor --ticket BOOK-OCR-PIPELINE-REDESIGN-001 --stale-after 30`

### Technical details

Key production write boundary added to the design:

```text
Diagnostic benchmark calls can generate warnings and reports, but they cannot write final page Markdown.
```

Proposed package split:

```text
internal/ocrpipeline/      production structured OCR pipeline
internal/vlmseparation/    diagnostic benchmark and reporting package
internal/ocrquality/       deterministic QA and figure embedding helpers
internal/ocrvalidation/    shared validation/oracle helpers, if needed
```

## Step 4: Land the first deterministic structured OCR contracts

I implemented the first code slice of the redesign without adding any live OCR calls. This slice creates the structured page contracts, deterministic Markdown renderer, Pinocchio turn-store wrapper, and validation helpers that future structured OCR calls will use.

The purpose of this step is to make the non-model parts of the redesign executable and testable. The next live OCR client should plug into these contracts rather than inventing output shapes inside a prompt.

### Prompt Context

**User prompt (verbatim):** (same as Step 3)

**Assistant interpretation:** Continue from the benchmark-informed redesign into the first implementation slice, keeping changes deterministic and committing at a coherent boundary.

**Inferred user intent:** The user wants the redesign to move from documentation into code while preserving the safety constraints learned from the VLM benchmark.

**Commit (code):** `5011269d876b65d5f2f30c791a26280ca87475e2` — "Add structured OCR pipeline contracts"

### What I did

- Added `internal/ocrpipeline/types.go` with:
  - `StructuredPageOCR`,
  - `OCRBlock`,
  - page/block type enums,
  - list/table/figure/warning contracts.
- Added `internal/ocrpipeline/renderer.go` with deterministic Markdown rendering for:
  - page markers,
  - headings,
  - paragraphs,
  - lists,
  - tables,
  - figures,
  - footnotes,
  - blank pages,
  - optional page footers.
- Added `internal/ocrpipeline/session.go` with an OCR-specific wrapper around Pinocchio `chatstore.TurnStore`.
- Added `internal/ocrvalidation` with:
  - expected/forbidden anchor evaluation,
  - adjacent duplicate figure-caption detection,
  - caption extraction and normalization.
- Added tests for renderer behavior, turn persistence, ID helpers, anchor matching, adjacent captions, and the page 116 figure-reference case.
- Updated the redesign guide with an implementation progress section.

### Why

- The pipeline should establish deterministic contracts before adding model calls.
- The renderer and validation gates are the parts that keep final Markdown stable and page-local.
- Production validation should not import the diagnostic `vlmseparation` command package.

### What worked

- `go test ./internal/ocrpipeline ./internal/ocrvalidation -count=1` passes.
- `go test ./... -count=1` passes.
- The turn-store test verifies that saving `input` and `final` phases records both phases in `turn_block_membership`.

### What didn't work

- My first turn-store test expected two rows in the `turns` table. Pinocchio's schema stores one row per `(conv_id, session_id, turn_id)` and records phase snapshots in `turn_block_membership`. I corrected the test to check membership phases instead.

### What I learned

- The Pinocchio turn store treats phases as snapshots of one turn, not as separate top-level turns.
- This is a good fit for page pipelines: one page turn ID can have input/final snapshots without duplicating turn identity.

### What was tricky to build

- The renderer must not accidentally reintroduce diagram text into final Markdown when an image link exists. I made `IncludeDiagramText` opt-in.
- The validation package had to distinguish prose references to a figure from rendered figure captions. `ExtractFigureCaptions` only treats line-start `Figure N-M:` captions as captions.

### What warrants a second pair of eyes

- Review whether `internal/ocrvalidation` should eventually own all Report 794 benchmark oracles or stay generic.
- Review renderer defaults, especially whether figure descriptions should render as `[FIGURE: ...]` when no crop exists.
- Review whether the turn ID format should include schema version or model profile.

### What should be done in the future

- Add the target-page-only structured OCR client using these contracts.
- Add fake client tests before any live OCR test.
- Wire deterministic validation into the structured workflow package once workflow steps exist.

### Code review instructions

- Start with:
  - `/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/ocrpipeline/types.go`
  - `/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/ocrpipeline/renderer.go`
  - `/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/ocrpipeline/session.go`
- Then review:
  - `/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/ocrvalidation/anchors.go`
  - `/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/ocrvalidation/adjacent.go`
- Validate with:
  - `go test ./internal/ocrpipeline ./internal/ocrvalidation -count=1`
  - `go test ./... -count=1`

### Technical details

Pinocchio phase persistence nuance:

```text
turns table primary key: conv_id, session_id, turn_id
phase snapshots:        turn_block_membership.phase
```

Therefore input/final snapshots for the same page turn should be queried from `turn_block_membership`, not counted as separate `turns` rows.

## Step 5: Update the Obsidian project report with oracle and redesign progress

I updated the Obsidian VLM benchmark article so the durable project report matches the latest implementation state. The article now reflects the refined page 59/116 oracles, the regenerated report score of 0.992, and the first structured OCR contract implementation.

The update also explains the new production packages at a high level: `internal/ocrpipeline` for structured page contracts, deterministic rendering, and turn persistence, and `internal/ocrvalidation` for anchor and adjacent-caption validation.

### Prompt Context

**User prompt (verbatim):** (same as Step 3)

**Assistant interpretation:** Keep the vault report current after improving oracles and landing the first structured OCR redesign code slice.

**Inferred user intent:** The user wants the Obsidian report to remain the readable source of truth for benchmark findings and redesign progress.

**Commit (code):** N/A — vault documentation update. Obsidian commit: `54c97ee160fe55a8a9a7812f908bb7b52de42713` — "Article: update OCR benchmark and redesign progress"

### What I did

- Updated the VLM benchmark article in the parc vault.
- Replaced stale report interpretation around pages 59 and 116.
- Added a section describing the first structured OCR implementation slice.
- Committed and pushed the vault.

### Why

- The article previously said pages 59 and 116 were remaining oracle improvement targets.
- Those oracles were improved and the report was regenerated, so the article needed to be corrected.
- The first structured OCR contracts also changed project status and next steps.

### What worked

- Vault commit `54c97ee160fe55a8a9a7812f908bb7b52de42713` was pushed.

### What didn't work

- N/A

### What I learned

- The article should distinguish benchmark progress from production pipeline progress. They are related but have different write boundaries.

### What was tricky to build

- The update needed to avoid overstating the benchmark: improved oracles and high scores do not make neighboring image context a production write path.

### What warrants a second pair of eyes

- Review whether the article's latest status section is clear that the next code step is a fake/dry-run target-page-only structured OCR client.

### What should be done in the future

- Update the article again after the structured OCR client exists and has a dry-run workflow.

### Code review instructions

- Review the final sections of:
  - `/home/manuel/code/wesen/go-go-golems/go-go-parc/Projects/2026/05/25/ARTICLE - VLM Separation Benchmark for Book OCR - Prompt Block Layouts and Turn Persistence.md`

### Technical details

Vault commit:

```text
54c97ee160fe55a8a9a7812f908bb7b52de42713 Article: update OCR benchmark and redesign progress
```
