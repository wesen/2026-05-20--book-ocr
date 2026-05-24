---
Title: Diary
Ticket: OCR-MVP-001
Status: active
Topics:
    - ocr
    - workflow
    - scraper
    - implementation-guide
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: 2026-05-20--book-ocr/ttmp/2026/05/24/OCR-MVP-001--implement-simple-mvp-ocr-workflow-with-new-workflow-api/design-doc/01-mvp-ocr-workflow-implementation-guide.md
      Note: Primary implementation guide written in this ticket
    - Path: pinocchio/pkg/cmds/profilebootstrap/engine_settings.go
      Note: Evidence inspected for profile-backed Geppetto engine construction
    - Path: scraper/pkg/workflow/runtime.go
      Note: Evidence inspected while writing diary and design
ExternalSources: []
Summary: Chronological diary for OCR-MVP-001 ticket creation and implementation-guide authoring.
LastUpdated: 2026-05-24T20:58:00-04:00
WhatFor: Record what was investigated, written, constrained, and delivered for OCR-MVP-001.
WhenToUse: Read before implementing or extending the MVP OCR workflow.
---


# Diary

## Goal

Capture the creation of `OCR-MVP-001`: a new docmgr ticket and intern-oriented design/implementation guide for a simple OCR workflow built on `scraper/pkg/workflow`, with Geppetto used for OCR and Pinocchio's default profile registry resolution used for model/profile selection.

## Step 1: Create ticket workspace and start evidence-backed design

I created the `OCR-MVP-001` ticket in the `2026-05-20--book-ocr` doc workspace and added the primary design document plus this diary. I also added task bookkeeping for ticket setup, analysis, design writing, validation, and reMarkable delivery.

Then I began mapping the implementation surface: the new workflow API in `scraper/pkg/workflow`, the prior AITR OCR scripts and diary, and the Geppetto/Pinocchio APIs needed to replace shell-based OCR calls with an embedded Go OCR client.

### Prompt Context

**User prompt (verbatim):** "Ok, Create a new ticket to implement a simple MVP OCR workflow using the new API, in 2026-05-20--book-ocr/ now.

Create  a detailed analysis / design / implementation guide that is for a new intern, explaining all the parts of the system needed to understand what it is, with prose paragraphs and bullet point sand pseudocode and diagrams and api references and file references. It should be very clear and technical. Store in the ticket and the nupload to remarkable.

Keep a diary as you work."

**Assistant interpretation:** Create a new docmgr ticket and write a detailed intern-oriented implementation guide for a simple OCR workflow that exercises the newly added scraper workflow API, then upload the docs to reMarkable and keep a chronological diary.

**Inferred user intent:** The user wants a new implementation-ready planning artifact, not immediate code, so that an intern can implement the OCR MVP correctly and with enough system context.

### What I did

- Ran `docmgr status --summary-only` from `/home/manuel/workspaces/2026-05-20/book-ocr`.
- Created ticket `OCR-MVP-001` with title `Implement Simple MVP OCR Workflow with New Workflow API`.
- Added design doc:
  - `/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/OCR-MVP-001--implement-simple-mvp-ocr-workflow-with-new-workflow-api/design-doc/01-mvp-ocr-workflow-implementation-guide.md`
- Added diary doc:
  - `/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/OCR-MVP-001--implement-simple-mvp-ocr-workflow-with-new-workflow-api/reference/01-diary.md`
- Added initial tasks for setup, analysis, guide writing, validation, and reMarkable upload.
- Removed the default placeholder task `Add tasks here` from the generated `tasks.md`.
- Inspected current workflow files, Geppetto turn/inference docs, Pinocchio profile bootstrap code, and prior AITR OCR scripts.

### Why

- A new ticket keeps the MVP OCR implementation planning separate from `SCRAPER-JOBS-001`, which designed and implemented the workflow runtime API itself.
- The new guide needs to be evidence-backed so implementation starts from the real APIs, not guessed abstractions.

### What worked

- `docmgr ticket create-ticket` created the expected workspace under `ttmp/2026/05/24/OCR-MVP-001--implement-simple-mvp-ocr-workflow-with-new-workflow-api`.
- `docmgr doc add` created the expected design and diary docs.
- Repository inspection found the exact APIs needed for runtime wiring, step contexts, artifact stores, projection stores, Geppetto multimodal blocks, and Pinocchio profile resolution.

### What didn't work

- The generated `tasks.md` contained the default placeholder task `Add tasks here`. I removed it manually so the checklist reflects real work only.

### What I learned

- The root workspace includes `scraper`, `geppetto`, and `pinocchio`, so the implementation guide can describe cross-module code paths concretely.
- The OCR MVP should not reproduce the old Python SQLite queue. The new workflow runtime already provides durable steps, scheduling, retries, artifacts, projections, and operator controls.

### What was tricky to build

- The task spans three modules: `scraper` for workflow runtime, `geppetto` for inference, and `pinocchio` for profile-registry defaults. The design had to be explicit about which module owns which responsibility to avoid an intern shelling out to `pinocchio` or reimplementing profile resolution.

