---
Title: VLM Multi-Page Separation Benchmark Design and Implementation Guide
Ticket: BOOK-OCR-VLM-SEPARATION-001
Status: active
Topics:
    - ocr
    - book-processing
    - experiments
    - geppetto
    - pinocchio
    - workflow
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../../../../../code/wesen/go-go-golems/geppetto/pkg/doc/topics/08-turns.md
      Note: Turns and blocks architecture reference
    - Path: ../../../../../../../../../../code/wesen/go-go-golems/geppetto/pkg/turns/helpers_blocks.go
      Note: Geppetto block constructors used to build benchmark scenarios
    - Path: ../../../../../../../../../../code/wesen/go-go-golems/pinocchio/pkg/persistence/chatstore/turn_store.go
      Note: TurnStore interface for benchmark turn snapshots
    - Path: ../../../../../../../../../../code/wesen/go-go-golems/pinocchio/pkg/persistence/chatstore/turn_store_sqlite.go
      Note: SQLite turns DB implementation
    - Path: cmd/book-ocr/main.go
      Note: Existing CLI root where the Glazed benchmark command will be registered
    - Path: internal/ocrmvp/geppetto_ocr.go
      Note: Current multi-image Geppetto OCR path and context-image packaging
    - Path: internal/ocrmvp/prompt.go
      Note: Current prompt contracts whose separation behavior needs benchmarking
    - Path: internal/vlmseparation/command.go
      Note: Implemented Glazed benchmark command
    - Path: internal/vlmseparation/scenarios.go
      Note: Implemented scenario turn construction
    - Path: internal/vlmseparation/turns.go
      Note: Implemented turns DB persistence wrapper
ExternalSources: []
Summary: Design and implementation guide for a Glazed benchmark tool that tests whether VLMs can keep target-page OCR separate from neighboring page images under different prompt and turn/block layouts.
LastUpdated: 2026-05-25T00:00:00-04:00
WhatFor: Use this guide to implement and run a benchmark before redesigning Book OCR around assumptions about multimodal page separation.
WhenToUse: Read before implementing or running the VLM multi-page separation benchmark, adding turns-db support, or changing OCR prompt/context strategy.
---



# VLM Multi-Page Separation Benchmark Design and Implementation Guide

## Executive Summary

The last full-book OCR run showed that page context can contaminate target-page OCR. The run used a `--context-window 1` mode that sent neighboring page PNG images together with the target page image. The prompt instructed the model to transcribe only the first/target image, but the output still duplicated adjacent figure captions and created false figure markers. Before we remove image context entirely or redesign the pipeline around text-only context, we should measure how much of this failure is caused by prompt wording versus how much is caused by the way image blocks are arranged inside Geppetto turns.

This ticket creates an investigation and benchmark tool. The tool should run controlled experiments over selected book page triplets and compare several multimodal prompt layouts:

- one user block with target and context images together,
- multiple user blocks separated by text labels,
- target image followed by explicit separator text and context images,
- context images first then target image as a negative/control scenario,
- target image only as a baseline,
- optional prior Markdown/text context instead of context images.

The tool must log every trial to files, SQLite tables, and a Pinocchio/Geppetto turns database. This gives us immediate observability for prompt debugging and also advances the planned turn-persistence work from `BOOK-OCR-PIPELINE-REDESIGN-001`.

The tool should be implemented as a Glazed command so results can be emitted as rows, rendered as tables/JSON/YAML, and integrated into the existing Go CLI style.

## Problem Statement

The current Book OCR system has a real, observed ambiguity: when multiple page images are sent to a VLM, the model may not consistently obey the instruction that only one image is the OCR target.

The immediate symptom came from the full-book run in:

```text
/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/25/BOOK-OCR-FULL-001--convert-full-presentation-based-user-interfaces-book/experiments/001-full-book-v5-mini
```

The raw and embedded artifacts contained adjacent duplicate captions such as:

```text
[12, 13] Figure 1-1: A Rudimentary User Interface
[31, 32] Figure 2-2: PPSCalc -- Formula Display
[31, 32] Figure 2-3: PPSCalc -- Value Display
[59, 60] Figure 3-2: Extension with Both Planning and Immediate Changes
[60, 61] Figure 3-3: Command Data Base Extension
```

A direct visual check showed that `page_012_figure_01.png` was a false figure crop of prose, while `page_013_figure_01.png` was the actual Figure 1-1 page. This strongly suggests that the VLM imported context-page visual content into the target-page transcript.

However, we do not yet know whether this is unavoidable, or whether better prompt and block layout can improve target isolation enough for some image-context use cases.

