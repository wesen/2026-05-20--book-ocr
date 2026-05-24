---
Title: Diary
Ticket: SCRAPER-JOBS-001
Status: active
Topics:
    - scraper
    - jobs
    - ocr
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: 2026-05-20--book-ocr/cmd/XXX/main.go
      Note: Evidence that the current book OCR repo is still a stub
    - Path: 2026-05-20--book-ocr/go.mod
      Note: Evidence that the current book OCR repo still has template module naming
    - Path: 2026-05-20--book-ocr/ttmp/2026/05/24/SCRAPER-JOBS-001--general-purpose-task-workflow-job-system-for-scraper/design-doc/01-scraper-task-workflow-job-system-design-and-book-ocr-mapping.md
      Note: Primary design deliverable produced in Step 1
    - Path: 2026-05-20--book-ocr/ttmp/2026/05/24/SCRAPER-JOBS-001--general-purpose-task-workflow-job-system-for-scraper/design-doc/02-embeddable-scraper-job-api-and-river-backend-design-considerations.md
      Note: Second design deliverable for embeddable River-like API and backend tradeoffs
    - Path: 2026-05-20--book-ocr/ttmp/2026/05/24/SCRAPER-JOBS-001--general-purpose-task-workflow-job-system-for-scraper/design-doc/02-embeddable-workflow-runtime-api-design.md
      Note: Rewritten second design document with workflow-native embeddable runtime API
    - Path: 2026-05-20--book-ocr/ttmp/2026/05/24/SCRAPER-JOBS-001--general-purpose-task-workflow-job-system-for-scraper/reference/01-diary.md
      Note: Chronological investigation diary for the ticket
    - Path: claw-stuff/ttmp/2026/05/20/AITR-794--clean-markdown-transcription-of-aitr-794-pdf/reference/01-diary.md
      Note: Source diary for Step 3 design refinement
    - Path: claw-stuff/ttmp/2026/05/20/AITR-794--clean-markdown-transcription-of-aitr-794-pdf/scripts/10-reprocess-universal.py
      Note: Prompt and page-attempt evidence used in Step 3
ExternalSources: []
Summary: Chronological diary for the scraper general-purpose workflow/job-system research and design ticket.
LastUpdated: 2026-05-24T16:24:00-04:00
WhatFor: Use this to resume or review the investigation and documentation work for SCRAPER-JOBS-001.
WhenToUse: Read before extending the design, implementing OCR workflows, or changing scheduler/backend architecture.
---





# Diary

## Goal

This diary records the creation of a docmgr ticket and the evidence-backed design work for evolving `scraper` into a general-purpose task/workflow/job system, with book OCR as the proving workload.

## Step 1: Create ticket, inspect architecture, and write the first design package

I created the `SCRAPER-JOBS-001` ticket, added an intern-oriented design document, and gathered evidence from both the mature `scraper/` repository and the current `2026-05-20--book-ocr/` repository. The main conclusion is that `scraper` already has a generic durable workflow core; the design recommends proving generalization with a book OCR package first, then adding production features such as heartbeats, cooperative cancellation, large artifact storage, and stage-level progress before considering River/Postgres.

I also captured River documentation into the ticket sources so the backend comparison is grounded in current River concepts rather than relying only on memory. The OCR repository currently appears to be a template stub, so the OCR mapping is intentionally a target design rather than a migration plan from existing OCR code.

### Prompt Context

**User prompt (verbatim):** "Create a new docmgr ticket to run @scraper/ into a more general purpose task / workflow / job system, which could for example be applie to running a book ocr like in @2026-05-20--book-ocr/ . It's scheduling part could potentially be tightened up by using something like river, although it doesn't necessarily need to be and maybe the current setup is just fine. Or maybe it could support different backends depend on its needs.

Create  a detailed analysis / design / implementation guide that is for a new intern, explaining all the parts of the system needed to understand what it is, with prose paragraphs and bullet point sand pseudocode and diagrams and api references and file references. It should be very clear and technical. Store in the ticket and the nupload to remarkable.

A aprt of this analysis would be to see how well it would be possible to map something like book ocr into the scraper system, and what that could look like, especially also from an admin/production point of view, where things need to be monitored, restartaed, signaled, shown to th euser, debugged.

Keep a diary as you work."

**Assistant interpretation:** Create a structured docmgr research/design ticket, deeply analyze `scraper/` as a generic workflow/job runtime, evaluate whether book OCR fits the model, compare the current scheduler with optional River-style backends, write an intern-ready implementation guide, maintain a diary, and upload the result to reMarkable.

**Inferred user intent:** The user wants a durable technical planning artifact that can guide a new intern through the existing system and de-risk a future implementation/refactor rather than jumping directly into code changes.

**Commit (code):** N/A — documentation-only ticket work; no runtime code was changed.

### What I did

