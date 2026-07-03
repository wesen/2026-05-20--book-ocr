#!/usr/bin/env python3
"""Experiment 02: minimal book-ocr NDJSON-stdio plugin (devctl protocol v2 shape).

Implements two candidate seams:
  - prompt.render : profile-aware prompt text for a page
  - ocr.page      : full page OCR returning StructuredPageOCR-shaped JSON
                    (fake here; a real plugin would call tesseract/OpenCV/an LLM)

Protocol rules (mirrors devctl):
  - first stdout line is the handshake, then only NDJSON frames
  - logs go to stderr
  - unsupported ops answer ok=false with code E_UNSUPPORTED
"""
import json
import os
import sys


def emit(obj):
    sys.stdout.write(json.dumps(obj) + "\n")
    sys.stdout.flush()


def log(msg):
    sys.stderr.write("[ocr-plugin] " + msg + "\n")
    sys.stderr.flush()


emit({
    "type": "handshake",
    "protocol_version": "v2",
    "plugin_name": "demo-ocr-strategy",
    "plugin_version": "0.0.1",
    "capabilities": {"ops": ["prompt.render", "ocr.page"]},
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
        # A plugin-owned prompt experiment: e.g. terser prompt for novels.
        book = inp.get("book_id", "unknown")
        page = inp.get("page_number", 0)
        emit({
            "type": "response", "request_id": rid, "ok": True,
            "output": {
                "system": "You are a precise structured OCR engine. JSON only.",
                "user": f"Transcribe page {page:03d} of {book}. "
                        "This plugin variant: prose-first, no figure synthesis.",
            },
        })
    elif op == "ocr.page":
        image_path = inp.get("image_path", "")
        page = inp.get("page_number", 0)
        if not os.path.exists(image_path):
            emit({
                "type": "response", "request_id": rid, "ok": False,
                "error": {"code": "E_IMAGE_NOT_FOUND",
                          "message": f"no such image: {image_path}",
                          "retryable": False},
            })
            continue
        size = os.path.getsize(image_path)
        log(f"ocr.page page={page} image={image_path} ({size} bytes)")
        # Progress events interleave with the eventual response.
        emit({"type": "event", "request_id": rid, "event": "progress",
              "data": {"stage": "segmenting", "page": page}})
        emit({
            "type": "response", "request_id": rid, "ok": True,
            "output": {
                "schema_version": "structured-ocr/v1",
                "book_id": inp.get("book_id", "unknown"),
                "page_number": page,
                "page_type": "body",
                "blocks": [
                    {"id": f"p{page:03d}-b001", "type": "paragraph",
                     "text": f"Demo plugin transcription of page {page} "
                             f"(source image: {size} bytes)."},
                ],
                "warnings": [],
                "engine": {"name": "demo-ocr-strategy", "version": "0.0.1"},
            },
        })
    else:
        emit({
            "type": "response", "request_id": rid, "ok": False,
            "error": {"code": "E_UNSUPPORTED", "message": f"unsupported op: {op}"},
        })
