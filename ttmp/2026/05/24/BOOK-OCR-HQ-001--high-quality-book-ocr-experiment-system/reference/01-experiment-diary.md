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
    - Path: 2026-05-20--book-ocr/ttmp/2026/05/24/BOOK-OCR-HQ-001--high-quality-book-ocr-experiment-system/experiments/001-baseline-single-page/manifest.yaml
      Note: Baseline experiment configuration
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
