# Tasks

## TODO

- [x] Add tasks here

- [x] Analyze current scraper book-specific OCR code and artifact boundaries
- [x] Write intern-facing design and implementation guide for moving book-specific OCR policy into 2026-05-20--book-ocr
- [x] Upload design guide to reMarkable
- [x] Phase 1: set up 2026-05-20--book-ocr as a Go module from go-template with book-ocr binary placeholders
- [x] Phase 2: copy OCR page workflow out of scraper into the book-ocr repository and make it compile externally
- [x] Phase 3: copy OCR quality/bookprofile workflow code and the OCR CLI into the book-ocr repository
- [x] Phase 4: wire module dependencies, replace scraper via local path, and run external book-ocr tests
- [x] Phase 5: smoke-test external book-ocr quality-pass against Report 794 artifacts
- [ ] Phase 6: remove OCR packages and ocr-mvp command from scraper after external parity is verified
- [ ] Phase 7: run scraper tests to verify it is workflow/job-runtime only with OCR removed
- [ ] Phase 8: update diary/changelog/design guide with implementation results and commits
