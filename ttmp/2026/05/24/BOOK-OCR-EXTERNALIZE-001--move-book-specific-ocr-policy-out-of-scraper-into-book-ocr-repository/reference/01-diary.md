---
Title: Diary
Ticket: BOOK-OCR-EXTERNALIZE-001
Status: active
Topics:
    - ocr
    - workflow
    - book-processing
    - implementation-guide
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../../scraper/pkg/workflows/bookprofile/profile.go
      Note: Inspected source for diary analysis
    - Path: ttmp/2026/05/24/BOOK-OCR-EXTERNALIZE-001--move-book-specific-ocr-policy-out-of-scraper-into-book-ocr-repository/design-doc/01-externalizing-book-ocr-policy-from-scraper-design-and-implementation-guide.md
      Note: Main design output for Step 1
ExternalSources: []
Summary: Diary for planning how to move book-specific OCR policy out of scraper and into the 2026-05-20--book-ocr repository.
LastUpdated: 2026-05-24T23:31:44-04:00
WhatFor: Use this to retrace the analysis and design decisions for BOOK-OCR-EXTERNALIZE-001.
WhenToUse: Read before implementing the external profile/prompt/policy migration.
---


# Diary

## Goal

This diary records the analysis work for moving Report 794/book-specific OCR policy out of `scraper/` and into the sibling `2026-05-20--book-ocr` repository, so that `scraper/` stays a generic workflow/OCR runtime.

## Step 1: Create the externalization ticket and write the migration design guide

I created `BOOK-OCR-EXTERNALIZE-001` as a focused planning ticket for separating generic OCR runtime code from concrete book policy. The main output is an intern-facing design and implementation guide that explains which pieces of the current system are generic, which pieces are Report 794-specific, where the specific data should move, and how to implement the migration safely.

The design recommends a file-contract boundary: `scraper/` should keep schemas, loaders, workflow execution, prompt rendering primitives, QA primitives, and figure algorithms; `2026-05-20--book-ocr/` should own concrete book profiles, prompts, QA lists, required figure expectations, discovery files, patch proposals, and experiments.

### Prompt Context

**User prompt (verbatim):** "Let's actually move the book-ocr specific parts into a separate ./2026-05-20--book-ocr/ repository, to keep scraper/ as generic as possible. 

Create a new docmgr ticket and do a detailed analysis of how to do that.

Create  a detailed analysis / design / implementation guide that is for a new intern, explaining all the parts of the system needed to understand what it is, with prose paragraphs and bullet point sand pseudocode and diagrams and api references and file references. It should be very clear and technical. Store in the ticket and the nupload to remarkable.

Keep a diary as you work"

**Assistant interpretation:** Create a new ticket, analyze the current coupling between `scraper/` and Report 794-specific OCR policy, write a detailed implementation guide for an intern, upload it to reMarkable, and keep a diary.

**Inferred user intent:** The user wants `scraper/` to become reusable infrastructure and wants the concrete book OCR project to own book-specific profiles, prompts, QA expectations, discoveries, and experiment artifacts.

**Commit (code):** N/A — analysis/design only in this step.

### What I did

- Created docmgr ticket `BOOK-OCR-EXTERNALIZE-001`.
- Added design doc `design-doc/01-externalizing-book-ocr-policy-from-scraper-design-and-implementation-guide.md`.
- Added this diary `reference/01-diary.md`.
- Added ticket tasks for analysis, guide writing, and reMarkable upload.
- Inspected current code locations that contain Report 794-specific policy:
  - `scraper/pkg/workflows/bookprofile/profile.go`
  - `scraper/pkg/workflows/ocrmvp/prompt.go`
  - `scraper/pkg/workflows/ocrquality/markdown.go`
- Wrote a migration design that moves concrete book policy into `2026-05-20--book-ocr/books/report-794/`.

### Why

- The earlier implementation intentionally moved quickly and embedded Report 794 knowledge in `scraper/` to prove the workflow.
- Long-term, `scraper/` should not accumulate one-off book facts.
- A file-based profile/prompt/QA contract lets `scraper/` remain reusable while `2026-05-20--book-ocr` preserves the concrete Report 794 project.

### What worked

- The boundary is clear: schemas and algorithms stay in `scraper`; profile data and prompt packs move to `2026-05-20--book-ocr`.
- The current stable-profile/discovery/patch model maps well to this repository split.
- Existing CLI flags such as `--book-profile`, `--discovery`, and `--profile-patch` already point in the right direction.

### What didn't work

- No code was moved in this step; the current `scraper` code still contains Report 794 policy.
- The current profile schema does not yet support relative file references such as `expected_strings_path` or `template_path`.
- `ocrmvp/prompt.go` still mixes generic prompt improvements with Report 794-specific vocabulary.

### What I learned

