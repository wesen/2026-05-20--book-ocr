# Tasks

## TODO

- [x] Add tasks here

- [x] Read current structured OCR CLI, workflow runtime usage, and old freeform OCR workflow package
- [x] Write intern-facing design and implementation guide with architecture, APIs, pseudocode, phases, and validation plan
- [x] Upload design guide bundle to reMarkable
- [x] Implement structured OCR workflow package with discover, page OCR, assemble, validate steps and retry policies
- [x] Wire book-ocr structured-run to workflow runtime and add status/resume/retry operator compatibility
- [x] Add workflow projection rows and artifact metadata for structured page outputs
- [x] Validate dry-run workflow over pages 1-50 and inspect turns/artifacts/projections
- [x] Run limited live structured workflow smoke and verify automatic retry/resume behavior
- [x] Phase 6 hardening: add deterministic workflow retry test for transient structured page failures
- [x] Phase 6 hardening: add structured page status command backed by structured_pages projection
- [x] Phase 6 hardening: add prose completeness and suspicious-short-page validation
- [ ] Phase 6 hardening: connect structured figure blocks to figure image embedding
