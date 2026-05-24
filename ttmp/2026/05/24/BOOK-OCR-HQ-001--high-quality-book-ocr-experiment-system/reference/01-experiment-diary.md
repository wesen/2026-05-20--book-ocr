---
Title: ""
Ticket: ""
Status: ""
Topics: []
DocType: ""
Intent: ""
Owners: []
RelatedFiles:
    - Path: ../../../../../../../../../../code/wesen/go-go-golems/go-go-parc/Projects/2026/05/24/ARTICLE - Building Book OCR on Scraper Job System - Workflow Runtime Deep Dive.md
      Note: Obsidian project report created before experiment ticket
    - Path: 2026-05-20--book-ocr/ttmp/2026/05/24/BOOK-OCR-HQ-001--high-quality-book-ocr-experiment-system/experiments/001-baseline-single-page/logs/01-run-clean-registry-summary.md
      Note: Compact summary of successful 30-page provider run
    - Path: 2026-05-20--book-ocr/ttmp/2026/05/24/BOOK-OCR-HQ-001--high-quality-book-ocr-experiment-system/experiments/001-baseline-single-page/manifest.yaml
      Note: Baseline experiment configuration
    - Path: 2026-05-20--book-ocr/ttmp/2026/05/24/BOOK-OCR-HQ-001--high-quality-book-ocr-experiment-system/experiments/001-baseline-single-page/outputs/01-final-baseline-clean.md
      Note: Successful baseline 30-page markdown artifact
    - Path: 2026-05-20--book-ocr/ttmp/2026/05/24/BOOK-OCR-HQ-001--high-quality-book-ocr-experiment-system/experiments/002-quality-v2-targeted/notes.md
      Note: Experiment 002 quality assessment and comparison notes
    - Path: 2026-05-20--book-ocr/ttmp/2026/05/24/BOOK-OCR-HQ-001--high-quality-book-ocr-experiment-system/experiments/007-quality-v4-mini-pages-001-030/notes.md
      Note: QA notes and experiment decision for current best run
    - Path: 2026-05-20--book-ocr/ttmp/2026/05/24/BOOK-OCR-HQ-001--high-quality-book-ocr-experiment-system/experiments/007-quality-v4-mini-pages-001-030/outputs/01-final-quality-v4-mini-pages-001-030.md
      Note: Best-current first 30 page OCR output
    - Path: 2026-05-20--book-ocr/ttmp/2026/05/24/BOOK-OCR-HQ-001--high-quality-book-ocr-experiment-system/experiments/008-deterministic-continuity-cleanup/notes.md
      Note: Experiment 008 QA and cleanup notes
    - Path: 2026-05-20--book-ocr/ttmp/2026/05/24/BOOK-OCR-HQ-001--high-quality-book-ocr-experiment-system/experiments/008-deterministic-continuity-cleanup/outputs/02-final-quality-v4-mini-pages-001-030-normalized.md
      Note: Normalized review artifact
    - Path: 2026-05-20--book-ocr/ttmp/2026/05/24/BOOK-OCR-HQ-001--high-quality-book-ocr-experiment-system/scripts/01-filter-ndjson-log-to-sqlite.py
      Note: SQLite log filtering script for noisy provider/SSE logs
    - Path: 2026-05-20--book-ocr/ttmp/2026/05/24/BOOK-OCR-HQ-001--high-quality-book-ocr-experiment-system/scripts/02-run-ocr-capture-log.py
      Note: Direct process-output-to-SQLite capture for noisy provider runs
    - Path: 2026-05-20--book-ocr/ttmp/2026/05/24/BOOK-OCR-HQ-001--high-quality-book-ocr-experiment-system/scripts/03-qa-ocr-markdown.py
      Note: Automated QA checks for OCR markdown
    - Path: 2026-05-20--book-ocr/ttmp/2026/05/24/BOOK-OCR-HQ-001--high-quality-book-ocr-experiment-system/scripts/04-normalize-ocr-markdown.py
      Note: Deterministic list-page continuity cleanup
    - Path: scraper/cmd/ocr-mvp/main.go
      Note: Added prompt-version and log-level CLI controls
    - Path: scraper/pkg/workflows/ocrmvp/prompt.go
      Note: |-
        Added ocr-quality-v2 prompt version
        Added v3 list-diplomatic and v4 Report 794 lexicon prompts
ExternalSources: []
Summary: ""
LastUpdated: 0001-01-01T00:00:00Z
WhatFor: ""
WhenToUse: ""
---






# Diary

## Goal

This diary records the high-quality book OCR experiment process for `BOOK-OCR-HQ-001`: setup, prompt experiments, model/profile choices, commands, failures, review findings, and decisions that lead toward a high-quality first-30-page result.