### What warrants a second pair of eyes

- Whether the final implementation should live directly in `scraper/pkg/workflows/ocrmvp` or in the `2026-05-20--book-ocr` app module as a consumer package. The guide recommends `scraper/pkg/workflows/ocrmvp` because the purpose is to validate the new scraper workflow API.

### What should be done in the future

- Implement the workflow package using the guide.
- Add a dry-run path and fake OCR client before live model calls.
- Add live OCR only behind explicit operator flags/environment variables.

### Code review instructions

- Start with the ticket docs:
  - `/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/OCR-MVP-001--implement-simple-mvp-ocr-workflow-with-new-workflow-api/design-doc/01-mvp-ocr-workflow-implementation-guide.md`
  - `/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/OCR-MVP-001--implement-simple-mvp-ocr-workflow-with-new-workflow-api/reference/01-diary.md`
- Validate ticket state with:
  - `cd /home/manuel/workspaces/2026-05-20/book-ocr && docmgr doctor --ticket OCR-MVP-001 --stale-after 30`

### Technical details

Ticket path:

```text
/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/OCR-MVP-001--implement-simple-mvp-ocr-workflow-with-new-workflow-api
```

Key source directories inspected:

```text
/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/workflow
/home/manuel/workspaces/2026-05-20/book-ocr/geppetto/pkg
/home/manuel/workspaces/2026-05-20/book-ocr/pinocchio/pkg/cmds/profilebootstrap
/home/manuel/workspaces/2026-05-20/book-ocr/claw-stuff/ttmp/2026/05/20/AITR-794--clean-markdown-transcription-of-aitr-794-pdf
```

## Step 2: Incorporate Geppetto OCR and Pinocchio profile-registry constraint

The user clarified that the OCR implementation should use `geppetto/` for OCR and should use proper profile registry resolution through Pinocchio's default registries. I updated the design around that constraint: the OCR executor should build a Geppetto engine through `pinocchio/pkg/cmds/profilebootstrap`, not shell out to `pinocchio` and not depend on a single `PINOCCHIO_PROFILE` environment variable.

This became one of the central design decisions in the implementation guide. The guide now explicitly identifies the API path: `profilebootstrap.NewCLISelectionValues`, `profilebootstrap.ResolveCLIEngineSettings`, `profilebootstrap.NewEngineFromResolvedCLIEngineSettings`, Geppetto `turns.NewUserMultimodalBlock`, and `engine.RunInference`.

### Prompt Context

**User prompt (verbatim):** "use geppetto/ for OCR stuff, use proper profile registry using pinocchio's default registries"

**Assistant interpretation:** Adjust the OCR MVP design so live OCR is embedded Geppetto inference and profile/model selection follows Pinocchio's normal profile registry defaults and precedence.

**Inferred user intent:** The user wants the MVP implementation to integrate with the real Go inference/profile stack rather than continuing the earlier script pattern of invoking the Pinocchio CLI with a hard-coded environment profile.

### What I did

- Inspected Pinocchio profile bootstrap code:
  - `/home/manuel/workspaces/2026-05-20/book-ocr/pinocchio/pkg/cmds/profilebootstrap/profile_selection.go`
  - `/home/manuel/workspaces/2026-05-20/book-ocr/pinocchio/pkg/cmds/profilebootstrap/engine_settings.go`
- Inspected Geppetto turn and inference documentation/code:
  - `/home/manuel/workspaces/2026-05-20/book-ocr/geppetto/pkg/turns/helpers_blocks.go`
  - `/home/manuel/workspaces/2026-05-20/book-ocr/geppetto/pkg/doc/topics/06-inference-engines.md`
  - `/home/manuel/workspaces/2026-05-20/book-ocr/geppetto/pkg/doc/topics/08-turns.md`
- Updated the implementation guide to state that live OCR should:
  - import Geppetto APIs;
  - construct a multimodal turn with page image bytes;
  - resolve engine settings through Pinocchio profile bootstrap;
  - build a Geppetto engine from the resolved settings;
  - call `RunInference` directly;
  - store selected profile/registry/config metadata in results/projections.

### Why

- Shelling out to `pinocchio code professional` would bypass the workflow runtime's structured error handling and make tests harder.
- Hard-coding `PINOCCHIO_PROFILE` would ignore Pinocchio's registry-source precedence and config document support.
- Using `profilebootstrap` gives the same default behavior as Pinocchio commands while keeping OCR execution embedded in Go.

### What worked

- Pinocchio already exposes a profile bootstrap package that wraps Geppetto's generic bootstrap with `AppName: pinocchio` and `EnvPrefix: PINOCCHIO`.
- Geppetto already supports multimodal user blocks with image maps containing `media_type`, `content`, `url`, and `detail`.
- The design could reuse real APIs rather than inventing a custom profile resolver.

### What didn't work

- I initially searched too broadly for `geppetto`/`pinocchio` symbols and got a lot of historical `ttmp` matches. I narrowed the evidence to source files under `geppetto/pkg`, `pinocchio/pkg`, and `scraper/pkg/workflow`.

### What I learned

