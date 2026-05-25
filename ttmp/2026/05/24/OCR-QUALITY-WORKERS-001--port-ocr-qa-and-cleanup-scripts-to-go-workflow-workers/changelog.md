# Changelog

## 2026-05-24

- Initial workspace created


## 2026-05-24

Created OCR quality worker follow-up ticket, wrote implementation guide/diary, ported Python QA/normalization/log-import ideas into Go workflow workers, added quality-pass CLI, and added context-window OCR plumbing (code commit eb19a4018ef5ebfbc89b730de597e686aeb5303f).

### Related Files

- /home/manuel/workspaces/2026-05-20/book-ocr/scraper/cmd/ocr-mvp/main.go — CLI integration
- /home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/workflows/ocrquality/package.go — Quality workers


## 2026-05-24

Added embedded figure extraction worker, validated crops with vision feedback, and preserved first embedded-figure quality pass outputs (code commit 509c8f5dd2b55e6cf88cd650f6c39896fede5a6d).

### Related Files

- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/OCR-QUALITY-WORKERS-001--port-ocr-qa-and-cleanup-scripts-to-go-workflow-workers/experiments/001-go-quality-pass-embedded-figures/outputs/02-embedded-figures.md — Embedded markdown
- /home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/workflows/ocrquality/figures.go — Figure worker


## 2026-05-24

Added frontmatter to generated OCR quality reports so workflow report artifacts are docmgr-valid Markdown (code commit 5c044e6dff17a0d779a08a5e906fcb18505e5a91).

### Related Files

- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/OCR-QUALITY-WORKERS-001--port-ocr-qa-and-cleanup-scripts-to-go-workflow-workers/experiments/001-go-quality-pass-embedded-figures/outputs/06-quality-report.md — Docmgr-valid generated report artifact
- /home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/workflows/ocrquality/package.go — Quality report frontmatter generation

