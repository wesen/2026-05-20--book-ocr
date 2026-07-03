#!/usr/bin/env bash
# Structure-rich sample: the pages the pilot's 1-24 range missed.
#   PDF p55  "Task network after failure to resolve the conflict" (diagram, book p.35)
#   PDF p59  "4.6 A Simple PANDORA Example" (code/trace, book p.37)
#   PDF p68  "the process of finding an explanation for an event" (diagram, book p.46)
# Subset = PDF pages 54-60 + 67-69 -> 10 pages renumbered 1-10:
#   subset 2 = diagram p55, subset 6 = code p59, subset 9 = diagram p68.
# Variants: A textlayer (structure-blind baseline), B vlm with --embed-figures.
set -euo pipefail

REPO=/home/manuel/code/wesen/2026-05-20--book-ocr
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PILOT=/tmp/wilensky-pilot
FULL="$HOME/Downloads/Planning and understanding a computational approach to human reasoning (Wilensky, Robert, 1951-) (z-library.sk, 1lib.sk, z-lib.sk).pdf"
SUBSET="$PILOT/wilensky-structure.pdf"
PAGES="$PILOT/pages-structure"
PLUGIN="$SCRIPT_DIR/textlayer_plugin.py"
PROFILE=gpt-5-mini-low
WHICH="${1:-all}"

cd "$REPO"
[[ -f "$SUBSET" ]] || qpdf "$FULL" --pages . 54-60 . 67-69 -- "$SUBSET"
[[ -d "$PAGES" ]] || go run ./cmd/book-ocr ingest --pdf "$SUBSET" --out-dir "$PAGES" --dpi 300 --grayscale

WRAPPER="$PILOT/textlayer-structure-wrapper.sh"
printf '#!/usr/bin/env bash\nexec python3 "%s" --pdf "%s"\n' "$PLUGIN" "$SUBSET" > "$WRAPPER"
chmod +x "$WRAPPER"

run() {
  local name="$1"; shift
  local work="$PILOT/run-structure-$name"
  rm -rf "$work"
  go run ./cmd/book-ocr structured-run \
    --book-id "wilensky-structure-$name" \
    --image-dir "$PAGES" --start-page 1 --end-page 10 \
    --work-dir "$work" --expected-pages 10 \
    --render-pdf --max-workers 4 --log-level warn "$@"
  echo "== $name done: $work"
}

case "$WHICH" in
  A|all) run textlayer --plugin "ocr.page=$WRAPPER" ;;&
  B|all) run vlm --profile "$PROFILE" --embed-figures ;;&
esac