- `pinocchio/pkg/cmds/profilebootstrap` is the right layer for an embedded Go consumer that wants Pinocchio defaults.
- `geppetto/pkg/cli/bootstrap` has a generic XDG profiles fallback, but using it directly would miss Pinocchio-specific unified config document handling.

### What was tricky to build

- The subtle distinction is that Geppetto owns inference and turns, while Pinocchio owns the user-facing profile/config defaulting story. The MVP should use both: Geppetto for OCR execution, Pinocchio profile bootstrap for default registries and active profile selection.

### What warrants a second pair of eyes

- Import-weight concerns: bringing `pinocchio/pkg/cmds/profilebootstrap` into `scraper` may pull in more Pinocchio dependencies than desired. If that is a problem, extract a smaller shared profile bootstrap package rather than duplicating the profile-resolution logic.

### What should be done in the future

- During implementation, add a fake OCR client first, then the Geppetto-backed client.
- Add a test that verifies explicit profile/profile-registry inputs are passed to `profilebootstrap.NewCLISelectionValues` without directly reading environment variables.

### Code review instructions

- Review the `Geppetto OCR Client Design` and `Pinocchio profile registry defaults` sections in the design doc.
- When code is implemented, start review in the future `scraper/pkg/workflows/ocrmvp/geppetto_ocr.go` file.

### Technical details

The guide's intended live OCR path is:

```text
OCR page executor
  -> profilebootstrap.NewCLISelectionValues(...)
  -> profilebootstrap.ResolveCLIEngineSettings(ctx, parsed)
  -> profilebootstrap.NewEngineFromResolvedCLIEngineSettings(resolved)
  -> turns.NewUserMultimodalBlock(prompt, image map)
  -> engine.RunInference(ctx, turn)
  -> extract last turns.BlockKindLLMText payload
```

## Step 3: Validate ticket and upload guide bundle to reMarkable

After writing the guide and relating source files, I validated the ticket with `docmgr doctor`. The first validation pass found missing vocabulary entries for the new ticket topics, so I added `workflow` and `implementation-guide` to the docmgr topic vocabulary and reran validation successfully.

I then uploaded the design guide and diary as a single PDF bundle to reMarkable under the ticket-specific `/ai/2026/05/24/OCR-MVP-001` folder. The upload completed successfully.

### Prompt Context

**User prompt (verbatim):** (same as Step 1 and Step 2)

**Assistant interpretation:** Complete ticket bookkeeping, validation, and reMarkable delivery for the OCR MVP planning guide.

**Inferred user intent:** The user wants the documentation not only written locally but also delivered to the reMarkable device for reading/review.

### What I did

- Related the design doc to the key workflow, Geppetto, Pinocchio, and prior OCR reference files.
- Related the diary to the design doc and key evidence files.
- Marked tasks 1-3 complete.
- Ran:
  - `docmgr doctor --ticket OCR-MVP-001 --stale-after 30`
- Added vocabulary entries:
  - `workflow`
  - `implementation-guide`
- Reran doctor successfully.
- Uploaded bundle:
  - `OCR MVP 001 Workflow Guide.pdf`
  - destination `/ai/2026/05/24/OCR-MVP-001`

### Why

- `docmgr doctor` keeps the ticket searchable and consistent.
- The reMarkable upload makes the long-form implementation guide easy to review away from the terminal.

### What worked

- After adding vocabulary entries, `docmgr doctor` reported all checks passed.
- `remarquee upload bundle` succeeded with:
  - `OK: uploaded OCR MVP 001 Workflow Guide.pdf -> /ai/2026/05/24/OCR-MVP-001`

### What didn't work

- Initial doctor warning:
  - `Unknown vocabulary value for Topics`
  - Values: `implementation-guide,workflow`
- Fixed by adding both topic slugs to the vocabulary.

### What I learned

- The doc workspace vocabulary did not yet include generic `workflow` or `implementation-guide` topics even though they are appropriate for this ticket.

### What was tricky to build

- The only tricky part was bookkeeping: because `docmgr doctor` validates vocabulary, topic additions had to be done before considering the ticket clean.

### What warrants a second pair of eyes

- Confirm the chosen topic slugs are acceptable for future tickets and not too broad.
- Confirm the reMarkable bundle name and destination are easy to find.

### What should be done in the future

- Implement the OCR MVP code following the guide.
- Add follow-up docs if implementation changes the recommended file layout or API usage.

### Code review instructions

- Validate docs with:
  - `cd /home/manuel/workspaces/2026-05-20/book-ocr && docmgr doctor --ticket OCR-MVP-001 --stale-after 30`
- Review uploaded source docs locally:
  - `design-doc/01-mvp-ocr-workflow-implementation-guide.md`
  - `reference/01-diary.md`

### Technical details

Upload command:

```bash
remarquee upload bundle \
  design-doc/01-mvp-ocr-workflow-implementation-guide.md \
  reference/01-diary.md \
  --name "OCR MVP 001 Workflow Guide" \
  --remote-dir "/ai/2026/05/24/OCR-MVP-001" \
  --toc-depth 2 \
  --non-interactive
```
