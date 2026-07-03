#!/usr/bin/env python3
"""Deterministic test plugin implementing all three book-ocr seams."""
import json
import sys


def emit(obj):
    sys.stdout.write(json.dumps(obj) + "\n")
    sys.stdout.flush()


emit({
    "type": "handshake",
    "protocol_version": "v2",
    "plugin_name": "test-plugin",
    "capabilities": {"ops": ["ocr.page", "prompt.render", "figures.segment",
                             "response.parse", "validate.page", "validate.book",
                             "page.classify"]},
    "declares": {"version": "test-1"},
})

for line in sys.stdin:
    line = line.strip()
    if not line:
        continue
    req = json.loads(line)
    rid = req.get("request_id", "")
    op = req.get("op", "")
    inp = req.get("input", {})

    if op == "prompt.render":
        emit({"type": "response", "request_id": rid, "ok": True, "output": {
            "system": "TEST SYSTEM PROMPT",
            "user": "TEST USER PROMPT for page %d" % inp.get("page_number", 0),
        }})
    elif op == "ocr.page":
        page = inp.get("page_number", 0)
        if inp.get("book_id") == "wrong-page-book":
            page += 1  # deliberately violate the page-number gate
        emit({"type": "response", "request_id": rid, "ok": True, "output": {
            "page": {
                "schema_version": "structured-ocr/v1",
                "book_id": inp.get("book_id", ""),
                "page_number": page,
                "page_type": "body",
                "blocks": [{"id": "p%03d-b001" % page, "type": "paragraph",
                            "text": "Plugin OCR output for page %d." % page},
                           {"id": "p%03d-b002" % page, "type": "code",
                            "text": "print('page %d')" % page}],
            },
            "engine": {"name": "test-plugin", "version": "test-1"},
        }})
    elif op == "figures.segment":
        emit({"type": "response", "request_id": rid, "ok": True, "output": {
            "crop": {"x": 10, "y": 20, "width": 30, "height": 40},
            "method": "test-seg-v1",
        }})
    elif op == "response.parse":
        raw = inp.get("raw_response", "")
        if raw.startswith("@@"):
            # Custom "@@page|text" format only this plugin understands.
            num, _, text = raw[2:].partition("|")
            page = int(num)
            emit({"type": "response", "request_id": rid, "ok": True, "output": {
                "page": {
                    "schema_version": "structured-ocr/v1",
                    "book_id": inp.get("book_id", ""),
                    "page_number": page,
                    "page_type": "body",
                    "blocks": [{"id": "p%03d-b001" % page, "type": "paragraph",
                                "text": text.strip()}],
                },
            }})
        else:
            emit({"type": "response", "request_id": rid, "ok": False,
                  "error": {"code": "E_DECLINED",
                            "message": "format not recognized; use builtin parser"}})
    elif op == "validate.page":
        warnings = []
        if "TYPO" in inp.get("markdown", ""):
            warnings.append({"code": "typo_found",
                             "message": "page contains the marker word TYPO",
                             "page": inp.get("page_number", 0)})
        emit({"type": "response", "request_id": rid, "ok": True,
              "output": {"warnings": warnings}})
    elif op == "validate.book":
        emit({"type": "response", "request_id": rid, "ok": True, "output": {
            "warnings": [{"code": "book_checked",
                          "message": "checked %d pages" % len(inp.get("page_numbers", []))}],
        }})
    elif op == "page.classify":
        page = inp.get("page_number", 0)
        out = {"page_type": "body", "confidence": 0.9}
        if page == 1:
            out["page_type"] = "title"
        elif page == 2:
            out["strategy"] = "alt-ocr"
        emit({"type": "response", "request_id": rid, "ok": True, "output": out})
    else:
        emit({"type": "response", "request_id": rid, "ok": False,
              "error": {"code": "E_UNSUPPORTED", "message": "unsupported op: " + op}})
