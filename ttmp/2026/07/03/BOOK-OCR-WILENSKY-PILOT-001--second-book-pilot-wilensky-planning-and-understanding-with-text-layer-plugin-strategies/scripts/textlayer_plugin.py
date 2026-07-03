#!/usr/bin/env python3
"""Text-layer plugin for book-ocr (NDJSON-stdio, devctl protocol v2).

The Wilensky scan is an Internet Archive digitization whose PDF carries a
hidden OCR text layer. This plugin leverages that layer through two seams:

  ocr.page       Zero-model-cost strategy: extract the page's text layer via
                 pdftotext, clean it deterministically, and emit paragraph/
                 heading blocks as structured-ocr/v1 JSON.

  prompt.render  Hybrid strategy: build a prompt that embeds the cleaned text
                 layer as a DRAFT the vision model corrects against the page
                 image. The host appends the non-negotiable schema contract,
                 so the draft cannot displace the JSON-only instruction.

Usage (bound via a book profile or --plugin flags):
  textlayer_plugin.py --pdf /path/to/subset.pdf
Page numbers map 1:1 onto PDF pages (the pilot ingests a qpdf page subset,
so ingested page_0001.png is PDF page 1).
"""
import argparse
import json
import re
import subprocess
import sys

parser = argparse.ArgumentParser()
parser.add_argument("--pdf", required=True)
args = parser.parse_args()


def emit(obj):
    sys.stdout.write(json.dumps(obj) + "\n")
    sys.stdout.flush()


def log(msg):
    sys.stderr.write("[textlayer] " + msg + "\n")
    sys.stderr.flush()


NOISE_LINE = re.compile(r"^[\W\d]*$")  # only punctuation/digits/whitespace
HEADING = re.compile(r"^(?:CHAPTER\s+\w+|\d+(?:\.\d+)*\s+\S.*|[A-Z][A-Z\s'&-]{3,60})$")


def extract_page_text(page):
    out = subprocess.run(
        ["pdftotext", "-f", str(page), "-l", str(page), "-layout", args.pdf, "-"],
        capture_output=True, text=True, timeout=60,
    )
    if out.returncode != 0:
        raise RuntimeError("pdftotext failed: " + out.stderr.strip())
    return out.stdout


def clean_lines(raw):
    lines = []
    # The IA OCR in this scan emits a literal backslash for unrecognized 'v'
    # ("reser\ed" for "reserved"); raw backslashes also break the pandoc PDF
    # step downstream, so this repair is both a correction and a sanitizer.
    raw = raw.replace("\\", "v")
    for line in raw.replace("\f", "\n").split("\n"):
        line = line.strip()
        if NOISE_LINE.match(line) and line != "":
            continue  # scanner noise like "- . ," and bare folio numbers
        lines.append(line)
    return lines


def to_blocks(raw, page):
    """Deterministic text-layer cleanup: join hyphenated breaks, split
    paragraphs on blank lines, classify short shouty lines as headings."""
    lines = clean_lines(raw)
    blocks = []
    para = []

    def flush():
        if not para:
            return
        text = " ".join(para)
        text = re.sub(r"(\w)-\s+(\w)", r"\1\2", text)  # de-hyphenate joins
        text = re.sub(r"\s+", " ", text).strip()
        if text:
            blocks.append({"id": "p%03d-b%03d" % (page, len(blocks) + 1),
                           "type": "paragraph", "text": text})
        para.clear()

    for line in lines:
        if line == "":
            flush()
            continue
        if HEADING.match(line) and len(line) < 70:
            flush()
            blocks.append({"id": "p%03d-b%03d" % (page, len(blocks) + 1),
                           "type": "heading", "level": 1 if line.isupper() else 2,
                           "text": line.title() if line.isupper() else line})
            continue
        para.append(line)
    flush()
    return blocks


def draft_text(raw):
    lines = clean_lines(raw)
    return "\n".join(line for line in lines if line != "")


emit({
    "type": "handshake",
    "protocol_version": "v2",
    "plugin_name": "textlayer",
    "capabilities": {"ops": ["ocr.page", "prompt.render"]},
    "declares": {"strategy": "pdf-text-layer", "source": args.pdf},
})

for line in sys.stdin:
    line = line.strip()
    if not line:
        continue
    req = json.loads(line)
    rid = req.get("request_id", "")
    op = req.get("op", "")
    inp = req.get("input", {})
    page = inp.get("page_number", 0)

    try:
        if op == "ocr.page":
            raw = extract_page_text(page)
            blocks = to_blocks(raw, page)
            page_type = "body" if blocks else "blank"
            log("ocr.page %d: %d blocks" % (page, len(blocks)))
            emit({"type": "response", "request_id": rid, "ok": True, "output": {
                "page": {
                    "schema_version": "structured-ocr/v1",
                    "book_id": inp.get("book_id", ""),
                    "page_number": page,
                    "page_type": page_type,
                    "blocks": blocks or [],
                },
                "engine": {"name": "textlayer", "strategy": "pdf-text-layer"},
            }})
        elif op == "prompt.render":
            raw = extract_page_text(page)
            draft = draft_text(raw)
            user = (
                "Transcribe exactly one target page image into structured OCR JSON.\n\n"
                "A DRAFT transcription extracted from the PDF's embedded OCR text layer "
                "follows. It is usually accurate for prose but may contain recognition "
                "errors, broken line flow, scanner noise, and running headers, and it "
                "contains no layout structure. Use the page IMAGE as the source of "
                "truth: correct the draft against the image, restore structure "
                "(headings, lists, tables, figures), and drop running headers and "
                "folio artifacts.\n\n"
                "--- DRAFT (text layer) ---\n" + draft + "\n--- END DRAFT ---"
            )
            emit({"type": "response", "request_id": rid, "ok": True, "output": {
                "system": "You are a precise structured OCR engine for scanned "
                          "technical books. Return strict JSON only. Do not return "
                          "Markdown, commentary, or code fences.",
                "user": user,
            }})
        else:
            emit({"type": "response", "request_id": rid, "ok": False,
                  "error": {"code": "E_UNSUPPORTED", "message": "unsupported op: " + op}})
    except Exception as exc:  # deliberate: any failure -> clean protocol error
        emit({"type": "response", "request_id": rid, "ok": False,
              "error": {"code": "E_RUNTIME", "message": str(exc),
                        "details": {"retryable": False}}})
