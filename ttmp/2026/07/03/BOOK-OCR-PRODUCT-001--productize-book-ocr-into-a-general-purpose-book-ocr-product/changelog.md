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


## 2026-07-03

Implementation session: golden renderer+prompt harness; Phase-1 wins (live-by-default with profile guard, usage fix, engine schema guard for rerun); plugin track P1 (internal/plugin on devctl runtime, ocr.page/prompt.render/figures.segment seams, --plugin CLI + profile plugins section, tests + E2E smoke); Phase 2 core (PolicyFromProfile, profiles/report-794.yaml + generic profile, byte-identical equivalence proofs, second-book E2E with zero Go changes); Phase 3 core (ingest via pdftoppm + report); CI build+vet+tests. Seven code commits.

### Related Files

- /home/manuel/code/wesen/2026-05-20--book-ocr/cmd/book-ocr/phase3.go — ingest and report commands
- /home/manuel/code/wesen/2026-05-20--book-ocr/internal/ocrpipeline/policy.go — Profile-to-policy compiler
- /home/manuel/code/wesen/2026-05-20--book-ocr/internal/plugin/manager.go — Plugin host core


## 2026-07-03

Second implementation session: F4 page-naming fix (glob-based image resolution + wide markers), book-ocr init (PDF to drafted profile workspace), releasable goreleaser + local v0.1.0 tag, cobra CLI tree (flags unchanged), plugin track P2 (response.parse with decline-to-builtin, tagged validate.page/book, page.classify with per-page strategy routing, plugin retryable-hint classification). Five code commits; E2E smoke: profile-declared plugin set routed page 2 to an alternate OCR strategy with book-validator warning in the report.

### Related Files

- /home/manuel/code/wesen/2026-05-20--book-ocr/cmd/book-ocr/root.go — Cobra command tree
- /home/manuel/code/wesen/2026-05-20--book-ocr/internal/ocrpipeline/hooks.go — Seam interfaces incl. RetryHinter
- /home/manuel/code/wesen/2026-05-20--book-ocr/internal/plugin/adapters_p2.go — P2 seam adapters


## 2026-07-03

Released v0.1.0 (OSS goreleaser, linux amd64+arm64 deb/rpm/tar.gz) after converting the Pro-only release pipeline; verified scraper's operator web UI serves a book-ocr engine.db via scraper api serve --engine-db.


## 2026-07-03

Added design doc 03: goja scripting as the in-process extension surface. Key findings: capability-allowlist sandbox by construction (xgoja app posture); geppetto JS module (in the already-pinned v0.11.28) provides LLM/VLM calls with user credentials + turn-store parity; goja-text supplies the parse/validate toolkit; interrupt watchdog must be host-built; no memory cap (honest gap). Scripts mirror the plugin seams one-to-one; plugins remain for polyglot/CV and memory-isolated hosted execution. Plan G1-G4.


## 2026-07-03

Design doc 03 amended: layered script context (page / book / run) staged by DAG determinism — source-derived book context (any page's text layer) available everywhere; run-derived cross-page output only in post-assembly hooks; new postProcessBook seam as the long-deferred second-pass cleanup home (hyphenation, running headers via frequency stats, figure-number continuity).


## 2026-07-03

Added glossary reference doc: borrowed terms with source literatures (seam/Feathers, oracle/testing theory, capability security, fencing/DDIA, policy-mechanism), project coinages with reasoning (source-derived vs run-derived, money bug, structure-blind, decline-to-builtin, driver, agent-first), and the compressed writing patterns unpacked (falls out, X-not-Y contrast, survives-intact-and-gets-sharper, worth naming, named-not-solved).