## Step 1: Create the high-quality OCR ticket and experiment plan

I created a new docmgr ticket for the quality-focused OCR work and set it up as an experiment lab rather than a single implementation task. The ticket contains a design guide, this diary, and experiment folders for baseline OCR, context-window OCR, and chunk-level continuity/style passes.

This step intentionally separates the quality problem from the already-completed OCR MVP plumbing. The workflow runtime can now process pages; this ticket asks how to make the output good enough to trust.

### Prompt Context

**User prompt (verbatim):** "Actually, before we start, let's write a project report in our parc obsidian vault, as a deep dive technical blog post in building a OCR setup on top of the scraper job system, in full detail. Use a textbook writing style. Commit and push when done.

Then, continue with :

                                                                                                   
 ok, so make a new ticket to create a high quality book OCR setup.                                   
                                                                                                     
 Your goal is to iterate on prompts and chunking and how to ensure proper continuinty in style and   
 such across many pages, extracting figures, etc... Start small, keep track of all your experiments  
 in a diary and in proper folders so I can inspect later, and try to get to a high quality result on 
 the first 30 pages. Everything is fair game, prompting, chunking, jobs, how many LLM calls (you can 
 use geppetto multiturn setup, register tools, do all kinds of stuff). Use logging to get feedback   
 etc...                                                                                              
                                                                                                     
 Create  a detailed analysis / design / implementation guide that is for a new intern, explaining    
 all the parts of the system needed to understand what it is, with prose paragraphs and bullet point 
 sand pseudocode and diagrams and api references and file references. It should be very clear and    
 technical. Store in the ticket and the nupload to remarkable.                                       
                                                                                                     
 Then get started, I'm going to do groceries and chores and I hope to return to a high quality       
 result. Use the gpt-5-mini-low or gpt-5-nano-low profiles, I think that should be enough for        
 starters."

**Assistant interpretation:** First write and push a textbook-style Obsidian project report about the OCR-on-scraper work, then create a new docmgr ticket for high-quality OCR experiments, document the system for an intern, upload the docs, and start running experiments on the first 30 pages.

**Inferred user intent:** The user wants both durable narrative documentation and a concrete autonomous experimentation loop that can improve OCR quality while preserving every artifact for later inspection.

### What I did

- Wrote an Obsidian article in the parc vault:
  - `/home/manuel/code/wesen/go-go-golems/go-go-parc/Projects/2026/05/24/ARTICLE - Building Book OCR on Scraper Job System - Workflow Runtime Deep Dive.md`
- Committed and pushed the Obsidian article:
  - `2a1d0c0` — `Article: document scraper OCR workflow runtime`
- Created docmgr ticket `BOOK-OCR-HQ-001`.
- Created design doc:
  - `design-doc/01-high-quality-book-ocr-experiment-system.md`
- Created diary:
  - `reference/01-experiment-diary.md`
- Created experiment folders:
  - `experiments/001-baseline-single-page/`
  - `experiments/002-context-window/`
  - `experiments/003-chunk-style-pass/`
- Added a baseline experiment manifest and prompt reference.
- Replaced the default task list with concrete phased experiment tasks.

### Why

- The previous OCR MVP ticket proved the runtime and provider path; this ticket needs an experiment structure to improve output quality.
- The user explicitly asked for inspectable folders and a diary so that prompt/chunking experiments can be audited later.

### What worked

- Ticket creation succeeded.
- The design guide now explains the runtime, current OCR package, experiment directory contract, prompt iteration strategy, and phased quality plan.
- The Obsidian report was committed and pushed before ticket work began.

### What didn't work

- N/A in this setup step.

### What I learned

- The correct next unit of work is a quality lab: manifests, logs, outputs, and review notes before deeper code changes.

### What was tricky to build

- The ticket needs to balance two goals: being detailed enough for an intern to understand the system, while also staying practical enough to drive experiments immediately.

### What warrants a second pair of eyes

- Whether the blank-page policy should be empty output or explicit `[BLANK PAGE]` during experiments.
- Whether `gpt-5-nano-low` should be the baseline or whether `gpt-5-mini-low` should be used immediately for better quality.

### What should be done in the future

- Upload the initial design guide to reMarkable.
- Run baseline pages 1-30 with a real provider profile.
- Copy logs, projections, artifacts, and final markdown into the experiment folder.

### Code review instructions

- Review the design guide first:
  - `/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/BOOK-OCR-HQ-001--high-quality-book-ocr-experiment-system/design-doc/01-high-quality-book-ocr-experiment-system.md`
