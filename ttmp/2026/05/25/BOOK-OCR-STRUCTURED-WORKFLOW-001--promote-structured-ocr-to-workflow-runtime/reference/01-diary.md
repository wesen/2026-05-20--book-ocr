---
Title: Diary
Ticket: ""
Status: active
Topics:
    - book-processing
    - ocr
    - workflow
    - geppetto
    - pinocchio
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ttmp/2026/05/25/BOOK-OCR-STRUCTURED-WORKFLOW-001--promote-structured-ocr-to-workflow-runtime/design-doc/01-workflow-backed-structured-ocr-design-and-implementation-guide.md
      Note: Primary workflow promotion design guide
    - Path: ttmp/2026/05/25/BOOK-OCR-STRUCTURED-WORKFLOW-001--promote-structured-ocr-to-workflow-runtime/reference/01-diary.md
      Note: Chronological implementation diary
ExternalSources: []
Summary: ""
LastUpdated: 0001-01-01T00:00:00Z
WhatFor: ""
WhenToUse: ""
---


# Diary

## Goal

This diary records the work to promote the structured OCR runner from a direct CLI loop into a workflow-runtime-backed package with durable retry, resume, projections, artifacts, and operator controls.

## Step 1: Create ticket and write workflow promotion design guide

I created a new ticket for Option B: promoting structured OCR into the existing `scraper/pkg/workflow` runtime instead of continuing to extend CLI-level artifact resume. The ticket now has a design guide that explains the current direct structured runner, the existing freeform workflow package, the workflow runtime APIs, and the planned structured workflow package.

The design guide is intentionally intern-facing. It explains the codebase boundaries, file references, current gaps, proposed step graph, retry policy, projection schema, artifact strategy, CLI behavior, implementation phases, and validation commands.

### Prompt Context

**User prompt (verbatim):** "ok, create a new ticket for option B, and create a design + implementatino guide. Create  a detailed analysis / design / implementation guide that is for a new intern, explaining all the parts of the system needed to understand what it is, with prose paragraphs and bullet point sand pseudocode and diagrams and api references and file references. It should be very clear and technical. Store in the ticket and the nupload to remarkable.


Then add detailed tasks to the ticket, and then work on them one by one, keeping a detailed diary, and committing at appropriate intervals."

**Assistant interpretation:** Create a new docmgr ticket for workflow-backed structured OCR, write a comprehensive design/implementation guide, upload it to reMarkable, add detailed tasks, then start implementing task-by-task with diary and commits.

**Inferred user intent:** The user wants to stop relying on CLI-level resume and use the workflow runtime's real retry/resume/operator capabilities for structured OCR, with enough documentation for a new engineer to pick up the work.

**Commit (code):** N/A — ticket setup and design documentation only.

### What I did

- Created ticket `BOOK-OCR-STRUCTURED-WORKFLOW-001`.
- Added design doc:
  - `design-doc/01-workflow-backed-structured-ocr-design-and-implementation-guide.md`
- Added this diary:
  - `reference/01-diary.md`
- Added initial tasks covering investigation, guide writing/upload, workflow package implementation, CLI wiring, projections, dry-run validation, and live smoke validation.
- Inspected current structured OCR CLI and page pipeline.
- Inspected old freeform OCR workflow package.
- Inspected the workflow runtime API surface in `scraper/pkg/workflow`.
- Wrote the long-form design and implementation guide.

### Why

- The structured OCR first-50 live run proved table and figure-boundary architecture, but operationally it still exits on errors.
- The project already has a workflow runtime that supports retries and operator controls, so structured OCR should use it rather than recreate retries in CLI code.

### What worked

- Ticket and docs were created successfully.
- The design guide now maps:
  - current direct structured runner,
  - reusable structured page pipeline,
  - freeform workflow template,
  - workflow runtime APIs,
  - desired structured workflow graph,
  - implementation phases.

### What didn't work

- N/A

### What I learned

- The structured page function `RunStructuredPage` is already a good workflow executor boundary. The missing layer is packaging, projections, workflow artifacts, and CLI runtime wiring.
- The old `internal/ocrmvp` workflow is the best immediate implementation template.

### What was tricky to build

