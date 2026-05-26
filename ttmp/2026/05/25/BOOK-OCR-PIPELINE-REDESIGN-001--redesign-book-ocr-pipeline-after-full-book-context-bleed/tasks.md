# Tasks

## TODO

- [x] Add tasks here

- [x] Analyze full-book OCR regression and identify lost invariants
- [x] Design structured target-page-only OCR pipeline with Geppetto turns and Pinocchio turn persistence
- [x] Write intern-facing implementation guide with file/API references, diagrams, pseudocode, and validation plan
- [x] Upload guide bundle to reMarkable
- [x] Phase 1 structured-page dry-run: add structured OCR prompt, fake client, single-page command, JSON artifacts, rendered Markdown, validation output, and turn input/final persistence
- [x] Phase 2 live structured-page smoke: run page 32 with gpt-5-mini-low, verify one-image input turn, table blocks, rendered Markdown tables, and saved raw/parsed outputs
- [x] Phase 3 figure boundary smoke: run pages 12,13,42,43 structured-page/live or controlled fake fixtures and verify page-local figure blocks and adjacent-caption validation
- [x] Phase 4 structured-run dry-run workflow: add multi-page workflow package/CLI, per-page artifacts, assembled Markdown, validation report, and turns DB
- [x] Phase 5 structured first-50 live run: compare against target-only freeform rerun, inspect Markdown tables/figures, and open final artifact with md-view
- [ ] Phase 6 production hardening: figure QA integration, text-only normalization, report command for structured QA, and full-book acceptance gates