- Review the baseline manifest:
  - `/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/BOOK-OCR-HQ-001--high-quality-book-ocr-experiment-system/experiments/001-baseline-single-page/manifest.yaml`

### Technical details

Important commands:

```bash
cd /home/manuel/workspaces/2026-05-20/book-ocr

docmgr ticket create-ticket \
  --ticket BOOK-OCR-HQ-001 \
  --title "High quality book OCR experiment system" \
  --topics ocr,workflow,experiments,book-processing,implementation-guide

docmgr doc add --ticket BOOK-OCR-HQ-001 \
  --doc-type design-doc \
  --title "High Quality Book OCR Experiment System"

docmgr doc add --ticket BOOK-OCR-HQ-001 \
  --doc-type reference \
  --title "Experiment Diary"
```

## Step 2: Upload the initial design guide to reMarkable

I uploaded the initial high-quality OCR design guide and experiment diary to reMarkable before starting the long-running provider experiment. This gives the user a readable design bundle immediately, even while experiments continue to produce artifacts.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Publish the new ticket's intern-oriented design guide and diary to reMarkable before running experiments.

**Inferred user intent:** The user wants durable documentation that can be read away from the terminal while the OCR experiments run.

### What I did

- Ran a dry-run bundle upload.
- Uploaded the real bundle:
  - `BOOK OCR HQ 001 Design Guide.pdf`
  - `/ai/2026/05/24/BOOK-OCR-HQ-001`

### Why

- The user explicitly requested that the design guide be stored in the ticket and uploaded to reMarkable.

### What worked

- Dry-run succeeded.
- Real upload succeeded:
  - `OK: uploaded BOOK OCR HQ 001 Design Guide.pdf -> /ai/2026/05/24/BOOK-OCR-HQ-001`

### What didn't work

- Initial `docmgr doctor` found missing topic vocabulary for `book-processing` and `experiments`, and a prompt note without frontmatter.
- I fixed this by adding topic vocabulary entries and replacing the prompt note with `prompts/01-page-prompt.md` with frontmatter.

### What I learned

- Experiment prompt notes under the ticket are checked by `docmgr doctor`, so even small markdown notes should have frontmatter and numeric prefixes.

### What was tricky to build

- The experiment folders are not standard docmgr doc-type folders, but `doctor` still sees markdown inside them. The safest pattern is to give every experiment markdown file valid frontmatter.

### What warrants a second pair of eyes

- Whether experiment subfolders should stay directly under the ticket root or move under `reference/` if docmgr hygiene becomes noisy.

### What should be done in the future

- Upload a refreshed bundle after the first 30-page experiment produces QA notes and a best-current artifact.

### Code review instructions

- Verify the uploaded bundle includes:
  - `design-doc/01-high-quality-book-ocr-experiment-system.md`
  - `reference/01-experiment-diary.md`

### Technical details

Upload command:

```bash
remarquee upload bundle \
  design-doc/01-high-quality-book-ocr-experiment-system.md \
  reference/01-experiment-diary.md \
  --name "BOOK OCR HQ 001 Design Guide" \
  --remote-dir "/ai/2026/05/24/BOOK-OCR-HQ-001" \
  --toc-depth 2 \
  --non-interactive
```

## Step 3: Add SQLite log filtering for noisy provider traces

The first 30-page provider run produced very noisy logs because Geppetto/OpenAI Responses emits trace-level SSE delta rows for streaming tokens. I added a small log-filtering script that loads NDJSON log lines into SQLite and writes compact markdown summaries. This keeps the full log for forensic debugging while making normal review use SQL queries and small summary files.

### Prompt Context

**User prompt (verbatim):** "use some sqlite filtering of log outputbecause things are pretty big, especially with SSE traces"

**Assistant interpretation:** Stop dumping huge provider trace logs into the conversation; preserve them on disk but create SQLite-backed filtered summaries for inspection.

**Inferred user intent:** The user wants long-running OCR experiments to remain inspectable without flooding the terminal or conversation with SSE token deltas.

### What I did

- Added `scripts/01-filter-ndjson-log-to-sqlite.py`.
- Loaded the failed duplicate-profile log into:
  - `experiments/001-baseline-single-page/logs/run-failed-duplicate-profile.sqlite`
- Loaded the successful clean-registry log into:
  - `experiments/001-baseline-single-page/logs/run-clean-registry.sqlite`
- Wrote compact markdown summaries:
  - `logs/02-run-failed-duplicate-profile-summary.md`
  - `logs/01-run-clean-registry-summary.md`
- Copied the successful final 30-page baseline artifact into:
  - `outputs/01-final-baseline-clean.md`