## Investigation Questions

The benchmark should answer these questions:

1. **Target isolation:** If the model receives multiple page images, can it reliably transcribe only the target page?
2. **Block-layout sensitivity:** Does separating images into different user blocks with text labels improve isolation compared to one multimodal block containing all images?
3. **Prompt sensitivity:** Do stricter contracts, sentinel text, or JSON outputs reduce bleed?
4. **Order sensitivity:** Is target-first safer than context-first? Is target-last unsafe? Does separator text matter?
5. **Context usefulness:** Does neighboring image context improve legitimate continuity enough to justify the risk?
6. **Model/profile sensitivity:** Do `gpt-5-mini-low`, `gpt-5-nano-low`, and other profiles behave differently?
7. **Turns persistence:** Can we persist every benchmark turn in a queryable turns DB for later prompt replay and debugging?

## Scope

This ticket includes:

- a docmgr ticket,
- this intern-facing guide,
- a diary,
- reMarkable upload,
- a Glazed benchmark command,
- file and SQLite logging,
- Pinocchio `chatstore` turn persistence,
- dry-run/fake mode for tests,
- live mode for opt-in VLM experiments.

This ticket does not require changing the production OCR pipeline yet. The benchmark informs that decision.

## Current System Context

### Book OCR repository

```text
/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr
```

Important files today:

```text
cmd/book-ocr/main.go
internal/ocrmvp/geppetto_ocr.go
internal/ocrmvp/prompt.go
internal/ocrmvp/types.go
internal/ocrquality/figures.go
```

`internal/ocrmvp/geppetto_ocr.go` contains the current Geppetto OCR client. The function `multimodalImages` currently builds one image list where the target page is first and context pages follow. That image list is passed to `turns.NewUserMultimodalBlock`.

### Geppetto turns

Geppetto uses `turns.Turn` as the canonical inference container:

```go
turn := &turns.Turn{}
turns.AppendBlock(turn, turns.NewSystemTextBlock(systemPrompt))
turns.AppendBlock(turn, turns.NewUserTextBlock(userText))
turns.AppendBlock(turn, turns.NewUserMultimodalBlock(prompt, images))
updated, err := engine.RunInference(ctx, turn)
```

Important docs/files:

```text
/home/manuel/code/wesen/go-go-golems/geppetto/pkg/doc/topics/08-turns.md
/home/manuel/code/wesen/go-go-golems/geppetto/pkg/turns/helpers_blocks.go
```

`NewUserMultimodalBlock` stores text plus an image list in one block. The benchmark should deliberately compare this current layout against a layout with multiple user blocks.

### Pinocchio turn persistence

Pinocchio already has SQLite turn persistence:

```go
import chatstore "github.com/go-go-golems/pinocchio/pkg/persistence/chatstore"
```

Important files:

```text
/home/manuel/code/wesen/go-go-golems/pinocchio/pkg/persistence/chatstore/turn_store.go
/home/manuel/code/wesen/go-go-golems/pinocchio/pkg/persistence/chatstore/turn_store_sqlite.go
/home/manuel/code/wesen/go-go-golems/pinocchio/pkg/cmds/cmdlayers/helpers.go
```

The benchmark should support:

```text
--turns-db PATH
--turns-dsn DSN
```

and use:

```go
dsn, err := chatstore.SQLiteTurnDSNForFile(path)
store, err := chatstore.NewSQLiteTurnStore(dsn)
store.Save(ctx, convID, sessionID, turnID, phase, createdAtMs, payloadYAML, opts)
```

### Pinocchio profile resolution

The benchmark should use the same profile resolution as current OCR:

```go
profilebootstrap.NewCLISelectionValues(...)
profilebootstrap.ResolveCLIEngineSettings(ctx, values)
profilebootstrap.NewEngineFromResolvedCLIEngineSettings(resolved)
```

This keeps model selection consistent with the OCR pipeline.

## Proposed Tool

Add a new Glazed command to `book-ocr`:

```text
book-ocr vlm-separation benchmark [flags]
```

For a minimal first implementation, it can be registered directly as:

```text
book-ocr vlm-separation-benchmark [flags]
```

but the preferred CLI tree is:

```text
book-ocr vlm-separation benchmark
book-ocr vlm-separation inspect
book-ocr vlm-separation scenarios
```

The first command runs benchmark trials.

### Example usage

