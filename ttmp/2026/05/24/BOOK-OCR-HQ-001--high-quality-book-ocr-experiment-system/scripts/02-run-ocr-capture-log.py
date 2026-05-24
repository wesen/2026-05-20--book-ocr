#!/usr/bin/env python3
"""Run an OCR command while storing stdout/stderr lines directly in SQLite.

This avoids blasting the terminal with provider SSE trace logs. The full output is
still preserved as rows in SQLite for later filtering.

Usage:
  02-run-ocr-capture-log.py OUT_DB -- COMMAND [ARGS...]
"""
from __future__ import annotations

import argparse
import json
import sqlite3
import subprocess
import sys
import time
from pathlib import Path


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument("out_db", type=Path)
    parser.add_argument("command", nargs=argparse.REMAINDER)
    ns = parser.parse_args()
    if ns.command and ns.command[0] == "--":
        ns.command = ns.command[1:]
    if not ns.command:
        parser.error("command required after --")
    return ns


def init_db(path: Path) -> sqlite3.Connection:
    path.parent.mkdir(parents=True, exist_ok=True)
    conn = sqlite3.connect(path)
    conn.execute("drop table if exists process_lines")
    conn.execute(
        """
        create table process_lines (
            line_no integer primary key,
            stream text not null,
            captured_at real not null,
            level text,
            event text,
            workflow_id text,
            op_id text,
            attempt integer,
            workflow_status text,
            message text,
            error_code text,
            raw text not null,
            parsed integer not null
        )
        """
    )
    conn.execute("drop table if exists process_meta")
    conn.execute("create table process_meta (key text primary key, value text not null)")
    conn.execute("create index idx_process_level on process_lines(level)")
    conn.execute("create index idx_process_event on process_lines(event)")
    conn.execute("create index idx_process_op on process_lines(op_id)")
    return conn


def row_from_line(line_no: int, stream: str, line: str):
    raw = line.rstrip("\n")
    try:
        obj = json.loads(raw)
        parsed = 1
    except json.JSONDecodeError:
        obj = {}
        parsed = 0
    return (
        line_no,
        stream,
        time.time(),
        obj.get("level"),
        obj.get("event"),
        obj.get("workflow_id"),
        obj.get("op_id"),
        obj.get("attempt"),
        obj.get("workflow_status"),
        obj.get("message") or (raw if not parsed else None),
        obj.get("error_code"),
        raw,
        parsed,
    )


def main() -> int:
    ns = parse_args()
    conn = init_db(ns.out_db)
    conn.execute("insert into process_meta values (?, ?)", ("command", json.dumps(ns.command)))
    conn.commit()

    proc = subprocess.Popen(
        ns.command,
        stdout=subprocess.PIPE,
        stderr=subprocess.STDOUT,
        text=True,
        bufsize=1,
    )
    assert proc.stdout is not None
    line_no = 0
    for line in proc.stdout:
        line_no += 1
        conn.execute(
            "insert into process_lines values (?,?,?,?,?,?,?,?,?,?,?,?,?)",
            row_from_line(line_no, "stdout", line),
        )
        if line_no % 100 == 0:
            conn.commit()
        # Print only compact, non-trace progress.
        try:
            obj = json.loads(line)
        except json.JSONDecodeError:
            if line.startswith("status=") or line.startswith("started run") or line.startswith("assemble result"):
                print(line.rstrip())
            continue
        if obj.get("level") != "trace" and obj.get("event"):
            print(
                f"{line_no}: {obj.get('level','')} {obj.get('event','')} "
                f"op={obj.get('op_id','')} attempt={obj.get('attempt','')} "
                f"msg={obj.get('message','')}"
            )
    rc = proc.wait()
    conn.execute("insert or replace into process_meta values (?, ?)", ("returncode", str(rc)))
    conn.execute("insert or replace into process_meta values (?, ?)", ("line_count", str(line_no)))
    conn.commit()
    conn.close()
    return rc


if __name__ == "__main__":
    raise SystemExit(main())
