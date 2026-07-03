# Changelog

## 2026-07-03

- Initial workspace created


## 2026-07-03

Investigated repo status (build fails standalone: F1; tests+E2E dry-run pass with go.work override), wrote productization design doc (findings F1–F9, 6 decision records, 4-phase plan), maintained diary, added reproducible experiments in scripts/.

### Related Files

- /home/manuel/code/wesen/2026-05-20--book-ocr/ttmp/2026/07/03/BOOK-OCR-PRODUCT-001--productize-book-ocr-into-a-general-purpose-book-ocr-product/design-doc/01-book-ocr-productization-analysis-design-and-implementation-guide.md — Primary deliverable


## 2026-07-03

Added plugin-seams design doc (02): 8 NDJSON-stdio seams ranked with op schemas, host architecture importing devctl pkg/protocol+pkg/runtime, decision records D1-D4, P1-P3 plan; proved protocol round-trip with Go host + Python plugin demo in scripts/02-plugin-protocol-demo (PLUGIN_PROTOCOL_DEMO_OK).

### Related Files

- /home/manuel/code/wesen/2026-05-20--book-ocr/ttmp/2026/07/03/BOOK-OCR-PRODUCT-001--productize-book-ocr-into-a-general-purpose-book-ocr-product/design-doc/02-plugin-seams-ndjson-stdio-plugins-for-recompile-free-ocr-experimentation.md — Plugin seams design doc
- /home/manuel/code/wesen/2026-05-20--book-ocr/ttmp/2026/07/03/BOOK-OCR-PRODUCT-001--productize-book-ocr-into-a-general-purpose-book-ocr-product/scripts/02-plugin-protocol-demo/host.go — Prototype NDJSON-stdio host


## 2026-07-03

Fixed F1: dropped replace ../scraper, pinned published scraper v0.0.4 (local checkout was exactly at that tag), go directive to 1.26.4; standalone build + full tests green; dry-run script no longer needs GOWORK.

### Related Files

- /home/manuel/code/wesen/2026-05-20--book-ocr/go.mod — F1 fix — published dependency instead of path replace