```bash
go run ./cmd/book-ocr vlm-separation benchmark \
  --book-id report-794 \
  --image-dir /home/manuel/code/wesen/claw-stuff/output/books/presentation-based-uis/pages \
  --target-pages 12,13,31,32,59,60 \
  --scenarios target-only,single-block-target-first,multi-block-labeled,target-text-context \
  --profile gpt-5-mini-low \
  --profile-registries /tmp/book-ocr-hq-001/profiles-clean.yaml \
  --out-dir /tmp/book-ocr-vlm-separation/report794 \
  --sqlite /tmp/book-ocr-vlm-separation/report794/results.sqlite \
  --turns-db /tmp/book-ocr-vlm-separation/report794/turns.db \
  --dry-run=false \
  --output table
```

Dry-run smoke test:

```bash
go run ./cmd/book-ocr vlm-separation benchmark \
  --book-id report-794 \
  --image-dir /home/manuel/code/wesen/claw-stuff/output/books/presentation-based-uis/pages \
  --target-pages 12,13 \
  --scenarios target-only,single-block-target-first,multi-block-labeled \
  --out-dir /tmp/book-ocr-vlm-separation-dry \
  --sqlite /tmp/book-ocr-vlm-separation-dry/results.sqlite \
  --turns-db /tmp/book-ocr-vlm-separation-dry/turns.db \
  --dry-run=true \
  --output json
```

## Benchmark Design

### Dataset selection

Start with pages where we know context bleed is risky:

| Target Page | Context Risk | Reason |
|---:|---|---|
| 12 | Figure 1-1 false marker | page 12 prose references figure on page 13 |
| 13 | actual Figure 1-1 | target has real figure |
| 31 | duplicate PPSCalc captions | page 31/32 repeated Figure 2-2 and 2-3 |
| 32 | actual table/figure region | neighbor overlap risk |
| 59 | duplicate Figure 3-2 | context page contamination candidate |
| 60 | duplicate Figure 3-2/Figure 3-3 | context page contamination candidate |
| 87 | Xerox Star duplicate | figure/caption adjacency |
| 97 | Steamer icons duplicate | visual icons/caption adjacency |
| 112 | command support duplicate | figure/caption adjacency |
| 115 | reference resolution duplicate | diagram-heavy page |

The command should accept arbitrary target pages, but provide defaults or a named preset:

```text
--preset report794-bleed-smoke
--preset report794-figure-adjacent
```

### Scenarios

Each scenario is a prompt and block-layout configuration. The benchmark should make the scenario explicit in SQLite and file output.

#### Scenario A: target-only

Purpose: baseline target-page OCR with no image context.

Turn layout:

```text
system: benchmark OCR contract
user multimodal block:
  text: transcribe target page only
  images: [target]
```

Expected: no context bleed possible.

#### Scenario B: single-block-target-first

Purpose: reproduce current OCR layout.

Turn layout:

```text
system: benchmark OCR contract
user multimodal block:
  text: first image is target, later images are context only
  images: [target, previous, next]
```

Expected: likely bleed on risky pages.

#### Scenario C: single-block-labeled-images

Purpose: test whether richer image metadata and prompt labels help when images still share one block.

Turn layout:

```text
system: benchmark OCR contract
user multimodal block:
  text: image labels and strict instructions
  images:
    - target with metadata role=target
    - previous with metadata role=context_previous
    - next with metadata role=context_next
```

Expected: may or may not improve, depending on provider behavior.

#### Scenario D: multi-block-labeled

Purpose: test whether text separators between user blocks help.

Turn layout:

```text
system: benchmark OCR contract
user text: The next image block is the ONLY target page.
user multimodal: TARGET PAGE ONLY + [target image]
user text: The following context blocks are NOT OCR targets. Use only for terminology.
user multimodal: PREVIOUS CONTEXT PAGE - DO NOT TRANSCRIBE + [previous image]
user multimodal: NEXT CONTEXT PAGE - DO NOT TRANSCRIBE + [next image]
user text: Now output only target page transcription as JSON.
```

Expected: if block separation matters, this should reduce bleed compared to scenario B.

#### Scenario E: context-first-negative-control

Purpose: measure worst-case order sensitivity.

Turn layout:

```text
images: [previous, next, target]
prompt: target is last image
```

Expected: likely worse; useful to quantify order sensitivity.

#### Scenario F: target-plus-text-context

Purpose: compare safe context design.

Turn layout:

```text
system: benchmark OCR contract
user text: previous page markdown summary, context only
user text: next page markdown summary, context only
user multimodal: target image only
```

Expected: should avoid visual context bleed while still testing style/terminology context.

## Output Contract

The model should not return freeform prose. It should return JSON that is easy to score.

Recommended benchmark response schema:

