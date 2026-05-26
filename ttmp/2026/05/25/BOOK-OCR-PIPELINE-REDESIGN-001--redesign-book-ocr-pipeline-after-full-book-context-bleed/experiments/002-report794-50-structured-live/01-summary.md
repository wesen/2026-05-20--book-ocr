---
Title: Report 794 First 50 Pages Structured OCR Live Run
DocType: analysis
Topics: [book-ocr, structured-ocr, geppetto]
Created: 2026-05-25
---

# Report 794 First 50 Pages Structured OCR Live Run

## Goal

Run the new structured OCR pipeline over the first 50 pages of Report 794 and compare it against the previous target-page-only freeform OCR rerun.

## Command

```bash
go run ./cmd/book-ocr structured-run \
  --book-id report-794-structured-50-v1 \
  --image-dir /home/manuel/code/wesen/claw-stuff/output/books/presentation-based-uis/pages \
  --start-page 1 \
  --end-page 50 \
  --work-dir /tmp/book-ocr-structured-50-live \
  --profile gpt-5-mini-low \
  --profile-registries /tmp/book-ocr-hq-001/profiles-clean.yaml \
  --dry-run=false \
  --log-level warn \
  --resume=true
```

The run needed two resumes after transient provider TLS failures:

```text
Post "https://api.openai.com/v1/responses": remote error: tls: bad record MAC
```

The `structured-run --resume=true` behavior reused completed page artifacts and continued from the next missing page.

## Output Artifacts

```text
/tmp/book-ocr-structured-50-live/assembled.md
/tmp/book-ocr-structured-50-live/validation-report.json
/tmp/book-ocr-structured-50-live/turns.db
/tmp/book-ocr-structured-50-live/pages/page_NNN/01-turn-input.yaml
/tmp/book-ocr-structured-50-live/pages/page_NNN/02-turn-final.yaml
/tmp/book-ocr-structured-50-live/pages/page_NNN/03-raw-response.json
/tmp/book-ocr-structured-50-live/pages/page_NNN/04-structured.json
/tmp/book-ocr-structured-50-live/pages/page_NNN/05-rendered.md
/tmp/book-ocr-structured-50-live/pages/page_NNN/06-validation.json
```

Opened with `md-view`:

```text
http://localhost:38789/render?file=/tmp/book-ocr-structured-50-live/assembled.md
```

## Counts

```text
pages: 50
assembled bytes: 79,281
table markdown lines: 45
structured table blocks: 9
structured figure blocks: 17
figure blocks with captions: 17
validation warnings: 0
turns table rows: 50
turn phase membership: input=204, final=204
```

## Figure Boundary Result

The structured target-page-only run did not reproduce the old adjacent-page figure bleed in the checked range. Page-local figure blocks appear on the figure pages, and prose neighbor pages do not get false figure blocks.

Detected figure captions in the first 50 pages:

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

A quick adjacent duplicate caption scan over `assembled.md` produced no adjacent duplicate captions.

## Table Result

Page 32 now renders Markdown tables from structured table blocks. Example from the assembled output:

```markdown
|  | A | B | C |
| --- | --- | --- | --- |
| 1 | 100 | 20 | A1*B1 |
| 2 | 75 | 5 | A2*B2 |
| 3 |  |  | C1+C2 |

|  | A | B | C |
| --- | --- | --- | --- |
| 1 | 100 | 20 | 2000 |
| 2 | 75 | 5 | 375 |
| 3 |  |  | 2375 |
```

## Interpretation

The structured pipeline now demonstrates the core improvement the freeform pipeline could not guarantee: the model returns structured table and figure blocks, and Go renders final Markdown deterministically. The first-50 live result has lower prose volume than the earlier freeform OCR, so it is not yet a final replacement for full-book OCR, but it proves the table and page-boundary architecture.

## Follow-ups

- Add figure image embedding/resolution for structured figure blocks.
- Add richer aggregate validation to `validation-report.json`.
- Add workflow-runtime integration if live structured runs need durable leases/retries rather than CLI resume.
- Improve prompt coverage for prose completeness before a full 202-page structured run.
