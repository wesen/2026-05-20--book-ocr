#!/usr/bin/env python3
"""Test plugin that only advertises prompt.render."""
import json
import sys


def emit(obj):
    sys.stdout.write(json.dumps(obj) + "\n")
    sys.stdout.flush()


emit({
    "type": "handshake",
    "protocol_version": "v2",
    "plugin_name": "prompt-only",
    "capabilities": {"ops": ["prompt.render"]},
})

for line in sys.stdin:
    line = line.strip()
    if not line:
        continue
    req = json.loads(line)
    rid = req.get("request_id", "")
    if req.get("op") == "prompt.render":
        emit({"type": "response", "request_id": rid, "ok": True, "output": {
            "system": "PROMPT ONLY SYSTEM",
            "user": "PROMPT ONLY USER",
        }})
    else:
        emit({"type": "response", "request_id": rid, "ok": False,
              "error": {"code": "E_UNSUPPORTED", "message": "unsupported"}})
