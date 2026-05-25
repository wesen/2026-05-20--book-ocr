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