```json
{
  "target_page": 13,
  "transcribed_page_identity": {
    "page_number_visible": "",
    "title_or_caption_lines": ["Figure 1-1: A Rudimentary User Interface"]
  },
  "content_markers": {
    "figure_captions": ["Figure 1-1: A Rudimentary User Interface"],
    "section_headings": [],
    "unique_phrases": ["Application Data Base", "observables", "queries"]
  },
  "transcription": "...target page markdown or plain text...",
  "suspected_context_copy": false,
  "notes": []
}
```

The benchmark should parse this JSON when possible. If parsing fails, it should store the raw response and mark the trial as `parse_failed`.

## Scoring Design

The benchmark should compute deterministic metrics from the model response.

### Page-specific expected and forbidden phrases

For each target page, define:

```go
type PageOracle struct {
    TargetPage       int
    ExpectedPhrases  []string
    ForbiddenPhrases []string
    ExpectedCaptions []string
    ForbiddenCaptions []string
}
```

Example for page 12:

```go
PageOracle{
    TargetPage: 12,
    ExpectedPhrases: []string{
        "A Rudimentary User Interface",
        "Representation Shift",
    },
    ForbiddenCaptions: []string{
        "Figure 1-1: A Rudimentary User Interface", // if the physical caption is only on page 13
    },
    ForbiddenPhrases: []string{
        "Application Data Base", // depending on exact source; verify oracle carefully
    },
}
```

Oracles must be curated carefully. The first implementation should include a small conservative set rather than overfitting.

### Metrics

Each trial should produce:

```go
type TrialMetrics struct {
    ExpectedPhraseHits   int
    ExpectedPhraseTotal  int
    ForbiddenPhraseHits  int
    ForbiddenCaptionHits int
    JSONParseOK          bool
    SuspectedBleed       bool
    TargetOnlyScore      float64
}
```

Suggested scoring:

```go
func ScoreTrial(resp BenchmarkResponse, oracle PageOracle) TrialMetrics {
    text := strings.ToLower(resp.Transcription + "\n" + strings.Join(resp.ContentMarkers.FigureCaptions, "\n"))
    expectedHits := countHits(text, oracle.ExpectedPhrases)
    forbiddenHits := countHits(text, append(oracle.ForbiddenPhrases, oracle.ForbiddenCaptions...))

    score := float64(expectedHits) / max(1, len(oracle.ExpectedPhrases))
    if forbiddenHits > 0 {
        score -= 0.5 * float64(forbiddenHits)
    }
    if score < 0 { score = 0 }

    return TrialMetrics{
        ExpectedPhraseHits: expectedHits,
        ExpectedPhraseTotal: len(oracle.ExpectedPhrases),
        ForbiddenPhraseHits: forbiddenHits,
        ForbiddenCaptionHits: countHits(text, oracle.ForbiddenCaptions),
        JSONParseOK: true,
        SuspectedBleed: forbiddenHits > 0 || resp.SuspectedContextCopy,
        TargetOnlyScore: score,
    }
}
```

## Logging and Persistence

The tool must log to three places.

### 1. Files

Directory layout:

```text
<out-dir>/
  manifest.json
  scenarios.json
  trials/
    trial-0001/
      request.yaml
      response.txt
      response.json
      metrics.json
      turn-input.yaml
      turn-final.yaml
    trial-0002/
      ...
  summary.json
  summary.md
```

### 2. SQLite benchmark database

SQLite path:

```text
--sqlite <out-dir>/results.sqlite
```

Schema:

```sql
CREATE TABLE benchmark_runs (
  id TEXT PRIMARY KEY,
  book_id TEXT NOT NULL,
  profile TEXT NOT NULL,
  prompt_version TEXT NOT NULL,
  started_at TEXT NOT NULL,
  completed_at TEXT,
  out_dir TEXT NOT NULL,
  turns_dsn TEXT,
  dry_run INTEGER NOT NULL
);

CREATE TABLE benchmark_trials (
  id TEXT PRIMARY KEY,
  run_id TEXT NOT NULL,
  scenario TEXT NOT NULL,
  target_page INTEGER NOT NULL,
  previous_page INTEGER,
  next_page INTEGER,
  status TEXT NOT NULL,
  request_path TEXT NOT NULL,
  response_path TEXT,
  parsed_response_path TEXT,
  metrics_path TEXT,
  turn_session_id TEXT NOT NULL,
  turn_id TEXT NOT NULL,
  started_at TEXT NOT NULL,
  completed_at TEXT,
  latency_ms INTEGER,
  error TEXT,
  FOREIGN KEY(run_id) REFERENCES benchmark_runs(id)
);

CREATE TABLE trial_metrics (
  trial_id TEXT PRIMARY KEY,
  json_parse_ok INTEGER NOT NULL,
  expected_phrase_hits INTEGER NOT NULL,
  expected_phrase_total INTEGER NOT NULL,
  forbidden_phrase_hits INTEGER NOT NULL,
  forbidden_caption_hits INTEGER NOT NULL,
  suspected_bleed INTEGER NOT NULL,
  target_only_score REAL NOT NULL,
  raw_json TEXT NOT NULL,
  FOREIGN KEY(trial_id) REFERENCES benchmark_trials(id)
);
```

