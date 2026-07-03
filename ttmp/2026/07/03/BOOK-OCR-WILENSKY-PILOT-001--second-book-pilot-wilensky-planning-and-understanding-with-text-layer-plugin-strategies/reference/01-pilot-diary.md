---
Title: Pilot diary
Ticket: BOOK-OCR-WILENSKY-PILOT-001
Status: active
Topics:
    - book-ocr
    - ocr
    - workflow
    - productization
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: Chronological diary of the Wilensky second-book pilot - PDF analysis, text-layer plugin build, the pandoc backslash failure, three strategy runs, and the hybrid schema-drift diagnosis.
LastUpdated: 2026-07-03T13:33:40.062590096-07:00
WhatFor: ""
WhenToUse: ""
---


# Pilot diary

## Goal

Record the Wilensky *Planning and Understanding* pilot: analyzing the Internet Archive PDF, building the text-layer plugin, running three OCR strategies over the first 24 pages, and diagnosing what broke.

## Step 1: PDF analysis

The source is a 200-page IA/LuraDocument digitization (2012) of the 1983 book: scanned page images with a hidden Courier OCR text layer. Sampling with pdftotext showed near-clean prose (page 18 preface), but also the defect signatures that shaped the plugin: literal backslashes for unrecognized 'v' ("reser\ed"), punctuation-only noise lines ("- . ,"), running headers and folios interleaved with content, and no structure of any kind.

### Prompt Context

**User prompt (verbatim):** "Try @~/Downloads/Planning and understanding a computational approach to human reasoning (Wilensky, Robert, 1951-) (z-library.sk, 1lib.sk, z-lib.sk).pdf . At least the first 20 pages. This one already has text extracted in it which might be useful as well. Potentially something that could be leveraged using a plugin. Analyze the pdf a bit and then create a new docmgr ticket and Create a detailed analysis / design / implementation guide that is for a new intern, explaining all the parts of the system needed to understand what it is, with prose paragraphs and bullet point sand pseudocode and diagrams and api references and file references. It should be very clear and technical. Store in the ticket and the nupload to remarkable."

**Assistant interpretation:** Run the second-book pilot on this PDF (>=20 pages), design a plugin that leverages the embedded text layer, and deliver an intern guide in a new ticket, uploaded to reMarkable.

**Inferred user intent:** Validate the whole productization stack (profiles, plugins, ingest) on a real foreign book, with the text layer as the plugin showcase.

### What I did
- pdfinfo/pdffonts/pdftotext sampling on pages 15, 18, 30, 45; confirmed gpt-5-mini-low still exists in the pinocchio registry.
- qpdf extracted pages 1-24 into /tmp/wilensky-pilot/wilensky-p1-24.pdf; book-ocr ingest rasterized them at 300 dpi grayscale.

### What worked / What didn't work
- Everything in this step worked; the text-layer quality assessment held up in the runs.

### What I learned
- IA text layers are transcription drafts, not structured documents - the strategy split (free prose source vs structure source) fell directly out of the sampling.

## Step 2: The text-layer plugin and the pandoc failure

One Python plugin (scripts/textlayer_plugin.py) serves two seams: ocr.page (pdftotext -> cleanup -> paragraph/heading blocks) and prompt.render (cleaned draft embedded in the prompt as untrusted reference). The first full run of variant A failed at the very last step: pandoc/XeLaTeX died on "reser\ed" - the text layer's literal backslash reached assembled.md and became a LaTeX control sequence.

### What I did
- Surveyed backslash occurrences (two in 24 pages, both misread 'v'); added backslash->v to the plugin cleanup (both a correction and a sanitizer); reran variant A to completion.
- The --plugin flag can't carry plugin arguments, so the run script generates a wrapper baking in --pdf (finding W5).

### What didn't work
- First A run: op assemble-structured-markdown failed structured_pdf_render_failed after all 24 pages had succeeded; it also retried the render once, pointlessly (render failures are deterministic - noted in W2).

### What I learned
- The renderer performs no Markdown/LaTeX escaping; any client can break the PDF step two stages downstream (finding W2, host fix specified in the design doc).

## Step 3: Three runs

- A textlayer: 24/24, zero warnings, book.pdf, zero model calls.
- B vlm (gpt-5-mini-low, live): 24/24, zero warnings, 24 persisted turns.
- C hybrid (gpt-5-mini-low + prompt.render draft): 24/24 "succeeded", zero warnings - but see Step 4.

### What worked
- The zero-Go-changes claim held: profile-free generic run, plugin bindings by flag, ingest layout consumed by the F4-fixed figure/discovery paths.

## Step 4: Comparison and the hybrid diagnosis

Prose pages converge across all three variants to within a few bytes (pages 17-19, 23). Structured pages diverge: textlayer over-detects headings (102 vs ~30 - its ALL-CAPS rule fires on contents rows); hybrid collapses on the title and copyright pages (30 and 393 bytes vs ~1,780 and ~940).

The hybrid collapse traced to schema drift: the model emitted heading blocks with a "lines": [...] field instead of "text". The prompt.render plugin had replaced the built-in prompt's detailed field-shape instruction, the host's appended contract names only block types, the tolerant decoder ignored the unknown field, and no validation flags empty headings. Silent end to end (finding W3, two-part fix specified).

### What warrants a second pair of eyes
- The W3 fix split (fatten the appended contract AND accept a lines alias in the decoder AND warn on empty text) - agreeing on all three vs a subset changes how safe prompt experiments are.

### What should be done in the future
- Rerun variant C after the W3 contract fix; sample a figure-bearing chapter; promote the plugin to examples/plugins/; route pages via page.classify using text-layer anomaly signals (W1).

### Code review instructions
- scripts/textlayer_plugin.py (cleanup rules), scripts/01-run-pilot.sh, scripts/02-compare-variants.sh; reproduce with variant A (free) and read /tmp/wilensky-pilot/run-hybrid/pages/page_002/03-raw-response.json for the W3 evidence.
