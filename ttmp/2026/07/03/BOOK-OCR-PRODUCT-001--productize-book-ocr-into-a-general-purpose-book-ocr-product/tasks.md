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

## Productization roadmap (from the design doc — future work)

- [ ] Phase 1: publish/tag workflow runtime, drop go.mod replace (F1); CI from clean clone (F9); flip --dry-run default + fix ocr-mvp usage strings (F6); goreleaser v0.1.0
- [ ] Phase 2: profile-driven generalization — extend bookprofile (CodePolicy/RenderPolicy/LexiconPolicy), thread profile through prompts/renderer/QA (F2–F4), golden Report-794 regression tests, generic dry-run fixtures, Cobra/Glazed CLI
- [ ] Phase 3: book-ocr ingest (pdftoppm) + init (profile bootstrap via discovery) + report (tokens/cost) + user docs (F8)
- [ ] Phase 4: first-class RequeueSteps rerun API (F5), audit command (F7), review-PDF mode, lease heartbeat/cancellation upstream requests, optional local review UI
- [ ] Plugin track P1: internal/plugin manager (import devctl pkg/runtime), ocr.page + prompt.render + figures.segment seams, profile plugins: section, reference plugins
- [ ] Plugin track P2: response.parse, validate.page/book chain, page.classify routing
- [ ] Plugin track P3: markdown.transform, ingest.pages hook, plugin cookbook doc
