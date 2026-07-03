#!/usr/bin/env bash
# Wilensky pilot: run the first 24 pages through three OCR strategies.
#
#   A  textlayer   ocr.page plugin, zero model calls (free baseline)
#   B  vlm         pure vision model (gpt-5-mini-low), the May approach
#   C  hybrid      vision model + prompt.render plugin embedding the PDF
#                  text layer as a draft to correct
#
# Usage: 01-run-pilot.sh [A|B|C|all]   (default: A)
set -euo pipefail

REPO=/home/manuel/code/wesen/2026-05-20--book-ocr
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PILOT=/tmp/wilensky-pilot
SUBSET="$PILOT/wilensky-p1-24.pdf"
PAGES="$PILOT/pages"
PLUGIN="$SCRIPT_DIR/textlayer_plugin.py"
PROFILE=gpt-5-mini-low
WHICH="${1:-A}"

cd "$REPO"
chmod +x "$PLUGIN"

if [[ ! -f "$SUBSET" ]]; then
  echo "subset PDF missing; create with:" >&2
  echo "  qpdf <full.pdf> --pages . 1-24 -- $SUBSET" >&2
  exit 1
fi
if [[ ! -d "$PAGES" ]]; then
  go run ./cmd/book-ocr ingest --pdf "$SUBSET" --out-dir "$PAGES" --dpi 300 --grayscale
fi

# The --plugin seam=path CLI flag carries no plugin arguments (profiles do,
# via args:), so wrap the python plugin with its --pdf argument baked in.
WRAPPER="$PILOT/textlayer-wrapper.sh"
cat > "$WRAPPER" <<WRAP
#!/usr/bin/env bash
exec python3 "$PLUGIN" --pdf "$SUBSET"
WRAP
chmod +x "$WRAPPER"

run() {
  local name="$1"; shift
  local work="$PILOT/run-$name"
  rm -rf "$work"
  go run ./cmd/book-ocr structured-run \
    --book-id "wilensky-$name" \
    --image-dir "$PAGES" \
    --start-page 1 --end-page 24 \
    --work-dir "$work" \
    --expected-pages 24 \
    --render-pdf \
    --max-workers 4 \
    --log-level warn \
    "$@"
  echo "== $name done: $work/assembled.md  $work/book.pdf"
}

case "$WHICH" in
  A|all) run textlayer --plugin "ocr.page=$WRAPPER" ;;&
  B|all) run vlm --profile "$PROFILE" ;;&
  C|all) run hybrid --profile "$PROFILE" --plugin "prompt.render=$WRAPPER" ;;&
esac
