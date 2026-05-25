---
Title: Diary
Ticket: OCR-QUALITY-WORKERS-001
Status: active
Topics:
    - ocr
    - workflow
    - experiments
    - book-processing
    - implementation-guide
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../../scraper/pkg/workflows/ocrmvp/prompt.go
      Note: Figure-aware OCR prompt version for Step 3
    - Path: ../../../../../../../scraper/pkg/workflows/ocrquality/figures.go
      Note: Caption-only diagram marker recovery for Step 3
    - Path: 2026-05-20--book-ocr/ttmp/2026/05/24/BOOK-OCR-HQ-001--high-quality-book-ocr-experiment-system/analysis/01-final-ocr-quality-report.md
      Note: Source for follow-up tasks
    - Path: 2026-05-20--book-ocr/ttmp/2026/05/24/OCR-QUALITY-WORKERS-001--port-ocr-qa-and-cleanup-scripts-to-go-workflow-workers/experiments/001-go-quality-pass-embedded-figures/notes.md
      Note: Embedded figure experiment notes
    - Path: scraper/pkg/workflows/ocrmvp/prompt.go
      Note: Context policy prompt update
    - Path: scraper/pkg/workflows/ocrquality/figures.go
      Note: Figure worker implementation for Step 2
    - Path: scraper/pkg/workflows/ocrquality/package.go
      Note: Main implementation for Step 1
    - Path: ttmp/2026/05/24/OCR-QUALITY-WORKERS-001--port-ocr-qa-and-cleanup-scripts-to-go-workflow-workers/experiments/002-figure-aware-marker-recovery/notes.md
      Note: Figure-aware recovery experiment notes
ExternalSources: []
Summary: Diary for porting OCR QA and cleanup scripts to Go workflow workers.
LastUpdated: 2026-05-24T20:40:00-04:00
WhatFor: Use this to understand what changed while porting OCR quality scripts to Go and adding context-aware OCR inputs.
WhenToUse: Read before continuing OCR quality worker implementation or debugging the quality-pass workflow.
---




# Diary

## Goal

This diary records the work to port OCR quality Python scripts into Go workflow workers, add tasks and an implementation guide, and begin exploring surrounding-page context for OCR continuity.

## Step 1: Create the worker-port ticket and implement the first Go quality workers

I created `OCR-QUALITY-WORKERS-001` as the follow-up ticket for turning the stabilized Python experiment scripts from `BOOK-OCR-HQ-001` into proper Go workflow workers. The first implementation ports the useful stable behavior: markdown QA, deterministic list-page normalization, log import summaries, a workflow package, and a CLI `quality-pass` entry point.

I also added an exploratory `--context-window` option to `ocr-mvp run`. This lets a page OCR call include previous/next page images for continuity, while the v4 prompt explicitly instructs the model to transcribe only the target page.

### Prompt Context

**User prompt (verbatim):** "port the ocr python things to go (add  implementation guide and tasks). And then add all the recommended next steps in @2026-05-20--book-ocr/ttmp/2026/05/24/BOOK-OCR-HQ-001--high-quality-book-ocr-experiment-system/analysis/01-final-ocr-quality-report.md  as tasks as well. also potentially explore giving surrounding context to the OCR to improve continutiny.

Then optimize the whole thing to get _stellar_ results, including embedded extracted images.

Don't stop until the quality is impeccable"

**Assistant interpretation:** Create the follow-up ticket/docs/tasks, port the Python QA/log/normalization scripts into Go workflow workers, add the final report's next steps as tasks, and start improving OCR continuity and future image extraction toward higher quality.

**Inferred user intent:** The user wants the experiment-prototype tooling promoted into production-ish workflow infrastructure, then wants the OCR quality loop to continue beyond prompt-only improvements.

**Commit (code):** `eb19a4018ef5ebfbc89b730de597e686aeb5303f` — "Add OCR quality workflow workers"

### What I did

