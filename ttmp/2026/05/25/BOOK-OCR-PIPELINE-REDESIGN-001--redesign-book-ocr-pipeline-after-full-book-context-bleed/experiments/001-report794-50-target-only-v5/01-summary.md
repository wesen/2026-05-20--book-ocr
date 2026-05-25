---
Title: Report 794 first-50 target-only OCR rerun
Ticket: BOOK-OCR-PIPELINE-REDESIGN-001
Status: complete
Topics:
  - ocr
  - book-processing
  - geppetto
DocType: experiment
Intent: evidence
---

# Report 794 first-50 target-only OCR rerun

## Purpose

Run the existing Book OCR workflow over the first 50 page images of Report 794 with the safer target-page-only setting (`--context-window 0`) and the figure-aware prompt. This is not the new structured OCR pipeline yet; it is an evidence run using the current workflow after the VLM separation benchmark clarified that neighboring page images should not be part of primary OCR.

## Command

```bash
go run ./cmd/book-ocr run \
  --book-id report-794-50-target-only-v5 \
  --image-dir /home/manuel/code/wesen/claw-stuff/output/books/presentation-based-uis/pages \
  --start-page 1 \
  --end-page 50 \
  --work-dir /tmp/book-ocr-report794-50-target-only/work \
  --profile gpt-5-mini-low \
  --profile-registries /tmp/book-ocr-hq-001/profiles-clean.yaml \
  --prompt-version ocr-quality-v5-figure-aware \
  --context-window 0 \
  --max-workers 2 \
  --log-level warn
```

A first attempt piped to `tee /tmp/book-ocr-report794-50-target-only/run.log` before creating the parent directory, so `tee` returned `No such file or directory` and the shell command exited non-zero. The workflow itself still completed successfully.

## Run result

```json
{
  "run_id": "ocr-mvp-d8701e29-d511-4d6a-9860-a44b75be1b20",
  "book_id": "report-794-50-target-only-v5",
  "page_count": 50,
  "char_count": 85873,
  "markdown_uri": "file:///tmp/book-ocr-report794-50-target-only/work/artifacts/assemble-markdown/artifact/001"
}
```

Copied raw Markdown:

```text
/tmp/book-ocr-report794-50-target-only/outputs/01-raw.md
```

## Quality pass

```bash
go run ./cmd/book-ocr quality-pass \
  --markdown /tmp/book-ocr-report794-50-target-only/outputs/01-raw.md \
  --output-dir /tmp/book-ocr-report794-50-target-only/outputs/quality-pass \
  --work-dir /tmp/book-ocr-report794-50-target-only/quality-work \
  --book-id report-794-50-target-only-v5 \
  --expected-pages 50 \
  --image-dir /home/manuel/code/wesen/claw-stuff/output/books/presentation-based-uis/pages \
  --embed-figures \
  --max-workers 2
```

Final review artifact opened with `md-view`:

```text
/tmp/book-ocr-report794-50-target-only/outputs/quality-pass/embedded-figures.md
```

`md-view` URL:

```text
http://localhost:38789/render?file=/tmp/book-ocr-report794-50-target-only/outputs/quality-pass/embedded-figures.md
```

## Counts

| Artifact | Bytes | Page markers | Markdown image links |
|---|---:|---:|---:|
| `01-raw.md` | 85,873 | 50 | 0 |
| `normalized.md` | 81,130 | 50 | 0 |
| `embedded-figures.md` | 81,555 | 50 | 17 |

Figure extraction output:

```text
PNG/debug PNG/sidecar files under /tmp/book-ocr-report794-50-target-only/outputs/quality-pass/figures
figure PNG count: 17
figure JSON sidecar count: 17
```

## Spot checks

- Page 12 contains prose referencing Figure 1-1 but has no figure image link.
- Page 13 contains Figure 1-1 and links `figures/page_013_figure_01.png`.
- Page 31 contains Figure 2-1 and links `figures/page_031_figure_01.png`.
- Page 32 contains Figures 2-2 and 2-3 and links two page-032 figure crops.
- Page 42 contains Figure 2-9 and links `figures/page_042_figure_01.png`.
- Page 43 is prose and has no Figure 2-9 image link.
- No adjacent duplicate rendered figure captions were detected in the first 50 pages by the quick caption scan.

## Interpretation

The target-page-only rerun materially improves the specific context-bleed symptom that motivated the redesign. The first-50 output still comes from the older freeform OCR workflow and therefore still includes diagram label text in the Markdown after image links. The structured pipeline work remains necessary to make final rendering deterministic and to move diagram text into sidecars/debug output by default.