- Created docmgr vocabulary topics `scraper`, `jobs`, and `ocr` when needed.
- Created ticket `SCRAPER-JOBS-001` with title `General Purpose Task Workflow Job System for Scraper`.
- Created the primary design doc:
  - `/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/SCRAPER-JOBS-001--general-purpose-task-workflow-job-system-for-scraper/design-doc/01-scraper-task-workflow-job-system-design-and-book-ocr-mapping.md`
- Created this diary:
  - `/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/SCRAPER-JOBS-001--general-purpose-task-workflow-job-system-for-scraper/reference/01-diary.md`
- Added ticket tasks for architecture mapping, OCR mapping, backend comparison, guide writing, validation, and reMarkable upload.
- Inspected current `scraper/` files, especially:
  - `pkg/engine/model/types.go`
  - `pkg/engine/scheduler/scheduler.go`
  - `pkg/engine/store/store.go`
  - `pkg/engine/store/sqlite/lease_store.go`
  - `pkg/engine/store/sqlite/result_store.go`
  - `pkg/sites/submitverbs/host.go`
  - `pkg/sites/submitverbs/runtime.go`
  - `pkg/sites/manifest/manifest.go`
  - `sites/nereval/scripts/extract_list.js`
  - `pkg/api/server/routes_engine.go`
  - `pkg/services/engineview/workflow_mutation_service.go`
  - `web/src/api/workflowApi.ts`
  - `pkg/metrics/metrics.go`
  - `ops/monitoring/prometheus/rules/scraper.yml`
  - `pkg/runtimeevents/backend.go`
- Inspected current `2026-05-20--book-ocr/` files and found a template/stub project:
  - `cmd/XXX/main.go`
  - `go.mod`
  - `README.md`
- Ran Kagi/Defuddle research for River and copied captured source markdown into:
  - `/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/SCRAPER-JOBS-001--general-purpose-task-workflow-job-system-for-scraper/sources/river/`

### Why

- The design needed to be evidence-backed, not speculative, because the user asked for a clear technical guide for a new intern.
- The existing `scraper` docs and code already explain much of the runtime; the design document connects those facts to a more general task/workflow/job vocabulary.
- River was mentioned as a possible scheduling improvement, so I captured current River documentation before recommending whether and how to use it.

### What worked

- `docmgr ticket create-ticket` created the ticket workspace successfully.
- `docmgr doc add` created both the design doc and diary doc successfully.
- Repository discovery showed `scraper/` has a mature engine, API, UI, metrics, runtime events, and monitoring rules.
- The existing `NEREVAL` workflow provided a concrete fan-out pattern that maps well to page-level OCR fan-out.
- Defuddle successfully captured River docs to local ticket sources.

### What didn't work

- The current working directory's docmgr configuration resolves to the `2026-05-20--book-ocr/ttmp` root even when commands are run from `scraper/`. I accepted that because the user explicitly referenced both `@scraper/` and `@2026-05-20--book-ocr/`, and the active docmgr root was the book OCR workspace.
- `2026-05-20--book-ocr/` does not yet contain a real OCR implementation. Evidence:
  - `cmd/XXX/main.go` is an empty `main` function.
  - `go.mod` still declares `module github.com/go-go-golems/XXX`.
  - `README.md` is the Go Go template ASCII-art README.
- Because of that, the OCR section is a proposed mapping/design rather than an analysis of an existing OCR code path.

### What I learned

- `scraper` is already a durable workflow engine in practice: the generic core is workflow/op/lease/retry/queue/result/artifact, while scraper-specific behavior mostly lives in manifests, JS scripts, submit verbs, and runner choices.
- The biggest production gap for OCR is not initial scheduling; it is long-running task correctness: heartbeats, cooperative cancellation, idempotency, and large artifact storage.
- River is best treated as a possible queue backend below the existing workflow model, not as a complete replacement for workflows, dependencies, artifacts, and user-facing admin surfaces.

### What was tricky to build

- The main tricky part was separating naming from architecture. Many concepts are named `site` or `scraper`, but the code paths are already generic enough for non-scraping workflows. I handled this by creating an explicit mapping table instead of proposing an immediate rename.
- Another tricky part was avoiding an over-eager River recommendation. River has strong queue features, but adopting it would introduce Postgres and duplicate retry/scheduling semantics unless the integration keeps scraper workflow state authoritative.
- The OCR repository being a stub required care: I could not claim evidence for existing OCR behavior, so I framed OCR as a target workflow mapping and called out the observed stub files directly.

### What warrants a second pair of eyes

- Whether the ticket should live in the book OCR docmgr root or be recreated/mirrored in `scraper/ttmp`, since the active docmgr config pointed to book OCR.
- Whether the proposed `site` -> `workflow package/domain` naming is the desired long-term vocabulary.
- Whether the first OCR implementation should be an in-repo `sites/bookocr` package or an external manifest directory loaded by `--sites-manifest-dir`.
- Whether River should remain a phase-4 option or be pulled forward if the intended production environment already assumes Postgres and multiple worker hosts.

### What should be done in the future