- The most important migration is not moving code; it is moving data and policy.
- The safest dependency boundary is a file contract, not a Go import from the docs/project repository.
- The migration should proceed in small commits: add external profile first, then extend loaders, then remove built-ins, then move prompts.

### What was tricky to build

- The design has to preserve reproducibility. If policy moves outside `scraper`, manifests and tickets must record exact profile paths and prompt files.
- There are two destinations for state: durable current book config under `books/report-794/`, and historical experiment evidence under ticket workspaces. The guide explicitly keeps both.
- Prompt versions are mixed: some parts are generic improvements, while other parts are concrete Report 794 vocabulary. The guide separates those concepts.

### What warrants a second pair of eyes

- Whether the durable book folder should be `books/report-794/` or another root layout.
- Whether generic family templates belong in `scraper` examples or in `2026-05-20--book-ocr/books/_templates/`.
- Whether prompt templates should use Go `text/template` or a simpler declarative format.

### What should be done in the future

- Implement Phase 1: create `2026-05-20--book-ocr/books/report-794/` with external profile, prompt, and QA files.
- Implement relative file expansion in `bookprofile.Load()`.
- Switch smoke tests to use `--book-profile` path.
- Remove production `Report794()` built-in profile from `scraper`.
- Move Report 794 prompt text out of `ocrmvp/prompt.go`.

### Code review instructions

- Review the design doc first:
  - `/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/BOOK-OCR-EXTERNALIZE-001--move-book-specific-ocr-policy-out-of-scraper-into-book-ocr-repository/design-doc/01-externalizing-book-ocr-policy-from-scraper-design-and-implementation-guide.md`
- Then inspect the current coupling points:
  - `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/workflows/bookprofile/profile.go`
  - `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/workflows/ocrmvp/prompt.go`
  - `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/workflows/ocrquality/markdown.go`

### Technical details

Target external folder proposed by the guide:

```text
2026-05-20--book-ocr/books/report-794/
├── README.md
├── book.profile.yaml
├── book.discovery.yaml
├── book.profile.patch.yaml
├── prompts/
│   ├── ocr-page.md
│   ├── ocr-page-figure-aware.md
│   └── continuity-pass.md
├── qa/
│   ├── expected-strings.txt
│   ├── known-bad-terms.txt
│   └── required-figures.yaml
├── manifests/
│   └── first-30-pages.yaml
└── experiments/
    └── README.md
```

Recommended runtime shape after migration:

```bash
ocr-mvp quality-pass \
  --book-profile /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/books/report-794/book.profile.yaml \
  --markdown RAW.md \
  --output-dir OUT \
  --image-dir /home/manuel/code/wesen/claw-stuff/output/books/presentation-based-uis/pages \
  --embed-figures
```

## Step 2: Correct the scope from policy externalization to full OCR extraction

The user corrected the architectural direction. The goal is not merely to move Report 794-specific profile data and prompt constants out of `scraper`. The goal is to move all OCR and book-OCR functionality out of `scraper`, leaving `scraper` as the workflow management, execution, job queue, artifact, projection, retry, and operator-control runtime.

I updated both the ticket design guide and the Obsidian project report to reflect this stronger boundary. The corrected design says that `2026-05-20--book-ocr` should become a Go workflow application that imports `scraper/pkg/workflow`, registers OCR workflow packages from outside `scraper`, and owns the OCR CLI, page OCR workflow, quality workflow, book profiles, prompt rendering, Geppetto client, figure extraction, sidecars, discovery files, and experiments.

### Prompt Context

**User prompt (verbatim):** "no, i wwant to move all the book-ocr functionality / all th eocr functionality out of scraper, and leave just the workflow management execution / job queue mechanisms there. All the book ocr functionality that we added should be in this separate repo. Update the document"

**Assistant interpretation:** Correct the externalization design to extract all OCR-related packages and commands from `scraper`, not only Report 794-specific policy.

**Inferred user intent:** The user wants `scraper` to remain a generic workflow/job runtime and wants Book OCR to become a separate workflow application repository.

**Commit (code):** N/A — documentation correction only.

### What I did

- Rewrote the `BOOK-OCR-EXTERNALIZE-001` design guide around the corrected boundary.
- Renamed and rewrote the Obsidian project report to focus on extracting OCR from `scraper`.
- Clarified that these current `scraper` paths should move to `2026-05-20--book-ocr`:
  - `scraper/pkg/workflows/ocrmvp`
  - `scraper/pkg/workflows/ocrquality`
  - `scraper/pkg/workflows/bookprofile`
  - `scraper/cmd/ocr-mvp`
- Clarified that `scraper` should keep `pkg/workflow` and runtime/job mechanisms only.
- Added an implementation plan where `2026-05-20--book-ocr` becomes a Go module importing `scraper/pkg/workflow`.

### Why