### 3. Pinocchio turns DB

Use Pinocchio `chatstore.SQLiteTurnStore`.

Default:

```text
<out-dir>/turns.db
```

Identifier scheme:

```text
convID    = vlm-separation:<book-id>:<run-id>
sessionID = page:<NNN>:scenario:<scenario-name>
turnID    = trial:<trial-id>:input
turnID    = trial:<trial-id>:final
phase     = input / final / parse-error
```

The persisted turn lets us inspect exactly how blocks were laid out.

## Glazed Command Design

### Package layout

Preferred layout:

```text
cmd/book-ocr/main.go                         existing root; add subcommand registration
internal/vlmseparation/
  types.go                                   scenarios, results, metrics
  scenarios.go                               turn construction for each scenario
  runner.go                                  benchmark orchestration
  sqlite.go                                  benchmark result database
  turns.go                                   Pinocchio chatstore wrapper
  oracle.go                                  page oracles and scoring
  files.go                                   manifest/trial file output
  command.go                                 Glazed command implementation
  command_test.go
  scoring_test.go
```

If root CLI refactor is too large, keep the existing manual CLI and add one Cobra subcommand built from a Glazed command. Do not rewrite the whole `book-ocr` CLI yet.

### Glazed command skeleton

```go
type BenchmarkCommand struct {
    *cmds.CommandDescription
}

type BenchmarkSettings struct {
    BookID            string   `glazed:"book-id"`
    ImageDir          string   `glazed:"image-dir"`
    TargetPages       string   `glazed:"target-pages"`
    Scenarios         []string `glazed:"scenarios"`
    Preset            string   `glazed:"preset"`
    OutDir            string   `glazed:"out-dir"`
    SQLitePath        string   `glazed:"sqlite"`
    TurnsDB           string   `glazed:"turns-db"`
    TurnsDSN          string   `glazed:"turns-dsn"`
    Profile           string   `glazed:"profile"`
    ProfileRegistries []string `glazed:"profile-registries"`
    DryRun            bool     `glazed:"dry-run"`
    MaxTrials         int      `glazed:"max-trials"`
}

func NewBenchmarkCommand() (*BenchmarkCommand, error) {
    glazedSection, err := settings.NewGlazedSchema()
    if err != nil { return nil, err }
    commandSettingsSection, err := cli.NewCommandSettingsSection()
    if err != nil { return nil, err }

    desc := cmds.NewCommandDescription(
        "benchmark",
        cmds.WithShort("Benchmark VLM target-page separation under multi-image prompt layouts"),
        cmds.WithFlags(
            fields.New("book-id", fields.TypeString, fields.WithHelp("Book identifier")),
            fields.New("image-dir", fields.TypeString, fields.WithHelp("Directory with page_NNN.png files")),
            fields.New("target-pages", fields.TypeString, fields.WithHelp("Comma-separated target pages")),
            fields.New("scenarios", fields.TypeStringList, fields.WithDefault([]string{"target-only"})),
            fields.New("out-dir", fields.TypeString, fields.WithDefault("/tmp/book-ocr-vlm-separation")),
            fields.New("sqlite", fields.TypeString, fields.WithDefault("")),
            fields.New("turns-db", fields.TypeString, fields.WithDefault("")),
            fields.New("turns-dsn", fields.TypeString, fields.WithDefault("")),
            fields.New("profile", fields.TypeString, fields.WithDefault("")),
            fields.New("profile-registries", fields.TypeStringList, fields.WithDefault([]string{})),
            fields.New("dry-run", fields.TypeBool, fields.WithDefault(true)),
            fields.New("max-trials", fields.TypeInteger, fields.WithDefault(0)),
        ),
        cmds.WithSections(glazedSection, commandSettingsSection),
    )
    return &BenchmarkCommand{CommandDescription: desc}, nil
}

func (c *BenchmarkCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
    s := &BenchmarkSettings{}
    if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
        return err
    }

    runner, err := NewRunner(RunnerConfigFromSettings(s))
    if err != nil { return err }

    results, err := runner.Run(ctx)
    if err != nil { return err }

    for _, r := range results.Trials {
        if err := gp.AddRow(ctx, types.NewRow(
            types.MRP("trial_id", r.ID),
            types.MRP("scenario", r.Scenario),
            types.MRP("target_page", r.TargetPage),
            types.MRP("status", r.Status),
            types.MRP("suspected_bleed", r.Metrics.SuspectedBleed),
            types.MRP("target_only_score", r.Metrics.TargetOnlyScore),
            types.MRP("turn_session_id", r.TurnSessionID),
            types.MRP("turn_id", r.TurnID),
        )); err != nil {
            return err
        }
    }
    return nil
}
```