- Validate the docmgr ticket and fix any doctor/frontmatter issues.
- Upload the final design bundle to reMarkable.
- If implementation begins, start with OCR as a normal scraper package/site before doing broad generic renames.
- Add lease heartbeat and cancellation design details to code-level tickets before running large OCR jobs.

### Code review instructions

- Start with the design doc at:
  - `/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/SCRAPER-JOBS-001--general-purpose-task-workflow-job-system-for-scraper/design-doc/01-scraper-task-workflow-job-system-design-and-book-ocr-mapping.md`
- Then inspect the source references listed in its `References` section.
- Validate docmgr with:
  - `cd /home/manuel/workspaces/2026-05-20/book-ocr && docmgr doctor --ticket SCRAPER-JOBS-001 --stale-after 30`
- There are no runtime tests to run because this step only created documentation and copied source references.

### Technical details

Commands run during this step included:

```bash
cd /home/manuel/workspaces/2026-05-20/book-ocr

docmgr ticket create-ticket \
  --ticket SCRAPER-JOBS-001 \
  --title "General Purpose Task Workflow Job System for Scraper" \
  --topics scraper,jobs,ocr

docmgr doc add \
  --ticket SCRAPER-JOBS-001 \
  --doc-type design-doc \
  --title "Scraper Task Workflow Job System Design and Book OCR Mapping"

docmgr doc add \
  --ticket SCRAPER-JOBS-001 \
  --doc-type reference \
  --title "Diary"

defuddle parse https://riverqueue.com/docs --md -o /tmp/.../river.md
```

Core architecture conclusion:

```text
Current scraper engine
  workflow + ops + dependencies + leases + retries + queues + artifacts
      |
      v
Generic workflow/job runtime candidate
      |
      +-- SQLite backend now
      +-- River/Postgres backend later if multi-host/Postgres requirements justify it
      |
      v
Book OCR as proving workload
  ingest -> split -> page OCR fan-out -> cleanup -> aggregate -> export
```

## Step 2: Validate docmgr hygiene and upload to reMarkable

I ran the docmgr doctor check, fixed the one hygiene issue it found, uploaded the design guide plus diary as a reMarkable bundle, and marked the final ticket task complete. This step closed the delivery loop: the ticket now has a validated design document, a chronological diary, source captures for River, and a successful reMarkable upload.

The only issue during validation was that copied Defuddle River source captures were plain Markdown files without docmgr frontmatter or numeric prefixes. I converted those captures into numbered reference files with frontmatter, updated the design document's source references, reran doctor successfully, and then uploaded the bundle.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Validate the documentation package, resolve docmgr hygiene issues, upload the requested deliverables to reMarkable, and record the delivery details.

**Inferred user intent:** The user wants the design package stored in docmgr and available on the reMarkable device for reading/review.

**Commit (code):** N/A — documentation-only ticket work; no runtime code was changed.

### What I did

- Ran `docmgr doctor --ticket SCRAPER-JOBS-001 --stale-after 30`.
- Fixed doctor findings by renaming River captures to numeric filenames and adding frontmatter:
  - `sources/river/01-river-getting-started.md`
  - `sources/river/02-river-scheduled-jobs.md`
  - `sources/river/03-river-periodic-jobs.md`
  - `sources/river/04-river-unique-jobs.md`
  - `sources/river/05-river-job-retries.md`
  - `sources/river/06-river-multiple-queues.md`
- Updated the design document to reference the renamed River source captures.
- Reran `docmgr doctor --ticket SCRAPER-JOBS-001 --stale-after 30` successfully.
- Uploaded the bundle with:
  - `remarquee upload bundle ... --name "SCRAPER JOBS 001 Design Guide" --remote-dir "/ai/2026/05/24/SCRAPER-JOBS-001" --toc-depth 2 --non-interactive`
- Marked the final task complete and updated the changelog.

### Why

- The ticket needed to pass docmgr validation before final handoff.
- The copied River docs are useful evidence, but docmgr expects Markdown files in the ticket workspace to have valid frontmatter and naming hygiene.
- The user explicitly requested upload to reMarkable.

### What worked

- After adding frontmatter and numeric prefixes to River captures, `docmgr doctor` returned:
  - `## Doctor Report (1 findings)`
  - `- ✅ All checks passed`
- reMarkable upload succeeded with:
  - `OK: uploaded SCRAPER JOBS 001 Design Guide.pdf -> /ai/2026/05/24/SCRAPER-JOBS-001`

### What didn't work

- First doctor run failed because River source captures lacked frontmatter. Representative error:
  - `YAML/frontmatter syntax error`
  - `Problem: frontmatter delimiters '---' not found`
  - Files affected: `sources/river/getting-started.md`, `sources/river/job-retries.md`, `sources/river/multiple-queues.md`, `sources/river/periodic-jobs.md`, `sources/river/scheduled-jobs.md`, and `sources/river/unique-jobs.md`.
- I initially attempted `docmgr task check --ticket SCRAPER-JOBS-001 --id 6`, but after removing the placeholder task the final task was renumbered to `5`. The command failed with:
  - `Error: task id(s) not found: [6]`

