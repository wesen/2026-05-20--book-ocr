# Tasks

## TODO

- [x] Create ticket workspace and initial diary
- [x] Analyze workflow API, existing scraper runtime, and OCR reference process
- [x] Write intern-oriented design and implementation guide
- [x] Relate key source files, update changelog, and validate docmgr ticket
- [x] Upload documentation bundle to reMarkable
- [x] Phase 1a: Add scraper/pkg/workflows/ocrmvp package skeleton, public inputs/results, prompt, projection schema, and registration API
- [x] Phase 1b: Implement fake-client workflow executors for discover-pages, ocr-page, and assemble-markdown using workflow artifacts and projections
- [x] Phase 1c: Add integration tests that run the OCR MVP workflow with temp SQLite stores, fake page images, fake OCR, artifacts, projections, and final assembly
- [x] Phase 2a: Implement Geppetto-backed OCR client using pinocchio/pkg/cmds/profilebootstrap default registry resolution
- [x] Phase 2b: Add focused tests for profile selection wiring and opt-in live OCR smoke test guard
- [x] Phase 3a: Add a small CLI/example command to run ocr-mvp with --book-id, --image-dir, --work-dir, --profile, --profile-registries, --dry-run, and --max-workers
- [x] Phase 3b: Document operator smoke flows for retrying failed page steps and canceling runs
- [x] Finalize: run full tests, update guide/diary/changelog, validate docmgr, optionally re-upload to reMarkable