### Root command registration

Add a `vlm-separation` group command and register the Glazed benchmark command underneath it.

Pseudocode:

```go
func newVLMSeparationRoot() (*cobra.Command, error) {
    root := &cobra.Command{
        Use:   "vlm-separation",
        Short: "Investigate VLM separation of target and context page images",
    }
    bench, err := vlmseparation.NewBenchmarkCommand()
    if err != nil { return nil, err }
    cobraBench, err := cli.BuildCobraCommandFromCommand(bench,
        cli.WithParserConfig(cli.CobraParserConfig{
            AppName: "book-ocr",
            ShortHelpSections: []string{schema.DefaultSlug},
            MiddlewaresFunc: cli.CobraCommandDefaultMiddlewares,
        }),
    )
    if err != nil { return nil, err }
    root.AddCommand(cobraBench)
    return root, nil
}
```

## Turn Construction Pseudocode

```go
func BuildTurnForScenario(s Scenario, input TrialInput) (*turns.Turn, error) {
    turn := &turns.Turn{ID: input.TurnID}
    turns.AppendBlock(turn, turns.NewSystemTextBlock(BenchmarkSystemPrompt()))

    switch s.Name {
    case "target-only":
        turns.AppendBlock(turn, turns.NewUserMultimodalBlock(
            TargetOnlyPrompt(input),
            []map[string]any{ImagePayload(input.TargetImagePath, "target", input.TargetPage)},
        ))

    case "single-block-target-first":
        images := []map[string]any{
            ImagePayload(input.TargetImagePath, "target", input.TargetPage),
            ImagePayload(input.PreviousImagePath, "context_previous", input.PreviousPage),
            ImagePayload(input.NextImagePath, "context_next", input.NextPage),
        }
        turns.AppendBlock(turn, turns.NewUserMultimodalBlock(SingleBlockPrompt(input), images))

    case "multi-block-labeled":
        turns.AppendBlock(turn, turns.NewUserTextBlock("The next image block is the ONLY OCR target."))
        turns.AppendBlock(turn, turns.NewUserMultimodalBlock(
            fmt.Sprintf("TARGET PAGE %03d. OCR this image only.", input.TargetPage),
            []map[string]any{ImagePayload(input.TargetImagePath, "target", input.TargetPage)},
        ))
        turns.AppendBlock(turn, turns.NewUserTextBlock("The following images are context only. Do not transcribe them."))
        if input.PreviousImagePath != "" {
            turns.AppendBlock(turn, turns.NewUserMultimodalBlock(
                fmt.Sprintf("PREVIOUS CONTEXT PAGE %03d. Do not transcribe.", input.PreviousPage),
                []map[string]any{ImagePayload(input.PreviousImagePath, "context_previous", input.PreviousPage)},
            ))
        }
        if input.NextImagePath != "" {
            turns.AppendBlock(turn, turns.NewUserMultimodalBlock(
                fmt.Sprintf("NEXT CONTEXT PAGE %03d. Do not transcribe.", input.NextPage),
                []map[string]any{ImagePayload(input.NextImagePath, "context_next", input.NextPage)},
            ))
        }
        turns.AppendBlock(turn, turns.NewUserTextBlock("Return JSON for the target page only."))
    }

    _ = turns.KeyTurnMetaSessionID.Set(&turn.Metadata, input.SessionID)
    _ = turns.KeyTurnMetaRuntime.Set(&turn.Metadata, "book-ocr/vlm-separation/"+s.Name)
    return turn, nil
}
```

## Runner Pseudocode