- Created ticket `OCR-QUALITY-WORKERS-001`.
- Added tasks for implementation guide, Go QA port, normalization port, workflow package, CLI, context-aware OCR, comparison run, and future embedded image extraction.
- Added the recommended next steps from `BOOK-OCR-HQ-001` final report as follow-up tasks in that ticket.
- Implemented `scraper/pkg/workflows/ocrquality`:
  - `QAResult`, `QAFinding`, `NormalizeResult`, `LogImportResult` types.
  - page-aware markdown splitting.
  - known bad term checks.
  - expected string checks.
  - adjacent duplicate line checks.
  - list-page markdown bullet/heading checks.
  - deterministic list-page dot-leader normalization.
  - NDJSON/plain log import summary with optional SQLite output.
  - workflow package with `qa-before`, `normalize-markdown`, `qa-after`, optional `import-log`, and `assemble-quality-report` steps.
- Added `ocr-mvp quality-pass` CLI command.
- Added `--context-window` to `ocr-mvp run`.
- Added context image plumbing through `PageOCRInput` and Geppetto multimodal calls.
- Added tests for QA, normalization, and context-window page selection.
- Ran the quality workflow against Experiment 007's raw markdown and confirmed it succeeded.

### Why

- The Python scripts had proven their value but were not first-class workflow workers.
- Typed QA findings and durable artifacts are needed before building operator/UI loops or targeted re-OCR.
- Surrounding-page context is a plausible next quality lever for continuity, but it must be explicit and bounded to avoid context leakage.

### What worked

- `go test ./cmd/ocr-mvp ./pkg/workflows/ocrquality ./pkg/workflows/ocrmvp -count=1` passed during development.
- The full pre-commit test/lint/security suite passed before the code commit.
- `ocr-mvp quality-pass` succeeded on the Experiment 007 markdown and wrote:
  - `/tmp/ocr-quality-go-pass/out/normalized.md`
  - `/tmp/ocr-quality-go-pass/out/cleanup.diff`
- The Go normalized output is semantically equivalent to the Python cleanup but has slightly tighter page-boundary whitespace.

### What didn't work

- The first commit attempt failed lint because `file.Close()` and `db.Close()` return values were not checked:

```text
pkg/workflows/ocrquality/logimport.go:41:18: Error return value of `file.Close` is not checked (errcheck)
pkg/workflows/ocrquality/logimport.go:52:17: Error return value of `db.Close` is not checked (errcheck)
```

- The first commit attempt also failed because a local variable was named `max`, shadowing the predeclared identifier:

```text
pkg/workflows/ocrquality/normalize.go:125:2: variable max has same name as predeclared identifier (predeclared)
```

- After fixing those, gosec flagged local artifact writes as potential path traversal (`G703`). I added narrow `#nosec G703` comments explaining that the paths are explicit operator/workflow inputs for local artifact export.

### What I learned

- The Python scripts mapped cleanly into three reusable worker concepts: QA, normalization, and log import.
- A workflow-native quality pass is more useful than a CLI-only port because it stores step results and artifacts in the same runtime model as OCR.
- Context-window OCR is easy to add mechanically, but it requires careful prompt wording and targeted experiments to ensure the model does not transcribe context pages.

### What was tricky to build

- The quality workflow needed to produce both local files and workflow artifacts. Local files are convenient for direct review, while artifacts preserve runtime provenance.
- The normalizer had to stay intentionally narrow. It should improve reviewability without becoming an invisible semantic editor.
- The Geppetto multimodal context path has to preserve target-page identity. The prompt now says the first image is always the target page and any additional images are context only.

### What warrants a second pair of eyes

- Whether the `quality-pass` CLI belongs under `ocr-mvp` long-term or should become a separate command.
- Whether `--context-window` should default to zero for safety, as it does now, or whether a future profile should enable it for list/continuation pages only.
- Whether log import should later be split into a pure library and a workflow executor with projection tables.

### What should be done in the future

- Add page-level QA projection tables instead of only result JSON/artifacts.
- Add targeted re-OCR workflow steps for failed QA pages.
- Add embedded figure/image extraction and markdown references to stored image artifacts.
- Run controlled context-window OCR experiments on pages 6-9, 13-15, and 29-31 before broad use.
- Compare context-window output against v4 mini baseline with both automated QA and vision checks.

### Code review instructions

- Start with:
  - `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/workflows/ocrquality/package.go`
  - `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/workflows/ocrquality/qa.go`
  - `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/workflows/ocrquality/normalize.go`
- Then review context OCR plumbing:
  - `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/workflows/ocrmvp/types.go`
  - `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/workflows/ocrmvp/discover.go`
  - `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/workflows/ocrmvp/geppetto_ocr.go`
  - `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/workflows/ocrmvp/prompt.go`