- Exported page projection rows and a compact non-trace timeline:
  - `outputs/pages-clean.tsv`
  - `outputs/timeline-clean.tsv`
- Added `experiments/001-baseline-single-page/notes.md`.

### Why

- The successful run log had 8687 lines, including 8443 trace-level SSE delta rows. The useful workflow timeline is only 69 non-trace workflow events.
- SQLite keeps the full raw logs queryable while allowing compact summaries and filtered timelines.

### What worked

- The clean-registry baseline run succeeded on pages 1-30:
  - run ID: `ocr-mvp-593bf5b6-19c6-4c8c-b631-b48a2d1aba78`
  - final markdown char count: 43857
- The log-filter summary reports:
  - total lines: 8687
  - trace lines: 8443
  - non-trace workflow events: 69
  - warning/error/failure rows: 0

### What didn't work

- The first baseline attempt failed before provider calls because the default Pinocchio profile registry has duplicate `gpt-5-nano-low` keys:
  - `yaml: unmarshal errors: line 278: mapping key "gpt-5-nano-low" already defined at line 181`
- I worked around it by creating `/tmp/book-ocr-hq-001/profiles-clean.yaml` with only `openai-responses-base`, `gpt-5-mini-low`, and one `gpt-5-nano-low` profile, then passing it via `--profile-registries`.

### What I learned

- The runtime and provider path can complete 30 pages with `gpt-5-nano-low`, but log handling must be filtered from the start.
- The local Pinocchio registry needs cleanup or all repeatable experiments should pass a clean registry file explicitly.

### What was tricky to build

- The raw logs mix JSON logs, status lines, and huge numbers of trace deltas. The script stores both parseable JSON rows and raw unparsed lines so no evidence is lost.

### What warrants a second pair of eyes

- The clean temporary profile registry was generated from the local profile file and includes API settings; it should stay under `/tmp`, not be committed.
- The duplicate profile entry in `~/.config/pinocchio/profiles.yaml` should be cleaned up separately, with care around secrets.

### What should be done in the future

- Add a CLI logging-level flag or environment configuration so trace logs can be suppressed at source.
- Use the SQLite summaries as the default inspection path for future experiments.
- Start manual quality review of `outputs/01-final-baseline-clean.md`.

### Code review instructions

- Review script:
  - `/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/BOOK-OCR-HQ-001--high-quality-book-ocr-experiment-system/scripts/01-filter-ndjson-log-to-sqlite.py`
- Review compact outputs:
  - `experiments/001-baseline-single-page/logs/01-run-clean-registry-summary.md`
  - `experiments/001-baseline-single-page/outputs/timeline-clean.tsv`
  - `experiments/001-baseline-single-page/outputs/pages-clean.tsv`
  - `experiments/001-baseline-single-page/outputs/01-final-baseline-clean.md`

### Technical details

Useful query:

```sql
select line_no, time, level, event, op_id, attempt, message
from log_events
where coalesce(event, '') != '' and level != 'trace'
order by line_no;
```

## Step 4: Assess baseline quality, add quality-v2 prompt, and stop log blast at source

I reviewed the baseline output myself and used the vision tool to validate the front-matter/list-page failures on pages 6-9. The baseline was technically successful, but it had consistency problems: table-of-contents continuation pages changed style, table-of-figures continuation pages changed style, blank pages emitted prose, and list pages drifted between markdown bullets, headings, and plain lines.

I implemented a new `ocr-quality-v2` prompt version and exposed `--prompt-version` through the CLI. After the next run still produced too much provider trace output, I also added `--log-level` to the OCR CLI and a direct stdout-to-SQLite capture script so future long-running runs do not blast SSE deltas into the terminal.

### Prompt Context

**User prompt (verbatim):** "ok, you do the quality assessment yourself, you can read the results, you can look at the pages too. 

Iterate, keep a detailed diary, make this stellar"

**Assistant interpretation:** Take ownership of reviewing OCR quality, use both text output and page images, iterate prompts, preserve evidence, and keep committing progress.

**Inferred user intent:** The user wants an autonomous quality loop with readable diary entries and commit boundaries, not just infrastructure setup.

**Commit (code):** `65b56d50bbb95778efd436dab08615686731f6b9fdbf99b40` — "Add OCR quality prompt version"

**Commit (code):** `c53cd92176dabeb366afdd0de56d1b9fdbf99b40` — "Add OCR quality prompt version" (CLI `--log-level` follow-up)

### What I did

