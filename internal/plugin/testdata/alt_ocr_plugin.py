#!/usr/bin/env python3
"""Alternate ocr.page strategy used by the page.classify routing tests."""
import json
import sys


def emit(obj):
    sys.stdout.write(json.dumps(obj) + "\n")
    sys.stdout.flush()


emit({
    "type": "handshake",
    "protocol_version": "v2",
    "plugin_name": "alt-ocr-strategy",
    "capabilities": {"ops": ["ocr.page"]},
})

for line in sys.stdin:
    line = line.strip()
    if not line:
        continue
    req = json.loads(line)
    rid = req.get("request_id", "")
    inp = req.get("input", {})
    if req.get("op") == "ocr.page":
        page = inp.get("page_number", 0)
        emit({"type": "response", "request_id": rid, "ok": True, "output": {
            "page": {
                "schema_version": "structured-ocr/v1",
                "book_id": inp.get("book_id", ""),
                "page_number": page,
                "page_type": inp.get("page_type_hint") or "body",
                "blocks": [{"id": "p%03d-b001" % page, "type": "paragraph",
                            "text": "ALT strategy output for page %d." % page}],
            },
            "engine": {"name": "alt-ocr-strategy"},
        }})
    else:
        emit({"type": "response", "request_id": rid, "ok": False,
              "error": {"code": "E_UNSUPPORTED", "message": "unsupported"}})