- Validate with:
  - `cd /home/manuel/workspaces/2026-05-20/book-ocr/scraper && go test ./cmd/ocr-mvp ./pkg/workflows/ocrquality ./pkg/workflows/ocrmvp -count=1`

### Technical details

Quality pass smoke command:

```bash
cd /home/manuel/workspaces/2026-05-20/book-ocr/scraper
rm -rf /tmp/ocr-quality-go-pass

go run ./cmd/ocr-mvp quality-pass \
  --markdown /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/BOOK-OCR-HQ-001--high-quality-book-ocr-experiment-system/experiments/007-quality-v4-mini-pages-001-030/outputs/01-final-quality-v4-mini-pages-001-030.md \
  --output-dir /tmp/ocr-quality-go-pass/out \
  --work-dir /tmp/ocr-quality-go-pass/work \
  --book-id presentation-based-uis-hq-007-v4-mini-30 \
  --expected-pages 30
```

Context-window OCR pattern for future experiments:

```bash
go run ./cmd/ocr-mvp run \
  --book-id presentation-based-uis-context-window-test \
  --image-dir /home/manuel/code/wesen/claw-stuff/output/books/presentation-based-uis/pages \
  --start-page 6 \
  --end-page 9 \
  --profile gpt-5-mini-low \
  --profile-registries /tmp/book-ocr-hq-001/profiles-clean.yaml \
  --prompt-version ocr-quality-v4-report794-lexicon \
  --context-window 1 \
  --log-level warn
```

## Step 2: Add embedded figure extraction to the Go quality pass

I extended the Go quality pass with an embedded figure extraction worker. This moves the first part of the “embedded extracted images” goal into the workflow runtime: markdown figure markers can now be replaced with image links to extracted PNG crops from the source page images.

The first crop implementation was too broad. I used the vision tool to inspect the extracted Figure 1-1 and Figure 1-4 crops, then tightened the algorithm until the page numbers/footers were removed while the complete diagrams remained visible.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Continue beyond QA/cleanup into embedded figure extraction as part of making the OCR output stellar.

**Inferred user intent:** The user wants final OCR markdown to preserve non-text figure information, not just textual `[FIGURE: ...]` markers.

**Commit (code):** `509c8f5dd2b55e6cf88cd650f6c39896fede5a6d` — "Add OCR figure embedding worker"

### What I did

- Added `scraper/pkg/workflows/ocrquality/figures.go`.
- Added `scraper/pkg/workflows/ocrquality/figures_test.go`.
- Added `EmbedFiguresInput` and `EmbedFiguresResult`.
- Added `ocr-quality/embed-figures` workflow executor.
- Added `ocr-mvp quality-pass --image-dir ... --embed-figures`.
- Ran the Go quality pass against Experiment 007 with embedded figure extraction.
- Created experiment folder `experiments/001-go-quality-pass-embedded-figures`.
- Preserved normalized markdown, embedded markdown, cleanup diff, QA reports, quality report, and extracted PNG figures.

### Why

- The best OCR output had two `[FIGURE: ...]` markers in the first 30 pages.
- Textual figure descriptions are useful, but a high-quality book OCR artifact should embed the actual extracted diagrams where possible.

### What worked

- The workflow extracted two figure images:
  - `page_013_figure_01.png`
  - `page_021_figure_01.png`
- The embedded markdown replaces figure markers with image links.
- The final vision check said page numbers/footers are removed and full diagrams remain present.
- Pre-commit tests/lint/security checks passed before the code commit.

### What didn't work

- The first crop used a broad non-white bounding box and included page margins, punch-hole artifacts, and page numbers.
- A second crop focused on the dominant ink band but cut off important parts of Figure 1-4.
- A third crop included all diagram content but still included page numbers.
- The final crop uses a meaningful ink-band union with a bottom cutoff to remove footer page numbers while preserving the diagrams.

### What I learned

- Figure extraction needs visual validation. A crop can be technically non-empty and still be bad.
- Page-level figure extraction can work for simple figure-only pages, but full-book extraction will need more structured segmentation.

### What was tricky to build

- The algorithm had to ignore scanner artifacts and footer page numbers without cutting off lower diagram boxes.
- The page images include margins and page-number ink that are visually far from the diagram but still count as non-white pixels.

### What warrants a second pair of eyes