- Reviewed baseline pages 1-30 summaries and inspected the baseline markdown.
- Used the vision tool on page images 6-9 to validate Table of Contents and Table of Figures formatting failures.
- Added `ocr-quality-v2` in `scraper/pkg/workflows/ocrmvp/prompt.go`.
- Added CLI support for `--prompt-version`.
- Added CLI support for `--log-level` using zerolog global level configuration.
- Added `scripts/02-run-ocr-capture-log.py`, which runs commands and stores stdout/stderr lines directly in SQLite while printing only compact non-trace progress.
- Ran Experiment 002 on pages 1-9 with `ocr-quality-v2`.
- Captured Experiment 002 outputs, page projections, timeline, filtered log summary, and notes.

### Why

- The first baseline established the current quality floor; the next step had to target observed failures.
- The user explicitly asked for consistency and prompt iteration.
- The provider emits huge SSE trace logs; logging had to be controlled at source and capturable in SQLite before longer experiments continue.

### What worked

- `ocr-quality-v2` improved several baseline failures:
  - Page 2 now emits `[BLANK PAGE]` instead of prose.
  - Page 8 no longer uses markdown bullets for the Table of Figures.
  - Page 9 keeps the same non-bullet figure-entry style as page 8.
  - Page 4 fixed `Ciccarrelli` to `Ciccarelli`.
  - Title pages no longer become `[IMAGE: ...]` markers.
- `--log-level warn` suppresses noisy zerolog trace/debug/info rows in a smoke run.
- `02-run-ocr-capture-log.py` stores process output in SQLite and prints only compact progress.

### What didn't work

- `ocr-quality-v2` is not good enough yet:
  - Page 1 became too visually literal with title lines split as `Presentation / Based User / Interfaces`.
  - Page 7 duplicated `Chapter Six: Constructing Presentation Systems`.
  - Page 7 still has continuation/list-style inconsistencies.
  - Page 6 may have an incorrect page number for section 1.1 (`9` vs baseline `8`), requiring visual verification.
- The first Experiment 002 run was captured with the older noisy method and produced a large terminal trace before I added `--log-level` and the direct SQLite capture script.

### What I learned

- The first prompt improvement should focus on page-type-specific list policy, not generic OCR accuracy.
- For front matter, a single prompt has to distinguish diplomatic visual transcription from readable markdown normalization.
- The log problem needs two controls:
  - suppress unwanted logs at source with `--log-level`;
  - capture full process output into SQLite when preservation matters.

### What was tricky to build

- The prompt had to be stricter without overfitting to one page. Stronger rules improved blank/list pages but introduced a title-page line-break issue.
- The CLI logging change needed to affect Geppetto/provider logs too, so it uses zerolog's global level.
- Direct capture to SQLite has to preserve unparsed status lines as well as JSON logs, because the OCR CLI emits both.

### What warrants a second pair of eyes

- Decide whether title pages should preserve visual line breaks or normalize to reading text.
- Decide whether continuation pages should include an explicit `[CONTINUED: ...]` marker.
- Verify page 6's section 1.1 page number against the scan.

### What should be done in the future

- Implement `ocr-quality-v3-list-diplomatic` focused on list pages and title-page normalization.
- Use `scripts/02-run-ocr-capture-log.py` plus `--log-level warn` for all future provider runs.
- Run targeted pages 1-9 again, compare against baseline and v2, then expand only if it improves.

### Code review instructions

- Review prompt code:
  - `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/workflows/ocrmvp/prompt.go`
- Review CLI logging:
  - `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/cmd/ocr-mvp/main.go`
- Review capture script:
  - `/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/BOOK-OCR-HQ-001--high-quality-book-ocr-experiment-system/scripts/02-run-ocr-capture-log.py`
- Review experiment notes:
  - `experiments/002-quality-v2-targeted/notes.md`

### Technical details

Experiment 002 command:

```bash
go run ./cmd/ocr-mvp run \
  --book-id presentation-based-uis-hq-002-quality-v2 \
  --image-dir /home/manuel/code/wesen/claw-stuff/output/books/presentation-based-uis/pages \
  --work-dir /tmp/book-ocr-hq-001/002-quality-v2-targeted \
  --start-page 1 \
  --end-page 9 \
  --profile gpt-5-nano-low \
  --profile-registries /tmp/book-ocr-hq-001/profiles-clean.yaml \
  --prompt-version ocr-quality-v2 \
  --max-workers 2
```

Future capture pattern:

```bash
python3 scripts/02-run-ocr-capture-log.py logs/run.sqlite -- \
  go run ./cmd/ocr-mvp run ... --log-level warn
```

## Step 5: Iterate to v4, compare nano versus mini, and produce the current best 30-page OCR