### What I learned

- Source captures inside docmgr ticket directories should either be frontmatter-valid Markdown documents or be stored in a way that docmgr ignores.
- Removing placeholder tasks changes docmgr task IDs, so it is safer to rerun `docmgr task list` before checking a task after manual task-file edits.

### What was tricky to build

- The validation issue came from treating `sources/river/*.md` as raw source captures while docmgr treated them as managed Markdown docs. The fix was to preserve them as useful ticket sources but make them valid docmgr Markdown files.
- The task ID mismatch was a small sequencing issue caused by manual cleanup of the placeholder task after task creation.

### What warrants a second pair of eyes

- Confirm that the uploaded bundle should include only the primary design guide and diary, not the raw River source captures. I intentionally kept the uploaded PDF focused on the authored deliverables.
- Confirm whether the ticket should be closed now that all tasks are complete, or left active for implementation follow-up.

### What should be done in the future

- If more external source captures are added, use numeric filenames and frontmatter immediately.
- Create implementation tickets for OCR package prototype, heartbeats/cancellation, artifact storage, and optional backend abstraction.

### Code review instructions

- Validate ticket hygiene with:
  - `cd /home/manuel/workspaces/2026-05-20/book-ocr && docmgr doctor --ticket SCRAPER-JOBS-001 --stale-after 30`
- Confirm task state with:
  - `docmgr task list --ticket SCRAPER-JOBS-001`
- Confirm upload from command output:
  - `OK: uploaded SCRAPER JOBS 001 Design Guide.pdf -> /ai/2026/05/24/SCRAPER-JOBS-001`

### Technical details

Final validation and upload commands:

```bash
cd /home/manuel/workspaces/2026-05-20/book-ocr

docmgr doctor --ticket SCRAPER-JOBS-001 --stale-after 30

remarquee upload bundle \
  2026-05-20--book-ocr/ttmp/2026/05/24/SCRAPER-JOBS-001--general-purpose-task-workflow-job-system-for-scraper/design-doc/01-scraper-task-workflow-job-system-design-and-book-ocr-mapping.md \
  2026-05-20--book-ocr/ttmp/2026/05/24/SCRAPER-JOBS-001--general-purpose-task-workflow-job-system-for-scraper/reference/01-diary.md \
  --name "SCRAPER JOBS 001 Design Guide" \
  --remote-dir "/ai/2026/05/24/SCRAPER-JOBS-001" \
  --toc-depth 2 \
  --non-interactive 2>&1
```

## Step 3: Fold concrete AITR-794 OCR evidence back into the design

I studied the AITR-794 OCR ticket under `claw-stuff/ttmp/2026/05/20/AITR-794--clean-markdown-transcription-of-aitr-794-pdf/` and updated the design guide accordingly. This ticket is the concrete implementation evidence that was missing from the first pass: it shows an actual scanned-PDF-to-clean-Markdown workflow implemented with Python scripts, Pinocchio VLM calls, SQLite progress tracking, tmux workers, an ANSI dashboard, prompt iteration, figure/table reprocessing, and final Markdown assembly.

The design direction did not fundamentally change: `scraper` is still a good base, and River should still remain optional. What changed is the OCR model: it should be described as an iterative page-processing campaign with page attempts, prompt versions, page-type classification, selective reprocessing, stale reset/retry operations, and filesystem/object artifacts, not merely as a one-shot page fan-out DAG.

### Prompt Context

**User prompt (verbatim):** "Look at ttmp/2026/05/20/AITR-794--clean-markdown-transcription-of-aitr-794-pdf/ for the concrete OCR actually, it looks like we might not have ported it to go just yet, although you found the design diary. maybe that will give some more insight, and potentially influence the design. Study it, then concisely tell me that you would potentially update, or if things are fine the way they are."

**Assistant interpretation:** Inspect the concrete OCR ticket, determine whether it changes the scraper/job-system design, and update the design document if useful.

**Inferred user intent:** The user wants the design guide to be grounded in the real OCR workflow that already happened, not only in a hypothetical OCR pipeline or the empty Go template repo.

**Commit (code):** N/A — documentation-only update; no runtime code changed.

### What I did

- Read the AITR-794 diary and key scripts:
  - `reference/01-diary.md`
  - `scripts/01-extract-pages.py`
  - `scripts/02-ocr-pages.py`
  - `scripts/03-ocr-batch.py`
  - `scripts/04-ocr-dashboard.py`
  - `scripts/09-build-final.py`
  - `scripts/10-reprocess-universal.py`
  - `scripts/11-postprocess.py`
- Updated the design guide with:
  - concrete AITR-794 OCR findings;
  - prompt/version/page-attempt provenance tables;
  - page-subset reprocessing API sketch;
  - stronger final recommendation that SQLite is fine initially if transactional leases, heartbeat/cancellation, and stale reset are handled;
  - references to the concrete AITR-794 diary and scripts.

### Why