- Whether the current crop whitespace is acceptable for reading or should be tightened further.
- Whether contrast normalization should be added for faint diagrams.
- Whether full-book figure extraction needs connected-component segmentation instead of page-level band selection.

### What should be done in the future

- Add figure QA checks that compare expected figure count to extracted image count.
- Add optional contrast enhancement.
- Add structured extraction for pages with multiple figures or figures embedded in prose.
- Run the embedded-figure quality pass as the default final review artifact for the first 30 pages.

### Code review instructions

- Review:
  - `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/workflows/ocrquality/figures.go`
  - `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/workflows/ocrquality/package.go`
  - `/home/manuel/workspaces/2026-05-20/book-ocr/scraper/cmd/ocr-mvp/main.go`
- Inspect outputs:
  - `experiments/001-go-quality-pass-embedded-figures/outputs/02-embedded-figures.md`
  - `experiments/001-go-quality-pass-embedded-figures/outputs/figures/`

### Technical details

Smoke command:

```bash
go run ./cmd/ocr-mvp quality-pass \
  --markdown /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/BOOK-OCR-HQ-001--high-quality-book-ocr-experiment-system/experiments/007-quality-v4-mini-pages-001-030/outputs/01-final-quality-v4-mini-pages-001-030.md \
  --output-dir /tmp/ocr-quality-go-figures/out \
  --work-dir /tmp/ocr-quality-go-figures/work \
  --book-id presentation-based-uis-hq-007-v4-mini-30 \
  --expected-pages 30 \
  --image-dir /home/manuel/code/wesen/claw-stuff/output/books/presentation-based-uis/pages \
  --embed-figures
```

## Step 3: Recover full-page diagram figures and add a figure-aware OCR prompt

I started the next self-optimization round from a concrete user observation: Figure 1-2 and Figure 1-3 are graphical full-page diagrams, but the current embedded artifact only linked Figure 1-1 and Figure 1-4 as images. The missing piece was the OCR contract. The v4 prompt told the model to mark figures, but it did not strongly state that a full-page flowchart/diagram still needs a `[FIGURE: ...]` marker even if the model transcribes all visible labels.

I fixed this in two layers. First, I added a new `ocr-quality-v5-figure-aware` prompt version for future OCR runs. Second, I made the figure embedding worker more robust for existing artifacts by synthesizing missing figure markers for caption-only diagram pages. This recovered Figure 1-2 and Figure 1-3 without rerunning the model.

### Prompt Context

**User prompt (verbatim):** "Ok, let's do the next round of self optimization, as before. Keep a detailed diary as you work so we can fully retrace your steps."

**Assistant interpretation:** Continue the OCR quality-improvement loop, implement the next fixes rather than just describing them, and keep a retraceable diary entry.

**Inferred user intent:** The user wants the system to iterate toward higher-quality OCR artifacts with explicit provenance, especially after noticing missing embedded full-page diagrams.

**Commit (code):** `aea38ef8f52a0354b4ef73acb045bdd1f998d825` — "Improve OCR figure detection"

### What I did

- Confirmed the current embedded artifact had Markdown image links for Figure 1-1 and Figure 1-4 only.
- Confirmed Figure 1-2 and Figure 1-3 were present as plain transcribed diagram text, not `[FIGURE: ...]` markers.
- Added prompt version `ocr-quality-v5-figure-aware` in `scraper/pkg/workflows/ocrmvp/prompt.go`.
- Added explicit prompt rules requiring `[FIGURE: ...]` markers for graphical pages, full-page diagrams, flowcharts, models, and architecture charts.
- Added figure-worker fallback detection for caption-only diagram pages in `scraper/pkg/workflows/ocrquality/figures.go`.
- Added tests for synthesized full-page diagram extraction and for not synthesizing markers from Table of Figures list rows.
- Ran the quality pass again against Experiment 007 raw OCR output.
- Preserved outputs under `experiments/002-figure-aware-marker-recovery`.
- Used the vision tool to validate the newly recovered Figure 1-2 and Figure 1-3 crops.

### Why

- The embedding worker uses `[FIGURE: ...]` as the extraction cue.
- The previous prompt did not consistently signal that a full-page diagram should emit that cue.
- Existing best OCR output should be salvageable without paying for a model rerun when the page already contains enough caption/diagram structure to infer a missing marker.

### What worked