After v2 improved the front matter but still left Table of Contents and Table of Figures problems, I narrowed the next iteration around list pages. The v3 prompt made list pages diplomatic plain text, stopped markdown heading drift, and normalized title-page line wrapping. I then compared `gpt-5-nano-low` with `gpt-5-mini-low` on the hard pages 6-9, because the remaining failures were no longer mostly formatting policy errors; they were visual accuracy errors in page numbers and book-specific terms.

The mini model was visibly better on the list-heavy pages, but still occasionally produced a `DiRed` case regression in a full front-matter run. I added a v4 Report 794 lexicon prompt with the known acronyms and names (`Dired`, `Steamer`, `PSBase`, `PPS`, `Zmacs`, `Xerox Star`, and the title/author forms), then ran the full first 30 pages with `gpt-5-mini-low`. That run is the current best candidate.

### Prompt Context

**User prompt (verbatim):** (same as Step 4)

**Assistant interpretation:** Continue the quality loop autonomously: identify remaining OCR defects, modify prompts/model selection, run targeted and then full-range experiments, preserve evidence, and keep commits at useful intervals.

**Inferred user intent:** Produce a high-quality, inspectable first-30-pages OCR output rather than stopping after a merely successful OCR run.

**Commit (code):** `6e05ee1230411711d1b280f4ac8c1b77ec86e699` — "Add OCR quality v3 list prompt"

**Commit (code):** `8e76419856fc87c865aa0088a4f67db00b673924` — "Add Report 794 OCR lexicon prompt"

### What I did

- Added `ocr-quality-v3-list-diplomatic` to make contents/figures pages plain-text diplomatic lists.
- Ran Experiment 003 on pages 1-9 with `gpt-5-nano-low` and v3.
- Used the vision tool to verify page 002 is not blank and contains the visible pagination sentence.
- Used the vision tool to verify page 006 Table of Contents numbers and page 008 figure entries:
  - `Chapter One: Introduction and Overview` is page 8.
  - `1.1 The Primitive Presentation System Model` is page 9.
  - `Figure 4-1: Dired Model` is page 72.
  - `Figure 4-9: Sample Steamer Schematic` is page 91.
  - `Figure 5-1: PSBase Support of PPS Components` is page 101.
- Ran Experiment 004 on pages 6-9 with `gpt-5-mini-low` and v3.
- Ran Experiment 005 on pages 1-9 with `gpt-5-mini-low` and v3.
- Added `ocr-quality-v4-report794-lexicon` with a small book-specific vocabulary and visible-intentionally-blank-page rule.
- Ran Experiment 006 on pages 6-9 with `gpt-5-mini-low` and v4.
- Ran Experiment 007 on pages 1-30 with `gpt-5-mini-low` and v4.
- Captured Experiment 007 output, projections, timeline, and SQLite process log.
- Used the vision tool on pages 13, 15, and 30 to spot-check figure captions/diagram labels and the Chapter Two opening.

### Why

- v2's remaining failures suggested the prompt needed a stricter list-page policy.
- v3 showed that prompt policy could fix structure, but model size mattered for visual accuracy.
- v4's lexicon was justified because the target is a specific book/report and repeated terms are known from the first pages.
- The full 30-page run was delayed until the hard front-matter pages had a better setting.

### What worked

- `scripts/02-run-ocr-capture-log.py` prevented provider logs from flooding the terminal during all later runs.
- `gpt-5-mini-low` improved page-number and spelling fidelity on list pages compared with `gpt-5-nano-low`.
- v4 fixed the known list-page regressions observed in earlier outputs:
  - no `DiRed`;
  - no `Streamer`;
  - no `PPSBase`;
  - no title-page `[IMAGE: ...]` marker;
  - no incorrect `[BLANK PAGE]` replacement for page 002.
- Experiment 007 completed all 30 pages successfully:
  - run id: `ocr-mvp-4c5c9406-926a-4ecd-a6b2-e8fedba847d8`
  - output: `experiments/007-quality-v4-mini-pages-001-030/outputs/01-final-quality-v4-mini-pages-001-030.md`
  - page markers: 30
  - projection status: 30 `done`
  - known regression grep hits: 0
- Vision spot-check on pages 13, 15, and 30 found the major headings, captions, and diagram labels broadly consistent with the OCR output.

### What didn't work

- v3 with `gpt-5-nano-low` still made visible list-page mistakes:
  - `Dired Model` page number drifted in one run;
  - `Steamer` became `Streamer`;
  - `PSBase` became `PPSBase`.
- v3 with `gpt-5-mini-low` improved targeted pages, but a full front-matter run still produced `DiRed` once.
- v4 still approximates dot-leader lengths; it does not exactly reproduce alignment.
- Single-page OCR still lacks a true continuity pass, so prose line wrapping and figure/list style can vary by page.