- The first design pass accurately identified the Go OCR repo as a stub, but it did not yet incorporate the real OCR process from `claw-stuff`.
- The real workflow revealed important requirements that a hypothetical DAG glosses over: prompt iteration, partial reruns, dashboard UX, page-level status, and output provenance.

### What worked

- The AITR-794 diary is very detailed and directly actionable.
- The concrete Python scripts map well to scraper concepts:
  - SQLite page queue -> durable ops/leases;
  - tmux workers -> worker processes;
  - dashboard -> frontend/operator API;
  - prompt versions -> workflow package config/artifacts;
  - per-page markdown and figure files -> artifacts/projection records.

### What didn't work

- No runtime code was changed or tested; this was a design-doc refinement only.
- The previous reMarkable upload now lacks this AITR-specific update unless a new bundle is uploaded.

### What I learned

- The real OCR workflow was iterative, not linear. A good production system must make reprocessing first-class.
- The original Python queue hit a duplicate-claim race and fixed it with `BEGIN IMMEDIATE`; this supports keeping scraper's stricter transactional lease path rather than replacing it immediately.
- Prompt design is a core artifact of OCR workflows and should be versioned like code.

### What was tricky to build

- The key nuance was not overcorrecting the original design. The AITR evidence does not imply “use River now”; it implies “model OCR attempts and operator reprocessing explicitly.”
- The concrete workflow has many ad-hoc script phases; the design needed to extract reusable runtime concepts without encoding every one-off script as a permanent product feature.

### What warrants a second pair of eyes

- Whether `page_attempts` and `prompts` should live in the OCR projection DB only, or whether prompt/attempt provenance should become a generic engine concept.
- Whether page-subset reprocessing should be a generic workflow mutation API or an OCR-specific service first.

### What should be done in the future

- If this ticket is used for implementation, start from the AITR scripts as executable requirements.
- Consider uploading a refreshed reMarkable bundle after this design-doc update.

### Code review instructions

- Review the updated sections in the design guide:
  - `Executive summary`
  - `Observed scope boundaries`
  - `Concrete AITR-794 OCR findings that affect the design`
  - `Operators need page-subset reprocessing controls`
  - `Final recommendation`
- Validate with:
  - `cd /home/manuel/workspaces/2026-05-20/book-ocr && docmgr doctor --ticket SCRAPER-JOBS-001 --stale-after 30`

### Technical details

Key design delta:

```text
Before:
  ingest -> split/render pages -> OCR pages -> cleanup -> aggregate -> export

After AITR-794 evidence:
  ingest -> render pages -> OCR attempt with prompt/version -> dashboard/QA
       -> subset reprocess for blank/image/table/code/error pages
       -> postprocess cross-page breaks -> aggregate/export
```

## Step 4: Design an embeddable River-like Go API for scraper

I added a second design document that studies whether scraper can expose a River-like embedded Go API. The answer is yes: the current runner registry, scheduler, store contract, workflow/op model, metrics, runtime events, and submission service already provide most of the lower-level machinery. What is missing is a public facade with typed `JobArgs`, typed `Worker[T]`, `AddWorker`, `Client.Insert`, `Client.Start`, `Client.Stop`, and workflow builder APIs.

The document also compares scraper and River directly. The main conclusion is that River is a better mature Postgres job queue, but scraper has richer workflow semantics: DAG dependencies, dynamic child op emission, artifacts, structured results, runtime events, JS/package support, local SQLite mode, and workflow/artifact-focused admin UI. If River is used, the design recommends using it as a ready-op queue backend under scraper's workflow model rather than replacing scraper's workflow model.

### Prompt Context

**User prompt (verbatim):** "can we structure scraper to have a similar API than river, where it can easily be imported / embedded into a go application, and thus allow easy registering of job types that are backed by go handlers, the same way river is? What are the advanctages anyway of our system over river / what are the pros and cons, especially if we for example use river itself as a backend / scheduler? 

Create a second document for these analysis and studies and design considerations.

Create  a detailed analysis / design / implementation guide that is for a new intern, explaining all the parts of the system needed to understand what it is, with prose paragraphs and bullet point sand pseudocode and diagrams and api references and file references. It should be very clear and technical. Store in the ticket and the nupload to remarkable."

**Assistant interpretation:** Add a second long-form ticket document that designs a River-like embeddable Go API for scraper, compares scraper with River, evaluates River as an optional backend, and provides intern-ready implementation guidance.

**Inferred user intent:** The user wants to know whether scraper can become a library-style workflow/job engine, not only a CLI app, and wants a grounded design before implementation.

**Commit (code):** N/A — documentation-only update; no runtime code changed.

### What I did

- Created a second design doc:
  - `/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/SCRAPER-JOBS-001--general-purpose-task-workflow-job-system-for-scraper/design-doc/02-embeddable-scraper-job-api-and-river-backend-design-considerations.md`
- Studied and referenced scraper internals:
  - `pkg/engine/runner/runner.go`
  - `pkg/engine/store/store.go`
  - `pkg/engine/model/types.go`
  - `pkg/engine/scheduler/scheduler.go`
  - `pkg/cmd/worker_runtime.go`
  - `pkg/cmd/runtime_helpers.go`
  - `pkg/services/submission/service.go`
