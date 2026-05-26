---
Title: Report 794 First 50 Pages Workflow Backed Structured OCR Live Run W4
DocType: analysis
Topics: [book-processing, ocr, workflow, geppetto, pinocchio]
Created: 2026-05-26
---

# Report 794 First 50 Pages Workflow Backed Structured OCR Live Run W4

## Goal

Validate the workflow-backed structured OCR runner with four concurrent workflow workers from the start. This tests whether the new workflow package can execute page OCR steps in parallel, assemble the result, and write validation/projection artifacts without the old direct CLI artifact-resume loop.

## Command

```bash
go run ./cmd/book-ocr structured-run \
  --book-id report-794-structured-workflow-live-50-w4 \
  --image-dir /home/manuel/code/wesen/claw-stuff/output/books/presentation-based-uis/pages \
  --start-page 1 \
  --end-page 50 \
  --work-dir /tmp/book-ocr-structured-workflow-live-50-w4 \
  --profile gpt-5-mini-low \
  --profile-registries /tmp/book-ocr-hq-001/profiles-clean.yaml \
  --dry-run=false \
  --expected-pages 50 \
  --max-workers 4 \
  --log-level warn
```

## Result

The run succeeded without provider failures or workflow retries.

```text
run_id: book-ocr/structured-f32e3de4-eb1f-49cd-b000-f4364241238b
work_dir: /tmp/book-ocr-structured-workflow-live-50-w4
assembled: /tmp/book-ocr-structured-workflow-live-50-w4/assembled.md
validation: /tmp/book-ocr-structured-workflow-live-50-w4/validation-report.json
projection: /tmp/book-ocr-structured-workflow-live-50-w4/projections/book_ocr_structured.db
engine: /tmp/book-ocr-structured-workflow-live-50-w4/engine.db
turns: /tmp/book-ocr-structured-workflow-live-50-w4/turns.db
```

## Counts

```text
page markers: 50
assembled bytes: 78,976
table markdown lines: 74
structured table blocks: 10
structured figure blocks: 17
figure blocks with captions: 17
validation warnings: 0
projection structured_pages: succeeded=50
engine ops: succeeded=53
turn rows: 50
turn phase memberships: input=200, final=200
```

The expected engine graph size is 53 ops:

```text
1 discover step
50 structured page OCR steps
1 assemble step
1 validate step
```

## Figure Captions Detected

```text
page 013 Figure 1-1: A Rudimentary User Interface
page 015 Figure 1-2: The Representation Shift Model
page 017 Figure 1-3: The Primitive Presentation System (PPS) Model
page 021 Figure 1-4: Structure of PSBase
page 031 Figure 2-1: The Primitive Presentation System (PPS) Model
page 032 Figure 2-2: PPSCalc -- Formula Display
page 032 Figure 2-3: PPSCalc -- Value Display
page 033 Figure 2-4: PPSCalc -- After Editing
page 033 Figure 2-5: PPSCalc -- After Recalculation
page 033 Figure 2-6: PPSCalc -- New Formulas
page 034 Figure 2-7: PPSCalc -- Values of New Formulas
page 036 Figure 2-8: World Model
page 042 Figure 2-9: Presenter Parts
page 046 Figure 2-10: Recognizer Parts
page 047 Figure 2-11: PPSCalc -- Value Moved
page 048 Figure 2-12: PPSCalc -- Formula Moved
page 048 Figure 2-13: PPSCalc -- Preparing to Copy Formula
```

## Interpretation

Four concurrent workflow workers worked for this 50-page live run. The workflow-backed implementation handled dynamic page step execution, page artifacts, external workflow artifacts, projection rows, assembly, validation, and turn persistence successfully.

No automatic retry was observed in this run because no page step failed. Automatic retry behavior is nevertheless configured on structured page OCR steps through the workflow retry policy. A future controlled test should induce a retryable fake client failure to prove retry accounting deterministically without depending on provider instability.

## Follow-ups

- Add an automated test/fake client that fails once with a retryable error and then succeeds, so retry behavior is proven deterministically.
- Add a `structured pages` status command or extend `pages` to query `structured_pages`.
- Continue Phase 6 hardening: figure embedding, prose completeness validation, and full-book acceptance gates.