- The main design tradeoff is whether to refactor shared discovery out of `internal/ocrmvp` before implementing structured workflow. The guide recommends reusing the old helper initially or moving it later only if necessary, to avoid blocking workflow promotion on cleanup.

### What warrants a second pair of eyes

- Review whether parse failures should be permanent or retryable.
- Review whether `structured-run` should keep a `--direct` fallback after workflow integration.
- Review the proposed structured projection schema before tests depend on it.

### What should be done in the future

- Upload the design guide to reMarkable.
- Implement Phase 1: `workflow_types.go` and `workflow_package.go`.

### Code review instructions

- Start with the design doc:
  - `/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/25/BOOK-OCR-STRUCTURED-WORKFLOW-001--promote-structured-ocr-to-workflow-runtime/design-doc/01-workflow-backed-structured-ocr-design-and-implementation-guide.md`
- Then inspect the implementation templates listed in the guide.

### Technical details

Ticket commands used:

```bash
docmgr ticket create-ticket \
  --ticket BOOK-OCR-STRUCTURED-WORKFLOW-001 \
  --title "Promote Structured OCR to Workflow Runtime" \
  --topics book-processing,ocr,workflow,geppetto,pinocchio

docmgr doc add \
  --ticket BOOK-OCR-STRUCTURED-WORKFLOW-001 \
  --doc-type design-doc \
  --title "Workflow Backed Structured OCR Design and Implementation Guide"

docmgr doc add \
  --ticket BOOK-OCR-STRUCTURED-WORKFLOW-001 \
  --doc-type reference \
  --title "Diary"
```

## Step 2: Upload workflow promotion guide to reMarkable

I uploaded the workflow-backed structured OCR guide and diary as a bundled PDF to reMarkable. This gives the intern-facing implementation guide an offline review copy before implementation begins.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** After writing the guide, publish it to reMarkable as requested.

**Inferred user intent:** The user wants the new ticket's design package available for reading outside the coding environment.

**Commit (code):** N/A — upload/bookkeeping only.

### What I did

- Uploaded the design guide and diary with `remarquee upload bundle`.
- Destination:
  - `/ai/2026/05/25/BOOK-OCR-STRUCTURED-WORKFLOW-001/BOOK OCR STRUCTURED WORKFLOW 001 Guide.pdf`

### Why

- The user explicitly requested upload to reMarkable after creating the design and implementation guide.

### What worked

- Upload succeeded with:
  - `OK: uploaded BOOK OCR STRUCTURED WORKFLOW 001 Guide.pdf -> /ai/2026/05/25/BOOK-OCR-STRUCTURED-WORKFLOW-001`

### What didn't work

- N/A

### What I learned

- The ticket now has both local docmgr documentation and a reMarkable PDF handoff artifact.

### What was tricky to build

- N/A

### What warrants a second pair of eyes

- Confirm the PDF renders the Mermaid-style ASCII diagrams and long code blocks readably on device.

### What should be done in the future

- Re-upload after major implementation updates if the guide changes materially.

### Code review instructions

- Review the uploaded sources:
  - `/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/25/BOOK-OCR-STRUCTURED-WORKFLOW-001--promote-structured-ocr-to-workflow-runtime/design-doc/01-workflow-backed-structured-ocr-design-and-implementation-guide.md`
  - `/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/25/BOOK-OCR-STRUCTURED-WORKFLOW-001--promote-structured-ocr-to-workflow-runtime/reference/01-diary.md`

### Technical details

Upload command:

```bash
remarquee upload bundle \
  ttmp/2026/05/25/BOOK-OCR-STRUCTURED-WORKFLOW-001--promote-structured-ocr-to-workflow-runtime/design-doc/01-workflow-backed-structured-ocr-design-and-implementation-guide.md \
  ttmp/2026/05/25/BOOK-OCR-STRUCTURED-WORKFLOW-001--promote-structured-ocr-to-workflow-runtime/reference/01-diary.md \
  --name "BOOK OCR STRUCTURED WORKFLOW 001 Guide" \
  --remote-dir "/ai/2026/05/25/BOOK-OCR-STRUCTURED-WORKFLOW-001" \
  --toc-depth 2 \
  --non-interactive
```