- Wrote proposed APIs and pseudocode for:
  - `JobArgs`
  - `Worker[T]`
  - `HandlerFunc[T]`
  - `Workers` registry
  - `Client.Insert`
  - workflow/DAG builder
  - River backend bridge with `ExecuteOpArgs`
- Related the new doc to source files and updated the changelog.

### Why

- River's biggest ergonomic advantage is its importable Go API. If scraper wants to be a general-purpose job/workflow system, it needs the same library-first developer experience.
- The design needed to explain where scraper is stronger than River so the project does not accidentally discard its unique workflow/artifact/admin capabilities.

### What worked

- Existing scraper internals map cleanly to a typed facade:
  - `runner.Runner` -> typed `Worker[T]` adapter.
  - `runner.Registry` -> `jobs.Workers`.
  - `store.Store` -> backend contract.
  - scheduler loop -> `Client.Start` / `RunOnce`.
  - `model.OpSpec` -> internal representation of inserted typed jobs.
- River docs already captured in ticket sources were enough to compare API shape and backend features.

### What didn't work

- This is still a design-only step. No `pkg/jobs` code exists yet.
- Current scheduler execution needs a second look before promising River-like concurrent worker behavior; current `RunOnce` leases and executes ops synchronously inside the cycle.

### What I learned

- Scraper can emulate River's ergonomics without becoming River.
- The best first implementation is a small typed worker adapter over `runner.Runner`, not a broad scheduler rewrite.
- River is most useful as an optional backend for Postgres/multi-host queue execution, while scraper should remain authoritative for workflow graph, results, artifacts, and UI.

### What was tricky to build

- The main design trap is duplicating semantics if River is used: retry, uniqueness, queue state, and job status can exist both in River and scraper. The document resolves this by recommending River `MaxAttempts: 1` initially and River uniqueness only for `{workflowID, opID}` bridge jobs.
- Another tricky point is preserving scraper's DAG workflow model while adding a simple one-job `Insert` API. The proposed solution is that `Insert` creates a one-op workflow by default, while `InsertWorkflow` and `Job.Emit` expose the richer model.

### What warrants a second pair of eyes

- Whether the public package should be named `jobs`, `workflow`, `task`, or something else.
- Whether `InsertTx` must be part of v1 or can wait for a Postgres backend.
- Whether River backend should live in the main module or an optional submodule/build tag.
- Whether typed Go workers should be allowed inside manifest-loaded workflow packages or only registered by host Go applications.

### What should be done in the future

- Implement a first small PR with `pkg/jobs.JobArgs`, `Worker[T]`, `HandlerFunc[T]`, `Workers`, and a typed runner adapter.
- Then add `Client.Insert` for one-op workflows using SQLite.
- Then add `Client.Start`/`RunOnce` over the current scheduler.
- Only after those are stable, spike a River backend.

### Code review instructions

- Start with the new document's sections:
  - `Current scraper pieces that map to a library API`
  - `Proposed scraper embeddable API`
  - `Advantages of scraper over River`
  - `Pros and cons if scraper uses River as backend`
  - `Implementation plan`
- Validate docs with:
  - `cd /home/manuel/workspaces/2026-05-20/book-ocr && docmgr doctor --ticket SCRAPER-JOBS-001 --stale-after 30`

### Technical details

Core API shape proposed:

```go
type JobArgs interface { Kind() string }
type Worker[T JobArgs] interface { Work(context.Context, *Job[T]) error }
type HandlerFunc[T JobArgs] func(context.Context, *Job[T]) error

workers := jobs.NewWorkers()
jobs.AddWorker(workers, &OCRPageWorker{})

client, _ := jobs.NewClient(jobs.SQLite("state/engine.db"), &jobs.Config{Workers: workers})
_, _ = client.Insert(ctx, OCRPageArgs{BookID: "aitr-794", Page: 47}, &jobs.InsertOpts{Queue: "ocr:vlm"})
_ = client.Start(ctx)
```

## Step 5: Rewrite the embeddable API document around the workflow-native model

I rewrote the second design document from the ground up after the user clarified that the API should fit our model rather than imitate another job queue API. The new document removes the queue-library framing and centers the runtime's own concepts: `Runtime`, `WorkflowPackage`, `Run`, `Step`, `Executor`, `ArtifactStore`, `ProjectionStore`, `EventSink`, and operator actions.

The rewritten design now treats typed Go execution as one executor style inside a broader durable workflow runtime. It foregrounds the capabilities that make this system distinct: dynamic DAG emission, per-step artifacts, domain projections, runtime events, reusable HTTP/operator APIs, and human-in-the-loop controls such as reprocess selectors.

### Prompt Context

**User prompt (verbatim):** "perfect, update the document from the ground up, no need to really mention river really."