- Keeping generic OCR workflows in `scraper` still makes `scraper` an OCR framework.
- OCR is a workload that uses the runtime; it is not part of the runtime itself.
- A separate `book-ocr` application can evolve rapidly without adding OCR-specific concerns to the workflow engine.

### What worked

- The corrected boundary is simpler and cleaner than the previous design.
- The existing workflow API is already close to supporting external packages.
- The migration can proceed package-by-package without changing OCR behavior first.

### What didn't work

- The previous design guide and article framed the target too narrowly.
- No code has been moved yet; the current implementation still has OCR packages inside `scraper`.

### What I learned

- The important distinction is not Report 794-specific versus generic OCR. The important distinction is runtime versus workflow application.
- Book OCR should import the workflow runtime. The workflow runtime should not contain Book OCR.

### What was tricky to build

- The design has to preserve the educational value of the OCR pipeline as an example of the Scraper job system while still removing the example from the runtime repo.
- The plan must avoid a hard module dependency in the wrong direction. The only correct direction is `book-ocr -> scraper/pkg/workflow`.

### What warrants a second pair of eyes

- Whether `scraper/pkg/workflow` exports enough runtime construction and operator APIs for an external CLI.
- Whether any OCR code currently depends on `scraper` internals outside `pkg/workflow`.
- Whether the package names should become `internal/ocr/workflow` or exported `pkg/ocr` packages in the book OCR repo.

### What should be done in the future

- Create `go.mod` in `2026-05-20--book-ocr`.
- Move `ocrmvp` first with minimal behavior changes.
- Move Geppetto client, quality workers, bookprofile, and CLI in separate commits.
- Delete OCR packages from `scraper` after external smoke tests pass.

### Code review instructions

- Review the updated design doc:
  - `/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/BOOK-OCR-EXTERNALIZE-001--move-book-specific-ocr-policy-out-of-scraper-into-book-ocr-repository/design-doc/01-externalizing-book-ocr-policy-from-scraper-design-and-implementation-guide.md`
- Review the updated Obsidian article:
  - `/home/manuel/code/wesen/go-go-golems/go-go-parc/Projects/2026/05/24/ARTICLE - Extracting Book OCR from Scraper - Workflow Runtime and External OCR Pipelines.md`

### Technical details

Corrected target:

```text
scraper/                  workflow runtime and job queue mechanisms
2026-05-20--book-ocr/     OCR application, profiles, prompts, QA, figures, experiments
```

## Step 3: Add implementation phases for moving OCR into the book-ocr repository

I added the concrete implementation phases to the ticket before moving code. The plan now breaks the extraction into repository setup, package copy, dependency wiring, external smoke tests, scraper removal, scraper verification, and final documentation.

The key sequencing decision is to prove the external OCR application first, then delete OCR from `scraper`. This keeps the move reversible during implementation and gives a clear parity check before removing the old packages.

### Prompt Context

**User prompt (verbatim):** "go ahead, add detailed phases and tasks to the ticket, then move everything over. First, setup a default go package from ~/code/wesen/go-go-golems/go-template . Then do everything task by task, committing at appropriate intervals, keeping a detailed diary as you work."

**Assistant interpretation:** Turn the corrected extraction design into phased ticket tasks, set up the book-ocr repo as a Go module from the template, move all OCR packages/CLI out of scraper, test/commit as each phase completes, and keep a retraceable diary.

**Inferred user intent:** The user wants implementation, not only design: `scraper` should become workflow-runtime-only, and `2026-05-20--book-ocr` should become the OCR application repository.

**Commit (code):** N/A — task planning only in this diary step.

### What I did

- Added implementation tasks for phases 1 through 8:
  - Go module/template setup.
  - OCR page workflow copy.
  - OCR quality/bookprofile/CLI copy.
  - Dependency wiring and tests.
  - External smoke test.
  - Scraper OCR removal.
  - Scraper verification.
  - Diary/changelog/design update.

### Why

- Moving code across repositories needs clear checkpoints.
- The old `scraper` packages should not be deleted until the copied external application compiles and passes a real smoke test.

### What worked

- The ticket now has implementation tasks that match the corrected design.

### What didn't work

- No code has been moved yet in this step.

### What I learned

- The existing `2026-05-20--book-ocr` repository already contained the go-template scaffold files, but still had `XXX` placeholders. The setup phase should normalize those placeholders rather than blindly overwriting ticket docs.

### What was tricky to build

- The template setup must not erase the existing docmgr ticket workspace and reports in the repository.

### What warrants a second pair of eyes

- Whether the module path should be `github.com/go-go-golems/book-ocr` or another final repository name.

### What should be done in the future

- Execute phases 1 through 8 and record commits/tests in this diary.

### Code review instructions

- Review `tasks.md` in the ticket to confirm phase ordering before reviewing code movement.

### Technical details

The external application should ultimately run as:

```bash
go run ./cmd/book-ocr quality-pass --book-profile ./books/report-794/book.profile.yaml ...
```
