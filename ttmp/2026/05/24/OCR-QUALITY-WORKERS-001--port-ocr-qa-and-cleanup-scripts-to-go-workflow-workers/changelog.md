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


## 2026-05-24

Added figure-aware prompt contract and marker-recovery fallback for caption-only full-page diagrams, recovering embedded Figure 1-2 and Figure 1-3 (code commit aea38ef8f52a0354b4ef73acb045bdd1f998d825).

### Related Files

- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/OCR-QUALITY-WORKERS-001--port-ocr-qa-and-cleanup-scripts-to-go-workflow-workers/experiments/002-figure-aware-marker-recovery/outputs/02-embedded-figures.md — Four-figure artifact
- /home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/workflows/ocrmvp/prompt.go — Prompt v5
- /home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/workflows/ocrquality/figures.go — Marker recovery


## 2026-05-24

Added intern-facing generic book OCR analysis/design guide explaining Report 794 specificity and a book-profile implementation plan.

### Related Files

- /home/manuel/workspaces/2026-05-20/book-ocr/2026-05-20--book-ocr/ttmp/2026/05/24/OCR-QUALITY-WORKERS-001--port-ocr-qa-and-cleanup-scripts-to-go-workflow-workers/design-doc/02-generic-book-ocr-system-analysis-and-implementation-guide.md — Genericization analysis and implementation guide


## 2026-05-24

Implemented stable book profile plus machine discovery/profile-patch layer and wired quality-pass to emit discovery artifacts without mutating canonical profiles (code commit c6e5bc2a03a990ee3131b5243110a1fdca95606a).

### Related Files

- /home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/workflows/bookprofile/discovery.go — Discovery and patch proposals
- /home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/workflows/bookprofile/profile.go — Stable profiles
- /home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/workflows/ocrquality/package.go — Quality-pass discovery step


## 2026-05-24

Added figure crop JSON sidecars and debug overlay PNGs so extracted figures are auditable, with smoke test output under /tmp/ocr-quality-sidecar-smoke (code commit d4eb1e36a3a7374fe4425354e9882dd3989b12a0).

### Related Files

- /home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/workflows/bookprofile/discovery.go — Crop metadata in discovery state
- /home/manuel/workspaces/2026-05-20/book-ocr/scraper/pkg/workflows/ocrquality/figures.go — Sidecar/debug overlay generation

