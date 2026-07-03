#!/usr/bin/env bash
# Compare the three Wilensky pilot runs: size, structure, and per-page bytes.
set -euo pipefail
PILOT=/tmp/wilensky-pilot

printf '%-10s %10s %8s %8s %8s %8s %8s\n' variant md_bytes heads figs tables code blanks
for v in textlayer vlm hybrid; do
  MD="$PILOT/run-$v/assembled.md"
  [[ -f "$MD" ]] || continue
  printf '%-10s %10s %8s %8s %8s %8s %8s\n' "$v" \
    "$(wc -c < "$MD")" \
    "$(grep -c '^#' "$MD" || true)" \
    "$(grep -c '\[FIGURE:\|!\[' "$MD" || true)" \
    "$(grep -c '^|' "$MD" || true)" \
    "$(grep -c '^```' "$MD" || true)" \
    "$(grep -c 'BLANK PAGE' "$MD" || true)"
done

echo
echo "per-page rendered bytes (page: textlayer / vlm / hybrid):"
for p in $(seq 1 24); do
  n=$(printf '%03d' "$p")
  a=$(wc -c < "$PILOT/run-textlayer/pages/page_$n/05-rendered.md" 2>/dev/null || echo -)
  b=$(wc -c < "$PILOT/run-vlm/pages/page_$n/05-rendered.md" 2>/dev/null || echo -)
  c=$(wc -c < "$PILOT/run-hybrid/pages/page_$n/05-rendered.md" 2>/dev/null || echo -)
  printf '  %s: %6s %6s %6s\n' "$n" "$a" "$b" "$c"
done