```go
func (r *Runner) Run(ctx context.Context) (*RunResult, error) {
    run := NewRunManifest(r.Config)
    if err := r.Files.WriteManifest(run); err != nil { return nil, err }
    if err := r.DB.InsertRun(ctx, run); err != nil { return nil, err }

    var rows []TrialResult
    for _, page := range r.TargetPages {
        for _, scenario := range r.Scenarios {
            trial := NewTrial(run.ID, page, scenario)
            if r.MaxTrials > 0 && len(rows) >= r.MaxTrials { break }

            result, err := r.RunTrial(ctx, trial)
            if err != nil {
                result = TrialResultFromError(trial, err)
            }
            rows = append(rows, result)
            _ = r.DB.InsertTrial(ctx, result)
            _ = r.Files.WriteTrial(result)
        }
    }

    summary := Summarize(rows)
    _ = r.Files.WriteSummary(summary)
    _ = r.DB.CompleteRun(ctx, run.ID, summary)
    return &RunResult{Run: run, Trials: rows, Summary: summary}, nil
}
```

## Trial Pseudocode

```go
func (r *Runner) RunTrial(ctx context.Context, trial Trial) (TrialResult, error) {
    input := r.BuildTrialInput(trial)
    turn, err := BuildTurnForScenario(trial.Scenario, input)
    if err != nil { return TrialResult{}, err }

    // Persist exact pre-inference turn.
    _ = r.TurnStore.Save(ctx, input.SessionID, input.TurnID, "input", turn)
    _ = r.Files.WriteTurn(input.TrialDir, "turn-input.yaml", turn)

    start := time.Now()
    var updated *turns.Turn
    if r.Config.DryRun {
        updated = FakeVLMResponse(turn, input)
    } else {
        updated, err = r.Engine.RunInference(ctx, turn)
        if err != nil { return TrialResult{}, err }
    }
    latency := time.Since(start)

    _ = r.TurnStore.Save(ctx, input.SessionID, input.TurnID, "final", updated)
    _ = r.Files.WriteTurn(input.TrialDir, "turn-final.yaml", updated)

    raw := LastLLMText(updated)
    parsed, parseErr := ParseBenchmarkResponse(raw)
    metrics := Score(raw, parsed, input.Oracle, parseErr)

    return TrialResult{
        ID: trial.ID,
        Scenario: trial.Scenario.Name,
        TargetPage: input.TargetPage,
        Status: statusFromParse(parseErr),
        LatencyMS: latency.Milliseconds(),
        RawResponse: raw,
        Parsed: parsed,
        Metrics: metrics,
        TurnSessionID: input.SessionID,
        TurnID: input.TurnID,
    }, nil
}
```

## Dry-Run Mode

Dry-run mode is required for normal tests. It should not call a provider.

A fake response can be deterministic per scenario:

```go
func FakeVLMResponse(turn *turns.Turn, input TrialInput) *turns.Turn {
    response := BenchmarkResponse{
        TargetPage: input.TargetPage,
        Transcription: fmt.Sprintf("DRY RUN target page %03d", input.TargetPage),
        ContentMarkers: ContentMarkers{
            FigureCaptions: input.Oracle.ExpectedCaptions,
        },
        SuspectedContextCopy: false,
    }
    if input.Scenario.Name == "context-first-negative-control" {
        response.SuspectedContextCopy = true
        response.Transcription += " CONTEXT_COPY_SENTINEL"
    }
    turns.AppendBlock(turn, turns.NewAssistantTextBlock(MustJSON(response)))
    return turn
}
```

Dry-run tests should verify:

- files are written,
- SQLite rows are written,
- turns DB rows are written,
- Glazed emits one row per trial,
- scenario and page IDs are stable.

## Implementation Plan

### Phase 1: Documentation and ticket setup

Already covered by this guide and diary.

### Phase 2: Core types and scoring

Add:

```text
internal/vlmseparation/types.go
internal/vlmseparation/oracle.go
internal/vlmseparation/scoring.go
internal/vlmseparation/scoring_test.go
```

Implement:

- scenario definitions,
- run/trial result structs,
- page oracle structs,
- JSON parser,
- deterministic scoring.

### Phase 3: Files and SQLite logging

Add:

```text
internal/vlmseparation/files.go
internal/vlmseparation/sqlite.go
internal/vlmseparation/sqlite_test.go
```

Implement:

- manifest writing,
- per-trial request/response/metrics files,
- SQLite schema migration,
- insert/update methods.

### Phase 4: Turns DB wrapper

Add:

```text
internal/vlmseparation/turns.go
internal/vlmseparation/turns_test.go
```

Implement wrapper around Pinocchio `chatstore.TurnStore`:

- derive DSN from `--turns-db`,
- open direct `--turns-dsn`,
- save input/final turns,
- close cleanly.

