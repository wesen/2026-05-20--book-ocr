---
docType: reference
status: active
intent: short-term
topics:
  - ocr
  - experiments
created: 2026-05-24
updated: 2026-05-24
title: 'Run Failed Duplicate Profile Log Summary'
---

# Filtered log summary for `run.log`

- Total lines loaded: 246
- `None` lines: 55
- `info` lines: 155
- `trace` lines: 6
- `warn` lines: 30
- Non-trace workflow events: 185
- Warning/error/failure rows: 30

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
1: 2026-05-24T18:45:20-04:00 info workflow_created op= attempt= code= msg=workflow created
2: 2026-05-24T18:45:20-04:00 info workflow_updated op= attempt= code= msg=workflow status updated
4: 2026-05-24T18:45:20-04:00 info op_leased op=discover-pages attempt=1 code= msg=leased op
5: 2026-05-24T18:45:20-04:00 info op_succeeded op=discover-pages attempt=1 code= msg=op succeeded
7: 2026-05-24T18:45:20-04:00 info op_leased op=ocr-page-001 attempt=1 code= msg=leased op
8: 2026-05-24T18:45:20-04:00 info op_retried op=ocr-page-001 attempt=1 code= msg=op scheduled for retry at 2026-05-24T22:45:21.547706295Z
9: 2026-05-24T18:45:20-04:00 info op_leased op=ocr-page-002 attempt=1 code= msg=leased op
10: 2026-05-24T18:45:20-04:00 info op_retried op=ocr-page-002 attempt=1 code= msg=op scheduled for retry at 2026-05-24T22:45:21.547706295Z
12: 2026-05-24T18:45:20-04:00 info op_leased op=ocr-page-003 attempt=1 code= msg=leased op
13: 2026-05-24T18:45:20-04:00 info op_retried op=ocr-page-003 attempt=1 code= msg=op scheduled for retry at 2026-05-24T22:45:21.825486345Z
14: 2026-05-24T18:45:20-04:00 info op_leased op=ocr-page-004 attempt=1 code= msg=leased op
15: 2026-05-24T18:45:20-04:00 info op_retried op=ocr-page-004 attempt=1 code= msg=op scheduled for retry at 2026-05-24T22:45:21.825486345Z
17: 2026-05-24T18:45:21-04:00 info op_leased op=ocr-page-005 attempt=1 code= msg=leased op
18: 2026-05-24T18:45:21-04:00 info op_retried op=ocr-page-005 attempt=1 code= msg=op scheduled for retry at 2026-05-24T22:45:22.106303226Z
19: 2026-05-24T18:45:21-04:00 info op_leased op=ocr-page-006 attempt=1 code= msg=leased op
20: 2026-05-24T18:45:21-04:00 info op_retried op=ocr-page-006 attempt=1 code= msg=op scheduled for retry at 2026-05-24T22:45:22.106303226Z
22: 2026-05-24T18:45:21-04:00 info op_leased op=ocr-page-007 attempt=1 code= msg=leased op
23: 2026-05-24T18:45:21-04:00 info op_retried op=ocr-page-007 attempt=1 code= msg=op scheduled for retry at 2026-05-24T22:45:22.38789876Z
24: 2026-05-24T18:45:21-04:00 info op_leased op=ocr-page-008 attempt=1 code= msg=leased op
25: 2026-05-24T18:45:21-04:00 info op_retried op=ocr-page-008 attempt=1 code= msg=op scheduled for retry at 2026-05-24T22:45:22.38789876Z
27: 2026-05-24T18:45:21-04:00 info op_leased op=ocr-page-001 attempt=2 code= msg=leased op
28: 2026-05-24T18:45:21-04:00 info op_retried op=ocr-page-001 attempt=2 code= msg=op scheduled for retry at 2026-05-24T22:45:23.666667598Z
29: 2026-05-24T18:45:21-04:00 info op_leased op=ocr-page-002 attempt=2 code= msg=leased op
30: 2026-05-24T18:45:21-04:00 info op_retried op=ocr-page-002 attempt=2 code= msg=op scheduled for retry at 2026-05-24T22:45:23.666667598Z
32: 2026-05-24T18:45:21-04:00 info op_leased op=ocr-page-003 attempt=2 code= msg=leased op
33: 2026-05-24T18:45:21-04:00 info op_retried op=ocr-page-003 attempt=2 code= msg=op scheduled for retry at 2026-05-24T22:45:23.944290044Z
34: 2026-05-24T18:45:21-04:00 info op_leased op=ocr-page-004 attempt=2 code= msg=leased op
35: 2026-05-24T18:45:21-04:00 info op_retried op=ocr-page-004 attempt=2 code= msg=op scheduled for retry at 2026-05-24T22:45:23.944290044Z
37: 2026-05-24T18:45:22-04:00 info op_leased op=ocr-page-005 attempt=2 code= msg=leased op
38: 2026-05-24T18:45:22-04:00 info op_retried op=ocr-page-005 attempt=2 code= msg=op scheduled for retry at 2026-05-24T22:45:24.222871899Z
39: 2026-05-24T18:45:22-04:00 info op_leased op=ocr-page-006 attempt=2 code= msg=leased op
40: 2026-05-24T18:45:22-04:00 info op_retried op=ocr-page-006 attempt=2 code= msg=op scheduled for retry at 2026-05-24T22:45:24.222871899Z
42: 2026-05-24T18:45:22-04:00 info op_leased op=ocr-page-007 attempt=2 code= msg=leased op
43: 2026-05-24T18:45:22-04:00 info op_retried op=ocr-page-007 attempt=2 code= msg=op scheduled for retry at 2026-05-24T22:45:24.504346756Z
44: 2026-05-24T18:45:22-04:00 info op_leased op=ocr-page-008 attempt=2 code= msg=leased op
45: 2026-05-24T18:45:22-04:00 info op_retried op=ocr-page-008 attempt=2 code= msg=op scheduled for retry at 2026-05-24T22:45:24.504346756Z
47: 2026-05-24T18:45:22-04:00 info op_leased op=ocr-page-009 attempt=1 code= msg=leased op
48: 2026-05-24T18:45:22-04:00 info op_retried op=ocr-page-009 attempt=1 code= msg=op scheduled for retry at 2026-05-24T22:45:23.78429842Z
49: 2026-05-24T18:45:22-04:00 info op_leased op=ocr-page-010 attempt=1 code= msg=leased op
50: 2026-05-24T18:45:22-04:00 info op_retried op=ocr-page-010 attempt=1 code= msg=op scheduled for retry at 2026-05-24T22:45:23.78429842Z
52: 2026-05-24T18:45:23-04:00 info op_leased op=ocr-page-011 attempt=1 code= msg=leased op
53: 2026-05-24T18:45:23-04:00 info op_retried op=ocr-page-011 attempt=1 code= msg=op scheduled for retry at 2026-05-24T22:45:24.062771379Z
54: 2026-05-24T18:45:23-04:00 info op_leased op=ocr-page-012 attempt=1 code= msg=leased op
55: 2026-05-24T18:45:23-04:00 info op_retried op=ocr-page-012 attempt=1 code= msg=op scheduled for retry at 2026-05-24T22:45:24.062771379Z
57: 2026-05-24T18:45:23-04:00 info op_leased op=ocr-page-013 attempt=1 code= msg=leased op
58: 2026-05-24T18:45:23-04:00 info op_retried op=ocr-page-013 attempt=1 code= msg=op scheduled for retry at 2026-05-24T22:45:24.343357421Z
59: 2026-05-24T18:45:23-04:00 info op_leased op=ocr-page-014 attempt=1 code= msg=leased op
60: 2026-05-24T18:45:23-04:00 info op_retried op=ocr-page-014 attempt=1 code= msg=op scheduled for retry at 2026-05-24T22:45:24.343357421Z
62: 2026-05-24T18:45:23-04:00 info op_leased op=ocr-page-015 attempt=1 code= msg=leased op
63: 2026-05-24T18:45:23-04:00 info op_retried op=ocr-page-015 attempt=1 code= msg=op scheduled for retry at 2026-05-24T22:45:24.624243918Z
64: 2026-05-24T18:45:23-04:00 info op_leased op=ocr-page-016 attempt=1 code= msg=leased op
65: 2026-05-24T18:45:23-04:00 info op_retried op=ocr-page-016 attempt=1 code= msg=op scheduled for retry at 2026-05-24T22:45:24.624243918Z
67: 2026-05-24T18:45:23-04:00 info op_leased op=ocr-page-001 attempt=3 code= msg=leased op
68: 2026-05-24T18:45:23-04:00 warn op_failed op=ocr-page-001 attempt=3 code=ocr_geppetto_failed msg=op failed
69: 2026-05-24T18:45:23-04:00 info op_leased op=ocr-page-002 attempt=3 code= msg=leased op
70: 2026-05-24T18:45:23-04:00 warn op_failed op=ocr-page-002 attempt=3 code=ocr_geppetto_failed msg=op failed
72: 2026-05-24T18:45:24-04:00 info op_leased op=ocr-page-003 attempt=3 code= msg=leased op
73: 2026-05-24T18:45:24-04:00 warn op_failed op=ocr-page-003 attempt=3 code=ocr_geppetto_failed msg=op failed
74: 2026-05-24T18:45:24-04:00 info op_leased op=ocr-page-004 attempt=3 code= msg=leased op
75: 2026-05-24T18:45:24-04:00 warn op_failed op=ocr-page-004 attempt=3 code=ocr_geppetto_failed msg=op failed
77: 2026-05-24T18:45:24-04:00 info op_leased op=ocr-page-005 attempt=3 code= msg=leased op
78: 2026-05-24T18:45:24-04:00 warn op_failed op=ocr-page-005 attempt=3 code=ocr_geppetto_failed msg=op failed
79: 2026-05-24T18:45:24-04:00 info op_leased op=ocr-page-006 attempt=3 code= msg=leased op
80: 2026-05-24T18:45:24-04:00 warn op_failed op=ocr-page-006 attempt=3 code=ocr_geppetto_failed msg=op failed
82: 2026-05-24T18:45:24-04:00 info op_leased op=ocr-page-007 attempt=3 code= msg=leased op
83: 2026-05-24T18:45:24-04:00 warn op_failed op=ocr-page-007 attempt=3 code=ocr_geppetto_failed msg=op failed
84: 2026-05-24T18:45:24-04:00 info op_leased op=ocr-page-008 attempt=3 code= msg=leased op
85: 2026-05-24T18:45:24-04:00 warn op_failed op=ocr-page-008 attempt=3 code=ocr_geppetto_failed msg=op failed
87: 2026-05-24T18:45:25-04:00 info op_leased op=ocr-page-009 attempt=2 code= msg=leased op
88: 2026-05-24T18:45:25-04:00 info op_retried op=ocr-page-009 attempt=2 code= msg=op scheduled for retry at 2026-05-24T22:45:27.022199281Z
89: 2026-05-24T18:45:25-04:00 info op_leased op=ocr-page-010 attempt=2 code= msg=leased op
90: 2026-05-24T18:45:25-04:00 info op_retried op=ocr-page-010 attempt=2 code= msg=op scheduled for retry at 2026-05-24T22:45:27.022199281Z
92: 2026-05-24T18:45:25-04:00 info op_leased op=ocr-page-011 attempt=2 code= msg=leased op
93: 2026-05-24T18:45:25-04:00 info op_retried op=ocr-page-011 attempt=2 code= msg=op scheduled for retry at 2026-05-24T22:45:27.299314661Z
94: 2026-05-24T18:45:25-04:00 info op_leased op=ocr-page-012 attempt=2 code= msg=leased op
95: 2026-05-24T18:45:25-04:00 info op_retried op=ocr-page-012 attempt=2 code= msg=op scheduled for retry at 2026-05-24T22:45:27.299314661Z
97: 2026-05-24T18:45:25-04:00 info op_leased op=ocr-page-013 attempt=2 code= msg=leased op
98: 2026-05-24T18:45:25-04:00 info op_retried op=ocr-page-013 attempt=2 code= msg=op scheduled for retry at 2026-05-24T22:45:27.580349991Z
99: 2026-05-24T18:45:25-04:00 info op_leased op=ocr-page-014 attempt=2 code= msg=leased op
100: 2026-05-24T18:45:25-04:00 info op_retried op=ocr-page-014 attempt=2 code= msg=op scheduled for retry at 2026-05-24T22:45:27.580349991Z
```

## Failures

```text
68: 2026-05-24T18:45:23-04:00 op=ocr-page-001 attempt=3 code=ocr_geppetto_failed msg=op failed
70: 2026-05-24T18:45:23-04:00 op=ocr-page-002 attempt=3 code=ocr_geppetto_failed msg=op failed
73: 2026-05-24T18:45:24-04:00 op=ocr-page-003 attempt=3 code=ocr_geppetto_failed msg=op failed
75: 2026-05-24T18:45:24-04:00 op=ocr-page-004 attempt=3 code=ocr_geppetto_failed msg=op failed
78: 2026-05-24T18:45:24-04:00 op=ocr-page-005 attempt=3 code=ocr_geppetto_failed msg=op failed
80: 2026-05-24T18:45:24-04:00 op=ocr-page-006 attempt=3 code=ocr_geppetto_failed msg=op failed
83: 2026-05-24T18:45:24-04:00 op=ocr-page-007 attempt=3 code=ocr_geppetto_failed msg=op failed
85: 2026-05-24T18:45:24-04:00 op=ocr-page-008 attempt=3 code=ocr_geppetto_failed msg=op failed
128: 2026-05-24T18:45:27-04:00 op=ocr-page-009 attempt=3 code=ocr_geppetto_failed msg=op failed
130: 2026-05-24T18:45:27-04:00 op=ocr-page-010 attempt=3 code=ocr_geppetto_failed msg=op failed
133: 2026-05-24T18:45:27-04:00 op=ocr-page-011 attempt=3 code=ocr_geppetto_failed msg=op failed
135: 2026-05-24T18:45:27-04:00 op=ocr-page-012 attempt=3 code=ocr_geppetto_failed msg=op failed
138: 2026-05-24T18:45:27-04:00 op=ocr-page-013 attempt=3 code=ocr_geppetto_failed msg=op failed
140: 2026-05-24T18:45:27-04:00 op=ocr-page-014 attempt=3 code=ocr_geppetto_failed msg=op failed
143: 2026-05-24T18:45:28-04:00 op=ocr-page-015 attempt=3 code=ocr_geppetto_failed msg=op failed
145: 2026-05-24T18:45:28-04:00 op=ocr-page-016 attempt=3 code=ocr_geppetto_failed msg=op failed
185: 2026-05-24T18:45:30-04:00 op=ocr-page-017 attempt=3 code=ocr_geppetto_failed msg=op failed
187: 2026-05-24T18:45:30-04:00 op=ocr-page-018 attempt=3 code=ocr_geppetto_failed msg=op failed
190: 2026-05-24T18:45:30-04:00 op=ocr-page-019 attempt=3 code=ocr_geppetto_failed msg=op failed
192: 2026-05-24T18:45:30-04:00 op=ocr-page-020 attempt=3 code=ocr_geppetto_failed msg=op failed
195: 2026-05-24T18:45:31-04:00 op=ocr-page-021 attempt=3 code=ocr_geppetto_failed msg=op failed
197: 2026-05-24T18:45:31-04:00 op=ocr-page-022 attempt=3 code=ocr_geppetto_failed msg=op failed
200: 2026-05-24T18:45:31-04:00 op=ocr-page-023 attempt=3 code=ocr_geppetto_failed msg=op failed
202: 2026-05-24T18:45:31-04:00 op=ocr-page-024 attempt=3 code=ocr_geppetto_failed msg=op failed
230: 2026-05-24T18:45:33-04:00 op=ocr-page-025 attempt=3 code=ocr_geppetto_failed msg=op failed
232: 2026-05-24T18:45:33-04:00 op=ocr-page-026 attempt=3 code=ocr_geppetto_failed msg=op failed
235: 2026-05-24T18:45:34-04:00 op=ocr-page-027 attempt=3 code=ocr_geppetto_failed msg=op failed
237: 2026-05-24T18:45:34-04:00 op=ocr-page-028 attempt=3 code=ocr_geppetto_failed msg=op failed
240: 2026-05-24T18:45:34-04:00 op=ocr-page-029 attempt=3 code=ocr_geppetto_failed msg=op failed
242: 2026-05-24T18:45:34-04:00 op=ocr-page-030 attempt=3 code=ocr_geppetto_failed msg=op failed
```
