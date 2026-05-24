#!/usr/bin/env python3
"""Load noisy NDJSON/console logs into SQLite and emit compact OCR run summaries.

Usage:
  01-filter-ndjson-log-to-sqlite.py LOGFILE OUT_DB [SUMMARY_MD]

The OCR provider emits many trace-level SSE delta rows. This script keeps all
parseable JSON rows in SQLite but makes it easy to query only workflow events,
warnings/errors, or non-trace rows.
"""
from __future__ import annotations

import json
import sqlite3
import sys
from collections import Counter
from pathlib import Path


def main() -> int:
    if len(sys.argv) < 3:
        print("usage: 01-filter-ndjson-log-to-sqlite.py LOGFILE OUT_DB [SUMMARY_MD]", file=sys.stderr)
        return 2
    log_path = Path(sys.argv[1])
    db_path = Path(sys.argv[2])
    summary_path = Path(sys.argv[3]) if len(sys.argv) > 3 else None
    db_path.parent.mkdir(parents=True, exist_ok=True)

    conn = sqlite3.connect(db_path)
    conn.execute("drop table if exists log_events")
    conn.execute(
        """
        create table log_events (
            line_no integer primary key,
            level text,
            event text,
            workflow_id text,
            op_id text,
            site text,
            queue text,
            attempt integer,
            workflow_status text,
            message text,
            error_code text,
            retryable text,
            time text,
            raw text not null,
            parsed integer not null
        )
        """
    )
    conn.execute("create index idx_log_level on log_events(level)")
    conn.execute("create index idx_log_event on log_events(event)")
    conn.execute("create index idx_log_op on log_events(op_id)")
    conn.execute("create index idx_log_workflow on log_events(workflow_id)")

    rows = []
    for line_no, line in enumerate(log_path.read_text(errors="replace").splitlines(), 1):
        raw = line.rstrip("\n")
        try:
            obj = json.loads(raw)
            parsed = 1
        except json.JSONDecodeError:
            obj = {}
            parsed = 0
        rows.append(
            (
                line_no,
                obj.get("level"),
                obj.get("event"),
                obj.get("workflow_id"),
                obj.get("op_id"),
                obj.get("site"),
                obj.get("queue"),
                obj.get("attempt"),
                obj.get("workflow_status"),
                obj.get("message") or (raw if not parsed else None),
                obj.get("error_code"),
                str(obj.get("retryable")) if "retryable" in obj else None,
                obj.get("time"),
                raw,
                parsed,
            )
        )
    conn.executemany("insert into log_events values (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)", rows)
    conn.commit()

    total = conn.execute("select count(*) from log_events").fetchone()[0]
    by_level = Counter(dict(conn.execute("select level, count(*) from log_events group by level")))
    workflow_events = conn.execute(
        """
        select line_no, time, level, event, workflow_id, op_id, attempt, message, error_code
        from log_events
        where coalesce(event, '') != ''
          and level != 'trace'
        order by line_no
        """
    ).fetchall()
    failures = conn.execute(
        """
        select line_no, time, op_id, attempt, error_code, message
        from log_events
        where level in ('warn','error') or event like '%failed%'
        order by line_no
        """
    ).fetchall()

    summary = []
    summary.append(f"# Filtered log summary for `{log_path.name}`\n")
    summary.append(f"- Total lines loaded: {total}")
    for level, count in sorted(by_level.items(), key=lambda kv: str(kv[0])):
        summary.append(f"- `{level}` lines: {count}")
    summary.append(f"- Non-trace workflow events: {len(workflow_events)}")
    summary.append(f"- Warning/error/failure rows: {len(failures)}")
    summary.append("\n## Useful SQLite queries\n")
    summary.append("```sql")
    summary.append("-- Compact workflow timeline")
    summary.append("select line_no, time, level, event, op_id, attempt, message from log_events where coalesce(event,'') != '' and level != 'trace' order by line_no;")
    summary.append("\n-- Failures only")
    summary.append("select line_no, time, op_id, attempt, error_code, message from log_events where level in ('warn','error') or event like '%failed%' order by line_no;")
    summary.append("\n-- Suppress SSE trace deltas")
    summary.append("select line_no, level, event, op_id, message from log_events where level != 'trace' order by line_no;")
    summary.append("```")
    summary.append("\n## First 80 compact workflow events\n")
    summary.append("```text")
    for row in workflow_events[:80]:
        line_no, t, level, event, workflow_id, op_id, attempt, message, error_code = row
        summary.append(f"{line_no}: {t or ''} {level or ''} {event or ''} op={op_id or ''} attempt={attempt or ''} code={error_code or ''} msg={message or ''}")
    summary.append("```")
    if failures:
        summary.append("\n## Failures\n")
        summary.append("```text")
        for row in failures[:120]:
            line_no, t, op_id, attempt, error_code, message = row
            summary.append(f"{line_no}: {t or ''} op={op_id or ''} attempt={attempt or ''} code={error_code or ''} msg={message or ''}")
        summary.append("```")

    text = "\n".join(summary) + "\n"
    if summary_path:
        summary_path.parent.mkdir(parents=True, exist_ok=True)
        summary_path.write_text(text)
    else:
        print(text)
    conn.close()
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
