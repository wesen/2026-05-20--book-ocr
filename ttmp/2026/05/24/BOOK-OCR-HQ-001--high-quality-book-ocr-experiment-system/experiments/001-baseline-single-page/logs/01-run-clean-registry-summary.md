---
docType: reference
status: active
intent: short-term
topics:
  - ocr
  - experiments
created: 2026-05-24
updated: 2026-05-24
title: 'Run Clean Registry Log Summary'
---

# Filtered log summary for `run-clean-registry.log`

- Total lines loaded: 8687
- `None` lines: 20
- `debug` lines: 155
- `info` lines: 69
- `trace` lines: 8443
- Non-trace workflow events: 69
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
1: 2026-05-24T18:47:05-04:00 info workflow_created op= attempt= code= msg=workflow created
2: 2026-05-24T18:47:05-04:00 info workflow_updated op= attempt= code= msg=workflow status updated
4: 2026-05-24T18:47:05-04:00 info op_leased op=discover-pages attempt=1 code= msg=leased op
5: 2026-05-24T18:47:05-04:00 info op_succeeded op=discover-pages attempt=1 code= msg=op succeeded
7: 2026-05-24T18:47:05-04:00 info op_leased op=ocr-page-001 attempt=1 code= msg=leased op
40: 2026-05-24T18:47:11-04:00 info op_succeeded op=ocr-page-001 attempt=1 code= msg=op succeeded
41: 2026-05-24T18:47:11-04:00 info op_leased op=ocr-page-002 attempt=1 code= msg=leased op
58: 2026-05-24T18:47:14-04:00 info op_succeeded op=ocr-page-002 attempt=1 code= msg=op succeeded
60: 2026-05-24T18:47:14-04:00 info op_leased op=ocr-page-003 attempt=1 code= msg=leased op
242: 2026-05-24T18:47:19-04:00 info op_succeeded op=ocr-page-003 attempt=1 code= msg=op succeeded
243: 2026-05-24T18:47:19-04:00 info op_leased op=ocr-page-004 attempt=1 code= msg=leased op
601: 2026-05-24T18:47:31-04:00 info op_succeeded op=ocr-page-004 attempt=1 code= msg=op succeeded
603: 2026-05-24T18:47:31-04:00 info op_leased op=ocr-page-005 attempt=1 code= msg=leased op
789: 2026-05-24T18:47:36-04:00 info op_succeeded op=ocr-page-005 attempt=1 code= msg=op succeeded
790: 2026-05-24T18:47:36-04:00 info op_leased op=ocr-page-006 attempt=1 code= msg=leased op
1133: 2026-05-24T18:47:46-04:00 info op_succeeded op=ocr-page-006 attempt=1 code= msg=op succeeded
1135: 2026-05-24T18:47:46-04:00 info op_leased op=ocr-page-007 attempt=1 code= msg=leased op
1142: 2026-05-24T18:47:47-04:00 info op_retried op=ocr-page-007 attempt=1 code= msg=op scheduled for retry at 2026-05-24T22:47:47.891606975Z
1143: 2026-05-24T18:47:47-04:00 info op_leased op=ocr-page-008 attempt=1 code= msg=leased op
1833: 2026-05-24T18:47:56-04:00 info op_succeeded op=ocr-page-008 attempt=1 code= msg=op succeeded
1835: 2026-05-24T18:47:57-04:00 info op_leased op=ocr-page-007 attempt=2 code= msg=leased op
1959: 2026-05-24T18:48:03-04:00 info op_succeeded op=ocr-page-007 attempt=2 code= msg=op succeeded
1960: 2026-05-24T18:48:03-04:00 info op_leased op=ocr-page-009 attempt=1 code= msg=leased op
2317: 2026-05-24T18:48:10-04:00 info op_succeeded op=ocr-page-009 attempt=1 code= msg=op succeeded
2319: 2026-05-24T18:48:10-04:00 info op_leased op=ocr-page-010 attempt=1 code= msg=leased op
2679: 2026-05-24T18:48:18-04:00 info op_succeeded op=ocr-page-010 attempt=1 code= msg=op succeeded
2680: 2026-05-24T18:48:18-04:00 info op_leased op=ocr-page-011 attempt=1 code= msg=leased op
3099: 2026-05-24T18:48:29-04:00 info op_succeeded op=ocr-page-011 attempt=1 code= msg=op succeeded
3101: 2026-05-24T18:48:29-04:00 info op_leased op=ocr-page-012 attempt=1 code= msg=leased op
3476: 2026-05-24T18:48:40-04:00 info op_succeeded op=ocr-page-012 attempt=1 code= msg=op succeeded
3477: 2026-05-24T18:48:40-04:00 info op_leased op=ocr-page-013 attempt=1 code= msg=leased op
3549: 2026-05-24T18:48:46-04:00 info op_succeeded op=ocr-page-013 attempt=1 code= msg=op succeeded
3551: 2026-05-24T18:48:47-04:00 info op_leased op=ocr-page-014 attempt=1 code= msg=leased op
3911: 2026-05-24T18:48:55-04:00 info op_succeeded op=ocr-page-014 attempt=1 code= msg=op succeeded
3912: 2026-05-24T18:48:55-04:00 info op_leased op=ocr-page-015 attempt=1 code= msg=leased op
3962: 2026-05-24T18:49:01-04:00 info op_succeeded op=ocr-page-015 attempt=1 code= msg=op succeeded
3964: 2026-05-24T18:49:01-04:00 info op_leased op=ocr-page-016 attempt=1 code= msg=leased op
4324: 2026-05-24T18:49:12-04:00 info op_succeeded op=ocr-page-016 attempt=1 code= msg=op succeeded
4325: 2026-05-24T18:49:12-04:00 info op_leased op=ocr-page-017 attempt=1 code= msg=leased op
4374: 2026-05-24T18:49:22-04:00 info op_succeeded op=ocr-page-017 attempt=1 code= msg=op succeeded
4376: 2026-05-24T18:49:22-04:00 info op_leased op=ocr-page-018 attempt=1 code= msg=leased op
4676: 2026-05-24T18:49:30-04:00 info op_succeeded op=ocr-page-018 attempt=1 code= msg=op succeeded
4677: 2026-05-24T18:49:30-04:00 info op_leased op=ocr-page-019 attempt=1 code= msg=leased op
5034: 2026-05-24T18:49:40-04:00 info op_succeeded op=ocr-page-019 attempt=1 code= msg=op succeeded
5036: 2026-05-24T18:49:40-04:00 info op_leased op=ocr-page-020 attempt=1 code= msg=leased op
5306: 2026-05-24T18:49:47-04:00 info op_succeeded op=ocr-page-020 attempt=1 code= msg=op succeeded
5307: 2026-05-24T18:49:47-04:00 info op_leased op=ocr-page-021 attempt=1 code= msg=leased op
5396: 2026-05-24T18:49:55-04:00 info op_succeeded op=ocr-page-021 attempt=1 code= msg=op succeeded
5398: 2026-05-24T18:49:56-04:00 info op_leased op=ocr-page-022 attempt=1 code= msg=leased op
5760: 2026-05-24T18:50:05-04:00 info op_succeeded op=ocr-page-022 attempt=1 code= msg=op succeeded
5761: 2026-05-24T18:50:05-04:00 info op_leased op=ocr-page-023 attempt=1 code= msg=leased op
6103: 2026-05-24T18:50:16-04:00 info op_succeeded op=ocr-page-023 attempt=1 code= msg=op succeeded
6105: 2026-05-24T18:50:16-04:00 info op_leased op=ocr-page-024 attempt=1 code= msg=leased op
6507: 2026-05-24T18:50:22-04:00 info op_succeeded op=ocr-page-024 attempt=1 code= msg=op succeeded
6508: 2026-05-24T18:50:22-04:00 info op_leased op=ocr-page-025 attempt=1 code= msg=leased op
6842: 2026-05-24T18:50:32-04:00 info op_succeeded op=ocr-page-025 attempt=1 code= msg=op succeeded
6844: 2026-05-24T18:50:32-04:00 info op_leased op=ocr-page-026 attempt=1 code= msg=leased op
7295: 2026-05-24T18:50:42-04:00 info op_succeeded op=ocr-page-026 attempt=1 code= msg=op succeeded
7296: 2026-05-24T18:50:42-04:00 info op_leased op=ocr-page-027 attempt=1 code= msg=leased op
7773: 2026-05-24T18:50:57-04:00 info op_succeeded op=ocr-page-027 attempt=1 code= msg=op succeeded
7775: 2026-05-24T18:50:57-04:00 info op_leased op=ocr-page-028 attempt=1 code= msg=leased op
8217: 2026-05-24T18:51:06-04:00 info op_succeeded op=ocr-page-028 attempt=1 code= msg=op succeeded
8218: 2026-05-24T18:51:06-04:00 info op_leased op=ocr-page-029 attempt=1 code= msg=leased op
8314: 2026-05-24T18:51:09-04:00 info op_succeeded op=ocr-page-029 attempt=1 code= msg=op succeeded
8316: 2026-05-24T18:51:10-04:00 info op_leased op=ocr-page-030 attempt=1 code= msg=leased op
8681: 2026-05-24T18:51:18-04:00 info op_succeeded op=ocr-page-030 attempt=1 code= msg=op succeeded
8683: 2026-05-24T18:51:18-04:00 info op_leased op=assemble-markdown attempt=1 code= msg=leased op
8684: 2026-05-24T18:51:18-04:00 info op_succeeded op=assemble-markdown attempt=1 code= msg=op succeeded
8685: 2026-05-24T18:51:18-04:00 info workflow_updated op= attempt= code= msg=workflow status updated
```