### What I learned

- The blank-page rule has to distinguish truly blank pages from intentionally blank pages with visible explanatory text. Page 002 visibly contains `This blank page was inserted to preserve pagination.`
- A targeted prompt can fix structural consistency, but visual fidelity of small list entries benefits from a stronger model.
- For historical technical reports, a small book-specific lexicon is a practical quality lever and should be captured as provenance.

### What was tricky to build

- The main difficulty was separating prompt-policy errors from OCR visual-recognition errors. v2 and v3 fixed policy drift but exposed visual errors in small list entries; switching from nano to mini addressed part of that.
- The lexicon prompt had to avoid overcorrecting `PPS` into `PSBase`: v4 explicitly says `PPS` remains correct for the Primitive Presentation System and `PPSCalc`.
- The intentionally blank page was counterintuitive: an earlier rule treated it as blank, but the page actually has visible text, so v4 now transcribes visible blank-page notices.

### What warrants a second pair of eyes

- Check whether dot leaders should be preserved approximately, normalized away, or reconstructed in a post-processing pass.
- Review pages 10-30 for subtle OCR substitutions that simple grep checks cannot catch.
- Decide whether figure diagrams need structured transcription beyond `[FIGURE: ...]` markers.

### What should be done in the future

- Add a second-pass continuity/style cleanup workflow over pages 1-30.
- Add automated QA checks for known vocabulary, page markers, duplicated headings, and figure/list consistency.
- Consider a targeted re-OCR loop for pages that fail QA rather than rerunning all pages.

### Code review instructions

- Review v3/v4 prompt changes in:
  - `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/workflows/ocrmvp/prompt.go`
- Review full 30-page output and QA notes in:
  - `/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/BOOK-OCR-HQ-001--high-quality-book-ocr-experiment-system/experiments/007-quality-v4-mini-pages-001-030/outputs/01-final-quality-v4-mini-pages-001-030.md`
  - `/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/BOOK-OCR-HQ-001--high-quality-book-ocr-experiment-system/experiments/007-quality-v4-mini-pages-001-030/notes.md`
- Validate with:
  - `go test ./cmd/ocr-mvp ./pkg/workflows/ocrmvp -count=1`
  - `grep -c '^<!-- page:' experiments/007-quality-v4-mini-pages-001-030/outputs/01-final-quality-v4-mini-pages-001-030.md`
  - `grep -nE 'DiRed|Streamer|PPSBase|Ciccarrelli|\[BLANK PAGE\]|\[IMAGE:' experiments/007-quality-v4-mini-pages-001-030/outputs/01-final-quality-v4-mini-pages-001-030.md || true`

### Technical details

Experiment 007 command pattern:

```bash
python3 scripts/02-run-ocr-capture-log.py experiments/007-quality-v4-mini-pages-001-030/logs/run-capture.sqlite -- \
  go run ./cmd/ocr-mvp run \
    --book-id presentation-based-uis-hq-007-v4-mini-30 \
    --image-dir /home/manuel/code/wesen/claw-stuff/output/books/presentation-based-uis/pages \
    --work-dir /tmp/book-ocr-hq-001/007-quality-v4-mini-pages-001-030 \
    --start-page 1 \
    --end-page 30 \
    --profile gpt-5-mini-low \
    --profile-registries /tmp/book-ocr-hq-001/profiles-clean.yaml \
    --prompt-version ocr-quality-v4-report794-lexicon \
    --max-workers 3 \
    --log-level warn
```

## Step 6: Add deterministic QA and a narrow continuity cleanup pass

After Experiment 007 produced the best current first-30-page OCR, I added an auditable second pass that does not call a model. The goal was to turn the quality observations into repeatable checks and make the list pages easier to review without hiding any model output or inventing content.

The result is Experiment 008: a deterministic QA report plus a narrow normalization pass for Table of Contents and Table of Figures dot leaders. The raw v4 mini OCR remains preserved as Experiment 007; Experiment 008 produces a review-oriented normalized copy and a patch showing every change.

### Prompt Context

**User prompt (verbatim):** (same as Step 4)

**Assistant interpretation:** Continue iterating on quality, but shift from single-page OCR prompting to auditable second-pass QA and cleanup.

**Inferred user intent:** Make the output easier to trust and review, with repeatable checks and preserved provenance.

### What I did

- Added `scripts/03-qa-ocr-markdown.py`.
- Added `scripts/04-normalize-ocr-markdown.py`.
- Created Experiment 008 under `experiments/008-deterministic-continuity-cleanup`.
- Ran QA on Experiment 007's raw output.
- Ran deterministic list-page normalization on Experiment 007's raw output.
- Wrote a diff showing the cleanup changes.
- Ran QA again on the normalized output.
- Wrote Experiment 008 notes.

