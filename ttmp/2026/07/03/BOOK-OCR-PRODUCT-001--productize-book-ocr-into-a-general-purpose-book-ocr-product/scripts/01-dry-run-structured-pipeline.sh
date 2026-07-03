#!/usr/bin/env bash
# Experiment 01: verify the structured OCR pipeline end-to-end in dry-run mode.
#
# Runs the structured workflow over the first 3 pages of Report 794 with the
# deterministic fake OCR backend (no model calls), renders the PDF, and checks
# that every documented artifact is actually produced. This is the baseline
# health check for the productization ticket: if this passes, the workflow
# runtime, renderer, assembly, validation, and Pandoc toolchain all work.
set -euo pipefail

REPO=/home/manuel/code/wesen/2026-05-20--book-ocr

# Finding F1 (go.mod `replace => ../scraper`) was fixed on 2026-07-03: the repo
# now requires the published github.com/go-go-golems/scraper v0.0.4 and builds
# standalone. book-ocr.go.work is kept only as a template for developing against
# a local scraper checkout (GOWORK=$SCRIPT_DIR/book-ocr.go.work).
export GOWORK=off
IMAGES=/home/manuel/code/wesen/claw-stuff/output/books/presentation-based-uis/pages
WORK=$(mktemp -d /tmp/book-ocr-product-001-dryrun-XXXX)

echo "work dir: $WORK"
cd "$REPO"

go run ./cmd/book-ocr structured-run \
  --book-id product-001-dryrun \
  --image-dir "$IMAGES" \
  --start-page 1 \
  --end-page 3 \
  --work-dir "$WORK" \
  --dry-run \
  --expected-pages 3 \
  --embed-figures \
  --render-pdf \
  --max-workers 2 \
  --log-level warn

echo "--- artifact check ---"
fail=0
for f in assembled.md embedded-figures.md book.pdf validation-report.json engine.db turns.db \
         pages/page_001/04-structured.json pages/page_001/05-rendered.md pages/page_001/06-validation.json; do
  if [[ -s "$WORK/$f" ]]; then
    echo "OK   $f ($(stat -c%s "$WORK/$f") bytes)"
  else
    echo "MISS $f"
    fail=1
  fi
done

echo "--- validation report summary ---"
python3 -c "import json,sys; d=json.load(open('$WORK/validation-report.json')); print(json.dumps(d, indent=2)[:1200])" || true

exit $fail