### Phase 5: Scenario turn construction

Add:

```text
internal/vlmseparation/scenarios.go
internal/vlmseparation/scenarios_test.go
```

Implement all scenario layouts and assert block counts/images:

- target-only: one multimodal block, one image,
- single-block-target-first: one multimodal block, target + context images,
- multi-block-labeled: several user blocks, image blocks separated by text blocks,
- context-first-negative-control: context before target,
- target-plus-text-context: target image plus textual context only.

### Phase 6: Glazed command

Add:

```text
internal/vlmseparation/command.go
```

Register in:

```text
cmd/book-ocr/main.go
```

Keep integration minimal: the existing CLI can still use manual `flag` subcommands, but the benchmark command itself must be a real Glazed command built with `cli.BuildCobraCommandFromCommand`.

### Phase 7: Dry-run validation

Run:

```bash
go test ./internal/vlmseparation -count=1
go run ./cmd/book-ocr vlm-separation benchmark \
  --dry-run \
  --book-id report-794 \
  --image-dir /home/manuel/code/wesen/claw-stuff/output/books/presentation-based-uis/pages \
  --target-pages 12,13 \
  --scenarios target-only,single-block-target-first,multi-block-labeled \
  --out-dir /tmp/book-ocr-vlm-separation-dry \
  --output json
```

### Phase 8: Opt-in live benchmark

Run a small live benchmark only after dry-run passes. Use 2 pages and 2 scenarios first.

```bash
go run ./cmd/book-ocr vlm-separation benchmark \
  --dry-run=false \
  --book-id report-794 \
  --image-dir /home/manuel/code/wesen/claw-stuff/output/books/presentation-based-uis/pages \
  --target-pages 12,13 \
  --scenarios target-only,multi-block-labeled \
  --profile gpt-5-mini-low \
  --profile-registries /tmp/book-ocr-hq-001/profiles-clean.yaml \
  --out-dir /tmp/book-ocr-vlm-separation-live-001 \
  --turns-db /tmp/book-ocr-vlm-separation-live-001/turns.db \
  --sqlite /tmp/book-ocr-vlm-separation-live-001/results.sqlite \
  --output table
```

## Acceptance Criteria

The implementation is acceptable when:

- `book-ocr vlm-separation benchmark` exists as a Glazed command.
- Dry-run mode passes tests and writes all three persistence layers.
- SQLite contains one run row, one trial row per scenario/page, and metrics rows.
- Turns DB contains input/final turn snapshots per trial.
- File output contains manifest, scenario descriptions, raw responses, parsed responses, and metrics.
- Scenario tests prove block layouts are distinct.
- The guide and diary are uploaded to reMarkable.

## File Reference Index

Current OCR files:

```text
/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/cmd/book-ocr/main.go
/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/ocrmvp/geppetto_ocr.go
/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/ocrmvp/prompt.go
/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/ocrmvp/types.go
```

New benchmark files to create:

```text
/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/vlmseparation/types.go
/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/vlmseparation/scenarios.go
/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/vlmseparation/runner.go
/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/vlmseparation/sqlite.go
/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/vlmseparation/turns.go
/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/internal/vlmseparation/command.go
```

Glazed references:

```text
/home/manuel/.pi/agent/skills/glazed-command-authoring/SKILL.md
/home/manuel/code/wesen/go-go-golems/glazed/pkg/cmds
/home/manuel/code/wesen/go-go-golems/glazed/pkg/cli
/home/manuel/code/wesen/go-go-golems/glazed/pkg/settings
```

Geppetto/Pinocchio references:

```text
/home/manuel/code/wesen/go-go-golems/geppetto/pkg/doc/topics/08-turns.md
/home/manuel/code/wesen/go-go-golems/geppetto/pkg/turns/helpers_blocks.go
/home/manuel/code/wesen/go-go-golems/pinocchio/pkg/persistence/chatstore/turn_store.go
/home/manuel/code/wesen/go-go-golems/pinocchio/pkg/persistence/chatstore/turn_store_sqlite.go
/home/manuel/code/wesen/go-go-golems/pinocchio/pkg/cmds/cmdlayers/helpers.go
```

Problem artifacts:

```text
/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/25/BOOK-OCR-FULL-001--convert-full-presentation-based-user-interfaces-book/experiments/001-full-book-v5-mini/outputs/01-full-book-raw.md
/home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/25/BOOK-OCR-FULL-001--convert-full-presentation-based-user-interfaces-book/experiments/001-full-book-v5-mini/outputs/quality-pass/03-embedded-figures.md
```