### Why

- Experiment 007 is good enough that broad prompt changes would risk regressions.
- The remaining obvious issue is review ergonomics: irregular dot leaders and the lack of automated repeatable checks.
- A deterministic pass is safer than an LLM rewrite at this stage because every change is diffable.

### What worked

- QA passes before and after cleanup:
  - 30 page markers found.
  - 30 expected page markers.
  - 2 figure markers.
  - no known bad term hits.
  - expected front-matter/list/chapter strings found.
  - no adjacent duplicate non-empty lines found.
  - list pages 006-009 have zero markdown bullet lines and zero markdown heading lines.
- The cleanup produces a normalized review artifact:
  - `experiments/008-deterministic-continuity-cleanup/outputs/02-final-quality-v4-mini-pages-001-030-normalized.md`
- The cleanup diff is preserved:
  - `experiments/008-deterministic-continuity-cleanup/outputs/03-cleanup-diff.patch`

### What didn't work

- The first QA script run failed because I over-escaped the `^[FIGURE:` regex in Python and produced:

```text
re.PatternError: unterminated character set at position 3
```

- I fixed the regex by using `r"^\[FIGURE:"` in a separate assignment before formatting the summary line.
- The first cleanup script version also introduced extra page-boundary whitespace. I tightened the page-body stripping and page-marker assembly, then reran the cleanup and QA.

### What I learned

- Even a small QA script is useful for catching regressions that are easy to miss during manual review.
- The normalized output should not replace raw OCR provenance; it should sit beside it with a diff.
- Deterministic cleanup is appropriate for punctuation/leader/spacing normalization, but not for semantic OCR repair.

### What was tricky to build

- The tricky part was making the cleanup narrow enough. It should improve list-page reviewability without rewriting prose or silently altering content.
- Page-marker whitespace needed care because markdown page boundaries are part of the review contract.
- The QA report had to distinguish automated pass/fail from actual OCR correctness: passing the script does not prove the transcription is perfect.

### What warrants a second pair of eyes

- Review `03-cleanup-diff.patch` to confirm that every normalized dot-leader line is acceptable.
- Decide whether normalized dot leaders (`...`) are preferred over approximate visual alignment in the final deliverable.
- Review whether page-boundary spacing in the normalized artifact is acceptable for downstream publication.

### What should be done in the future

- Add an LLM or rule-based semantic continuity pass only after the deterministic QA baseline is stable.
- Add more expected-string checks for pages 10-30.
- Add figure-specific QA for expected captions and marker counts.

### Code review instructions

- Review scripts:
  - `/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/BOOK-OCR-HQ-001--high-quality-book-ocr-experiment-system/scripts/03-qa-ocr-markdown.py`
  - `/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/BOOK-OCR-HQ-001--high-quality-book-ocr-experiment-system/scripts/04-normalize-ocr-markdown.py`
- Review outputs:
  - `experiments/008-deterministic-continuity-cleanup/outputs/01-qa-before-cleanup.md`
  - `experiments/008-deterministic-continuity-cleanup/outputs/02-final-quality-v4-mini-pages-001-030-normalized.md`
  - `experiments/008-deterministic-continuity-cleanup/outputs/03-cleanup-diff.patch`
  - `experiments/008-deterministic-continuity-cleanup/outputs/04-qa-after-cleanup.md`

### Technical details

Commands:

```bash
python3 scripts/03-qa-ocr-markdown.py \
  experiments/007-quality-v4-mini-pages-001-030/outputs/01-final-quality-v4-mini-pages-001-030.md \
  --out experiments/008-deterministic-continuity-cleanup/outputs/01-qa-before-cleanup.md \
  --expected-pages 30

python3 scripts/04-normalize-ocr-markdown.py \
  experiments/007-quality-v4-mini-pages-001-030/outputs/01-final-quality-v4-mini-pages-001-030.md \
  experiments/008-deterministic-continuity-cleanup/outputs/02-final-quality-v4-mini-pages-001-030-normalized.md \
  --diff experiments/008-deterministic-continuity-cleanup/outputs/03-cleanup-diff.patch

python3 scripts/03-qa-ocr-markdown.py \
  experiments/008-deterministic-continuity-cleanup/outputs/02-final-quality-v4-mini-pages-001-030-normalized.md \
  --out experiments/008-deterministic-continuity-cleanup/outputs/04-qa-after-cleanup.md \
  --expected-pages 30
```