- `go test ./pkg/workflows/ocrquality ./pkg/workflows/ocrmvp -count=1` passed during development.
- `go test ./cmd/ocr-mvp ./pkg/workflows/ocrquality ./pkg/workflows/ocrmvp -count=1` passed before commit.
- The pre-commit hook passed full tests, web unit tests, lint, gosec, and govulncheck.
- The rerun quality pass succeeded with run ID `ocr-quality-f29626cb-d734-4c0b-8ab1-e3874ad1fc8c`.
- The new embedded artifact now contains four figure images:
  - `page_013_figure_01.png` — Figure 1-1
  - `page_015_figure_01.png` — Figure 1-2
  - `page_017_figure_01.png` — Figure 1-3
  - `page_021_figure_01.png` — Figure 1-4
- Vision validation said Figure 1-2 and Figure 1-3 include the full title/diagram/labels and avoid visible page numbers/footers.

### What didn't work

- The original v4 OCR output had no `[FIGURE: ...]` markers for Figure 1-2 and Figure 1-3.
- This meant the first embedded-figure pass silently missed graphical pages that looked like text to the prompt.
- Figure 1-3's crop has noticeable background speckling/noise, although the crop is complete.

### What I learned

- Figure detection needs both a model-side contract and a deterministic fallback.
- Full-page diagrams are easy for a multimodal model to transcribe as labels, but downstream extraction needs an explicit machine-readable image marker.
- Table of Figures pages are a trap for naive caption matching; rows such as `Figure 1-2: ... 13` must not become extraction markers.

### What was tricky to build

- The fallback must distinguish actual figure pages from Table of Figures list entries.
- The heuristic therefore ignores caption lines containing dot leaders and requires diagram-like page structure: arrows, short label lines, mostly uppercase labels, or sparse non-prose layout.
- The fallback must not insert markers into pages that already have `[FIGURE: ...]` or Markdown image links.

### What warrants a second pair of eyes

- Whether the synthesized descriptions are good enough as alt text, e.g. `Full-page diagram showing The Representation Shift Model`.
- Whether the diagram-page heuristic is too permissive for later chapters with short prose fragments.
- Whether Figure 1-3 should get an immediate deterministic denoise/contrast enhancement pass.

### What should be done in the future

- Add figure QA counts: detected figure captions, markers, extracted images, and mismatches.
- Add sidecar JSON with crop rectangles and detection method.
- Add debug overlay PNGs for candidate regions and selected crop rectangles.
- Add optional raw/enhanced crop output with contrast/denoise.
- Run a targeted live OCR test using `ocr-quality-v5-figure-aware` on pages 13, 15, 17, and 21 to verify markers are emitted by the model without fallback synthesis.

### Code review instructions

- Start with `scraper/pkg/workflows/ocrmvp/prompt.go` and review `PromptVersionQualityV5FigureAware`.
- Then review `scraper/pkg/workflows/ocrquality/figures.go`, especially `synthesizeMissingFigureMarkers`, `addMissingFigureMarkerToPage`, and `looksLikeDiagramPage`.
- Review tests in `scraper/pkg/workflows/ocrquality/figures_test.go`.
- Inspect the new artifact:
  - `/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/OCR-QUALITY-WORKERS-001--port-ocr-qa-and-cleanup-scripts-to-go-workflow-workers/experiments/002-figure-aware-marker-recovery/outputs/02-embedded-figures.md`

### Technical details

Quality-pass command:

```bash
go run ./cmd/ocr-mvp quality-pass \
  --markdown /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/BOOK-OCR-HQ-001--high-quality-book-ocr-experiment-system/experiments/007-quality-v4-mini-pages-001-030/outputs/01-final-quality-v4-mini-pages-001-030.md \
  --output-dir /tmp/ocr-quality-go-figures-v2/out \
  --work-dir /tmp/ocr-quality-go-figures-v2/work \
  --book-id presentation-based-uis-hq-007-v4-mini-30 \
  --expected-pages 30 \
  --image-dir /home/manuel/code/wesen/claw-stuff/output/books/presentation-based-uis/pages \
  --embed-figures
```

New figure links observed in the output:

```markdown
![Full-page diagram showing The Representation Shift Model](figures/page_015_figure_01.png)
![Full-page diagram showing The Primitive Presentation System (PPS) Model](figures/page_017_figure_01.png)
```
