---
docType: reference
status: active
intent: short-term
topics:
  - ocr
  - experiments
created: 2026-05-24
updated: 2026-05-24
---

# Filtered log summary for `run.log`

- Total lines loaded: 2216
- `None` lines: 9
- `debug` lines: 50
- `info` lines: 27
- `trace` lines: 2130
- Non-trace workflow events: 27
- Warning/error/failure rows: 0

## Useful SQLite queries

```sql
-- Compact workflow timeline
select line_no, time, level, event, op_id, attempt, message from log_events where coalesce(event,'') != '' and level != 'trace' order by line_no;

-- Failures only
select line_no, time, op_id, attempt, error_code, message from log_events where level in ('warn','error') or event like '%failed%' order by line_no;

-- Suppress SSE trace deltas
select line_no, level, event, op_id, message from log_events where level != 'trace' order by line_no;
```

## First 80 compact workflow events

```text
1: 2026-05-24T19:07:10-04:00 info workflow_created op= attempt= code= msg=workflow created
2: 2026-05-24T19:07:10-04:00 info workflow_updated op= attempt= code= msg=workflow status updated
4: 2026-05-24T19:07:10-04:00 info op_leased op=discover-pages attempt=1 code= msg=leased op
5: 2026-05-24T19:07:10-04:00 info op_succeeded op=discover-pages attempt=1 code= msg=op succeeded
7: 2026-05-24T19:07:10-04:00 info op_leased op=ocr-page-001 attempt=1 code= msg=leased op
41: 2026-05-24T19:07:18-04:00 info op_succeeded op=ocr-page-001 attempt=1 code= msg=op succeeded
42: 2026-05-24T19:07:18-04:00 info op_leased op=ocr-page-002 attempt=1 code= msg=leased op
55: 2026-05-24T19:07:20-04:00 info op_succeeded op=ocr-page-002 attempt=1 code= msg=op succeeded
57: 2026-05-24T19:07:21-04:00 info op_leased op=ocr-page-003 attempt=1 code= msg=leased op
242: 2026-05-24T19:07:26-04:00 info op_succeeded op=ocr-page-003 attempt=1 code= msg=op succeeded
243: 2026-05-24T19:07:26-04:00 info op_leased op=ocr-page-004 attempt=1 code= msg=leased op
599: 2026-05-24T19:07:35-04:00 info op_succeeded op=ocr-page-004 attempt=1 code= msg=op succeeded
601: 2026-05-24T19:07:35-04:00 info op_leased op=ocr-page-005 attempt=1 code= msg=leased op
608: 2026-05-24T19:07:35-04:00 info op_retried op=ocr-page-005 attempt=1 code= msg=op scheduled for retry at 2026-05-24T23:07:36.569332926Z
609: 2026-05-24T19:07:35-04:00 info op_leased op=ocr-page-006 attempt=1 code= msg=leased op
938: 2026-05-24T19:07:45-04:00 info op_succeeded op=ocr-page-006 attempt=1 code= msg=op succeeded
940: 2026-05-24T19:07:45-04:00 info op_leased op=ocr-page-005 attempt=2 code= msg=leased op
1126: 2026-05-24T19:07:54-04:00 info op_succeeded op=ocr-page-005 attempt=2 code= msg=op succeeded
1127: 2026-05-24T19:07:54-04:00 info op_leased op=ocr-page-007 attempt=1 code= msg=leased op
1259: 2026-05-24T19:08:01-04:00 info op_succeeded op=ocr-page-007 attempt=1 code= msg=op succeeded
1261: 2026-05-24T19:08:01-04:00 info op_leased op=ocr-page-008 attempt=1 code= msg=leased op
1852: 2026-05-24T19:08:13-04:00 info op_succeeded op=ocr-page-008 attempt=1 code= msg=op succeeded
1853: 2026-05-24T19:08:13-04:00 info op_leased op=ocr-page-009 attempt=1 code= msg=leased op
2210: 2026-05-24T19:08:20-04:00 info op_succeeded op=ocr-page-009 attempt=1 code= msg=op succeeded
2212: 2026-05-24T19:08:21-04:00 info op_leased op=assemble-markdown attempt=1 code= msg=leased op
2213: 2026-05-24T19:08:21-04:00 info op_succeeded op=assemble-markdown attempt=1 code= msg=op succeeded
2214: 2026-05-24T19:08:21-04:00 info workflow_updated op= attempt= code= msg=workflow status updated
```
