# Tasks

## TODO

- [x] Map current scraper workflow, scheduler, store, API, UI, metrics, and runtime-event architecture
- [x] Analyze how a book OCR pipeline would map onto scraper workflows and identify production/admin gaps
- [x] Compare current SQLite scheduler with optional River/Postgres backend and propose backend abstraction
- [x] Write intern-oriented design and implementation guide with file references, diagrams, APIs, pseudocode, and rollout plan
- [x] Validate docmgr ticket and upload final bundle to reMarkable
- [x] Phase 1: Implement pkg/workflow executor facade over runner.Runner with StepContext input/result/artifact/record/emit helpers
- [x] Phase 1 validation: Add unit tests for typed executor adapter, StepContext helpers, and error classification
- [ ] Phase 2: Add Runtime skeleton and SQLite-backed StartRun/RunOnce design implementation
- [ ] Phase 3: Add artifact/projection/operator APIs after executor facade stabilizes