**Assistant interpretation:** Rewrite the second design document so it is no longer framed around River-like API design and instead describes a native embeddable workflow runtime API.

**Inferred user intent:** The user wants the design to guide the project's own API identity and future naming, not position the system as a clone of another queue library.

**Commit (code):** N/A — documentation-only update; no runtime code changed.

### What I did

- Renamed the second design file from:
  - `02-embeddable-scraper-job-api-and-river-backend-design-considerations.md`
- To:
  - `02-embeddable-workflow-runtime-api-design.md`
- Rewrote the document with a new title and frontmatter:
  - `Embeddable Workflow Runtime API Design`
- Removed the queue-library API framing.
- Added a workflow-native API design around:
  - `Runtime`
  - `WorkflowPackage`
  - `Entrypoint`
  - `RunBuilder`
  - `StepContext`
  - `Executor`
  - `ArtifactStore`
  - `ProjectionStore`
  - `EventSink`
  - operator APIs and step selectors
- Related the rewritten doc to current scraper source files.
- Updated the changelog and ran `docmgr doctor`.

### Why

- The previous document answered the earlier prompt, but it over-indexed on matching a job queue API shape.
- The future system's API should expose durable workflow runs, artifacts, projections, events, and operator controls as first-class concepts.

### What worked

- The current internals still map cleanly to the new API:
  - `WorkflowRun` -> `Run`
  - `OpSpec` -> `Step`
  - `Runner` -> `Executor`
  - `site.yaml`/site package -> `WorkflowPackage`
  - site DB -> `ProjectionStore`
  - runtime events -> `EventSink`
- `docmgr doctor --ticket SCRAPER-JOBS-001 --stale-after 30` passed after the rewrite.

### What didn't work

- The file rename means any old external reference to the previous filename is stale.
- The reMarkable upload from the previous step still contains the old framing unless a refreshed bundle is uploaded.

### What I learned

- The clearest public API starts from `Runtime.StartRun` and `RegisterPackage`, not one-off job insertion.
- Typed inputs are still valuable, but they should be attached to workflow steps and executors rather than become the top-level conceptual model.

### What was tricky to build

- The main challenge was keeping the design concrete while not falling back into queue terminology. I resolved this by making the API examples use book OCR and by making artifacts/projections/operator actions appear in the core examples.

### What warrants a second pair of eyes

- Naming: `workflow`, `runtime`, `flows`, or another package name.
- Whether `WorkflowPackage` is the right future replacement for `site`.
- Whether `Step` or `Task` should replace `op` in user-facing APIs.

### What should be done in the future

- Upload a refreshed reMarkable bundle if the user wants the latest version there.
- Start implementation with a small `Executor` facade over `runner.Runner`.

### Code review instructions

- Review the rewritten design doc at:
  - `/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/SCRAPER-JOBS-001--general-purpose-task-workflow-job-system-for-scraper/design-doc/02-embeddable-workflow-runtime-api-design.md`
- Validate with:
  - `cd /home/manuel/workspaces/2026-05-20/book-ocr && docmgr doctor --ticket SCRAPER-JOBS-001 --stale-after 30`

### Technical details

Core direction after rewrite:

```text
Runtime
  registers WorkflowPackages
  starts Runs
  dispatches Steps to Executors
  stores Artifacts
  updates Projections
  emits Events
  exposes Operator controls
```

## Step 6: Implement Phase 1 workflow executor facade

I implemented the first code phase in `scraper`: a small `pkg/workflow` facade that introduces workflow-native executor names without changing the scheduler/store internals. This is intentionally narrow. It proves that the new public model can wrap the existing `runner.Runner` interface while giving future executors a friendlier `StepContext` with typed input decoding, result data, records, artifacts, emitted child steps, and stable error classification.

This step also fixed a local validation issue: the checked-in `go.work` referenced `../go-go-goja`, which does not exist in this workspace. I removed that workspace entry and upgraded `github.com/go-go-golems/go-go-goja` to `v0.4.16`, the released version that contains the `database.QueryExecer` API used by current scraper packages. After that, the full scraper test suite passed with normal `go test ./...`.

### Prompt Context

**User prompt (verbatim):** "Add detailed phases + tasks (if necessary) and implement them step by step, commit at appropriate intervals, keep a detailed diary as you work"

**Assistant interpretation:** Turn the workflow-runtime API design into phased implementation tasks, start with the first safe implementation phase, test it, commit at appropriate boundaries, and keep the ticket diary current.

**Inferred user intent:** The user wants incremental implementation rather than just more design, with reviewable commits and a continuation-friendly diary.

**Commit (code):** `bc6baa26d8fb7eb3a78b8d9e32ab544ee6deaf43` — "Add workflow executor facade"

### What I did

- Added ticket implementation tasks for:
  - Phase 1 executor facade;
  - Phase 1 tests;
  - Phase 2 runtime skeleton;
  - Phase 3 artifact/projection/operator APIs.
- Added `scraper/pkg/workflow/errors.go`:
  - `Error` type implementing the scheduler's `OpError()` carrier pattern;
  - `Retryable(code, err)`;
  - `Permanent(code, err)`.
