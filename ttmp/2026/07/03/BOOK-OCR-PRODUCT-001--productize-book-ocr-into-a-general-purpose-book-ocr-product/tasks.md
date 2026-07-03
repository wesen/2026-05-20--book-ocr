# Tasks

## Investigation and deliverable (this ticket's work)

- [x] Create ticket workspace and docs (design doc + diary)
- [x] Survey repository, git history, and all nine prior ttmp tickets
- [x] Deep-read internal packages (ocrpipeline, ocrquality, ocrvalidation, bookprofile, ocrmvp, vlmseparation) and cmd/book-ocr
- [x] Deep-read scraper/pkg/workflow runtime and its persistence/execution model
- [x] Experiment: standalone build check (found F1: missing ../scraper replace target)
- [x] Experiment: build + full test suite via scripts/book-ocr.go.work (all packages pass)
- [x] Experiment: 3-page end-to-end dry run incl. figures + PDF (scripts/01-dry-run-structured-pipeline.sh, all artifacts verified)
- [x] Write productization analysis / design / implementation guide (findings F1–F9, 6 decision records, 4 phases)
- [x] Maintain investigation diary
- [x] Validate with docmgr doctor
- [x] Upload bundle to reMarkable
- [x] Analyze devctl NDJSON-stdio plugin protocol (pkg/protocol + pkg/runtime importable)
- [x] Experiment: protocol round-trip demo — Go host + Python plugin (scripts/02-plugin-protocol-demo, PLUGIN_PROTOCOL_DEMO_OK)
- [x] Write plugin-seams design doc (02): seams S1–S8, decisions D1–D4, plugin track P1–P3
- [x] Re-upload updated bundle to reMarkable

## Implementation queue (2026-07-03 session, in order)

- [x] Golden-file renderer regression harness (fixtures for every block type incl. figure suppression; pins behavior before Phase-2 refactor)
- [x] Phase-1 quick wins: flip --dry-run default to false (+ --live docs), fix ocr-mvp usage strings, engine schema-version guard in structured-rerun-pages
- [x] Plugin P1: internal/plugin manager on devctl pkg/runtime+protocol; ocr.page + prompt.render seams; FigureSegmenter extraction + figures.segment; profile plugins: section; --plugin override; adversarial + identity tests
- [x] Phase 2 (core): CodePolicy/RenderPolicy/PromptSpec in bookprofile, prompts/renderer threaded via PolicyFromProfile + discover-time stamping, profiles/report-794.yaml pinned in sync; goldens stayed byte-identical (leftovers: ocrquality QA/page-naming F4, generic dry-run fixtures, Cobra CLI)
- [x] Phase 3: book-ocr ingest (pdftoppm) + report (tokens/warnings from projection+turns)
- [x] CI workflow (build + vet + tests; E2E dry-run covered by package tests)

## Productization roadmap (from the design doc — future work)

- [x] Phase 1 (complete; v0.1.0 tagged locally, push pending): scraper v0.0.4 published dep (F1), CI green from clean clone (F9), --dry-run default flipped + usage fixed (F6); TODO: goreleaser v0.1.0 tag
- [x] Phase 2 (core): profile-driven prompts+renderer via PolicyFromProfile, profiles/report-794.yaml + generic-technical-book.yaml, golden equivalence proofs (leftovers: ocrquality QA/page-naming F4, Cobra/Glazed CLI, generic dry-run fixtures)
- [x] Phase 3 (core+init): ingest + init (drafted profile, review checklist, next-command) + report (leftovers: cost table, user docs)
- [ ] Phase 4: first-class RequeueSteps rerun API (F5), audit command (F7), review-PDF mode, lease heartbeat/cancellation upstream requests, optional local review UI
- [x] Plugin track P1: internal/plugin manager (import devctl pkg/runtime), ocr.page + prompt.render + figures.segment seams, profile plugins: section, reference plugins
- [x] Plugin track P2: response.parse (decline-to-builtin), validate.page/book (tagged, additive), page.classify with per-page strategy routing, plugin retryable-hint classification
- [ ] Plugin track P3: markdown.transform, ingest.pages hook, plugin cookbook doc

## Addressed 2026-07-03 (second implementation session)

- [x] F4: page-image resolution by number (figures, vlmseparation, marker regex) — ingest layout no longer breaks figure extraction
- [x] book-ocr init: PDF -> workspace + drafted profile + checklist
- [x] goreleaser trimmed to secret-free release; v0.1.0 tagged locally (push to trigger release workflow)
- [x] Cobra CLI migration (flags unchanged via DisableFlagParsing; vlm-separation mounted natively)
