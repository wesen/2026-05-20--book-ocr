#!/usr/bin/env bash
# Experiment 02: prove the NDJSON-stdio plugin protocol round-trip for the
# proposed book-ocr plugin seams (prompt.render, ocr.page) with a Python
# plugin and a stdlib-only Go host. Exit 0 = protocol viable.
set -euo pipefail
DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
IMAGE="${1:-/home/manuel/code/wesen/claw-stuff/output/books/presentation-based-uis/pages/page_012.png}"
chmod +x "$DIR/ocr_plugin.py"
cd "$DIR"
go run host.go "$DIR/ocr_plugin.py" "$IMAGE"