- Added `scraper/pkg/workflow/context.go`:
  - `StepContext` wrapping `runner.RunContext`;
  - input decoding;
  - result data;
  - record writes;
  - artifact writes;
  - dynamic child-step emission.
- Added `scraper/pkg/workflow/executor.go`:
  - `Executor`;
  - `ExecutorFunc`;
  - `TypedExecutor[I]`;
  - `TypedExecutorFunc[I]`;
  - `NewExecutor`;
  - `NewTypedExecutor`;
  - `ToRunner` adapter;
  - workflow-native executor `Registry` backed by `runner.Registry`.
- Added `scraper/pkg/workflow/executor_test.go` covering:
  - typed input decoding;
  - result assembly;
  - record writes;
  - artifact writes;
  - dynamic child emission;
  - duplicate executor kind rejection;
  - retryable error classification;
  - input decode errors.
- Updated `scraper/go.work` to remove the missing `../go-go-goja` workspace entry.
- Updated `scraper/go.mod` / `go.sum` via `go get github.com/go-go-golems/go-go-goja@v0.4.16` and `go mod tidy`.
- Committed the scraper code as `bc6baa26d8fb7eb3a78b8d9e32ab544ee6deaf43`.

### Why

- The design recommends the first implementation be a small facade over the existing runner interface. This proves the naming and context API without destabilizing scheduler or store code.
- `StepContext` is the bridge from low-level `OpResult` assembly to a future public executor API.
- The dependency/workspace update was necessary because the repository could not compile in this workspace while `go.work` referenced a missing local module.

### What worked

- `pkg/workflow` can adapt typed executors into existing `runner.Runner` implementations.
- Tests confirm that an executor can decode input, write results/records/artifacts, and emit child steps that the existing store completion path can persist.
- Full test suite passed:
  - `go test ./... -count=1`

### What didn't work

- Initial validation failed with:
  - `go: cannot load module ../go-go-goja listed in go.work file: open ../go-go-goja/go.mod: no such file or directory`
- Running with `GOWORK=off` then revealed that `go-go-goja v0.4.5` did not expose the `database.QueryExecer` API expected by current scraper code:
  - `undefined: databasemod.QueryExecer`
  - `undefined: databasemod.New`
  - `undefined: databasemod.WithName`
  - `undefined: databasemod.WithPreconfiguredDB`
- Updating to `go-go-goja v0.4.16` fixed those compile errors.

### What I learned

- The existing `runner.Runner` interface is already a good internal seam for the future `Executor` API.
- `StepContext` can remain lightweight in Phase 1 while still demonstrating the important public capabilities: typed input, artifacts, records, dynamic emission, and error classification.
- The current workspace assumed a sibling local `go-go-goja` checkout that is not present; relying on the released module makes normal validation more reproducible.

### What was tricky to build

- The API needed to be future-facing without prematurely renaming all internals. I kept `pkg/workflow` as the new public facade and converted to/from `model.OpSpec` and `model.OpResult` internally.
- Dynamic child emission needed careful defaults: same workflow ID, same site/package, parent set to current step, JSON input marshaling, and emitted IDs collected in the final result.
- Error classification needed to match the existing scheduler behavior. The scheduler detects errors with an `OpError() model.OpError` method, so `workflow.Error` implements that shape.

### What warrants a second pair of eyes

- Whether `StepContext.Artifact` should return a richer `ArtifactRef` instead of `model.ArtifactID` before the API becomes public.
- Whether `StepOpts.Kind` should be a positional argument to `Emit` rather than an option field.
- Whether `Workflow()`, `Step()`, and `Lease()` should return public wrapper types instead of current model types before external consumers use the package.
- Whether upgrading `go-go-goja` and removing `../go-go-goja` from `go.work` is acceptable for everyone using this workspace.

### What should be done in the future

- Phase 2: add `Runtime` skeleton, package registration, SQLite store construction, `StartRun`, and `RunOnce` over the existing scheduler.
- Add `StepContext.Dependency` and richer artifact/projection APIs.
- Decide public names before broad adoption: `workflow`, `Runtime`, `Step`, `Executor`, `WorkflowPackage`.

### Code review instructions

- Start with:
  - `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/workflow/executor.go`
  - `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/workflow/context.go`
  - `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/workflow/errors.go`
- Then read tests:
  - `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/workflow/executor_test.go`
- Validate with:
  - `cd /home/manuel/workspaces/2026-05-20/book-ocr/scraper && go test ./... -count=1`

### Technical details

Core adapter shape:

```go
type Executor interface {
    Kind() string
    Execute(ctx context.Context, step *StepContext) error
}

func ToRunner(executor Executor) runner.Runner
```

Typed executor shape:

```go
workflow.NewTypedExecutor("ocr/vlm-page", func(ctx context.Context, step *workflow.StepContext, input OCRPageInput) error {
    // write result, artifacts, records, emitted child steps
    return nil
})
```
