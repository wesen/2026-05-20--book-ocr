# Tasks

- [x] Analyze the PDF (provenance, text-layer quality, defect signatures)
- [x] Create ticket + subset PDF (pages 1-24) + ingest at 300 dpi
- [x] Build textlayer plugin (ocr.page + prompt.render seams, deterministic cleanup)
- [x] Run variant A (text-layer only, 0 model calls) — 24/24, zero warnings
- [x] Run variant B (pure VLM, gpt-5-mini-low) — 24/24
- [x] Run variant C (hybrid draft-correction) — 24/24 but schema drift (W3)
- [x] Three-way comparison + findings W1–W5
- [x] Write intern guide; diary; upload to reMarkable

## Follow-ups (feed BOOK-OCR-PRODUCT-001)

- [ ] W2: renderer escapes Markdown/LaTeX-active characters; render failures classified permanent
- [ ] W3: fatten StructuredOCRSchemaContract with field shapes + example; decoder accepts lines alias; empty-text warning in page validation
- [ ] W4: move textlayer cleanup heuristics into profile plugin args
- [ ] W5: decide --plugin arg support vs documented wrapper pattern
- [ ] Rerun variant C post-W3; sample a figure-bearing chapter; consider full 200-page run
- [ ] Promote textlayer plugin to examples/plugins/ with W1's classify-routing sketch
